package mongodb

import (
	kubeClientModel "github.com/containerum/kube-client/pkg/model"
	"github.com/globalsign/mgo"
	"github.com/sirupsen/logrus"
)

const (
	ResourceQuotasCollection = "namespaces"
	EventsCollection         = "events"
	DeploymentCollection     = "deployments"
	ServiceCollection        = "services"
	IngressCollection        = "ingresses"
	PVCCollection            = "volumes"
	SecretsCollection        = "secrets"
	ConfigMapsCollection     = "configmaps"
	UserCollection           = "user"
	SystemCollection         = "system"
	CRDCollection            = "crd"
)

var Collections = []string{
	ResourceQuotasCollection,
	EventsCollection,
	DeploymentCollection,
	ServiceCollection,
	IngressCollection,
	PVCCollection,
	SecretsCollection,
	ConfigMapsCollection,
	UserCollection,
	SystemCollection,
	CRDCollection,
}

type Storage struct {
	db  *mgo.Database
	log *logrus.Entry
}

func OpenConnection(cfg *mgo.DialInfo) (*Storage, error) {
	log := logrus.WithField("component", "mongo-storage")
	log.WithFields(logrus.Fields{
		"addrs":    cfg.Addrs,
		"user":     cfg.Username,
		"database": cfg.Database,
	}).Info("Opening connection with MongoDB")
	session, err := mgo.DialWithInfo(cfg)
	if err != nil {
		return nil, err
	}
	db := session.DB(cfg.Database)

	storage := &Storage{
		db:  db,
		log: log,
	}

	if err := storage.createCollectionIfNotExist(EventsCollection); err != nil {
		return nil, err
	}
	if err := storage.createCollectionIfNotExist(DeploymentCollection); err != nil {
		return nil, err
	}
	if err := storage.createCollectionIfNotExist(ResourceQuotasCollection); err != nil {
		return nil, err
	}
	if err := storage.createCollectionIfNotExist(ServiceCollection); err != nil {
		return nil, err
	}
	if err := storage.createCollectionIfNotExist(IngressCollection); err != nil {
		return nil, err
	}
	if err := storage.createCollectionIfNotExist(PVCCollection); err != nil {
		return nil, err
	}
	if err := storage.createCollectionIfNotExist(SecretsCollection); err != nil {
		return nil, err
	}
	if err := storage.createCollectionIfNotExist(ConfigMapsCollection); err != nil {
		return nil, err
	}
	if err := storage.createCollectionIfNotExist(UserCollection); err != nil {
		return nil, err
	}
	if err := storage.createCollectionIfNotExist(SystemCollection); err != nil {
		return nil, err
	}
	if err := storage.createCollectionIfNotExist(CRDCollection); err != nil {
		return nil, err
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

func (s *Storage) Close() error {
	s.log.Debugf("Closing storage")
	s.db.Session.Close()
	return nil
}
