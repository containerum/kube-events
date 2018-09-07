package mongodb

import (
	"time"

	kubeClientModel "github.com/containerum/kube-client/pkg/model"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/sirupsen/logrus"
)

const (
	ResourceQuotasCollection = "namespaces"
	EventsCollection         = "events"
	DeploymentCollection     = "deployments"
	ServiceCollection        = "services"
	IngressCollection        = "ingresses"
	PVCCollection            = "volumes"
	UserCollection           = "user"
	SystemCollection         = "system"
)

var collections = []string{
	ResourceQuotasCollection,
	DeploymentCollection,
	ServiceCollection,
	IngressCollection,
	PVCCollection,
	EventsCollection,
	UserCollection,
	SystemCollection,
}

type Config struct {
	mgo.DialInfo

	CollectionSize uint64 // in bytes
	MaxDocuments   uint
}

type Storage struct {
	db  *mgo.Database
	log *logrus.Entry
}

func OpenConnection(cfg *Config) (*Storage, error) {
	log := logrus.WithField("component", "mongo-storage")
	log.WithFields(logrus.Fields{
		"addrs":           cfg.Addrs,
		"user":            cfg.Username,
		"database":        cfg.Database,
		"collection_size": cfg.CollectionSize,
		"max_docs":        cfg.MaxDocuments,
	}).Info("Opening connection with MongoDB")
	session, err := mgo.DialWithInfo(&cfg.DialInfo)
	if err != nil {
		return nil, err
	}
	db := session.DB(cfg.DialInfo.Database)

	storage := &Storage{
		db:  db,
		log: log,
	}

	if cfg.CollectionSize > 0 {
		if err := storage.createCappedCollectionIfNotExist(EventsCollection, cfg.CollectionSize, cfg.MaxDocuments); err != nil {
			return nil, err
		}
		if err := storage.createCappedCollectionIfNotExist(DeploymentCollection, cfg.CollectionSize, cfg.MaxDocuments); err != nil {
			return nil, err
		}
		if err := storage.createCappedCollectionIfNotExist(ResourceQuotasCollection, cfg.CollectionSize, cfg.MaxDocuments); err != nil {
			return nil, err
		}
		if err := storage.createCappedCollectionIfNotExist(ServiceCollection, cfg.CollectionSize, cfg.MaxDocuments); err != nil {
			return nil, err
		}
		if err := storage.createCappedCollectionIfNotExist(IngressCollection, cfg.CollectionSize, cfg.MaxDocuments); err != nil {
			return nil, err
		}
		if err := storage.createCappedCollectionIfNotExist(PVCCollection, cfg.CollectionSize, cfg.MaxDocuments); err != nil {
			return nil, err
		}
		if err := storage.createCappedCollectionIfNotExist(UserCollection, cfg.CollectionSize, cfg.MaxDocuments); err != nil {
			return nil, err
		}
		if err := storage.createCappedCollectionIfNotExist(SystemCollection, cfg.CollectionSize, cfg.MaxDocuments); err != nil {
			return nil, err
		}
	}

	if err := storage.ensureIndexes(); err != nil {
		return nil, err
	}

	return storage, nil
}

func (s *Storage) Insert(r *kubeClientModel.Event, collection string) error {
	s.log.Debugf("Insert single record")
	return s.db.C(collection).Insert(r)
}

func (s *Storage) BulkInsert(r []kubeClientModel.Event, collection string) error {
	s.log.WithField("record_count", len(r)).Debugf("Bulk insert")
	docs := make([]interface{}, len(r))
	for i, record := range r {
		docs[i] = record
	}
	bulk := s.db.C(collection).Bulk()
	bulk.Unordered()
	bulk.Insert(docs...)
	result, err := bulk.Run()
	if err != nil {
		return err
	}
	s.log.WithFields(logrus.Fields{
		"matched":  result.Matched,
		"modified": result.Modified,
	}).Debug("Bulk insert run")
	return nil
}

func (s *Storage) Cleanup(deleteBefore time.Time) error {
	s.log.WithField("delete_before", deleteBefore).Debugf("Cleanup")
	for _, col := range collections {
		bulk := s.db.C(col).Bulk()
		bulk.Unordered()
		bulk.RemoveAll(bson.M{"timestamp": bson.M{"$lte": deleteBefore}})
		result, err := bulk.Run()
		if err != nil {
			return err
		}
		s.log.WithFields(logrus.Fields{
			"matched":    result.Matched,
			"modified":   result.Modified,
			"collection": col,
		}).Debug("Cleanup run")
	}
	return nil
}

func (s *Storage) Close() error {
	s.log.Debugf("Closing storage")
	s.db.Session.Close()
	return nil
}
