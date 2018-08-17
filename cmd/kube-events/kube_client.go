package main

import (
	"strings"

	"github.com/containerum/kube-events/pkg/mergedwatch"
	"github.com/containerum/kube-events/pkg/transform"
	"github.com/sirupsen/logrus"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Kube struct {
	*kubernetes.Clientset
	config *rest.Config
}

func (k *Kube) WatchSupportedResources(listOptions meta_v1.ListOptions) (watch.Interface, error) {
	rqWatch, err := k.CoreV1().ResourceQuotas("").Watch(listOptions)
	if err != nil {
		return nil, err
	}
	deplWatch, err := k.AppsV1().Deployments("").Watch(listOptions)
	if err != nil {
		return nil, err
	}
	eventWatch, err := k.CoreV1().Events("").Watch(listOptions)
	if err != nil {
		return nil, err
	}
	serviceWatch, err := k.CoreV1().Services("").Watch(listOptions)
	if err != nil {
		return nil, err
	}
	ingressWatch, err := k.ExtensionsV1beta1().Ingresses("").Watch(listOptions)
	if err != nil {
		return nil, err
	}
	pvWatch, err := k.CoreV1().PersistentVolumes().Watch(listOptions)
	if err != nil {
		return nil, err
	}
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
