package kube

import (
	"context"
	"log"
	"log/slog"
	"os"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

func WatchPodStreamLogsAndCleanup(ctx context.Context, clientset *kubernetes.Clientset, pod corev1.Pod, cleanupPod bool) int {
	var podName = pod.GetName()
	var podNamespace = pod.GetNamespace()
	var returnCode = -1
	if clientset == nil {
		log.Fatal("Clientset provided to WatchPods cannot be nil.")
	}
	log.Println("Starting Pod watchers for specified namespaces...")

	stopChannel := make(chan struct{})

	cleanup := func() {
		if cleanupPod {
			slog.Info("Deleting pod", "name", podName, "namespace", podNamespace)
			err := deletePod(clientset, podName, podNamespace)
			if err != nil {
				slog.Error("Error deleting pod", "error", err)
				close(stopChannel)
			}
		} else {
			close(stopChannel)
		}
	}

	streamPodLogs := func() {
		err := StreamPodLogsTo(ctx, pod.GetName(), pod.GetNamespace(), os.Stdout, clientset)
		if err != nil {
			slog.Error("Encountered error streaming pod logs", "error", err)
		}
	}

	go func(namespace, name string) {
		slog.Debug("Setting up watcher for Pod", "name", name, "namespace", namespace)

		//See https://jiminbyun.medium.com/getting-started-with-client-go-building-a-kubernetes-pod-watcher-in-go-caa2be8623eb
		//for more information on Informers and how to use them
		factory := informers.NewSharedInformerFactoryWithOptions(
			clientset,
			5*time.Minute,
			informers.WithNamespace(namespace),
			informers.WithTweakListOptions(func(opt *metav1.ListOptions) {
				opt.FieldSelector = fields.OneTermEqualSelector("metadata.name", name).String()
			}),
		)

		informer := factory.Core().V1().Pods().Informer()

		_, err := informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
			DeleteFunc: func(obj interface{}) {
				pod := obj.(*corev1.Pod)
				slog.Info("Pod deleted", "name", pod.GetName(), "namespace", pod.GetNamespace())
				close(stopChannel)
			},
			// You can also add an UpdateFunc here if you want to react to Pod modifications:
			UpdateFunc: func(oldObj, newObj interface{}) {
				oldPod := oldObj.(*corev1.Pod)
				newPod := newObj.(*corev1.Pod)
				if oldPod.ResourceVersion != newPod.ResourceVersion {
					slog.Debug("Pod updated", "name", pod.GetName(), "namespace", pod.GetNamespace(), "status", newPod.Status.String())
				}
				if oldPod.Status.Phase == newPod.Status.Phase {
					slog.Debug("Pod update remained in same phase", "phase", newPod.Status.Phase)
					return //Avoid performing actions too often
				}
				switch newPod.Status.Phase {
				case corev1.PodPending, corev1.PodUnknown:
					return
				case corev1.PodRunning:
					slog.Info("Pod entered Running state.")
					streamPodLogs()
				case corev1.PodSucceeded:
					returnCode = 0
					streamPodLogs()
					cleanup()
				case corev1.PodFailed:
					returnCode = 1

					slog.Error("Pod failed", "podStatus", newPod.Status.Message, "containerStatus", newPod.Status.ContainerStatuses)
					streamPodLogs()
					cleanup()
				}

			},
		})
		if err != nil {
			slog.Error("Cannot add event handler", "error", err)
			return
		}

		go factory.Start(stopChannel)

		// Wait for the informer's caches to be synced. This is important!
		// It ensures the informer has retrieved the initial state of all Pods
		// before it starts processing real-time events. This prevents missing initial events.
		// It will block until caches are synced or stopCh is closed.
		factory.WaitForCacheSync(stopChannel)
		slog.Debug("Cache synced", "name", name, "namespace", namespace)

		select {
		case <-ctx.Done():
			slog.Error("Context has expired so we need to perform our cleanup")
			cleanup()
		case <-stopChannel:
			slog.Debug("Normal stop channel signaled end")
		}
	}(podNamespace, podName)

	// Here we also wait the stopCh to be closed
	<-stopChannel
	return returnCode
}
