package storage

import "time"

type EventCleaner interface {
	Cleanup(deleteBefore time.Time) error
}

type RecordCleanerConfig struct {
	Storage          EventCleaner
	CleanupRunPeriod time.Duration
	RetentionPeriod  time.Duration
}

type RecordCleaner struct {
	cfg RecordCleanerConfig

	cleanupTimer *time.Ticker
	stop         chan struct{}
}

func NewRecordCleaner(cfg RecordCleanerConfig) *RecordCleaner {
	return &RecordCleaner{
		cfg:          cfg,
		cleanupTimer: time.NewTicker(cfg.CleanupRunPeriod),
		stop:         make(chan struct{}),
	}
}

func (rc *RecordCleaner) cleanup() {
	for {
		select {
		case <-rc.cleanupTimer.C:
			rc.cfg.Storage.Cleanup(time.Now().UTC().Add(-rc.cfg.RetentionPeriod))
		case <-rc.stop:
			return
		}
	}
}

func (rc *RecordCleaner) RunPeriodicCleanup() {
	go rc.cleanup()
}

func (rc *RecordCleaner) Stop() {
	rc.stop <- struct{}{}
	rc.cleanupTimer.Stop()
	return
}
