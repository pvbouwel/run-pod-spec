# run-pod-spec

A binary that helps to run a Kubernetes pod manifest, stream the logs and cleanup the pod at the end.


## Why?

`kubectl run` does not have a usable way to mount secrets/configmaps. This binary allows to use a full Pod spec.

## Usage

### CLI documentation
Just invoke the CLI with -h to get the usage. For example usign the container image:
```sh
podman run -it --rm localhost/run-pod-spec:latest -h
``` 

### Configuration
If you run it on K8s it resolves API credentials using its service account.
If you run outside of K8S you are expected to have set `KUBECONFIG` environment variable.

For example you first run kubie to set a context for a specific image then you can run the container and make sure to pass the environment variable as well as the files.

```sh
POD_SPEC_FILE="/tmp/testpod.yaml"
podman run -it -v $POD_SPEC_FILE:$POD_SPEC_FILE -v $KUBECONFIG:$KUBECONFIG -e KUBECONFIG=$KUBECONFIG --rm localhost/run-pod-spec:latest -f $POD_SPEC_FILE
```