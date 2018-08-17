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
	Generation int64
	Phase      string

	model.Deployment `bson:",inline"`

	*Event `bson:",omitempty"`
}

// Pod represents observable deployment data.
type Pod struct {
	*Event `bson:",inline,omitempty"`
}

// Service represents observable service data.
type Service struct {
	model.Service `bson:",inline"`
}

// Ingress represents observable ingress data.
type Ingress struct {
	model.Ingress `bson:",inline"`
}

// PersistentVolume represents observable pv data.
type PersistentVolume struct {
	Phase       string
	Capacity    int // GB
	AccessModes []v1.PersistentVolumeAccessMode
}

// Node represents observable kubernetes node data.
type Node struct {
	Role       string
	Addresses  []v1.NodeAddress
	Conditions []v1.NodeCondition
}
