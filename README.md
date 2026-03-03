 Kube-Sherlock

Instant Root Cause Analysis Context for Kubernetes

Kube-Sherlock replaces the manual, sequential `kubectl` debugging loop. It fires concurrent requests to the Kubernetes API to aggregate Node conditions, multi-container logs, and events into a single, chronologically sorted timeline.

If your pod crashed, Kube-Sherlock gives you the exact sequence of events that led to the failure, in under a second.

The Problem
Debugging modern Kubernetes workloads requires checking the pod logs, the sidecar logs, the node conditions, and the events. Doing this manually takes minutes and requires mentally correlating timestamps. Kube-Sherlock does it concurrently in milliseconds.

Usage

Investigate a crashing pod:
```bash
kubectl sherlock investigate pod/api-gateway-7b8c9d
```


Kube-Sherlock

Kube-Sherlock is a high-speed CLI tool for Kubernetes root cause analysis. It aggregates node status, pod events, and container logs into a single, chronologically sorted timeline—making it easy to diagnose pod failures and infrastructure issues in seconds.

What does it do?

Kube-Sherlock replaces the manual, sequential `kubectl` debugging loop. Instead of running multiple commands and correlating timestamps by hand, it concurrently fetches all relevant signals and presents them in a unified, readable format.

How does it work?

- Phase 1: Synchronously fetches the target pod and its metadata (node, UID, containers).
- Phase 2: Spawns three concurrent workers:
	- Node Worker: Fetches node conditions (skips if pod is pending; handles RBAC errors gracefully).
	- Events Worker: Fetches only events tied to the pod (using a strict field selector).
	- Logs Worker: Fan-out fetch for all containers (init and app), with fallback to previous logs if current logs are empty or unavailable.

All results are sent through dedicated buffered channels—no shared mutable state. Timeline entries are normalized and sorted stably, with synthetic entries injected for pending pods or RBAC errors.

Features

- Concurrent Fetching: Eliminates manual API latency.
- Node Reality Checks: Surfaces node health before you look at application logs.
- Sidecar Native: Tags logs and events by their origin container.
- Graceful Degradation: Skips protected resources automatically if your RBAC is restricted.
- Minimal Footprint: Output is optimized for piping to `less` or grep; minimal ANSI formatting.

Usage

Investigate a crashing pod:
```bash
kubectl sherlock investigate pod/api-gateway-7b8c9d
```

Specify a namespace:
```bash
kubectl sherlock investigate api-gateway-7b8c9d -n production
```

Example Output

```
INVESTIGATING: api-gateway-7b8c9d (Namespace: default)

INFRASTRUCTURE CHECK:
[Node: ip-10-0-2-4] Status: Ready | Pressure: None

TIMELINE OF EVENTS:
10:00:01 [SYSTEM]      Pod Scheduled on node ip-10-0-2-4
10:00:02 [init-vault]  Exit 0 (Completed)
10:00:05 [main-app]    Connecting to database...
10:00:15 [main-app]    WARN: Connection timeout.
10:00:16 [SYSTEM]      Liveness probe failed: HTTP GET /healthz 500
10:00:18 [main-app]    Container exited with code 137 (OOMKilled)
```

CLI Flags

- `--namespace`, `-n`: Kubernetes namespace (defaults to current context or 'default')
- `--kubeconfig`: Path to kubeconfig file (defaults to `$HOME/.kube/config`)

Design Constraints

- No heuristics, config diffing, or AI integrations.
- Defensive API usage: handles missing environments, empty logs, throttling, and RBAC errors.
- All concurrency is controlled and dependency-aware.
- Timeline entries are always sorted stably; zero-value timestamps are never used.
- Output is readable and minimal, with color only for source prefixes.

Installation

Via Krew (Recommended)
```bash
kubectl krew install sherlock
```

Building from Source

Requires Go 1.21+ and access to a Kubernetes cluster.

```bash
git clone https://github.com/Sahil0114/sherlock-k8s.git
cd sherlock-k8s
go build -o kube-sherlock.exe .
```

License

MIT
