# Pod Ashiato

Pod Ashiato ("足跡" - footprint in Japanese) is a Kubernetes utility that logs which nodes are running specific pods, providing visibility into pod-to-node mapping over time.

## Features

- Tracks which Kubernetes nodes are running pods
- Outputs structured logs to stdout
- Can run in one-shot mode or continuously with a configurable interval
- Works both in-cluster (as a pod) or out-of-cluster (from your local machine)

## Usage

### Building

```bash
go build -o pod-ashiato cmd/ashiato/main.go
```

### Running

From your local machine:

```bash
# Run once and exit
./pod-ashiato --oneshot

# Run continuously with 1 minute interval
./pod-ashiato --interval 1m
```

### In a Kubernetes Cluster

You can deploy pod-ashiato as a pod in your cluster. Check the `deploy` directory for example manifests.

## Command Line Options

- `--kubeconfig`: Path to the kubeconfig file (default: `~/.kube/config`)
- `--interval`: Interval between pod checks (default: 30s)
- `--oneshot`: Run only once and exit (default: false)

## Output Format

The tool outputs JSON logs to stdout with the following structure:

```json
{
  "namespace": "default",
  "pod_name": "example-pod",
  "node_name": "worker-node-1",
  "pod_ip": "10.1.2.3",
  "phase": "Running",
  "timestamp": "2023-01-01T12:00:00Z",
  "conditions": [
    {
      "type": "Ready",
      "status": "True",
      "last_transition_time": "2023-01-01T11:55:00Z"
    }
  ]
}
```