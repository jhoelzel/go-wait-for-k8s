package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	// command-line flags and their default values
	var (
		namespace          = flag.String("namespace", "", "The namespace to monitor")
		labelSelector      = flag.String("label-selector", "", "The label selector to filter resources")
		resourceType       = flag.String("resource-type", "", "The resource type to monitor: 'pod', 'job', 'deployment', 'statefulset', 'daemonset', or 'replicaset'")
		kubeconfigPath     = flag.String("kubeconfig", "", "Path to the kubeconfig file")
		timeout            = flag.Int("timeout", 0, "The maximum amount of time to wait for resources to become ready, default is infinite")
		interval           = flag.Int("interval", 5, "The interval between checks for resource readiness, default is 5 seconds")
		validResourceTypes = map[string]bool{
			"pod":         true,
			"job":         true,
			"deployment":  true,
			"statefulset": true,
			"daemonset":   true,
			"replicaset":  true,
		}
	)
	flag.Parse()

	// Update values of flags from environment variables if they are not provided as command-line arguments
	if *namespace == "" {
		if ns := os.Getenv("NAMESPACE"); ns != "" {
			*namespace = ns
		}
	}

	if *labelSelector == "" {
		if ls := os.Getenv("LABEL_SELECTOR"); ls != "" {
			*labelSelector = ls
		}
	}

	if *resourceType == "" {
		if rt := os.Getenv("RESOURCE_TYPE"); rt != "" {
			*resourceType = rt
		}
	}

	if *kubeconfigPath == "" {
		if kc := os.Getenv("KUBECONFIG"); kc != "" {
			*kubeconfigPath = kc
		}
	}
	if *timeout == 0 {
		if tOut := os.Getenv("TIMEOUT_SECONDS"); tOut != "" {
			i, err := strconv.Atoi(tOut)
			if err != nil {
				log.Fatalf("expected an integer value for timeout but got: %s, with error: %v", tOut, err)
			}
			*timeout = i
		}
	}
	if *interval == 5 {
		if ival := os.Getenv("INTERVAL_SECONDS"); ival != "" {
			i, err := strconv.Atoi(ival)
			if err != nil {
				log.Fatalf("expected an integer value for timeout but got: %s, with error: %v", ival, err)
			}
			*interval = i
		}
	}

	// Validate the provided resource type
	if !validResourceTypes[*resourceType] {
		log.Fatalf("Invalid resource type: %s. Supported resource types are: 'pod', 'job', 'deployment', 'statefulset', 'daemonset', and 'replicaset'.", *resourceType)
	}

	//kubeconfig := os.Getenv("KUBECONFIG")
	*kubeconfigPath = "/home/vscode/.kube/config"
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfigPath)
	if err != nil {
		config, err = rest.InClusterConfig()
		if err != nil {
			log.Fatalf("Failed to load Kubernetes config: %v", err)
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Failed to create Kubernetes client: %v", err)
	}

	ctx := context.Background()
	// Create a new context with the provided timeout if provided
	if *timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(*timeout)*time.Minute)
		defer cancel()
	}
	// Continuously check for resource readiness until the timeout is reached
	for {
		select {
		case <-ctx.Done():
			log.Fatalf("Timeout reached while waiting for resources to become ready.")
		default:
			ready, err := checkResourceReadiness(ctx, clientset, *namespace, *labelSelector, *resourceType)
			if err != nil {
				log.Fatalf("Error checking resource readiness: %v", err)
			} else if ready {
				return // All resources are ready!
			}
			time.Sleep(time.Duration(*interval) * time.Second)
		}
	}
}

// checkResourceReadiness checks the readiness of the specified resources
func checkResourceReadiness(ctx context.Context, clientset *kubernetes.Clientset, namespace, labelSelector, resourceType string) (bool, error) {
	var (
		listObject runtime.Object
		err        error
	)

	switch resourceType {
	case "pod":
		listObject, err = clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
			LabelSelector: labelSelector,
		})
	case "job":
		listObject, err = clientset.BatchV1().Jobs(namespace).List(ctx, metav1.ListOptions{
			LabelSelector: labelSelector,
		})
	case "deployment":
		listObject, err = clientset.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{
			LabelSelector: labelSelector,
		})
	case "statefulset":
		listObject, err = clientset.AppsV1().StatefulSets(namespace).List(ctx, metav1.ListOptions{
			LabelSelector: labelSelector,
		})
	case "daemonset":
		listObject, err = clientset.AppsV1().DaemonSets(namespace).List(ctx, metav1.ListOptions{
			LabelSelector: labelSelector,
		})
	case "replicaset":
		listObject, err = clientset.AppsV1().ReplicaSets(namespace).List(ctx, metav1.ListOptions{
			LabelSelector: labelSelector,
		})
	}

	if err != nil {
		return false, fmt.Errorf("failed to list %ss: %v", resourceType, err)
	}

	items, err := meta.ExtractList(listObject)
	if err != nil {
		return false, fmt.Errorf("failed to extract list of %ss: %v", resourceType, err)
	}

	// Check if there are no resources found with the provided label selector
	if len(items) == 0 {
		log.Printf("No %ss found with label selector '%s', waiting...\n", resourceType, labelSelector)
		return false, nil
	}
	// Iterate through the resources and check their readiness
	allReady := true
	accessor := meta.NewAccessor()
	for _, item := range items {
		name, _ := accessor.Name(item)
		ready, err := isResourceReady(item)
		if err != nil {
			return false, fmt.Errorf("error checking readiness for %s %s: %v", resourceType, name, err)
		}
		if !ready {
			allReady = false
			fmt.Printf("%s %s is not ready, waiting...\n", resourceType, name)
		} else {
			fmt.Printf("%s %s is ready.\n", resourceType, name)
		}
	}

	if allReady {
		fmt.Printf("All %ss are ready!\n", resourceType)
		return true, nil
	}
	return false, nil
}

// isResourceReady checks if the given resource object is ready
func isResourceReady(obj runtime.Object) (bool, error) {
	switch resource := obj.(type) {
	case *corev1.Pod:
		for _, condition := range resource.Status.Conditions {
			if condition.Type == corev1.PodReady && condition.Status == corev1.ConditionTrue {
				return true, nil
			}
		}
		return false, nil
	case *batchv1.Job:
		return resource.Status.Succeeded > 0, nil
	case *appsv1.Deployment:
		return resource.Status.UpdatedReplicas == *resource.Spec.Replicas &&
			resource.Status.AvailableReplicas == *resource.Spec.Replicas, nil
	case *appsv1.StatefulSet:
		return resource.Status.ReadyReplicas == *resource.Spec.Replicas, nil
	case *appsv1.DaemonSet:
		return resource.Status.DesiredNumberScheduled == resource.Status.NumberReady, nil
	case *appsv1.ReplicaSet:
		return resource.Status.ReadyReplicas == *resource.Spec.Replicas, nil
	default:
		return false, fmt.Errorf("unsupported resource type: %T", obj)
	}
}
