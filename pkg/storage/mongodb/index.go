package mongodb

import (
	"errors"
	"strings"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"k8s.io/apimachinery/pkg/watch"
)

var (
	addedDeletedIndex = mgo.Index{
		Name:     "unique_resource_added",
		Key:      []string{"eventtype", "uid"},
		DropDups: true,
		PartialFilter: bson.M{
			"eventtype": watch.Added,
		},
		Unique: true,
	}
)

func (s *Storage) ensureIndexes() error {
	s.log.Debugf("Ensure indexes")
	var errs []string

	{
		collection := s.db.C(DeploymentCollection)
		if err := collection.EnsureIndex(addedDeletedIndex); err != nil {
			errs = append(errs, err.Error())
		}
		if err := collection.EnsureIndexKey("eventtype"); err != nil {
			errs = append(errs, err.Error())
		}
		if err := collection.EnsureIndexKey("name"); err != nil {
			errs = append(errs, err.Error())
		}
		if err := collection.EnsureIndexKey("eventtype", "name"); err != nil {
			errs = append(errs, err.Error())
		}
	}

	{
		collection := s.db.C(EventsCollection)
		if err := collection.EnsureIndexKey("eventtype"); err != nil {
			errs = append(errs, err.Error())
		}
		if err := collection.EnsureIndexKey("name"); err != nil {
			errs = append(errs, err.Error())
		}
		if err := collection.EnsureIndexKey("eventtype", "name"); err != nil {
			errs = append(errs, err.Error())
		}
	}

	{
		collection := s.db.C(ResourceQuotasCollection)
		if err := collection.EnsureIndex(addedDeletedIndex); err != nil {
			errs = append(errs, err.Error())
		}
		if err := collection.EnsureIndexKey("eventtype"); err != nil {
			errs = append(errs, err.Error())
		}
		if err := collection.EnsureIndexKey("name"); err != nil {
			errs = append(errs, err.Error())
		}
		if err := collection.EnsureIndexKey("eventtype", "name"); err != nil {
			errs = append(errs, err.Error())
		}
	}

	{
		collection := s.db.C(IngressCollection)
		if err := collection.EnsureIndex(addedDeletedIndex); err != nil {
			errs = append(errs, err.Error())
		}
		if err := collection.EnsureIndexKey("eventtype"); err != nil {
			errs = append(errs, err.Error())
		}
		if err := collection.EnsureIndexKey("name"); err != nil {
			errs = append(errs, err.Error())
		}
		if err := collection.EnsureIndexKey("eventtype", "name"); err != nil {
			errs = append(errs, err.Error())
		}
	}

	{
		collection := s.db.C(ServiceCollection)
		if err := collection.EnsureIndex(addedDeletedIndex); err != nil {
			errs = append(errs, err.Error())
		}
		if err := collection.EnsureIndexKey("eventtype"); err != nil {
			errs = append(errs, err.Error())
		}
		if err := collection.EnsureIndexKey("name"); err != nil {
			errs = append(errs, err.Error())
		}
		if err := collection.EnsureIndexKey("eventtype", "name"); err != nil {
			errs = append(errs, err.Error())
		}
	}

	{
		collection := s.db.C(PVCollection)
		if err := collection.EnsureIndex(addedDeletedIndex); err != nil {
			errs = append(errs, err.Error())
		}
		if err := collection.EnsureIndexKey("eventtype"); err != nil {
			errs = append(errs, err.Error())
		}
		if err := collection.EnsureIndexKey("name"); err != nil {
			errs = append(errs, err.Error())
		}
		if err := collection.EnsureIndexKey("eventtype", "name"); err != nil {
			errs = append(errs, err.Error())
		}
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, ","))
	}

	return nil
}
