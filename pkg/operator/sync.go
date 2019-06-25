package operator

import (
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	appsclientv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	"k8s.io/utils/pointer"

	"github.com/golang/glog"
)

const (
	deploymentRolloutPollInterval = time.Second
	deploymentRolloutTimeout      = 5 * time.Minute
)

func (optr *Operator) syncAll(config *Config) error {
	controller := newDeployment(config, config.TechPreviewEnabled)
	updated, err := applyDeploymentReplicas(optr.kubeClient.AppsV1(), controller)
	if err != nil {
		return err
	}
	if updated {
		glog.V(4).Infof("Update deployment %s with replicas %d", controller.Name, *controller.Spec.Replicas)
		return optr.waitForDeploymentRollout(controller)
	}
	return nil
}

func (optr *Operator) waitForDeploymentRollout(resource *appsv1.Deployment) error {
	return wait.Poll(deploymentRolloutPollInterval, deploymentRolloutTimeout, func() (bool, error) {
		d, err := optr.deployLister.Deployments(resource.Namespace).Get(resource.Name)
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		if err != nil {
			// Do not return error here, as we could be updating the API Server itself, in which case we
			// want to continue waiting.
			glog.Errorf("Error getting Deployment %q during rollout: %v", resource.Name, err)
			return false, nil
		}

		if d.DeletionTimestamp != nil {
			return false, fmt.Errorf("deployment %q is being deleted", resource.Name)
		}

		if d.Generation <= d.Status.ObservedGeneration && d.Status.UpdatedReplicas == d.Status.Replicas && d.Status.UnavailableReplicas == 0 {
			return true, nil
		}
		glog.V(4).Infof("Deployment %q is not ready. status: (replicas: %d, updated: %d, ready: %d, unavailable: %d)", d.Name, d.Status.Replicas, d.Status.UpdatedReplicas, d.Status.ReadyReplicas, d.Status.UnavailableReplicas)
		return false, nil
	})
}

func newDeployment(config *Config, techPreviewEnabled bool) *appsv1.Deployment {
	replicas := int32(1)
	if techPreviewEnabled {
		replicas = int32(0)
	}

	template := newPodTemplateSpec(config)

	// we do not want to create deployment when it does not have any containers
	if template == nil {
		return nil
	}

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "machine-health-check-controller",
			Namespace: config.TargetNamespace,
			Labels: map[string]string{
				ManagedByLabel: ManagedByLabelOperatorValue,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					ManagedByLabel: ManagedByLabelOperatorValue,
				},
			},
			Template: *template,
		},
	}
}

func newPodTemplateSpec(config *Config) *corev1.PodTemplateSpec {
	containers := newContainers(config)

	// we do not want to create deployment when it does not have any containers
	if len(containers) == 0 {
		return nil
	}

	tolerations := []corev1.Toleration{
		{
			Key:    "node-role.kubernetes.io/master",
			Effect: corev1.TaintEffectNoSchedule,
		},
		{
			Key:      "CriticalAddonsOnly",
			Operator: corev1.TolerationOpExists,
		},
		{
			Key:               "node.kubernetes.io/not-ready",
			Effect:            corev1.TaintEffectNoExecute,
			Operator:          corev1.TolerationOpExists,
			TolerationSeconds: pointer.Int64Ptr(120),
		},
		{
			Key:               "node.kubernetes.io/unreachable",
			Effect:            corev1.TaintEffectNoExecute,
			Operator:          corev1.TolerationOpExists,
			TolerationSeconds: pointer.Int64Ptr(120),
		},
	}

	return &corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				ManagedByLabel: ManagedByLabelOperatorValue,
			},
		},
		Spec: corev1.PodSpec{
			Containers:        containers,
			PriorityClassName: "system-node-critical",
			NodeSelector:      map[string]string{"node-role.kubernetes.io/master": ""},
			SecurityContext: &corev1.PodSecurityContext{
				RunAsNonRoot: pointer.BoolPtr(true),
				RunAsUser:    pointer.Int64Ptr(65534),
			},
			ServiceAccountName: "machine-api-controllers",
			Tolerations:        tolerations,
		},
	}
}

func newContainers(config *Config) []corev1.Container {
	resources := corev1.ResourceRequirements{
		Requests: map[corev1.ResourceName]resource.Quantity{
			corev1.ResourceMemory: resource.MustParse("20Mi"),
			corev1.ResourceCPU:    resource.MustParse("10m"),
		},
	}
	args := []string{
		"--logtostderr=true",
		"--v=3",
		// Available only in 4.2
		//fmt.Sprintf("--namespace=%s", config.TargetNamespace),
	}

	return []corev1.Container{
		corev1.Container{
			Name:      "machine-health-check-controller",
			Image:     config.Controllers.MachineHealthCheck,
			Command:   []string{"/machine-healthcheck"},
			Args:      args,
			Resources: resources,
		},
	}
}

// applyDeploymentReplicas applies the required deployment replicas to the cluster
func applyDeploymentReplicas(client appsclientv1.DeploymentsGetter, deployment *appsv1.Deployment) (bool, error) {
	existing, err := client.Deployments(deployment.Namespace).Get(deployment.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err := client.Deployments(deployment.Namespace).Create(deployment)
		return true, err
	}
	if err != nil {
		return false, err
	}

	modified := false
	if *existing.Spec.Replicas != *deployment.Spec.Replicas {
		_, err = client.Deployments(deployment.Namespace).Update(deployment)
		if err != nil {
			return modified, err
		}
		modified = true
	}
	return modified, nil
}
