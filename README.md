# Pod Ashiato

Pod Ashiato ("足跡" - footprint in Japanese) is a Kubernetes utility that logs which nodes are running specific pods, providing visibility into pod-to-node mapping over time.

## Features

- Tracks which Kubernetes nodes are running pods
- Outputs structured logs to stdout
- Can run in one-shot mode or continuously with a configurable interval
- Works both in-cluster (as a pod) or out-of-cluster (from your local machine)
- Automatically published to GitHub Container Registry (ghcr.io)

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

# Filter pods by namespace
./pod-ashiato --namespace kube-system

# Filter pods by name prefix (matches all pods starting with "nginx-")
./pod-ashiato --name nginx-

# Filter pods by name prefix (matches all pods starting with "app-backend-")
./pod-ashiato --name app-backend-

# Filter pods by label
./pod-ashiato --label app=nginx,env=production

# Combine filters
./pod-ashiato --namespace default --label app=backend --interval 2m
```

### In a Kubernetes Cluster

You can deploy pod-ashiato as a pod in your cluster using the pre-built container image from GitHub Container Registry:

```bash
# Update your deployment manifest to use the ghcr.io image
# Example:
kubectl apply -f deploy/manifests/deployment.yaml
```

#### Using ghcr.io Images

The container images are automatically published to GitHub Container Registry when a new release is created:

```bash
# Pull the latest released version
docker pull ghcr.io/takutakahashi/pod-ashiato:latest

# Pull a specific version
docker pull ghcr.io/takutakahashi/pod-ashiato:v1.0.0

# Use in Kubernetes manifest
# image: ghcr.io/takutakahashi/pod-ashiato:v1.0.0
```

You can also check the `deploy` directory for example manifests.

## Command Line Options

- `--kubeconfig`: Path to the kubeconfig file (default: `~/.kube/config`)
- `--interval`: Interval between pod checks (default: 30s)
- `--oneshot`: Run only once and exit (default: false)
- `--namespace`: Filter pods by namespace (default: all namespaces)
- `--name`: Filter pods by name prefix (matches pods starting with the specified string)
- `--label`: Filter pods by label selector (e.g., 'app=nginx,env=prod')

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

## Documentation

- [Architecture](ARCH.md) - Technical architecture and design decisions
- [Release Process](docs/RELEASE.md) - How releases are managed and automated