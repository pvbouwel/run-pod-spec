package kube

import (
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
)

var podPhaseLookup map[string]corev1.PodPhase = map[string]corev1.PodPhase{
	"":          corev1.PodPhase(""),
	"Pending":   corev1.PodPending,
	"Running":   corev1.PodRunning,
	"Succeeded": corev1.PodSucceeded,
}

func AllowedPodPhases() string {
	keys := make([]string, 0, len(podPhaseLookup))
	for k := range podPhaseLookup {
		if k == "" {
			k = "''"
		}
		keys = append(keys, k)
	}
	return strings.Join(keys, ", ")
}

func PodPhasesFromString(s string) (podPhases []corev1.PodPhase, err error) {
	podPhase := make([]corev1.PodPhase, 0)

	for _, p := range strings.Split(s, ",") {
		typedP, ok := podPhaseLookup[p]
		if !ok {
			return nil, fmt.Errorf("unknown pod phase: %s", p)
		}
		podPhase = append(podPhase, typedP)
	}
	return podPhase, nil
}
