package main

import (
	"k8s.io/kubernetes/pkg/kubelet/events"
)

type errorSet map[string]interface{}

var errorReasons = errorSet{events.FailedToCreateContainer: nil,
	events.FailedToKillPod:            nil,
	events.FailedToCreatePodContainer: nil,
	events.NetworkNotReady:            nil,
	events.FailedToInspectImage:       nil,
	events.ErrImageNeverPullPolicy:    nil,
	events.BackOffPullImage:           nil,
	events.NodeNotReady:               nil,
	events.NodeNotSchedulable:         nil,
	events.FailedAttachVolume:         nil,
	events.FailedDetachVolume:         nil,
	events.FailedMountVolume:          nil,
	events.VolumeResizeFailed:         nil,
	events.FileSystemResizeFailed:     nil,
	events.FailedUnMountVolume:        nil,
	events.FailedMapVolume:            nil,
	events.FailedUnmapDevice:          nil,
	events.WarnAlreadyMountedVolume:   nil,
	events.HostPortConflict:           nil,
	events.NodeSelectorMismatching:    nil,
	events.InsufficientFreeCPU:        nil,
	events.InsufficientFreeMemory:     nil,
	events.HostNetworkNotSupported:    nil,
	events.UnsupportedMountOption:     nil,
	events.FailedCreatePodSandBox:     nil,
	events.FailedStatusPodSandBox:     nil,
	events.FailedSync:                 nil,
	events.FailedValidation:           nil,
}

func (errs errorSet) isErrorReason(reason string) bool {
	_, isErr := errs[reason]
	return isErr
}
