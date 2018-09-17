package main

import (
	volume_events "k8s.io/kubernetes/pkg/controller/volume/events"
	kubelet_events "k8s.io/kubernetes/pkg/kubelet/events"
)

type errorSet map[string]interface{}

var errorReasons = errorSet{kubelet_events.FailedToCreateContainer: nil,
	kubelet_events.FailedToKillPod:            nil,
	kubelet_events.FailedToCreatePodContainer: nil,
	kubelet_events.NetworkNotReady:            nil,
	kubelet_events.FailedToInspectImage:       nil,
	kubelet_events.ErrImageNeverPullPolicy:    nil,
	kubelet_events.BackOffPullImage:           nil,
	kubelet_events.NodeNotReady:               nil,
	kubelet_events.NodeNotSchedulable:         nil,
	kubelet_events.FailedAttachVolume:         nil,
	kubelet_events.FailedDetachVolume:         nil,
	kubelet_events.FailedMountVolume:          nil,
	kubelet_events.VolumeResizeFailed:         nil,
	kubelet_events.FileSystemResizeFailed:     nil,
	kubelet_events.FailedUnMountVolume:        nil,
	kubelet_events.FailedMapVolume:            nil,
	kubelet_events.FailedUnmapDevice:          nil,
	kubelet_events.WarnAlreadyMountedVolume:   nil,
	kubelet_events.HostPortConflict:           nil,
	kubelet_events.NodeSelectorMismatching:    nil,
	kubelet_events.InsufficientFreeCPU:        nil,
	kubelet_events.InsufficientFreeMemory:     nil,
	kubelet_events.HostNetworkNotSupported:    nil,
	kubelet_events.UnsupportedMountOption:     nil,
	kubelet_events.FailedCreatePodSandBox:     nil,
	kubelet_events.FailedStatusPodSandBox:     nil,
	kubelet_events.FailedSync:                 nil,
	kubelet_events.FailedValidation:           nil,

	volume_events.FailedBinding:             nil,
	volume_events.VolumeMismatch:            nil,
	volume_events.VolumeFailedRecycle:       nil,
	volume_events.VolumeRecycled:            nil,
	volume_events.VolumeFailedDelete:        nil,
	volume_events.ProvisioningFailed:        nil,
	volume_events.ProvisioningCleanupFailed: nil,
}

func (errs errorSet) isErrorReason(reason string) bool {
	_, isErr := errs[reason]
	return isErr
}
