package storage

import (
	"sync"
	"time"

	"github.com/containerum/kube-events/pkg/model"
)

type EventInserter interface {
	Insert(r *model.Record) error
}

type EventBulkInserter interface {
	BulkInsert(r []model.Record) error
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
}

func NewRecordBuffer(cfg RecordBufferConfig) *RecordBuffer {
	return &RecordBuffer{
		cfg:         cfg,
		buffer:      make([]model.Record, 0, cfg.BufferCap),
		readStop:    make(chan struct{}),
		insertStop:  make(chan struct{}),
		insertTimer: time.NewTicker(cfg.InsertPeriod),
	}
}

func (rb *RecordBuffer) readRecords() {
	for {
		select {
		case record := <-rb.cfg.Collector:
			rb.bufferMu.Lock()
			rb.buffer = append(rb.buffer, record)
			rb.bufferMu.Unlock()
		case <-rb.readStop:
			break
		}
	}
}

func (rb *RecordBuffer) insertRecords() {
	for {
		select {
		case <-rb.insertStop:
			break
		case <-rb.insertTimer.C:
			// get a buffer length and copy slice pointer (it may be replaced in RecordBuffer)
			rb.bufferMu.Lock()
			oldBuf := rb.buffer
			bufLen := len(rb.buffer)
			rb.bufferMu.Unlock()

			if bufLen < rb.cfg.MinInsertEvents {
				continue
			}

			// replace a buffer with empty one
			rb.bufferMu.Lock()
			rb.buffer = make([]model.Record, 0, rb.cfg.BufferCap)
			rb.bufferMu.Unlock()

			// perform bulk insert
			go func() {
				rb.cfg.Storage.BulkInsert(oldBuf)
			}()
		}
	}
}

func (rb *RecordBuffer) RunCollection() {
	go rb.readRecords()
	go rb.insertRecords()
}

func (rb *RecordBuffer) Stop() {
	rb.readStop <- struct{}{}
	rb.insertStop <- struct{}{}
	rb.insertTimer.Stop()
}
