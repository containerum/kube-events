package storage

import (
	"time"

	"github.com/sirupsen/logrus"
)

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

	log *logrus.Entry
}

func NewRecordCleaner(cfg RecordCleanerConfig) *RecordCleaner {
	log := logrus.WithField("component", "record_cleaner")
	log.WithFields(logrus.Fields{
		"cleanup_run_period": cfg.CleanupRunPeriod,
		"retention_period":   cfg.RetentionPeriod,
	}).Info("Initialized record cleaner")
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
			rc.log.Info("Run cleanup")
			err := rc.cfg.Storage.Cleanup(time.Now().UTC().Add(-rc.cfg.RetentionPeriod))
			if err != nil {
				rc.log.WithError(err).Error("Cleanup failed")
			}
		case <-rc.stop:
			return
		}
	}
}

func (rc *RecordCleaner) RunPeriodicCleanup() {
	rc.log.Debug("Run periodic cleanup")
	go rc.cleanup()
}

func (rc *RecordCleaner) Stop() {
	rc.log.Debug("Stop periodic cleanup")
	rc.stop <- struct{}{}
	rc.cleanupTimer.Stop()
	return
}
