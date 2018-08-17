package main

import (
	"fmt"
	"time"

	kubeAPIModel "git.containerum.net/ch/kube-api/pkg/model"
	kubeClientModel "github.com/containerum/kube-client/pkg/model"
	"github.com/containerum/kube-events/pkg/model"
	"github.com/globalsign/mgo/bson"
	apps_v1 "k8s.io/api/apps/v1"
	core_v1 "k8s.io/api/core/v1"
	extensions_v1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
)

func ObservableTypeFromObject(object runtime.Object) model.ObservableResource {
	switch object.(type) {
	case *core_v1.Namespace:
		return model.ObservableNamespace
	case *apps_v1.Deployment:
		return model.ObservableDeployment
	case *core_v1.Service:
		return model.ObservableService
	case *extensions_v1beta1.Ingress:
		return model.ObservableIngress
	case *core_v1.PersistentVolume:
		return model.ObservablePersistentVolume
	case *core_v1.Node:
		return model.ObservableNode
	case *core_v1.Event:
		event := object.(*core_v1.Event)
		switch event.InvolvedObject.Kind {
		case "Pod":
			return model.ObservablePod
		case "Deployment":
			return model.ObservableDeployment
		default:
			panic("Unsupported event involved object kind " + event.InvolvedObject.Kind)
		}
	default:
		panic(fmt.Sprintf("Unsupported object type %T", object))
	}
}

func ExtractMetadata(event watch.Event) model.Metadata {
	mdAccessor := meta.NewAccessor()
	namespace, err := mdAccessor.Namespace(event.Object)
	if err != nil {
		panic(err)
	}
	name, err := mdAccessor.Name(event.Object)
	if err != nil {
		panic(err)
	}
	uid, err := mdAccessor.UID(event.Object)
	if err != nil {
		panic(err)
	}
	return model.Metadata{
		ID:           bson.NewObjectId(),
		EventType:    event.Type,
		ResourceType: ObservableTypeFromObject(event.Object),
		Timestamp:    time.Now().UTC(),
		UID:          string(uid),
		Namespace:    namespace,
		Name:         name,
	}
}

func extractResourceList(krl core_v1.ResourceList) kubeClientModel.Resource {
	return kubeClientModel.Resource{
		CPU:    uint(krl.Cpu().MilliValue()),
		Memory: uint(krl.Memory().ScaledValue(resource.Mega)),
	}
}

func extractEvent(e *core_v1.Event) *model.Event {
	return &model.Event{
		Reason:  e.Reason,
		Type:    e.Type,
		Message: e.Message,
		Count:   int(e.Count),
	}
}

func MakeNamespaceRecord(event watch.Event) model.Record {
	var ret model.Record
	ret.Metadata = ExtractMetadata(event)
	rq := event.Object.(*core_v1.ResourceQuota)
	ret.Object = &model.Namespace{
		Quota: kubeClientModel.Resources{
			Hard: extractResourceList(rq.Spec.Hard),
			Used: func() *kubeClientModel.Resource {
				rl := extractResourceList(rq.Status.Used)
				return &rl
			}(),
		},
	}
	return ret
}

func MakeDeployRecord(event watch.Event) model.Record {
	var ret model.Record
	ret.Metadata = ExtractMetadata(event)
	var deploy model.Deployment
	switch event.Object.(type) {
	case *apps_v1.Deployment:
		var err error
		kubeDeploy := event.Object.(*apps_v1.Deployment)
		parsedDeploy, err := kubeAPIModel.ParseKubeDeployment(kubeDeploy, false)
		if err != nil {
			panic(err)
		}
		deploy = model.Deployment{
			Generation: kubeDeploy.Generation,

			Deployment: *parsedDeploy,
		}
	case *core_v1.Event:
		kubeEvent := event.Object.(*core_v1.Event)
		deploy = model.Deployment{
			Event: extractEvent(kubeEvent),
		}

		// fix object reference
		ret.UID = string(kubeEvent.InvolvedObject.UID)
		ret.Name = kubeEvent.InvolvedObject.Name
		ret.Namespace = kubeEvent.InvolvedObject.Namespace
	default:
		panic(fmt.Sprintf("Invalid event type %T for deploy", event.Object))
	}

	ret.Object = &deploy
	return ret
}

func MakePodRecord(event watch.Event) model.Record {
	var ret model.Record
	ret.Metadata = ExtractMetadata(event)
	kubeEvent := event.Object.(*core_v1.Event)
	ret.Object = &model.Pod{
		Event: extractEvent(kubeEvent),
	}

	// fix object reference
	ret.UID = string(kubeEvent.InvolvedObject.UID)
	ret.Name = kubeEvent.InvolvedObject.Name
	ret.Namespace = kubeEvent.InvolvedObject.Namespace
	return ret
}

func MakeServiceRecord(event watch.Event) model.Record {
	var ret model.Record
	ret.Metadata = ExtractMetadata(event)

	parsedService, err := kubeAPIModel.ParseKubeService(event.Object.(*core_v1.Service), false)
	if err != nil {
		panic(err)
	}
	ret.Object = &model.Service{
		Service: *parsedService.Service,
	}
	return ret
}

func MakeIngressRecord(event watch.Event) model.Record {
	var ret model.Record
	ret.Metadata = ExtractMetadata(event)

	parsedIngress, err := kubeAPIModel.ParseKubeIngress(event.Object.(*extensions_v1beta1.Ingress), false)
	if err != nil {
		panic(err)
	}
	ret.Object = &model.Ingress{
		Ingress: *parsedIngress,
	}
	return ret
}

func MakePVRecord(event watch.Event) model.Record {
	var ret model.Record
	ret.Metadata = ExtractMetadata(event)
	pv := event.Object.(*core_v1.PersistentVolume)
	storage := pv.Spec.Capacity["storage"]

	ret.Object = &model.PersistentVolume{
		Phase:       string(pv.Status.Phase),
		Capacity:    int(storage.ScaledValue(resource.Giga)),
		AccessModes: pv.Spec.AccessModes,
	}
	return ret
}

func MakeNodeRecord(event watch.Event) model.Record {
	var ret model.Record
	ret.Metadata = ExtractMetadata(event)
	node := event.Object.(*core_v1.Node)

	ret.Object = &model.Node{
		Role:       node.Labels["role"],
		Addresses:  node.Status.Addresses,
		Conditions: node.Status.Conditions,
	}
	return ret
}
