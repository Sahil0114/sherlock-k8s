package fetcher

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

// GetPodEvents fetches events strictly tied to the given pod UID.
func GetPodEvents(ctx context.Context, clientset *kubernetes.Clientset, namespace string, podUID types.UID) ([]corev1.Event, error) {
	fieldSelector := fields.OneTermEqualSelector("involvedObject.uid", string(podUID)).String()

	eventList, err := clientset.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{
		FieldSelector: fieldSelector,
	})
	if err != nil {
		return nil, err
	}

	return eventList.Items, nil
}
