package kube

import (
	"os"
	"path"
	"testing"
)

var example_pod = `
apiVersion: v1
kind: Pod
metadata:
  name: original-name
  namespace: my-space
  annotations:
    www.example.org/URL: https://allinthemiddle.com/justatest?a=dfaf.12
spec:
  containers:
    - command:
        - /bin/bash
        - -c
        - /etc/scripts/copy_all_images.sh
      image: quay.io/skopeo/stable:v1.18.0
      imagePullPolicy: IfNotPresent
      name: copy-images
      volumeMounts:
        - mountPath: /etc/auths
          mountPropagation: None
          name: auths
          readOnly: true
        - mountPath: /etc/scripts
          mountPropagation: None
          name: images-script
          readOnly: true
  tolerations:
    - effect: NoExecute
      key: node.kubernetes.io/not-ready
      operator: Exists
      tolerationSeconds: 300
    - effect: NoExecute
      key: node.kubernetes.io/unreachable
      operator: Exists
      tolerationSeconds: 300
  volumes:
    - name: auths
      secret:
        defaultMode: 420
        optional: false
        secretName: auth-files
    - configMap:
        defaultMode: 365
        name: images-scripts
        optional: false
      name: images-script
`

func TestGetPodSpecFile(t *testing.T) {
	var tempdir = t.TempDir()
	testFile := path.Join(tempdir, "test.yaml")
	err := os.WriteFile(testFile, []byte(example_pod), 0644)
	if err != nil {
		t.Errorf("Error creating testfile: %s: %s", testFile, err)
		t.FailNow()
	}
	p, err := GetPodSpec(testFile)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if p.Name != "original-name" {
		t.Errorf("Name: Expected: %s, got: %s", "original-name", p.Name)
	}
	if p.Namespace != "my-space" {
		t.Errorf("Namespace: Expected: %s, got: %s", "my-space", p.Namespace)
	}
	if p.Kind != "Pod" {
		t.Errorf("Kind: Expected: %s, got: %s", "Pod", p.Kind)
	}
}
