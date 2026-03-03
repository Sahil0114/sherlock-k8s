package timeline

import (
	"sort"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
)

// TimelineEntry represents a single event in the diagnostic timeline.
type TimelineEntry struct {
	Time    time.Time
	Kind    string // "log", "event", "system"
	Source  string // Container name or Node name
	Message string
}

// EventsToEntries converts Kubernetes events to timeline entries.
func EventsToEntries(events []corev1.Event, fallbackTime time.Time) []TimelineEntry {
	var entries []TimelineEntry
	for _, e := range events {
		t := resolveEventTime(e, fallbackTime)
		entries = append(entries, TimelineEntry{
			Time:    t,
			Kind:    "event",
			Source:  "SYSTEM",
			Message: e.Message,
		})
	}
	return entries
}

// LogsToEntries converts container log lines to timeline entries.
func LogsToEntries(containerLogs map[string][]string, fallbackTime time.Time) []TimelineEntry {
	var entries []TimelineEntry
	for containerName, lines := range containerLogs {
		for _, line := range lines {
			t, msg := parseLogLine(line, fallbackTime)
			entries = append(entries, TimelineEntry{
				Time:    t,
				Kind:    "log",
				Source:  containerName,
				Message: msg,
			})
		}
	}
	return entries
}

// SortAndCap sorts entries chronologically using stable sort, then caps to maxEntries.
func SortAndCap(entries []TimelineEntry, maxEntries int) []TimelineEntry {
	sort.SliceStable(entries, func(i, j int) bool {
		return entries[i].Time.Before(entries[j].Time)
	})

	if len(entries) > maxEntries {
		// Keep the most recent entries
		entries = entries[len(entries)-maxEntries:]
	}

	return entries
}

// resolveEventTime returns the best available timestamp for a Kubernetes event.
// Priority: EventTime -> LastTimestamp -> FirstTimestamp -> fallback
func resolveEventTime(e corev1.Event, fallback time.Time) time.Time {
	if !e.EventTime.Time.IsZero() {
		return e.EventTime.Time
	}
	if !e.LastTimestamp.Time.IsZero() {
		return e.LastTimestamp.Time
	}
	if !e.FirstTimestamp.Time.IsZero() {
		return e.FirstTimestamp.Time
	}
	if fallback.IsZero() {
		return time.Now()
	}
	return fallback
}

// parseLogLine extracts a timestamp and message from a Kubernetes log line.
// Kubernetes logs with Timestamps: true have format: "2006-01-02T15:04:05.999999999Z message..."
func parseLogLine(line string, fallback time.Time) (time.Time, string) {
	parts := strings.SplitN(line, " ", 2)
	if len(parts) < 2 {
		return safeFallback(fallback), line
	}

	t, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil {
		return safeFallback(fallback), line
	}

	return t, parts[1]
}

// safeFallback returns the fallback time, or time.Now() if fallback is zero.
func safeFallback(fallback time.Time) time.Time {
	if fallback.IsZero() {
		return time.Now()
	}
	return fallback
}
