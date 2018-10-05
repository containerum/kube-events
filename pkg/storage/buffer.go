package storage

import (
	"sync"
	"time"

	kubeClientModel "github.com/containerum/kube-client/pkg/model"
	log "github.com/sirupsen/logrus"
)

type EventInserter interface {
	Insert(r *kubeClientModel.Event, collection string) error
}

type EventBulkInserter interface {
	BulkInsert(r []kubeClientModel.Event, collection string) error
}

type RecordBufferConfig struct {
	Storage         EventBulkInserter
	BufferCap       int
	InsertPeriod    time.Duration
	MinInsertEvents int
	Collector       <-chan kubeClientModel.Event
}

type RecordBuffer struct {
	cfg RecordBufferConfig

	bufferMu sync.Mutex
	buffer   []kubeClientModel.Event

	readStop    chan struct{}
	insertStop  chan struct{}
	insertTimer *time.Ticker

	log *log.Entry
}

func NewRecordBuffer(cfg RecordBufferConfig) *RecordBuffer {
	rbLog := log.WithField("component", "record_buffer")
	rbLog.WithFields(log.Fields{
		"capacity":          cfg.BufferCap,
		"insert_period":     cfg.InsertPeriod,
		"min_insert_events": cfg.MinInsertEvents,
	}).Info("Initialized record buffer")

	return &RecordBuffer{
		cfg:         cfg,
		buffer:      make([]kubeClientModel.Event, 0, cfg.BufferCap),
		readStop:    make(chan struct{}),
		insertStop:  make(chan struct{}),
		insertTimer: time.NewTicker(cfg.InsertPeriod),
		log:         rbLog,
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
				rb.log.Debugf("Wanted minimum %d records to be inserted, collected %d",
					rb.cfg.MinInsertEvents, bufLen)
				continue
			}

			// replace a buffer with empty one
			newBuf := make([]kubeClientModel.Event, 0, rb.cfg.BufferCap)
			rb.bufferMu.Lock()
			rb.buffer = newBuf
			rb.bufferMu.Unlock()

			// perform bulk insert
			go func() {
				rb.log.Debugf("Inserting %d events", bufLen)
				dateAdded := time.Now()
				for i := range oldBuf {
					oldBuf[i].DateAdded = dateAdded
				}
				err := rb.cfg.Storage.BulkInsert(oldBuf, collection)
				if err != nil {
					rb.log.WithError(err).Debug("BulkInsert failed")
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
