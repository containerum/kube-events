package main

import (
	"fmt"
	"strings"
	"time"

	kubeClientModel "github.com/containerum/kube-client/pkg/model"
	"github.com/containerum/kube-events/pkg/model"
	apiApps "k8s.io/api/apps/v1"
	apiCore "k8s.io/api/core/v1"
	apiExtensions "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
)

func ObservableTypeFromObject(object runtime.Object) model.ObservableResource {
	switch object.(type) {
	case *apiCore.ResourceQuota:
		return model.ObservableNamespace
	case *apiApps.Deployment:
		return model.ObservableDeployment
	case *apiCore.Service:
		return model.ObservableService
	case *apiExtensions.Ingress:
		return model.ObservableIngress
	case *apiCore.PersistentVolumeClaim:
		return model.ObservablePersistentVolumeClaim
	case *apiCore.Secret:
		return model.ObservableSecret
	case *apiCore.ConfigMap:
		return model.ObservableConfigMap
	case *apiCore.Node:
		return model.ObservableNode
	case *apiCore.Event:
		event := object.(*apiCore.Event)
		switch event.InvolvedObject.Kind {
		case "Pod", "PersistentVolumeClaim", "Node":
			return model.ObservableEvent
		default:
			panic("Unsupported event involved object kind " + event.InvolvedObject.Kind)
		}
	default:
		panic(fmt.Sprintf("Unsupported object type %T", object))
	}
}

func MakeNamespaceRecord(event watch.Event) kubeClientModel.Event {
	rq := event.Object.(*apiCore.ResourceQuota)
	ret := kubeClientModel.Event{
		Time:              time.Now().Format(time.RFC3339),
		Kind:              kubeClientModel.EventInfo,
		ResourceName:      rq.Namespace,
		ResourceNamespace: rq.Namespace,
		ResourceUID:       string(rq.UID),
		ResourceType:      kubeClientModel.TypeNamespace,
	}
	switch event.Type {
	case watch.Added:
		ret.Name = kubeClientModel.ResourceCreated
		ret.Time = rq.CreationTimestamp.Format(time.RFC3339)
	case watch.Modified:
		ret.Name = kubeClientModel.ResourceModified
	case watch.Deleted:
		ret.Name = kubeClientModel.ResourceDeleted
	case watch.Error:
		ret.Kind = kubeClientModel.EventError
	}
	return ret
}

func MakeDeployRecord(event watch.Event) kubeClientModel.Event {
	depl := event.Object.(*apiApps.Deployment)
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
	case watch.Error:
		ret.Kind = kubeClientModel.EventError
	}
	return ret
}

func MakeEventRecord(event watch.Event) kubeClientModel.Event {
	kubeEvent := event.Object.(*apiCore.Event)
	ret := kubeClientModel.Event{
		Time:              kubeEvent.FirstTimestamp.Time.Format(time.RFC3339),
		ResourceName:      kubeEvent.InvolvedObject.Name,
		ResourceUID:       string(kubeEvent.UID),
		ResourceNamespace: kubeEvent.Namespace,
		Message:           kubeEvent.Message,
		Details:           map[string]string{},
	}

	switch kubeEvent.InvolvedObject.Kind {
	case "Pod":
		//Resource type
		ret.ResourceType = kubeClientModel.TypePod
		//Container name
		if kubeEvent.InvolvedObject.FieldPath != "" {
			ret.Details["container"] = strings.TrimSuffix(strings.TrimPrefix(kubeEvent.InvolvedObject.FieldPath, "spec.containers{"), "}")
		}
		//Event kind and name
		switch {
		case podFailedReasons.check(kubeEvent.Reason):
			ret.Kind = kubeClientModel.EventWarning
			ret.Name = "PodFailed"
		case podKillFailedReasons.check(kubeEvent.Reason):
			ret.Kind = kubeClientModel.EventWarning
			ret.Name = "PodKillFailed"
		}
	case "PersistentVolumeClaim":
		//Resource type
		ret.ResourceType = kubeClientModel.TypeVolume
		//Event kind and name
		if volumeProvisionSuccessfulReasons.check(kubeEvent.Reason) {
			ret.Kind = kubeClientModel.EventInfo
			ret.Name = "VolumeSuccessful"
		}
	case "Node":
		//Resource type
		ret.ResourceType = kubeClientModel.TypeNode
	}

	if ret.Name == "" || ret.Kind == "" {
		//Default event kind and name
		if event.Type == watch.Error {
			ret.Kind = kubeClientModel.EventError
		} else if errorReasons.check(kubeEvent.Reason) {
			ret.Kind = kubeClientModel.EventWarning
		} else {
			ret.Kind = kubeClientModel.EventInfo
		}
		ret.Name = kubeEvent.Reason
	}

	return ret
}

func MakeServiceRecord(event watch.Event) kubeClientModel.Event {
	svc := event.Object.(*apiCore.Service)
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
	case watch.Error:
		ret.Kind = kubeClientModel.EventError
	}
	return ret
}

func MakeIngressRecord(event watch.Event) kubeClientModel.Event {
	ingr := event.Object.(*apiExtensions.Ingress)
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
	case watch.Error:
		ret.Kind = kubeClientModel.EventError
	}
	return ret
}

func MakePVCRecord(event watch.Event) kubeClientModel.Event {
	ret := kubeClientModel.Event{}
	switch pvc := event.Object.(type) {
	case *apiCore.PersistentVolumeClaim:
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
		case watch.Error:
			ret.Kind = kubeClientModel.EventError
		}
	default:
		panic("unknown resource type!")
	}
	return ret
}

func MakeSecretRecord(event watch.Event) kubeClientModel.Event {
	sct := event.Object.(*apiCore.Secret)
	ret := kubeClientModel.Event{
		Time:              time.Now().Format(time.RFC3339),
		Kind:              kubeClientModel.EventInfo,
		ResourceName:      sct.Name,
		ResourceUID:       string(sct.UID),
		ResourceNamespace: sct.Namespace,
		ResourceType:      kubeClientModel.TypeSecret,
	}
	switch event.Type {
	case watch.Added:
		ret.Name = kubeClientModel.ResourceCreated
		ret.Time = sct.CreationTimestamp.Format(time.RFC3339)
	case watch.Modified:
		ret.Name = kubeClientModel.ResourceModified
	case watch.Deleted:
		ret.Name = kubeClientModel.ResourceDeleted
	case watch.Error:
		ret.Kind = kubeClientModel.EventError
	}
	return ret
}

func MakeConfigMapRecord(event watch.Event) kubeClientModel.Event {
	cm := event.Object.(*apiCore.ConfigMap)
	ret := kubeClientModel.Event{
		Time:              time.Now().Format(time.RFC3339),
		Kind:              kubeClientModel.EventInfo,
		ResourceName:      cm.Name,
		ResourceUID:       string(cm.UID),
		ResourceNamespace: cm.Namespace,
		ResourceType:      kubeClientModel.TypeConfigMap,
	}
	switch event.Type {
	case watch.Added:
		ret.Name = kubeClientModel.ResourceCreated
		ret.Time = cm.CreationTimestamp.Format(time.RFC3339)
	case watch.Modified:
		ret.Name = kubeClientModel.ResourceModified
	case watch.Deleted:
		ret.Name = kubeClientModel.ResourceDeleted
	case watch.Error:
		ret.Kind = kubeClientModel.EventError
	}
	return ret
}

func MakeNodeRecord(event watch.Event) kubeClientModel.Event {
	node := event.Object.(*apiCore.Node)
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
	case watch.Error:
		ret.Kind = kubeClientModel.EventError
	}
	return ret
}
