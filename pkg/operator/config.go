package operator

import (
	"encoding/json"
	"fmt"

	corev1 "k8s.io/api/core/v1"
)

const imageJSON = "images.json"

// Provider contains provider type
type Provider string

// Config contains configuration for MHCO
type Config struct {
	TargetNamespace    string
	TechPreviewEnabled bool
	Controllers        Controllers
}

// Controllers contains controllers images
type Controllers struct {
	MachineHealthCheck string
}

// Images allows build systems to inject images for MAO components
type Images struct {
	MachineAPIOperator            string `json:"machineAPIOperator"`
	ClusterAPIControllerAWS       string `json:"clusterAPIControllerAWS"`
	ClusterAPIControllerOpenStack string `json:"clusterAPIControllerOpenStack"`
	ClusterAPIControllerLibvirt   string `json:"clusterAPIControllerLibvirt"`
	ClusterAPIControllerBareMetal string `json:"clusterAPIControllerBareMetal"`
	ClusterAPIControllerAzure     string `json:"clusterAPIControllerAzure"`
}

func getImagesFromConfigMap(cmImages *corev1.ConfigMap) (*Images, error) {
	data, ok := cmImages.Data[imageJSON]
	if !ok {
		return nil, fmt.Errorf("config map %s does not have data with key %s", cmImages.Name, imageJSON)
	}

	var i Images
	if err := json.Unmarshal([]byte(data), &i); err != nil {
		return nil, err
	}
	return &i, nil
}

func getMachineAPIOperatorFromConfigMap(cmImages *corev1.ConfigMap) (string, error) {
	images, err := getImagesFromConfigMap(cmImages)
	if err != nil {
		return "", err
	}

	if images.MachineAPIOperator == "" {
		return "", fmt.Errorf("failed gettingMachineAPIOperator image. It is empty")
	}
	return images.MachineAPIOperator, nil
}
