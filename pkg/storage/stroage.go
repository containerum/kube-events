package storage

import (
	"time"

	"github.com/containerum/kube-events/pkg/model"
)

type EventStorage interface {
	Insert(r *model.Record) error
	BulkInsert(r []model.Record) error
	Cleanup(deleteBefore time.Time) error
}
