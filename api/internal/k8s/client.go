package k8s

import (
	"context"
	"fmt"
	"log"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var monitorGVR = schema.GroupVersionResource{
	Group:    "monitoring.statuspage.io",
	Version:  "v1alpha1",
	Resource: "monitors",
}

type K8sClient struct {
	kubeClient    *kubernetes.Clientset
	dynamicClient dynamic.Interface
}

func NewK8sClient(kubeconfig string) (*K8sClient, error) {
	var cfg *rest.Config
	var err error

	if kubeconfig != "" {
		cfg, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	} else {
		// Try in-cluster first, fallback to default kubeconfig path
		cfg, err = rest.InClusterConfig()
		if err != nil {
			rules := clientcmd.NewDefaultClientConfigLoadingRules()
			cfg, err = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, &clientcmd.ConfigOverrides{}).ClientConfig()
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to build kubeconfig: %w", err)
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes clientset: %w", err)
	}

	dynClient, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %w", err)
	}

	log.Println("Kubernetes client successfully initialized.")
	return &K8sClient{
		kubeClient:    kubeClient,
		dynamicClient: dynClient,
	}, nil
}

// EnsureTenantNamespace creates namespace `tenant-<tenant_id>` if it doesn't exist
func (k *K8sClient) EnsureTenantNamespace(ctx context.Context, namespace string) error {
	_, err := k.kubeClient.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if err == nil {
		return nil // Namespace already exists
	}

	if errors.IsNotFound(err) {
		ns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: namespace,
				Labels: map[string]string{
					"app.kubernetes.io/managed-by": "statuspage-api",
					"statuspage.io/tenant":         namespace,
				},
			},
		}
		_, err = k.kubeClient.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
		if err != nil && !errors.IsAlreadyExists(err) {
			return fmt.Errorf("failed to create namespace %s: %w", namespace, err)
		}
		log.Printf("Created Kubernetes namespace %s for tenant", namespace)
		return nil
	}

	return err
}

// CreateMonitorCR creates a Monitor CR in the specified namespace
func (k *K8sClient) CreateMonitorCR(ctx context.Context, namespace, name, targetURL string, interval, expectedStatus, timeout int32) error {
	if err := k.EnsureTenantNamespace(ctx, namespace); err != nil {
		return err
	}

	monitorCR := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "monitoring.statuspage.io/v1alpha1",
			"kind":       "Monitor",
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": namespace,
				"labels": map[string]interface{}{
					"app.kubernetes.io/name": "statuspage-operator",
				},
			},
			"spec": map[string]interface{}{
				"url":             targetURL,
				"intervalSeconds": interval,
				"expectedStatus":  expectedStatus,
				"timeoutSeconds":  timeout,
			},
		},
	}

	_, err := k.dynamicClient.Resource(monitorGVR).Namespace(namespace).Create(ctx, monitorCR, metav1.CreateOptions{})
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create Monitor CR %s/%s: %w", namespace, name, err)
	}

	log.Printf("Successfully created Monitor CR %s/%s", namespace, name)
	return nil
}

// DeleteMonitorCR deletes a Monitor CR from the specified namespace
func (k *K8sClient) DeleteMonitorCR(ctx context.Context, namespace, name string) error {
	err := k.dynamicClient.Resource(monitorGVR).Namespace(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to delete Monitor CR %s/%s: %w", namespace, name, err)
	}
	log.Printf("Successfully deleted Monitor CR %s/%s", namespace, name)
	return nil
}
