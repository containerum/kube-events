package storage

import (
	"sync"
	"time"

	"github.com/containerum/kube-events/pkg/model"
	"github.com/sirupsen/logrus"
)

type EventInserter interface {
	Insert(r *model.Record, collection string) error
}

type EventBulkInserter interface {
	BulkInsert(r []model.Record, collection string) error
}

type RecordBufferConfig struct {
	Storage         EventBulkInserter
	BufferCap       int
	InsertPeriod    time.Duration
	MinInsertEvents int
	Collector       <-chan model.Record
}

type RecordBuffer struct {
	cfg RecordBufferConfig

	bufferMu sync.Mutex
	buffer   []model.Record

	readStop    chan struct{}
	insertStop  chan struct{}
	insertTimer *time.Ticker

	log *logrus.Entry
}

func NewRecordBuffer(cfg RecordBufferConfig) *RecordBuffer {
	log := logrus.WithField("component", "record_buffer")
	log.WithFields(logrus.Fields{
		"capacity":          cfg.BufferCap,
		"insert_period":     cfg.InsertPeriod,
		"min_insert_events": cfg.MinInsertEvents,
	}).Info("Initialized record buffer")

	return &RecordBuffer{
		cfg:         cfg,
		buffer:      make([]model.Record, 0, cfg.BufferCap),
		readStop:    make(chan struct{}),
		insertStop:  make(chan struct{}),
		insertTimer: time.NewTicker(cfg.InsertPeriod),
		log:         log,
	}
}

func (rb *RecordBuffer) readRecords() {
	for {
		select {
		case record, ok := <-rb.cfg.Collector:
			if !ok {
				return
			}
			rb.log.Debugf("Collected record %+v", record)
			rb.bufferMu.Lock()
			rb.buffer = append(rb.buffer, record)
			rb.bufferMu.Unlock()
		case <-rb.readStop:
			return
		}
	}
}

func (rb *RecordBuffer) insertRecords(collection string) {
	for {
		select {
		case <-rb.insertStop:
			return
		case <-rb.insertTimer.C:
			// get a buffer length and copy slice pointer (it may be replaced in RecordBuffer)
			rb.bufferMu.Lock()
			oldBuf := rb.buffer
			bufLen := len(rb.buffer)
			rb.bufferMu.Unlock()

			if bufLen < rb.cfg.MinInsertEvents {
				rb.log.Infof("Wanted minimum %d records to be inserted, collected %d",
					rb.cfg.MinInsertEvents, bufLen)
				continue
			}

			// replace a buffer with empty one
			newBuf := make([]model.Record, 0, rb.cfg.BufferCap)
			rb.bufferMu.Lock()
			rb.buffer = newBuf
			rb.bufferMu.Unlock()

			// perform bulk insert
			go func() {
				rb.log.Infof("Inserting %d events", bufLen)
				err := rb.cfg.Storage.BulkInsert(oldBuf, collection)
				if err != nil {
					rb.log.WithError(err).Error("BulkInsert failed")
				}
			}()
		}
	}
}

func (rb *RecordBuffer) RunCollection(collection string) {
	rb.log.Debug("Starting reading/inserting records")
	go rb.readRecords()
	go rb.insertRecords(collection)
}

func (rb *RecordBuffer) Stop() {
	rb.log.Debug("Stopping reading/inserting records")
	rb.readStop <- struct{}{}
	rb.insertStop <- struct{}{}
	rb.insertTimer.Stop()
}
