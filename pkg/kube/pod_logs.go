package kube

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"sync"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type podLogFetcher struct {
	mu sync.Mutex

	logFetchers map[string]*rest.Request
}

var PodLogFetcher = &podLogFetcher{
	mu:          sync.Mutex{},
	logFetchers: map[string]*rest.Request{},
}

// Get an identifier for a pod based on name and namespace
func podId(name, namespace string) string {
	return fmt.Sprintf("%s@%s", name, namespace)
}

func (plf *podLogFetcher) streamPodLogsTo(ctx context.Context, name, namespace string, o io.Writer, c *kubernetes.Clientset) error {
	var tailLines int64 = 100000

	podLogOpts := corev1.PodLogOptions{
		TailLines: &tailLines,
		Follow:    true,
	}
	podId := podId(name, namespace)
	PodLogFetcher.mu.Lock()
	defer PodLogFetcher.mu.Unlock()

	_, ok := PodLogFetcher.logFetchers[podId]
	if ok {
		slog.Debug("Already fetching logs", "podId", podId)
		return nil
	}

	req := c.CoreV1().Pods(namespace).GetLogs(name, &podLogOpts)
	plf.logFetchers[podId] = req
	readCloser, err := req.Stream(ctx)
	if err != nil {
		return err
	}

	go func() {
		_, err = io.Copy(o, readCloser)
		if err != nil {
			slog.Error("Encountered error while reading pod logs", "error", err)
		}
	}()
	return nil
}

func StreamPodLogsTo(ctx context.Context, name, namespace string, w io.Writer, clientset *kubernetes.Clientset) error {
	return PodLogFetcher.streamPodLogsTo(ctx, name, namespace, w, clientset)
}
