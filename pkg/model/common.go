package model

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/globalsign/mgo/bson"
	"k8s.io/apimachinery/pkg/watch"
)

// ObservableResource is set of constants representing kubernetes resource for watching.
type ObservableResource string

const (
	ObservableNamespace        ObservableResource = "namespace"
	ObservableDeployment       ObservableResource = "deployment"
	ObservablePod              ObservableResource = "pod"
	ObservableService          ObservableResource = "service"
	ObservableIngress          ObservableResource = "ingress"
	ObservablePersistentVolume ObservableResource = "pv"
	ObservableNode             ObservableResource = "node"
)

// Metadata represents common data for MongoDB record.
type Metadata struct {
	ID           bson.ObjectId `bson:"_id"`
	EventType    watch.EventType
	ResourceType ObservableResource
	Timestamp    time.Time
	UID          string
	Namespace    string
	Name         string
}

// Object represents observable kubernetes resource.
type Object interface{}

// Record represents MongoDB record (document).
type Record struct {
	Metadata
	Object Object `json:"object"`
}

// UnmarshalJSON implements Unmarshaller.
func (r *Record) UnmarshalJSON(b []byte) error {
	if err := json.Unmarshal(b, &r.Metadata); err != nil {
		return err
	}
	switch r.ResourceType {
	case ObservableNamespace:
		r.Object = new(Namespace)
	case ObservableDeployment:
		r.Object = new(Deployment)
	case ObservablePod:
		r.Object = new(Pod)
	case ObservableService:
		r.Object = new(Service)
	case ObservableIngress:
		r.Object = new(Ingress)
	case ObservablePersistentVolume:
		r.Object = new(PersistentVolume)
	case ObservableNode:
		r.Object = new(Node)
	default:
		return errors.New("unknown resource type " + string(r.ResourceType))
	}
	return json.Unmarshal(b, &struct {
		Object Object `json:"object"`
	}{Object: r.Object})
}

// Event represents common fields for kubernetes events.
type Event struct {
	Reason  string
	Type    string
	Message string
	Count   int
}
