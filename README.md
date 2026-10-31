![Horus logo](./docs/horus_logo.png)

# Horus

**A Kubernetes-native ML workload control plane — under active development.**

Horus is a Go project for discovering ML workloads and checking their status through workload-specific plugins. The current prototype discovers Kubernetes Job targets from local JSON files. Kubernetes and AWS configuration types, plus the plugin and registry foundations, are present, but cluster discovery and live workload inspection are not implemented yet.

## Current status

| Capability | Status |
| --- | --- |
| Load and validate YAML application configuration | Implemented |
| Discover targets from a local directory of JSON files | Implemented |
| Kubernetes Job target format and validation | Implemented |
| Kubernetes cluster and AWS discovery | Not yet wired into the executable |
| Read live Kubernetes Job status | Not yet implemented |
| CLI output and reporting | Prototype output |

## Requirements

- Go **1.26.4** or newer

## Get started

Run the test suite:

```sh
go test ./...
```

To try local discovery, create a temporary config file whose `path` points to the repository's example target directory. The path must be absolute:

```sh
cat > /tmp/horus.yaml <<EOF
discovery:
  local:
    - name: local-dev
      path: "$(pwd)/examples/targets"
EOF

APP_CONFIG_PATH=/tmp/horus.yaml go run ./cmd/monitor
```

`APP_CONFIG_PATH` is optional; when omitted, Horus reads `configs/application.yaml`. The monitor currently constructs local discoverers from the configuration and prints the discovered result as development output.
The checked-in config points at `/examples/targets`; override that path if this directory does not exist on your machine.

## Configuration

Configuration is YAML. At least one discovery source must be declared. The local source path must be absolute:

```yaml
discovery:
  local:
    - name: local-dev
      path: /absolute/path/to/targets
```

Configuration types for Kubernetes and AWS sources also exist. However, the current monitor does not create discoverers for those source types.

## Local target format

Each `.json` file in the configured directory describes one target. The current supported type is `kubernetes-job`:

```json
{
  "id": "job-001",
  "name": "train-job-001",
  "type": "kubernetes-job",
  "spec": {
    "cluster": "local-dev",
    "namespace": "ml-workloads",
    "jobName": "training-demo"
  }
}
```

Target files are decoded strictly: unknown fields, malformed JSON, unsupported target types, and invalid required fields are reported as discovery failures. The included sample is [`examples/targets/kubernetes-job.json`](./examples/targets/kubernetes-job.json).

## How it fits together

1. The monitor loads and validates the YAML configuration.
2. It creates a discoverer for each configured local source.
3. The discovery manager combines the discovered targets and per-file failures.
4. A plugin can check a target and return its workload status. The Kubernetes Job plugin currently returns `unknown` because live inspection is not implemented.

## Project layout

```text
cmd/monitor/          Application entry point
configs/              Example application configuration
docs/                 Project and architecture images
examples/targets/     Sample target definitions
internal/config/      Configuration loading and validation
internal/discovery/   Discovery interfaces, manager, and local discoverer
internal/health/       Workload status types
internal/plugin/      Workload plugin contract and implementations
internal/registry/    Plugin registry
internal/target/      Target types and specifications
```

## Development

Run all Go tests with:

```sh
go test ./...
```

The packages under `internal/` are implementation details and are not intended as stable public APIs yet.
