package formatter

import (
	"fmt"
	"strings"

	"github.com/kube-sherlock/pkg/timeline"
	corev1 "k8s.io/api/core/v1"
)

const (
	colorReset  = "\033[0m"
	colorCyan   = "\033[36m"
	colorYellow = "\033[33m"
)

// Render outputs the investigation results to stdout.
func Render(pod *corev1.Pod, node *corev1.Node, namespace string, entries []timeline.TimelineEntry) {
	fmt.Printf("INVESTIGATING: %s (Namespace: %s)\n\n", pod.Name, namespace)

	// Infrastructure check
	renderNodeStatus(node, pod.Spec.NodeName)

	// Timeline
	fmt.Println("TIMELINE OF EVENTS:")
	for _, e := range entries {
		renderEntry(e)
	}
}

func renderNodeStatus(node *corev1.Node, nodeName string) {
	fmt.Println("INFRASTRUCTURE CHECK:")
	if nodeName == "" {
		fmt.Printf("  %s[SYSTEM]%s No node assigned (pod may be Pending)\n\n", colorYellow, colorReset)
		return
	}

	if node == nil {
		fmt.Printf("  %s[%s]%s Status unavailable\n\n", colorYellow, nodeName, colorReset)
		return
	}

	var conditions []string
	ready := "Unknown"
	var pressures []string

	for _, c := range node.Status.Conditions {
		switch c.Type {
		case corev1.NodeReady:
			if c.Status == corev1.ConditionTrue {
				ready = "Ready"
			} else {
				ready = "NotReady"
			}
		case corev1.NodeMemoryPressure:
			if c.Status == corev1.ConditionTrue {
				pressures = append(pressures, "MemoryPressure")
			}
		case corev1.NodeDiskPressure:
			if c.Status == corev1.ConditionTrue {
				pressures = append(pressures, "DiskPressure")
			}
		case corev1.NodePIDPressure:
			if c.Status == corev1.ConditionTrue {
				pressures = append(pressures, "PIDPressure")
			}
		}
	}

	pressureStr := "None"
	if len(pressures) > 0 {
		pressureStr = strings.Join(pressures, ", ")
	}

	conditions = append(conditions, fmt.Sprintf("Status: %s", ready))
	conditions = append(conditions, fmt.Sprintf("Pressure: %s", pressureStr))

	fmt.Printf("  [Node: %s] %s\n\n", nodeName, strings.Join(conditions, " | "))
}

func renderEntry(e timeline.TimelineEntry) {
	ts := e.Time.Format("15:04:05")
	source := formatSource(e)
	fmt.Printf("%s %s %s\n", ts, source, e.Message)
}

func formatSource(e timeline.TimelineEntry) string {
	padded := fmt.Sprintf("[%-12s]", e.Source)
	switch e.Kind {
	case "system", "event":
		return fmt.Sprintf("%s%s%s", colorYellow, padded, colorReset)
	case "log":
		return fmt.Sprintf("%s%s%s", colorCyan, padded, colorReset)
	default:
		return padded
	}
}
