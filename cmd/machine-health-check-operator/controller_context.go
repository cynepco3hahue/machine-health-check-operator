package main

import (
	"time"

	"github.com/openshift/machine-health-check-operator/pkg/operator"
	configinformersv1 "github.com/openshift/client-go/config/informers/externalversions"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/informers"
)

// ControllerContext stores all the informers for a variety of kubernetes objects.
type ControllerContext struct {
	ClientBuilder *ClientBuilder

	DeploymentInformerFactory informers.SharedInformerFactory
	ConfigMapInformerFactory  informers.SharedInformerFactory
	ConfigInformerFactory     configinformersv1.SharedInformerFactory

	AvailableResources map[schema.GroupVersionResource]bool

	Stop <-chan struct{}

	InformersStarted chan struct{}

	ResyncPeriod func() time.Duration
}

// CreateControllerContext creates the ControllerContext with the ClientBuilder.
func CreateControllerContext(cb *ClientBuilder, stop <-chan struct{}, targetNamespace string) *ControllerContext {
	kubeClient := cb.KubeClientOrDie("kube-shared-informer")
	configClient := cb.OpenshiftClientOrDie("config-shared-informer")

	configMapInformerFactory := informers.NewSharedInformerFactoryWithOptions(kubeClient, resyncPeriod()(), informers.WithNamespace(targetNamespace))
	tweakListOptions := func(listOptions *metav1.ListOptions) {
		listOptions.LabelSelector = operator.ManagedByLabel + "=" + operator.ManagedByLabelOperatorValue
	}
	deploymentInformerFactory := informers.NewSharedInformerFactoryWithOptions(kubeClient, resyncPeriod()(), informers.WithTweakListOptions(tweakListOptions), informers.WithNamespace(targetNamespace))

	configInformerFactory := configinformersv1.NewSharedInformerFactoryWithOptions(configClient, resyncPeriod()(), configinformersv1.WithNamespace(targetNamespace))

	return &ControllerContext{
		ClientBuilder:             cb,
		DeploymentInformerFactory: deploymentInformerFactory,
		ConfigMapInformerFactory:  configMapInformerFactory,
		ConfigInformerFactory:     configInformerFactory,
		Stop:                      stop,
		InformersStarted:          make(chan struct{}),
		ResyncPeriod:              resyncPeriod(),
	}
}
