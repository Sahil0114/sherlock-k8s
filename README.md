# 🕵️ Kube-Sherlock

**Instant Root Cause Analysis Context for Kubernetes.**

Kube-Sherlock replaces the manual, sequential `kubectl` debugging loop. It fires concurrent requests to the Kubernetes API to aggregate Node conditions, multi-container logs, and events into a single, chronologically sorted timeline.

If your pod crashed, Kube-Sherlock gives you the exact sequence of events that led to the failure, in under a second.

## The Problem
Debugging modern Kubernetes workloads requires checking the pod logs, the sidecar logs, the node conditions, and the events. Doing this manually takes minutes and requires mentally correlating timestamps. Kube-Sherlock does it concurrently in milliseconds.

## Usage

Investigate a crashing pod:
```bash
kubectl sherlock investigate pod/api-gateway-7b8c9d
```

Specify a namespace:
```bash
kubectl sherlock investigate api-gateway-7b8c9d -n production
```

### Example Output
```text
🔍 INVESTIGATING: api-gateway-7b8c9d (Namespace: default)

🧱 INFRASTRUCTURE CHECK:
[Node: ip-10-0-2-4] Status: Ready | Pressure: None

⏱️ TIMELINE OF EVENTS:
10:00:01 [SYSTEM]      Pod Scheduled on node ip-10-0-2-4
10:00:02 [init-vault]  Exit 0 (Completed)
10:00:05 [main-app]    Connecting to database...
10:00:15 [main-app]    WARN: Connection timeout.
10:00:16 [SYSTEM]      Liveness probe failed: HTTP GET /healthz 500
10:00:18 [main-app]    Container exited with code 137 (OOMKilled)
```

## Features

* **Concurrent Fetching:** Bypasses manual API latency.
* **Node Reality Checks:** Surfaces hardware health before you look at application logs.
* **Sidecar Native:** Explicitly tags logs and events by their origin container.
* **Graceful Degradation:** Skips protected resources automatically if your RBAC is restricted.
* **Minimal Footprint:** Output is optimized for `less` and grepping; no excessive ANSI formatting.

## Installation

**Via Krew (Recommended)**

```bash
kubectl krew install sherlock
```
