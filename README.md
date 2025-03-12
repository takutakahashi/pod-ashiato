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

# Store pod-to-node mapping in ConfigMaps (created in default namespace)
./pod-ashiato --store-in-cm

# Store mappings in a specific namespace
./pod-ashiato --store-in-cm --cm-namespace monitoring
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
- `--store-in-cm`: Store pod-to-node mapping in ConfigMaps (default: false)
- `--cm-namespace`: Namespace where ConfigMaps will be created (default: `default`)

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

## ConfigMap Storage

When the `--store-in-cm` flag is enabled, Pod Ashiato will store pod-to-node mapping information in ConfigMaps:

- ConfigMaps are created with the name pattern `pod-ashiato-YYYYMMDDHH` (where YYYY=year, MM=month, DD=day, HH=hour)
- A new ConfigMap is created every hour
- Each ConfigMap contains data in the format `namespace_podname: nodename` (using underscore as separator for Kubernetes compatibility)
- ConfigMaps are labeled with `app: pod-ashiato` and `type: pod-node-mapping`
- By default, ConfigMaps are created in the `default` namespace, but this can be changed with `--cm-namespace`

You can retrieve the stored mapping information with:

```bash
kubectl get cm -l app=pod-ashiato
kubectl get cm pod-ashiato-2023010112 -o yaml
```

This feature is useful for:

- Historical tracking of pod placement
- Audit and compliance purposes
- Integration with other tools that can consume ConfigMap data