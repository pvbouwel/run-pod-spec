package kube

// https://jiminbyun.medium.com/getting-started-with-client-go-building-a-kubernetes-pod-watcher-in-go-caa2be8623eb

import (
	"fmt" // For formatting error messages
	"log" // For logging messages

	// For joining file paths
	"os"

	"k8s.io/client-go/kubernetes"      // The core client-go package that provides the Clientset
	"k8s.io/client-go/rest"            // Used for in-cluster config
	"k8s.io/client-go/tools/clientcmd" // Used for loading kubeconfig files (out-of-cluster)
	// Utility to find user's home directory
)

// GetClientSet returns a *kubernetes.Clientset.
// It tries to get the configuration in this order:
// 1. In-cluster configuration (if running inside a Kubernetes Pod).
// 2. Out-of-cluster configuration (from the lcoal kubeconfig file, usually ~/.kube/config).
func GetClientSet() (*kubernetes.Clientset, error) {
	// First, try to get in-cluster config.
	// This is how your program connects if it's deployed as a Pod inside a Kubernetes cluster.
	config, err := rest.InClusterConfig()
	if err == nil {
		log.Println("Using in-cluster Kubernetes configuration.")
		// If successful, create and return the Clientset using this config.
		return kubernetes.NewForConfig(config)
	}

	// If in-cluster config failed (meaning we're likely running locally),
	// fall back to using the local kubeconfig file.
	log.Println("Using out-of-cluster Kubernetes configuration (kubeconfig).")
	var kubeconfigPath = os.Getenv("KUBECONFIG")
	// Find the user's home directory to lcoate the .kube/config file.
	if kubeconfigPath == "" {
		return nil, fmt.Errorf("KUBECONFIG not set (%s)", kubeconfigPath)
	}

	// Build the configuration from the kubeconfig file.
	// The first argument "" means "use the current context in the kubeconfig file".
	config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		// If there's an error building the config, return it.
		return nil, fmt.Errorf("failed to build kubeconfig from %s: %w", kubeconfigPath, err)
	}

	// Finally, create and return the Clientset using the loaded config.
	return kubernetes.NewForConfig(config)
}
