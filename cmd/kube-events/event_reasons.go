package main

import (
	"regexp"

	core_v1 "k8s.io/api/core/v1"
	volume_events "k8s.io/kubernetes/pkg/controller/volume/events"
	kubelet_events "k8s.io/kubernetes/pkg/kubelet/events"
)

type eventSet map[string]interface{}

var errorReasons = eventSet{
	kubelet_events.FailedToCreateContainer:    nil,
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

	"FailedScheduling": nil,
}

var podFailedReasons = eventSet{
	kubelet_events.FailedToCreateContainer:    nil,
	kubelet_events.BackOffStartContainer:      nil,
	kubelet_events.FailedToCreatePodContainer: nil,
	kubelet_events.FailedCreatePodSandBox:     nil,
	"FailedScheduling":                        nil,
}

var podKillFailedReasons = eventSet{
	kubelet_events.FailedToKillPod:     nil,
	kubelet_events.ExceededGracePeriod: nil,
}

var volumeProvisionSuccessfulReasons = eventSet{
	kubelet_events.VolumeResizeSuccess:  nil,
	volume_events.ProvisioningSucceeded: nil,
}

func (errs eventSet) check(reason string) bool {
	_, isErr := errs[reason]
	return isErr
}

//Whitelist of events reasons with blacklist of messages (blacklist entry is regexp)
type wlblReasonsMessages map[string][]string

var eventsWhitelist = wlblReasonsMessages{
	kubelet_events.FailedToStartContainer:  []string{"Error: ImagePullBackOff", "Error: ErrImagePull"},
	kubelet_events.PullingImage:            nil,
	kubelet_events.PulledImage:             nil,
	kubelet_events.FailedToInspectImage:    nil,
	kubelet_events.ErrImageNeverPullPolicy: nil,
	kubelet_events.BackOffPullImage:        nil,
	kubelet_events.ImageGCFailed:           nil,
	kubelet_events.InvalidDiskCapacity:     nil,
	kubelet_events.FreeDiskSpaceFailed:     nil,

	kubelet_events.StartedContainer:    nil,
	kubelet_events.PreemptContainer:    nil,
	kubelet_events.ExceededGracePeriod: nil,
	kubelet_events.ContainerUnhealthy:  nil,

	kubelet_events.FailedToKillPod:            nil,
	kubelet_events.FailedToCreatePodContainer: nil,
	kubelet_events.NetworkNotReady:            nil,
	kubelet_events.FailedSync:                 nil,
	kubelet_events.FailedCreatePodSandBox:     nil,
	"Scheduled":                               nil,
	"FailedScheduling":                        nil,

	volume_events.FailedBinding:             nil,
	volume_events.VolumeMismatch:            nil,
	volume_events.VolumeFailedRecycle:       nil,
	volume_events.VolumeFailedDelete:        nil,
	volume_events.ExternalProvisioning:      nil,
	volume_events.ProvisioningFailed:        nil,
	volume_events.ProvisioningCleanupFailed: nil,
	volume_events.ProvisioningSucceeded:     nil,
	kubelet_events.FailedAttachVolume:       nil,
	kubelet_events.FailedDetachVolume:       nil,
	kubelet_events.FailedMountVolume:        nil,
	kubelet_events.FailedUnMountVolume:      nil,
	kubelet_events.VolumeResizeFailed:       nil,
	kubelet_events.VolumeResizeSuccess:      nil,
	kubelet_events.FailedMapVolume:          nil,
	kubelet_events.WarnAlreadyMountedVolume: nil,
	kubelet_events.SuccessfulMountVolume:    nil,
	kubelet_events.SuccessfulUnMountVolume:  nil,

	kubelet_events.NodeReady:               nil,
	kubelet_events.NodeNotReady:            nil,
	kubelet_events.NodeSelectorMismatching: nil,
}

func (errs wlblReasonsMessages) check(event *core_v1.Event) bool {
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
