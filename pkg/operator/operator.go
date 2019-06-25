package operator

import (
	"fmt"
	"time"

	"github.com/golang/glog"
	osev1 "github.com/openshift/api/config/v1"
	osclientset "github.com/openshift/client-go/config/clientset/versioned"
	configinformersv1 "github.com/openshift/client-go/config/informers/externalversions/config/v1"
	configlistersv1 "github.com/openshift/client-go/config/listers/config/v1"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	appsinformersv1 "k8s.io/client-go/informers/apps/v1"
	coreinformersv1 "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	appslisterv1 "k8s.io/client-go/listers/apps/v1"
	corelistersv1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
)

const (
	// maxRetries is the number of times a key will be retried before it is dropped out of the queue.
	// With the current rate-limiter in use (5ms*2^(maxRetries-1)) the following numbers represent the times
	// a machineconfig pool is going to be requeued:
	//
	// 5ms, 10ms, 20ms, 40ms, 80ms, 160ms, 320ms, 640ms, 1.3s, 2.6s, 5.1s, 10.2s, 20.4s, 41s, 82s
	maxRetries = 15
	// machineAPIOperatorImages contains the name of the config map with machine-api-operator images
	machineAPIOperatorImages = "machine-api-operator-images"
	// ManagedByLabel contains machine-health-check-operator label key
	ManagedByLabel = "app.kubernetes.io/managed-by"
	// ManagedByLabelOperatorValue contains machine-health-check-operator label value
	ManagedByLabelOperatorValue = "machine-health-check-operator"
)

// Operator defines machine api operator.
type Operator struct {
	namespace, name string
	config          string

	kubeClient    kubernetes.Interface
	osClient      osclientset.Interface
	eventRecorder record.EventRecorder

	syncHandler func(ic string) error

	deployLister       appslisterv1.DeploymentLister
	deployListerSynced cache.InformerSynced

	featureGateLister      configlistersv1.FeatureGateLister
	featureGateCacheSynced cache.InformerSynced

	configMapLister      corelistersv1.ConfigMapLister
	configMapCacheSynced cache.InformerSynced

	// queue only ever has one item, but it has nice error handling backoff/retry semantics
	queue workqueue.RateLimitingInterface
}

// New returns a new machine config operator.
func New(
	namespace, name string,
	config string,

	configMapInformer coreinformersv1.ConfigMapInformer,
	deployInformer appsinformersv1.DeploymentInformer,
	featureGateInformer configinformersv1.FeatureGateInformer,

	kubeClient kubernetes.Interface,
	osClient osclientset.Interface,

	recorder record.EventRecorder,
) *Operator {
	optr := &Operator{
		namespace:     namespace,
		name:          name,
		kubeClient:    kubeClient,
		osClient:      osClient,
		eventRecorder: recorder,
		queue:         workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "machinehealthcheckoperator"),
	}

	deployInformer.Informer().AddEventHandler(optr.eventHandler())
	featureGateInformer.Informer().AddEventHandler(optr.eventHandler())
	configMapInformer.Informer().AddEventHandler(optr.eventHandler())

	optr.config = config
	optr.syncHandler = optr.sync

	optr.deployLister = deployInformer.Lister()
	optr.deployListerSynced = deployInformer.Informer().HasSynced

	optr.featureGateLister = featureGateInformer.Lister()
	optr.featureGateCacheSynced = featureGateInformer.Informer().HasSynced

	optr.configMapLister = configMapInformer.Lister()
	optr.configMapCacheSynced = configMapInformer.Informer().HasSynced

	return optr
}

// Run runs the machine config operator.
func (optr *Operator) Run(workers int, stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer optr.queue.ShutDown()

	glog.Info("Starting Machine Health Check Operator")
	defer glog.Info("Shutting down Machine Health Check Operator")

	if !cache.WaitForCacheSync(stopCh,
		optr.deployListerSynced,
		optr.featureGateCacheSynced,
		optr.configMapCacheSynced) {
		glog.Error("Failed to sync caches")
		return
	}
	glog.Info("Synced up caches")
	for i := 0; i < workers; i++ {
		go wait.Until(optr.worker, time.Second, stopCh)
	}

	<-stopCh
}

func (optr *Operator) eventHandler() cache.ResourceEventHandler {
	workQueueKey := fmt.Sprintf("%s/%s", optr.namespace, optr.name)
	return cache.ResourceEventHandlerFuncs{
		AddFunc:    func(obj interface{}) { optr.queue.Add(workQueueKey) },
		UpdateFunc: func(old, new interface{}) { optr.queue.Add(workQueueKey) },
		DeleteFunc: func(obj interface{}) { optr.queue.Add(workQueueKey) },
	}
}

func (optr *Operator) worker() {
	for optr.processNextWorkItem() {
	}
}

func (optr *Operator) processNextWorkItem() bool {
	key, quit := optr.queue.Get()
	if quit {
		return false
	}
	defer optr.queue.Done(key)

	glog.V(4).Infof("Processing key %s", key)
	err := optr.syncHandler(key.(string))
	optr.handleErr(err, key)

	return true
}

func (optr *Operator) handleErr(err error, key interface{}) {
	if err == nil {
		optr.queue.Forget(key)
		return
	}

	if optr.queue.NumRequeues(key) < maxRetries {
		glog.V(1).Infof("Error syncing operator %v: %v", key, err)
		optr.queue.AddRateLimited(key)
		return
	}

	utilruntime.HandleError(err)
	glog.V(1).Infof("Dropping operator %q out of the queue: %v", key, err)
	optr.queue.Forget(key)
}

func (optr *Operator) sync(key string) error {
	startTime := time.Now()
	glog.V(4).Infof("Started syncing operator %q (%v)", key, startTime)
	defer func() {
		glog.V(4).Infof("Finished syncing operator %q (%v)", key, time.Since(startTime))
	}()

	operatorConfig, err := optr.configFromInfrastructure()
	if err != nil {
		glog.Errorf("Failed getting operator config: %v", err)
		return err
	}
	return optr.syncAll(operatorConfig)
}

func (optr *Operator) configFromInfrastructure() (*Config, error) {
	cmImages, err := optr.configMapLister.ConfigMaps(optr.namespace).Get(machineAPIOperatorImages)
	if err != nil {
		return nil, err
	}

	machineAPIOperatorImage, err := getMachineAPIOperatorFromConfigMap(cmImages)
	if err != nil {
		return nil, err
	}
	glog.V(4).Infof("machine API operator images %s", machineAPIOperatorImage)

	techPreviewEnabled, err := optr.isTechPreviewEnabled()
	if err != nil {
		return nil, err
	}
	glog.V(4).Infof("tech preview enabled: %t", techPreviewEnabled)

	return &Config{
		TargetNamespace:    optr.namespace,
		TechPreviewEnabled: techPreviewEnabled,
		Controllers: Controllers{
			MachineHealthCheck: machineAPIOperatorImage,
		},
	}, nil
}

func (optr *Operator) isTechPreviewEnabled() (bool, error) {
	// Fetch the Feature
	featureGate, err := optr.featureGateLister.Get(MachineAPIFeatureGateName)

	var featureSet osev1.FeatureSet
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return false, err
		}
		glog.V(2).Infof("Failed to find feature gate %q, will use default feature set", MachineAPIFeatureGateName)
		featureSet = osev1.Default
	} else {
		featureSet = featureGate.Spec.FeatureSet
	}

	features, err := generateFeatureMap(featureSet)
	if err != nil {
		return false, err
	}

	if enabled, ok := features[FeatureGateMachineHealthCheck]; ok && enabled {
		return true, nil
	}

	return false, nil
}
