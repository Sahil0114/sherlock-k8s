package fetcher

import (
	"context"
	"net/http"

	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// GetNode fetches the node by name.
func GetNode(ctx context.Context, clientset *kubernetes.Clientset, nodeName string) (*corev1.Node, error) {
	node, err := clientset.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return node, nil
}

// IsForbidden returns true if the error is a 403 Forbidden API error.
func IsForbidden(err error) bool {
	if statusErr, ok := err.(*k8serrors.StatusError); ok {
		return statusErr.ErrStatus.Code == http.StatusForbidden
	}
	return false
}
