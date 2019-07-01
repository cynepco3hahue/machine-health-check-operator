package main

import (
	"context"
	"flag"

	"github.com/openshift/machine-health-check-operator/pkg/operator"
	"github.com/openshift/machine-health-check-operator/pkg/version"
	"github.com/golang/glog"
	osconfigv1 "github.com/openshift/api/config/v1"
	"github.com/spf13/cobra"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	coreclientsetv1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/record"
)

var (
	startCmd = &cobra.Command{
		Use:   "start",
		Short: "Starts Machine Health Check Operator",
		Long:  "",
		Run:   runStartCmd,
	}

	startOpts struct {
		kubeconfig string
	}
)

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.PersistentFlags().StringVar(&startOpts.kubeconfig, "kubeconfig", "", "Kubeconfig file to access a remote cluster (testing only)")
}

func runStartCmd(cmd *cobra.Command, args []string) {
	flag.Set("logtostderr", "true")
	flag.Parse()

	// To help debugging, immediately log version
	glog.Infof("Version: %+v", version.Get())

	cb, err := NewClientBuilder(startOpts.kubeconfig)
	if err != nil {
		glog.Fatalf("error creating clients: %v", err)
	}
	stopCh := make(chan struct{})

	leaderelection.RunOrDie(context.TODO(), leaderelection.LeaderElectionConfig{
		Lock:          CreateResourceLock(cb, componentNamespace, componentName),
		LeaseDuration: LeaseDuration,
		RenewDeadline: RenewDeadline,
		RetryPeriod:   RetryPeriod,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(ctx context.Context) {
				ctrlCtx := CreateControllerContext(cb, stopCh, componentNamespace)
				startControllers(ctrlCtx)
				ctrlCtx.ConfigMapInformerFactory.Start(ctrlCtx.Stop)
				ctrlCtx.DeploymentInformerFactory.Start(ctrlCtx.Stop)
				ctrlCtx.ConfigInformerFactory.Start(ctrlCtx.Stop)
				close(ctrlCtx.InformersStarted)

				select {}
			},
			OnStoppedLeading: func() {
				glog.Fatalf("Leader election lost")
			},
		},
	})
	panic("unreachable")
}

func initRecorder(kubeClient kubernetes.Interface) record.EventRecorder {
	eventRecorderScheme := runtime.NewScheme()
	osconfigv1.Install(eventRecorderScheme)
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(glog.Infof)
	eventBroadcaster.StartRecordingToSink(&coreclientsetv1.EventSinkImpl{Interface: kubeClient.CoreV1().Events("")})
	return eventBroadcaster.NewRecorder(eventRecorderScheme, v1.EventSource{Component: "machinehealthcheckoperator"})
}

func startControllers(ctx *ControllerContext) {
	kubeClient := ctx.ClientBuilder.KubeClientOrDie(componentName)
	recorder := initRecorder(kubeClient)
	go operator.New(
		componentNamespace, componentName,
		config,
		ctx.ConfigMapInformerFactory.Core().V1().ConfigMaps(),
		ctx.DeploymentInformerFactory.Apps().V1().Deployments(),
		ctx.ConfigInformerFactory.Config().V1().FeatureGates(),
		ctx.ClientBuilder.KubeClientOrDie(componentName),
		ctx.ClientBuilder.OpenshiftClientOrDie(componentName),
		recorder,
	).Run(2, ctx.Stop)
}
