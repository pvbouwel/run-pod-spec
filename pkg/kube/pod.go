package kube

import (
	"context"
	"fmt"
	"os"
	"strings"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"
)

func GetPodSpec(filepath string) (*corev1.Pod, error) {
	if strings.HasSuffix(filepath, ".yaml") {
		return getPodSpecYamlFile(filepath)
	}
	return nil, fmt.Errorf("unsupported file format: %s", filepath)
}

func getPodSpecYamlFile(filepath string) (*corev1.Pod, error) {
	dat, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	return getPodSpecYaml(dat)
}

func getPodSpecYaml(podSpec []byte) (*corev1.Pod, error) {
	var pod corev1.Pod
	err := yaml.Unmarshal(podSpec, &pod)
	if err != nil {
		return nil, err
	}
	if pod.Kind != "Pod" {
		return nil, fmt.Errorf("manifest should have pod kind but got %s", pod.Kind)
	}
	return &pod, nil
}

func deletePod(clientset *kubernetes.Clientset, name, namespace string) error {
	return clientset.CoreV1().Pods(namespace).Delete(context.TODO(), name, v1.DeleteOptions{})

}
