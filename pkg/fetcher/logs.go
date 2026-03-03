package fetcher

import (
	"bufio"
	"context"
	"io"
	"sync"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

// GetContainerLogs fetches logs for all containers (init + app) using an inner fan-out pattern.
// Returns a map of container name to log lines.
func GetContainerLogs(ctx context.Context, clientset *kubernetes.Clientset, namespace, podName string, pod *corev1.Pod) map[string][]string {
	result := make(map[string][]string)
	var mu sync.Mutex
	var innerWg sync.WaitGroup

	// Collect all container names (init + app)
	type containerInfo struct {
		name    string
		isWaiting bool
	}

	var containers []containerInfo

	for _, c := range pod.Spec.InitContainers {
		waiting := isContainerWaiting(pod, c.Name)
		containers = append(containers, containerInfo{name: c.Name, isWaiting: waiting})
	}
	for _, c := range pod.Spec.Containers {
		waiting := isContainerWaiting(pod, c.Name)
		containers = append(containers, containerInfo{name: c.Name, isWaiting: waiting})
	}

	for _, ci := range containers {
		innerWg.Add(1)
		go func(cInfo containerInfo) {
			defer innerWg.Done()

			// Skip waiting containers silently
			if cInfo.isWaiting {
				return
			}

			lines := fetchLogs(ctx, clientset, namespace, podName, cInfo.name, false)

			// If current logs fail or are empty, try previous
			if len(lines) == 0 {
				lines = fetchLogs(ctx, clientset, namespace, podName, cInfo.name, true)
			}

			if len(lines) > 0 {
				mu.Lock()
				result[cInfo.name] = lines
				mu.Unlock()
			}
		}(ci)
	}

	innerWg.Wait()
	return result
}

func fetchLogs(ctx context.Context, clientset *kubernetes.Clientset, namespace, podName, containerName string, previous bool) []string {
	var tailLines int64 = 200

	opts := &corev1.PodLogOptions{
		Container:  containerName,
		TailLines:  &tailLines,
		Timestamps: true,
		Previous:   previous,
	}

	req := clientset.CoreV1().Pods(namespace).GetLogs(podName, opts)
	stream, err := req.Stream(ctx)
	if err != nil {
		return nil
	}
	defer stream.Close()

	var lines []string
	scanner := bufio.NewScanner(stream)
	// Increase buffer size for potentially long log lines
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	// Silently discard scanner errors
	_, _ = io.Copy(io.Discard, stream)

	return lines
}

func isContainerWaiting(pod *corev1.Pod, containerName string) bool {
	for _, cs := range pod.Status.InitContainerStatuses {
		if cs.Name == containerName && cs.State.Waiting != nil {
			return true
		}
	}
	for _, cs := range pod.Status.ContainerStatuses {
		if cs.Name == containerName && cs.State.Waiting != nil {
			return true
		}
	}
	return false
}
