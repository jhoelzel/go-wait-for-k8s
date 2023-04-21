# go-wait-for-k8s
go-wait-for-k8s is a utility program written in Go that monitors the readiness of Kubernetes resources like Pods, Jobs, Deployments, StatefulSets, DaemonSets, and ReplicaSets. The program checks the readiness status of the specified resources in a given namespace at a specified interval and exits once all resources are ready or if the optional timeout is reached.

## Usage
go-wait-for-k8s -namespace=<namespace> -label-selector=<label-selector> -resource-type=<resource-type> [flags]

```sh
go-wait-for-k8s -namespace=<namespace> -label-selector=<label-selector> -resource-type=<resource-type> [flags]
```

## Flags
* -namespace: The namespace to monitor. Can also be set using the NAMESPACE environment variable.
* -label-selector: The label selector to filter resources. Can also be set using the LABEL_SELECTOR environment variable.
* -resource-type: The resource type to monitor. Must be one of the following: pod, job, deployment, statefulset, daemonset, or replicaset. Can also be set using the RESOURCE_TYPE environment variable.
* -kubeconfig: Path to the kubeconfig file. Can also be set using the KUBECONFIG environment variable.
* -timeout: The maximum amount of time (in minutes) to wait for resources to become ready; default is infinite (0). Can also be set using the TIMEOUT_SECONDS environment variable.
* -interval: The interval (in seconds) between checks for resource readiness; default is 5 seconds. Can also be set using the INTERVAL_SECONDS environment variable.

## Installation
Make sure you have Go installed.

Clone the repository:

```sh
git clone https://github.com/jhoelzel/go-wait-for-k8s.git
```

Build the binary:

```sh
make build
```

Move the binary to your desired location, for example:

```sh
sudo mv go-wait-for-k8s /usr/local/bin/
```

## Use with init-container

This Deployment creates a single replica of the BusyBox container. The go-wait-for-k8s program runs as an init container and waits for all Deployments with the label app=readiness-test to become ready before starting the BusyBox container.

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: busybox-deployment
spec:
  replicas: 1
  selector:
    matchLabels:
      app: busybox
  template:
    metadata:
      labels:
        app: busybox
    spec:
      initContainers:
        - name: go-wait-for-k8s
          image: ghcr.io/jhoelzel/go-wait-for-k8s:latest
          args:
            - "--label-selector=app=readiness-test"
            - "--resource-type=deployment"
          env:
            - name: NAMESPACE
              valueFrom:
                fieldRef:
                  fieldPath: metadata.namespace
      containers:
        - name: busybox
          image: busybox:latest
          command: ["sh", "-c", "while true; do sleep 3600; done"]

```

## Examples

Wait for all Pods in the default namespace with the label app=myapp:

```sh
go-wait-for-k8s -namespace=default -label-selector="app=myapp" -resource-type=pod
```

Wait for all Deployments in the production namespace with the label tier=frontend to be ready within 10 minutes, and check readiness every 10 seconds:

```sh
go-wait-for-k8s -namespace=production -label-selector="tier=frontend" -resource-type=deployment -timeout=10 -interval=10
```

## Testing

In order to test this program, some manifests have been created in ./kube-manifests/tests

In order to apply success tests apply:

```sh
kubectl -f ./kube-manifests/tests/success
```

In order to apply failure tests apply:

```sh
kubectl -f ./kube-manifests/tests/fail
```

then run the tests with:

```sh
./go-wait-for --namespace=default --label-selector=app=readiness-test --resource-type=deployment
./go-wait-for --namespace=default --label-selector=app=readiness-test --resource-type=job
./go-wait-for --namespace=default --label-selector=app=readiness-test --resource-type=statefulset
./go-wait-for --namespace=default --label-selector=app=readiness-test --resource-type=daemonset
./go-wait-for --namespace=default --label-selector=app=readiness-test --resource-type=replicaset
```


# License
This project is released under the [MIT License]<https://opensource.org/license/mit/>.

