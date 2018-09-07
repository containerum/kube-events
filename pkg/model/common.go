package model

// ObservableResource is set of constants representing kubernetes resource for watching.
type ObservableResource string

const (
	ObservableNamespace             ObservableResource = "namespace"
	ObservableDeployment            ObservableResource = "deployment"
	ObservableEvent                 ObservableResource = "event"
	ObservableService               ObservableResource = "service"
	ObservableIngress               ObservableResource = "ingress"
	ObservablePersistentVolumeClaim ObservableResource = "pvc"
	ObservableNode                  ObservableResource = "node"
)
