package kube

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type RunPodOptions struct {
	//The amount of time we allow to pass for creating the pod object on Kubernetes
	CreateTimeout time.Duration

	//The amount of time we allow to run the created pod upon reaching completion
	RunTimeout time.Duration

	//Whether the Pod should be cleaned up at the end of the run
	CleanupPod bool

	//Whether a pre-existing pod should be replaced
	ReplaceOldPod bool
}

func deleteLingeringPod(ctx context.Context, clientset *kubernetes.Clientset, namespace, name string, createTimeout time.Duration) error {
	getDeadlineCtx, cancelGet := context.WithTimeout(ctx, createTimeout)
	defer cancelGet()
	pod, err := clientset.CoreV1().Pods(namespace).Get(getDeadlineCtx, name, metav1.GetOptions{})
	if err != nil {
		errMsg := err.Error()
		if strings.HasPrefix(errMsg, "pods \"") && strings.HasSuffix(errMsg, " not found") {
			slog.Debug("No old pod", "name", name, "namespace", namespace)
			return nil
		} else {
			return err
		}
	}
	switch pod.Status.Phase {
	case v1.PodSucceeded, v1.PodFailed:
		slog.Info("Will cleanup old pod", "phase", pod.Status.Phase)
		err := clientset.CoreV1().Pods(namespace).Delete(getDeadlineCtx, name, metav1.DeleteOptions{})
		if err != nil {
			slog.Error("Failed deleting existing pod")
			return err
		}
		return nil
	default:
		slog.Error("Encountered pod in non-final phase, cleanup would be dangerous. Manual cleanup is required.")
		return fmt.Errorf("cannot cleanup pod in phase %s", pod.Status.Phase)
	}
}

func RunPod(ctx context.Context, filepath string, opts *RunPodOptions) error {

	clientset, err := GetClientSet()
	if err != nil {
		return err
	}

	podSpec, err := GetPodSpec(filepath)
	if err != nil {
		slog.Error("Failed to load podSpec", "filepath", filepath, "error", err)
		return err
	}
	name := podSpec.GetName()
	namespace := podSpec.GetNamespace()

	if opts.ReplaceOldPod {
		err := deleteLingeringPod(ctx, clientset, namespace, name, opts.CreateTimeout)
		if err != nil {
			slog.Error("Error deleting lingering pod", "error", err)
		}
	}

	createDeadlineCtx, cancelCreate := context.WithTimeout(ctx, opts.CreateTimeout)
	defer cancelCreate()
	createPod, err := clientset.CoreV1().Pods(namespace).Create(createDeadlineCtx, podSpec, metav1.CreateOptions{})

	if err != nil {
		slog.Error("Failed to create pod", "pod", createPod, "error", err)
		return err
	}

	runDeadlineCtx, cancelRun := context.WithTimeout(ctx, opts.RunTimeout)
	defer cancelRun()
	returnCode := WatchPodStreamLogsAndCleanup(runDeadlineCtx, clientset, *createPod, opts.CleanupPod)
	if returnCode != 0 {
		return fmt.Errorf("encountered non-zero return code from watching pods %d", returnCode)
	}
	return nil
}
