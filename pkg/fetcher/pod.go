package fetcher

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// GetPod fetches a pod by name from the given namespace.
// Returns an error if the pod cannot be found.
func GetPod(ctx context.Context, clientset *kubernetes.Clientset, namespace, name string) (*corev1.Pod, error) {
	pod, err := clientset.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("pod %s not found in namespace %s: %w", name, namespace, err)
	}
	return pod, nil
}
