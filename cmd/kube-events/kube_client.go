package main

import (
	"strings"

	"github.com/containerum/kube-events/pkg/transform"

	"github.com/containerum/kube-events/pkg/informerwatch"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Kube struct {
	*kubernetes.Clientset
	config *rest.Config
}

type Watchers struct {
	ResourceQuotas watch.Interface //Namespaces
	Deployments    watch.Interface
	PodEvents      watch.Interface //Pods
	Services       watch.Interface
	Ingresses      watch.Interface
	PVCs           watch.Interface //Volumes
	PVCEvents      watch.Interface
}

func (k *Kube) WatchSupportedResources() Watchers {
	informerFactory := informers.NewSharedInformerFactory(k.Clientset, 0)

	rqWatch := informerwatch.NewInformerWatch(informerFactory.Core().V1().ResourceQuotas().Informer())
	deplWatch := informerwatch.NewInformerWatch(informerFactory.Apps().V1().Deployments().Informer())
	eventWatch := informerwatch.NewInformerWatch(informerFactory.Core().V1().Events().Informer())
	serviceWatch := informerwatch.NewInformerWatch(informerFactory.Core().V1().Services().Informer())
	ingressWatch := informerwatch.NewInformerWatch(informerFactory.Extensions().V1beta1().Ingresses().Informer())
	pvcWatch := informerwatch.NewInformerWatch(informerFactory.Core().V1().PersistentVolumeClaims().Informer())

	logrus.Infof("Watching for: %s", strings.Join([]string{
		"ResourceQuota",
		"Deployment",
		"Event",
		"Service",
		"Ingress",
		"PersistentVolume",
	}, ","))

	return Watchers{
		ResourceQuotas: transform.NewFilteredWatch(rqWatch, ResourceQuotaFilter, ErrorFilter),
		Deployments:    transform.NewFilteredWatch(deplWatch, NewDeployFilter().Filter, ErrorFilter),
		PodEvents:      transform.NewFilteredWatch(eventWatch, PodEventsFilter, ErrorFilter),
		Services:       transform.NewFilteredWatch(serviceWatch, ErrorFilter),
		Ingresses:      transform.NewFilteredWatch(ingressWatch, ErrorFilter),
		PVCs:           transform.NewFilteredWatch(pvcWatch, PVCFilter, ErrorFilter),
		PVCEvents:      transform.NewFilteredWatch(eventWatch, PVCEventsFilter, ErrorFilter),
	}
}
