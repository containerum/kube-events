package mongodb

import (
	"time"

	"github.com/containerum/kube-events/pkg/model"
	"github.com/globalsign/mgo"
	"gopkg.in/mgo.v2/bson"
)

const EventsCollection = "events"

type Config struct {
	mgo.DialInfo

	Database string
}

type Storage struct {
	db *mgo.Database
}

func OpenConnection(cfg *Config) (*Storage, error) {
	session, err := mgo.DialWithInfo(&cfg.DialInfo)
	if err != nil {
		return nil, err
	}
	db := session.DB(cfg.Database)

	return &Storage{
		db: db,
	}, nil
}

func (s *Storage) Insert(r *model.Record) error {
	return s.db.C(EventsCollection).Insert(r)
}

func (s *Storage) BulkInsert(r []model.Record) error {
	docs := make([]interface{}, len(r))
	for i, record := range r {
		docs[i] = record
	}
	bulk := s.db.C(EventsCollection).Bulk()
	bulk.Insert(docs...)
	_, err := bulk.Run()
	return err
}

func (s *Storage) Cleanup(deleteBefore time.Time) error {
	bulk := s.db.C(EventsCollection).Bulk()
	bulk.RemoveAll(bson.M{"timestamp": bson.M{"$lte": deleteBefore}})
	_, err := bulk.Run()
	return err
}
