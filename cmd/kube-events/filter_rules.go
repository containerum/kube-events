package main

import (
	"sync"

	core_v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
)

type DeployFilter struct {
	generationMu  sync.Mutex
	generationMap map[types.UID]int64
}

func (df *DeployFilter) compareAndSwapGeneration(uid types.UID, newGeneration int64) bool {
	df.generationMu.Lock()
	defer df.generationMu.Unlock()

	if df.generationMap[uid] < newGeneration {
		df.generationMap[uid] = newGeneration
		return true
	}
	return false
}

func (df *DeployFilter) Filter(event watch.Event) bool {
	deploy, ok := event.Object.(*core_v1.Event)
	if !ok {
		return false
	}
	return df.compareAndSwapGeneration(deploy.UID, deploy.Generation)
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

func EventFilter(event watch.Event) bool {
	kubeEvent, ok := event.Object.(*core_v1.Event)
	if !ok {
		return false
	}

	switch kubeEvent.InvolvedObject.Kind {
	case "Pod", "Deployment":
		return true
	default:
		return false
	}
}

func PVFilter(event watch.Event) bool {
	pv, ok := event.Object.(*core_v1.PersistentVolume)
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
