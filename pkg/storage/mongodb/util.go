package mongodb

import (
	"github.com/globalsign/mgo"
	log "github.com/sirupsen/logrus"
)

func (s *Storage) isCollectionExist(name string) (bool, error) {
	colls, err := s.db.CollectionNames()
	if err != nil {
		return false, err
	}
	for _, v := range colls {
		if name == v {
			return true, nil
		}
	}
	return false, nil
}

func (s *Storage) createCollectionIfNotExist(name string) error {
	s.log.WithFields(log.Fields{
		"name": name,
	}).Debugf("Create collection if not exists")
	exist, err := s.isCollectionExist(name)
	if err != nil {
		return err
	}
	if exist {
		return nil
	}
	return s.db.C(name).Create(&mgo.CollectionInfo{})
}
