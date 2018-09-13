package mongodb

import (
	"errors"
	"strings"
	"time"

	kubeClientModel "github.com/containerum/kube-client/pkg/model"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

var (
	uniqueAddedIndex = mgo.Index{
		Name:     "unique_resource_added",
		Key:      []string{"eventname", "resourceuid"},
		DropDups: true,
		PartialFilter: bson.M{
			"eventname": kubeClientModel.ResourceCreated,
		},
		Unique: true,
	}

	uniqueEventsIndex = mgo.Index{
		Name:     "unique_resourceuid",
		Key:      []string{"resourceuid"},
		DropDups: true,
		Unique:   true,
	}

	dateExpirationIndex = mgo.Index{
		Name:        "date_expiration",
		Key:         []string{"dateadded"},
		ExpireAfter: 30 * 24 * time.Hour,
	}
)

func (s *Storage) ensureIndexes() error {
	s.log.Debugf("Ensure indexes")
	var errs []string

	for _, collectionName := range Сollections {
		collection := s.db.C(collectionName)
		if err := collection.EnsureIndex(dateExpirationIndex); err != nil {
			errs = append(errs, err.Error())
		}
		if err := collection.EnsureIndexKey("eventname"); err != nil {
			errs = append(errs, err.Error())
		}
		if err := collection.EnsureIndexKey("resourcename"); err != nil {
			errs = append(errs, err.Error())
		}
		if err := collection.EnsureIndexKey("eventname", "resourcename"); err != nil {
			errs = append(errs, err.Error())
		}
	}

	{
		collection := s.db.C(DeploymentCollection)
		if err := collection.EnsureIndex(uniqueAddedIndex); err != nil {
			errs = append(errs, err.Error())
		}
	}

	{
		collection := s.db.C(EventsCollection)
		if err := collection.EnsureIndex(uniqueEventsIndex); err != nil {
			errs = append(errs, err.Error())
		}
	}
	{
		collection := s.db.C(ResourceQuotasCollection)
		if err := collection.EnsureIndex(uniqueAddedIndex); err != nil {
			errs = append(errs, err.Error())
		}
	}

	{
		collection := s.db.C(IngressCollection)
		if err := collection.EnsureIndex(uniqueAddedIndex); err != nil {
			errs = append(errs, err.Error())
		}
	}

	{
		collection := s.db.C(ServiceCollection)
		if err := collection.EnsureIndex(uniqueAddedIndex); err != nil {
			errs = append(errs, err.Error())
		}
	}

	{
		collection := s.db.C(PVCCollection)
		if err := collection.EnsureIndex(uniqueAddedIndex); err != nil {
			errs = append(errs, err.Error())
		}
	}

	{
		collection := s.db.C(SecretsCollection)
		if err := collection.EnsureIndex(uniqueAddedIndex); err != nil {
			errs = append(errs, err.Error())
		}
	}

	{
		collection := s.db.C(ConfigMapsCollection)
		if err := collection.EnsureIndex(uniqueAddedIndex); err != nil {
			errs = append(errs, err.Error())
		}
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, ","))
	}

	return nil
}
