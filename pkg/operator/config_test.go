package operator

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const images = `{
	"clusterAPIControllerAWS": "docker.io/openshift/origin-aws-machine-controllers:v4.0.0",
	"clusterAPIControllerOpenStack": "docker.io/openshift/origin-openstack-machine-controllers:v4.0.0",
	"clusterAPIControllerLibvirt": "docker.io/openshift/origin-libvirt-machine-controllers:v4.0.0",
	"machineAPIOperator": "docker.io/openshift/origin-machine-api-operator:v4.0.0",
	"clusterAPIControllerBareMetal": "quay.io/openshift/origin-baremetal-machine-controllers:v4.0.0",
	"clusterAPIControllerAzure": "quay.io/openshift/origin-azure-machine-controllers:v4.0.0"
}`

func TestGetMachineAPIOperatorFromConfigMap(t *testing.T) {
	expectedImage := "docker.io/openshift/origin-machine-api-operator:v4.0.0"
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      machineAPIOperatorImages,
			Namespace: "openshift-machine-api",
		},
		Data: map[string]string{"images.json": images},
	}
	machineAPIOperatorImage, err := getMachineAPIOperatorFromConfigMap(cm)
	if err != nil {
		t.Errorf("failed getMachineAPIOperatorFromConfigMap")
	}
	if machineAPIOperatorImage != expectedImage {
		t.Errorf("failed getMachineAPIOperatorFromConfigMap. Expected: %s, got: %s", expectedImage, machineAPIOperatorImage)
	}
}
