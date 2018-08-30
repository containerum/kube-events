package mongodb

import "github.com/sirupsen/logrus"

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

func (s *Storage) createCappedCollectionIfNotExist(name string, size uint64, maxDocs uint) error {
	s.log.WithFields(logrus.Fields{
		"name":     name,
		"size":     size,
		"max_docs": maxDocs,
	}).Debugf("Create capped collection if not exists")
	exist, err := s.isCollectionExist(name)
	if err != nil {
		return err
	}
	if exist {
		return nil
	}
	// bson.D not working here (gives "command not found: '0'"), why?
	return s.db.Run(struct {
		Create string `bson:"create"`
		Capped bool   `bson:"capped"`
		Size   uint64 `bson:"size"`
		Max    uint   `bson:"max,omitempty"`
	}{
		Create: name,
		Capped: true,
		Size:   size,
		Max:    maxDocs,
	}, nil)
}
