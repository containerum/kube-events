package main

import (
	"regexp"

	apiCore "k8s.io/api/core/v1"
	volumeEvents "k8s.io/kubernetes/pkg/controller/volume/events"
	kubeletEvents "k8s.io/kubernetes/pkg/kubelet/events"
)

type eventSet map[string]interface{}

var errorReasons = eventSet{
	kubeletEvents.FailedToCreateContainer:    nil,
	kubeletEvents.FailedToKillPod:            nil,
	kubeletEvents.FailedToCreatePodContainer: nil,
	kubeletEvents.NetworkNotReady:            nil,
	kubeletEvents.FailedToInspectImage:       nil,
	kubeletEvents.ErrImageNeverPullPolicy:    nil,
	kubeletEvents.BackOffPullImage:           nil,
	kubeletEvents.NodeNotReady:               nil,
	kubeletEvents.NodeNotSchedulable:         nil,
	kubeletEvents.FailedAttachVolume:         nil,
	kubeletEvents.FailedDetachVolume:         nil,
	kubeletEvents.FailedMountVolume:          nil,
	kubeletEvents.VolumeResizeFailed:         nil,
	kubeletEvents.FileSystemResizeFailed:     nil,
	kubeletEvents.FailedUnMountVolume:        nil,
	kubeletEvents.FailedMapVolume:            nil,
	kubeletEvents.FailedUnmapDevice:          nil,
	kubeletEvents.WarnAlreadyMountedVolume:   nil,
	kubeletEvents.HostPortConflict:           nil,
	kubeletEvents.NodeSelectorMismatching:    nil,
	kubeletEvents.InsufficientFreeCPU:        nil,
	kubeletEvents.InsufficientFreeMemory:     nil,
	kubeletEvents.HostNetworkNotSupported:    nil,
	kubeletEvents.UnsupportedMountOption:     nil,
	kubeletEvents.FailedCreatePodSandBox:     nil,
	kubeletEvents.FailedStatusPodSandBox:     nil,
	kubeletEvents.FailedSync:                 nil,
	kubeletEvents.FailedValidation:           nil,

	volumeEvents.FailedBinding:             nil,
	volumeEvents.VolumeMismatch:            nil,
	volumeEvents.VolumeFailedRecycle:       nil,
	volumeEvents.VolumeRecycled:            nil,
	volumeEvents.VolumeFailedDelete:        nil,
	volumeEvents.ProvisioningFailed:        nil,
	volumeEvents.ProvisioningCleanupFailed: nil,

	"FailedScheduling": nil,
}

var podFailedReasons = eventSet{
	kubeletEvents.FailedToCreateContainer:    nil,
	kubeletEvents.BackOffStartContainer:      nil,
	kubeletEvents.FailedToCreatePodContainer: nil,
	kubeletEvents.FailedCreatePodSandBox:     nil,
	"FailedScheduling":                       nil,
}

var podKillFailedReasons = eventSet{
	kubeletEvents.FailedToKillPod:     nil,
	kubeletEvents.ExceededGracePeriod: nil,
}

var volumeProvisionSuccessfulReasons = eventSet{
	kubeletEvents.VolumeResizeSuccess:  nil,
	volumeEvents.ProvisioningSucceeded: nil,
}

func (errs eventSet) check(reason string) bool {
	_, isErr := errs[reason]
	return isErr
}

//Whitelist of events reasons with blacklist of messages (blacklist entry is regexp)
type wlblReasonsMessages map[string][]string

var eventsWhitelist = wlblReasonsMessages{
	kubeletEvents.FailedToStartContainer:  []string{"Error: ImagePullBackOff", "Error: ErrImagePull"},
	kubeletEvents.PullingImage:            nil,
	kubeletEvents.PulledImage:             nil,
	kubeletEvents.FailedToInspectImage:    nil,
	kubeletEvents.ErrImageNeverPullPolicy: nil,
	kubeletEvents.BackOffPullImage:        nil,
	kubeletEvents.ImageGCFailed:           nil,
	kubeletEvents.InvalidDiskCapacity:     nil,
	kubeletEvents.FreeDiskSpaceFailed:     nil,

	kubeletEvents.StartedContainer:    nil,
	kubeletEvents.PreemptContainer:    nil,
	kubeletEvents.ExceededGracePeriod: nil,
	kubeletEvents.ContainerUnhealthy:  nil,

	kubeletEvents.FailedToKillPod:            nil,
	kubeletEvents.FailedToCreatePodContainer: nil,
	kubeletEvents.NetworkNotReady:            nil,
	kubeletEvents.FailedSync:                 nil,
	kubeletEvents.FailedCreatePodSandBox:     nil,
	"FailedScheduling":                       nil,

	volumeEvents.FailedBinding:             nil,
	volumeEvents.VolumeMismatch:            nil,
	volumeEvents.VolumeFailedRecycle:       nil,
	volumeEvents.VolumeFailedDelete:        nil,
	volumeEvents.ExternalProvisioning:      nil,
	volumeEvents.ProvisioningFailed:        nil,
	volumeEvents.ProvisioningCleanupFailed: nil,
	volumeEvents.ProvisioningSucceeded:     nil,
	kubeletEvents.FailedAttachVolume:       nil,
	kubeletEvents.FailedDetachVolume:       nil,
	kubeletEvents.FailedMountVolume:        nil,
	kubeletEvents.FailedUnMountVolume:      nil,
	kubeletEvents.VolumeResizeFailed:       nil,
	kubeletEvents.VolumeResizeSuccess:      nil,
	kubeletEvents.FailedMapVolume:          nil,
	kubeletEvents.WarnAlreadyMountedVolume: nil,
	kubeletEvents.SuccessfulMountVolume:    nil,
	kubeletEvents.SuccessfulUnMountVolume:  nil,

	kubeletEvents.NodeReady:               nil,
	kubeletEvents.NodeNotReady:            nil,
	kubeletEvents.NodeSelectorMismatching: nil,
}

func (errs wlblReasonsMessages) check(event *apiCore.Event) bool {
	bl, inWl := errs[event.Reason]
	if !inWl {
		return false
	}

	for _, blEntry := range bl {
		if match, _ := regexp.Match(blEntry, []byte(event.Message)); match {
			return false
		}
	}

	return true
}
