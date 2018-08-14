package model

import (
	"github.com/containerum/kube-client/pkg/model"
	"k8s.io/api/core/v1"
)

// Namespace represents observable namespace data.
type Namespace struct {
	Quota model.Resources
}

// Deployment represents observable deployment data.
type Deployment struct {
	Generation int
	Phase      string

	model.Deployment

	*Event
}

// Pod represents observable deployment data.
type Pod struct {
	*Event
}

// Service represents observable service data.
type Service struct {
	model.Service
}

// Ingress represents observable ingress data.
type Ingress struct {
	model.Ingress
}

// PersistentVolume represents observable pv data.
type PersistentVolume struct {
	Phase string

	model.Volume
}

// Node represents observable kubernetes node data.
type Node struct {
	Role       string
	Addresses  []v1.NodeAddress
	Conditions []v1.NodeCondition
}
