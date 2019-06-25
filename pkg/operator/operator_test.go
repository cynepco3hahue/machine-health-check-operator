package operator

import (
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	fakekube "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/utils/pointer"

	v1 "github.com/openshift/api/config/v1"
	fakeos "github.com/openshift/client-go/config/clientset/versioned/fake"
	configinformersv1 "github.com/openshift/client-go/config/informers/externalversions"
)

const (
	deploymentName  = "machine-health-check-controller"
	targetNamespace = "test-namespace"
)

func newFeatureGate(featureSet v1.FeatureSet) *v1.FeatureGate {
	return &v1.FeatureGate{
		ObjectMeta: metav1.ObjectMeta{
			Name: MachineAPIFeatureGateName,
		},
		Spec: v1.FeatureGateSpec{
			FeatureSet: featureSet,
		},
	}
}

func newOperatorConfig(techPreviewEnabled bool) *Config {
	return &Config{
		targetNamespace,
		techPreviewEnabled,
		Controllers{
			"docker.io/openshift/origin-machine-api-operator:v4.0.0",
		},
	}
}

func newImagesConfigMap() *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      machineAPIOperatorImages,
			Namespace: targetNamespace,
		},
		Data: map[string]string{"images.json": images},
	}
}

func newFakeOperator(kubeObjects []runtime.Object, osObjects []runtime.Object, stopCh <-chan struct{}) *Operator {
	kubeClient := fakekube.NewSimpleClientset(kubeObjects...)
	osClient := fakeos.NewSimpleClientset(osObjects...)

	configMapInformerFactory := informers.NewSharedInformerFactoryWithOptions(kubeClient, 2*time.Minute, informers.WithNamespace(targetNamespace))
	tweakListOptions := func(listOptions *metav1.ListOptions) {
		listOptions.LabelSelector = ManagedByLabel + "=" + ManagedByLabelOperatorValue
	}
	deploymentInformerFactory := informers.NewSharedInformerFactoryWithOptions(kubeClient, 2*time.Minute, informers.WithTweakListOptions(tweakListOptions), informers.WithNamespace(targetNamespace))
	configInformerFactory := configinformersv1.NewSharedInformerFactoryWithOptions(osClient, 2*time.Minute, configinformersv1.WithNamespace(targetNamespace))

	configMapInformer := configMapInformerFactory.Core().V1().ConfigMaps()
	featureGateInformer := configInformerFactory.Config().V1().FeatureGates()
	deploymentInformer := deploymentInformerFactory.Apps().V1().Deployments()

	optr := &Operator{
		kubeClient:             kubeClient,
		osClient:               osClient,
		configMapLister:        configMapInformer.Lister(),
		featureGateLister:      featureGateInformer.Lister(),
		deployLister:           deploymentInformer.Lister(),
		namespace:              targetNamespace,
		eventRecorder:          record.NewFakeRecorder(50),
		queue:                  workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "machineapioperator"),
		configMapCacheSynced:   configMapInformer.Informer().HasSynced,
		deployListerSynced:     deploymentInformer.Informer().HasSynced,
		featureGateCacheSynced: featureGateInformer.Informer().HasSynced,
	}

	configMapInformerFactory.Start(stopCh)
	deploymentInformerFactory.Start(stopCh)
	configInformerFactory.Start(stopCh)

	optr.syncHandler = optr.sync
	configMapInformer.Informer().AddEventHandler(optr.eventHandler())
	deploymentInformer.Informer().AddEventHandler(optr.eventHandler())
	featureGateInformer.Informer().AddEventHandler(optr.eventHandler())

	return optr
}

func TestOperatorSyncClusterAPIControllerHealthCheckController(t *testing.T) {
	cmImages := newImagesConfigMap()

	tests := []struct {
		featureGate      *v1.FeatureGate
		expectedReplicas *int32
	}{{
		featureGate:      newFeatureGate(v1.Default),
		expectedReplicas: pointer.Int32Ptr(1),
	}, {
		featureGate:      &v1.FeatureGate{},
		expectedReplicas: pointer.Int32Ptr(1),
	}, {
		featureGate:      newFeatureGate(v1.TechPreviewNoUpgrade),
		expectedReplicas: pointer.Int32Ptr(0),
	}}

	for _, tc := range tests {
		stopCh := make(<-chan struct{})
		optr := newFakeOperator([]runtime.Object{cmImages}, []runtime.Object{tc.featureGate}, stopCh)
		go optr.Run(2, stopCh)

		if err := wait.PollImmediate(1*time.Second, 5*time.Second, func() (bool, error) {
			d, err := optr.deployLister.Deployments(targetNamespace).Get(deploymentName)
			if err != nil {
				t.Logf("Failed to get %q deployment: %v", deploymentName, err)
				return false, nil
			}
			if *d.Spec.Replicas != *tc.expectedReplicas {
				t.Logf("Replicas tests failed. Expected: %d, got: %d", *tc.expectedReplicas, *d.Spec.Replicas)
				return false, nil
			}
			return true, nil
		}); err != nil {
			t.Errorf("Failed to verify %q deployment", deploymentName)
		}
	}
}
