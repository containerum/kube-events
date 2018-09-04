package model

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
