package main

import (
	"strings"

	"github.com/containerum/kube-events/pkg/informerwatch"
	"github.com/containerum/kube-events/pkg/mergedwatch"
	"github.com/containerum/kube-events/pkg/transform"
	"github.com/sirupsen/logrus"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Kube struct {
	*kubernetes.Clientset
	config *rest.Config
}

func (k *Kube) WatchSupportedResources(listOptions meta_v1.ListOptions) (watch.Interface, error) {
	informerFactory := informers.NewSharedInformerFactory(k.Clientset, 0)

	rqWatch := informerwatch.NewInformerWatch(informerFactory.Core().V1().ResourceQuotas().Informer())
	deplWatch := informerwatch.NewInformerWatch(informerFactory.Apps().V1().Deployments().Informer())
	eventWatch := informerwatch.NewInformerWatch(informerFactory.Core().V1().Events().Informer())
	serviceWatch := informerwatch.NewInformerWatch(informerFactory.Core().V1().Services().Informer())
	ingressWatch := informerwatch.NewInformerWatch(informerFactory.Extensions().V1beta1().Ingresses().Informer())
	pvWatch := informerwatch.NewInformerWatch(informerFactory.Core().V1().PersistentVolumes().Informer())
	/*nodeWatch, err := k.CoreV1().Nodes().Watch(listOptions)
	if err != nil {
		return nil, err
	}*/

	logrus.Infof("Watching for: %s", strings.Join([]string{
		"ResourceQuota",
		"Deployment",
		"Event",
		"Service",
		"Ingress",
		"PersistentVolume",
		//"Node",
	}, ","))

	mw := mergedwatch.NewMergedWatch(
		transform.NewFilteredWatch(rqWatch, ResourceQuotaFilter),
		transform.NewFilteredWatch(deplWatch, NewDeployFilter().Filter),
		transform.NewFilteredWatch(eventWatch, EventFilter),
		serviceWatch,
		ingressWatch,
		transform.NewFilteredWatch(pvWatch, PVFilter),
		//nodeWatch,
	)
	return mw, nil
}
