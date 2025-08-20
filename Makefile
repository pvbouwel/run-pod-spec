BENCH_COUNT ?= 10


build-container:
	podman build -t run-pod-spec -f Dockerfile .

precommit:
	golangci-lint run
	go vet
	go fmt run-pod-spec/pkg/kube run-pod-spec/pkg/logging .
