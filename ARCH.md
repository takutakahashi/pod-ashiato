# Architecture

Pod Ashiato is designed with simplicity in mind to track pod-to-node mapping in a Kubernetes cluster.

## Components

### Main Application (`cmd/ashiato/main.go`)

The entry point for the application that:
- Sets up Kubernetes client configuration
- Initializes the Pod controller
- Handles signals for graceful shutdown
- Manages command line flags

### Pod Controller (`pkg/controller/pod_controller.go`)

Core component responsible for:
- Periodically querying the Kubernetes API for pod information
- Converting pod data to structured log format
- Outputting logs to stdout
- Supporting both continuous and one-shot operation modes
- Filtering pods by namespace, name prefix, and labels
- Optionally storing pod-to-node mappings in ConfigMaps

### ConfigMap Store (`pkg/controller/configmap_store.go`)

Component for persisting pod-to-node mapping information:
- Creates hourly ConfigMaps with format `pod-ashiato-YYYYMMDDHH`
- Stores pod-to-node mappings as key-value pairs
- Handles ConfigMap creation and updates
- Manages data format and storage operations

## Flow

1. The application initializes and connects to the Kubernetes API
2. The Pod controller queries for pods based on filter criteria
3. For each pod, it extracts relevant information (namespace, name, node, IP, etc.)
4. This information is formatted as JSON and output to stdout
5. If configured, the information is also stored in ConfigMaps by the ConfigMap Store
6. If in continuous mode, the process repeats after the configured interval

## Design Decisions

- **Structured Logging**: JSON format enables easy integration with log aggregation and analysis tools
- **stdout Output**: Following the Kubernetes and cloud-native best practice of logging to stdout
- **In-cluster/Out-cluster Support**: Can run as a pod in the cluster or from a developer's machine
- **Simple Architecture**: Single-purpose tool with minimal dependencies
- **ConfigMap Storage**: Provides a simple Kubernetes-native way to store historical pod placement data
- **Hourly Partitions**: ConfigMaps are created on an hourly basis to provide a balance between storage granularity and manageability

## Extension Points

Future enhancements could include:
- More sophisticated filtering of pods
- Additional storage backends beyond ConfigMaps
- Metrics reporting on pod movement between nodes
- Alerting on pod reschedules or node changes
- Integration with monitoring systems