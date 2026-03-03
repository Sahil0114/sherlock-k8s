package cmd

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/kube-sherlock/pkg/fetcher"
	"github.com/kube-sherlock/pkg/formatter"
	"github.com/kube-sherlock/pkg/timeline"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
)

func init() {
	rootCmd.AddCommand(investigateCmd)
}

var investigateCmd = &cobra.Command{
	Use:   "investigate <pod-name>",
	Short: "Investigate a pod and produce a diagnostic timeline",
	Args:  cobra.ExactArgs(1),
	Run:   runInvestigate,
}

func runInvestigate(cmd *cobra.Command, args []string) {
	podName := args[0]
	// Strip "pod/" prefix if provided
	podName = strings.TrimPrefix(podName, "pod/")

	clientset, ns, err := buildClient()
	if err != nil {
		exitWithError(err.Error())
	}

	ctx := context.Background()

	// Phase 1: Synchronous Pod fetch
	pod, err := fetcher.GetPod(ctx, clientset, ns, podName)
	if err != nil {
		exitWithError(fmt.Sprintf("failed to fetch pod %s/%s: %v", ns, podName, err))
	}

	podStartTime := time.Now()
	if pod.Status.StartTime != nil {
		podStartTime = pod.Status.StartTime.Time
	}

	// Phase 2: Concurrent fetch
	eventsCh := make(chan []corev1.Event, 1)
	nodeCh := make(chan *corev1.Node, 1)
	logsCh := make(chan map[string][]string, 1)
	rbacErrCh := make(chan string, 1)

	var wg sync.WaitGroup

	// Node Worker
	wg.Add(1)
	go func() {
		defer wg.Done()
		nodeName := pod.Spec.NodeName
		if nodeName == "" {
			nodeCh <- nil
			rbacErrCh <- ""
			return
		}
		node, fetchErr := fetcher.GetNode(ctx, clientset, nodeName)
		if fetchErr != nil {
			nodeCh <- nil
			if fetcher.IsForbidden(fetchErr) {
				rbacErrCh <- "[SYSTEM] Node status unavailable (insufficient permissions)"
			} else {
				rbacErrCh <- fmt.Sprintf("[SYSTEM] Node fetch error: %v", fetchErr)
			}
			return
		}
		nodeCh <- node
		rbacErrCh <- ""
	}()

	// Events Worker
	wg.Add(1)
	go func() {
		defer wg.Done()
		events, fetchErr := fetcher.GetPodEvents(ctx, clientset, ns, pod.UID)
		if fetchErr != nil {
			eventsCh <- nil
			return
		}
		eventsCh <- events
	}()

	// Logs Worker (Fan-Out)
	wg.Add(1)
	go func() {
		defer wg.Done()
		logs := fetcher.GetContainerLogs(ctx, clientset, ns, podName, pod)
		logsCh <- logs
	}()

	wg.Wait()

	// Collect results
	events := <-eventsCh
	node := <-nodeCh
	containerLogs := <-logsCh
	rbacMsg := <-rbacErrCh

	// Build timeline
	var entries []timeline.TimelineEntry

	// Inject synthetic entries
	if pod.Status.Phase == corev1.PodPending {
		entries = append(entries, timeline.TimelineEntry{
			Time:    podStartTime,
			Kind:    "system",
			Source:  "SYSTEM",
			Message: "Pod is Pending (not yet scheduled)",
		})
	}

	if rbacMsg != "" {
		entries = append(entries, timeline.TimelineEntry{
			Time:    podStartTime,
			Kind:    "system",
			Source:  "SYSTEM",
			Message: strings.TrimPrefix(rbacMsg, "[SYSTEM] "),
		})
	}

	// Convert events to timeline entries
	eventEntries := timeline.EventsToEntries(events, podStartTime)
	entries = append(entries, eventEntries...)

	// Convert logs to timeline entries
	logEntries := timeline.LogsToEntries(containerLogs, podStartTime)
	entries = append(entries, logEntries...)

	// Sort and cap
	entries = timeline.SortAndCap(entries, 100)

	// Render
	formatter.Render(pod, node, ns, entries)
}
