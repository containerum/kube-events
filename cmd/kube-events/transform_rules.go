package main

import (
	"fmt"
	"time"

	kubeClientModel "github.com/containerum/kube-client/pkg/model"
	"github.com/containerum/kube-events/pkg/model"
	apps_v1 "k8s.io/api/apps/v1"
	core_v1 "k8s.io/api/core/v1"
	extensions_v1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
)

func ObservableTypeFromObject(object runtime.Object) model.ObservableResource {
	switch object.(type) {
	case *core_v1.ResourceQuota:
		return model.ObservableNamespace
	case *apps_v1.Deployment:
		return model.ObservableDeployment
	case *core_v1.Service:
		return model.ObservableService
	case *extensions_v1beta1.Ingress:
		return model.ObservableIngress
	case *core_v1.PersistentVolumeClaim:
		return model.ObservablePersistentVolumeClaim
	case *core_v1.Node:
		return model.ObservableNode
	case *core_v1.Event:
		event := object.(*core_v1.Event)
		switch event.InvolvedObject.Kind {
		case "Pod", "PersistentVolumeClaim":
			return model.ObservableEvent
		default:
			panic("Unsupported event involved object kind " + event.InvolvedObject.Kind)
		}
	default:
		panic(fmt.Sprintf("Unsupported object type %T", object))
	}
}

func MakeNamespaceRecord(event watch.Event) kubeClientModel.Event {
	rq := event.Object.(*core_v1.ResourceQuota)
	ret := kubeClientModel.Event{
		Time:         time.Now().Format(time.RFC3339),
		Kind:         kubeClientModel.EventInfo,
		ResourceName: rq.Namespace,
		ResourceUID:  string(rq.UID),
		ResourceType: kubeClientModel.TypeNamespace,
	}
	switch event.Type {
	case watch.Added:
		ret.Name = kubeClientModel.ResourceCreated
		ret.Time = rq.CreationTimestamp.Format(time.RFC3339)
	case watch.Modified:
		ret.Name = kubeClientModel.ResourceModified
	case watch.Deleted:
		ret.Name = kubeClientModel.ResourceDeleted
	}
	return ret
}

func MakeDeployRecord(event watch.Event) kubeClientModel.Event {
	depl := event.Object.(*apps_v1.Deployment)
	ret := kubeClientModel.Event{
		Time:              time.Now().Format(time.RFC3339),
		Kind:              kubeClientModel.EventInfo,
		ResourceName:      depl.Name,
		ResourceUID:       string(depl.UID),
		ResourceNamespace: depl.Namespace,
		ResourceType:      kubeClientModel.TypeDeployment,
	}
	switch event.Type {
	case watch.Added:
		ret.Name = kubeClientModel.ResourceCreated
		ret.Time = depl.CreationTimestamp.Format(time.RFC3339)
	case watch.Modified:
		ret.Name = kubeClientModel.ResourceModified
	case watch.Deleted:
		ret.Name = kubeClientModel.ResourceDeleted
	}
	return ret
}

func MakeEventRecord(event watch.Event) kubeClientModel.Event {
	kubeEvent := event.Object.(*core_v1.Event)
	ret := kubeClientModel.Event{
		Time:              kubeEvent.FirstTimestamp.Time.Format(time.RFC3339),
		ResourceName:      kubeEvent.InvolvedObject.Name,
		ResourceUID:       string(kubeEvent.UID),
		ResourceNamespace: kubeEvent.Namespace,
		Message:           kubeEvent.Message,
		Details: map[string]string{
			"reason": kubeEvent.Reason,
		},
	}

	switch kubeEvent.InvolvedObject.Kind {
	case "Pod":
		ret.ResourceType = kubeClientModel.TypePod
	case "PersistentVolumeClaim":
		ret.ResourceType = kubeClientModel.TypeVolume
	}

	if errorReasons.isErrorReason(kubeEvent.Reason) {
		ret.Kind = kubeClientModel.EventWarning
	} else {
		ret.Kind = kubeClientModel.EventInfo
	}
	return ret
}

func MakeServiceRecord(event watch.Event) kubeClientModel.Event {
	svc := event.Object.(*core_v1.Service)
	ret := kubeClientModel.Event{
		Time:              time.Now().Format(time.RFC3339),
		Kind:              kubeClientModel.EventInfo,
		ResourceName:      svc.Name,
		ResourceUID:       string(svc.UID),
		ResourceNamespace: svc.Namespace,
		ResourceType:      kubeClientModel.TypeService,
	}
	switch event.Type {
	case watch.Added:
		ret.Name = kubeClientModel.ResourceCreated
		ret.Time = svc.CreationTimestamp.Format(time.RFC3339)
	case watch.Modified:
		ret.Name = kubeClientModel.ResourceModified
	case watch.Deleted:
		ret.Name = kubeClientModel.ResourceDeleted
	}
	return ret
}

func MakeIngressRecord(event watch.Event) kubeClientModel.Event {
	ingr := event.Object.(*extensions_v1beta1.Ingress)
	ret := kubeClientModel.Event{
		Time:              time.Now().Format(time.RFC3339),
		Kind:              kubeClientModel.EventInfo,
		ResourceName:      ingr.Name,
		ResourceUID:       string(ingr.UID),
		ResourceNamespace: ingr.Namespace,
		ResourceType:      kubeClientModel.TypeIngress,
	}
	switch event.Type {
	case watch.Added:
		ret.Name = kubeClientModel.ResourceCreated
		ret.Time = ingr.CreationTimestamp.Format(time.RFC3339)
	case watch.Modified:
		ret.Name = kubeClientModel.ResourceModified
	case watch.Deleted:
		ret.Name = kubeClientModel.ResourceDeleted
	}
	return ret
}

func MakePVCRecord(event watch.Event) kubeClientModel.Event {
	ret := kubeClientModel.Event{}
	switch pvc := event.Object.(type) {
	case *core_v1.PersistentVolumeClaim:
		ret = kubeClientModel.Event{
			Time:              time.Now().Format(time.RFC3339),
			Kind:              kubeClientModel.EventInfo,
			ResourceName:      pvc.Name,
			ResourceUID:       string(pvc.UID),
			ResourceNamespace: pvc.Namespace,
			ResourceType:      kubeClientModel.TypeVolume,
		}
		switch event.Type {
		case watch.Added:
			ret.Name = kubeClientModel.ResourceCreated
			ret.Time = pvc.CreationTimestamp.Format(time.RFC3339)
		case watch.Modified:
			ret.Name = kubeClientModel.ResourceModified
		case watch.Deleted:
			ret.Name = kubeClientModel.ResourceDeleted
		}
	default:
		panic("unknown resource type!")
	}
	return ret
}

func MakeNodeRecord(event watch.Event) kubeClientModel.Event {
	//TODO
	node := event.Object.(*core_v1.Node)
	ret := kubeClientModel.Event{
		Time:         time.Now().Format(time.RFC3339),
		Kind:         kubeClientModel.EventInfo,
		ResourceName: node.Name,
		ResourceType: kubeClientModel.TypeNode,
	}
	switch event.Type {
	case watch.Added:
		ret.Name = kubeClientModel.ResourceCreated
	case watch.Modified:
		ret.Name = kubeClientModel.ResourceModified
	case watch.Deleted:
		ret.Name = kubeClientModel.ResourceDeleted
	}
	return ret
}
