package main

import (
	"sync"

	"k8s.io/api/apps/v1"

	core_v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
)

type DeployFilter struct {
	generationMu  sync.Mutex
	generationMap map[types.UID]int64
}

func NewDeployFilter() *DeployFilter {
	return &DeployFilter{
		generationMap: make(map[types.UID]int64),
	}
}

func (df *DeployFilter) compareAndSwapGeneration(uid types.UID, newGeneration int64) bool {
	df.generationMu.Lock()
	defer df.generationMu.Unlock()

	if df.generationMap[uid] < newGeneration {
		df.generationMap[uid] = newGeneration
		if newGeneration > 1 {
			return true
		}
	}
	return false
}

func (df *DeployFilter) Filter(event watch.Event) bool {
	deploy, ok := event.Object.(*v1.Deployment)
	if !ok {
		return false
	}
	if event.Type == watch.Modified {
		return df.compareAndSwapGeneration(deploy.UID, deploy.Generation)
	}
	return true
}

func ResourceQuotaFilter(event watch.Event) bool {
	rq, ok := event.Object.(*core_v1.ResourceQuota)
	if !ok {
		return false
	}
	if event.Type == watch.Modified {
		return rq.Spec.Hard.Memory().Cmp(*rq.Status.Hard.Memory()) == 0 &&
			rq.Spec.Hard.Cpu().Cmp(*rq.Status.Hard.Cpu()) == 0
	}
	return true
}

func EventsFilter(event watch.Event) bool {
	switch event.Type {
	case watch.Added, watch.Error:
		//pass
	default:
		return false
	}

	kubeEvent, ok := event.Object.(*core_v1.Event)
	if !ok {
		return false
	}

	switch kubeEvent.InvolvedObject.Kind {
	case "Pod", "PersistentVolumeClaim", "Node":
		return true
	default:
		return false
	}
}

func PVCFilter(event watch.Event) bool {
	pv, ok := event.Object.(*core_v1.PersistentVolumeClaim)
	if !ok {
		return false
	}
	if event.Type == watch.Modified {
		if len(pv.Finalizers) == 0 {
			return false
		}
		if pv.DeletionTimestamp != nil {
			return false
		}
	}
	return true
}

func ErrorFilter(event watch.Event) bool {
	return event.Type != watch.Error
}
