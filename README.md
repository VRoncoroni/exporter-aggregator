# Exporter Aggregator

A lightweight Go-based HTTP server that aggregates Prometheus metrics from multiple exporter targets into a single `/metrics` endpoint. It also provides a `/health` endpoint for monitoring the status of all configured targets.

## Features

- **Metrics Aggregation**: Collects and combines metrics from multiple Prometheus exporters
- **Health Monitoring**: Provides health status for all configured targets
- **Configurable Targets**: Easily configure multiple targets via JSON configuration
- **Timeout Handling**: Configurable request timeouts for target fetching
- **Concurrent Fetching**: Fetches metrics from all targets concurrently for better performance

## Installation

### Prerequisites

- Go 1.16 or later

### Build

Clone the repository and build the binary:

```bash
git clone <repository-url>
cd exporter-aggregator
$env:GOOS="linux"
$env:GOARCH="amd64"
go build -o ./build/exporter-aggregator main.go
```

## Configuration

Create a `config.json` file with the following structure:

```json
{
  "port": 9999,
  "timeout_ms": 2000,
  "targets": [
    {
      "name": "podman-app-1",
      "url": "http://localhost:9888/metrics"
    },
    {
      "name": "podman-app-2",
      "url": "http://localhost:9888/metrics"
    },
    {
      "name": "systemd-app-1",
      "url": "http://localhost:9558/metrics"
    }
  ]
}
```

### Configuration Options

- `port`: The port on which the aggregator server will listen (default: 9999)
- `timeout_ms`: Timeout in milliseconds for HTTP requests to targets (default: 2000)
- `targets`: Array of target objects, each containing:
  - `name`: A descriptive name for the target
  - `url`: The full URL to the target's `/metrics` endpoint

## Usage

Run the aggregator with the default config:

```bash
./exporter-aggregator
```

Run with a custom config file:

```bash
./exporter-aggregator -config /apps/proxy-monitoring/config.json
```

### Command Line Flags

- `-config`: Path to the configuration file (default: "config.json")

## Endpoints

### `/metrics`

Returns aggregated Prometheus metrics from all configured targets. Each target's metrics are prefixed with a comment indicating the target name and URL, followed by an `aggregator_target_up` metric indicating the target's status (1 for up, 0 for down).

Example output:
```
# TARGET podman-app-1 (http://localhost:9888/metrics)
# HELP go_gc_duration_seconds A summary of the pause duration of garbage collection cycles.
...

aggregator_target_up{name="podman-app-1",url="http://localhost:9888/metrics"} 1
```

### `/health`

Returns a JSON response with the health status of all targets.

- **Status Code**: 200 if all targets are healthy, 503 if any target is down
- **Response Body**: JSON object containing overall status and individual target statuses

Example response:
```json
{
  "status": "ok",
  "targets": [
    {"name": "podman-app-1", "url": "http://localhost:9888/metrics", "status": "up"},
    {"name": "podman-app-2", "url": "http://localhost:9888/metrics", "status": "up"}
  ]
}
```

## Monitoring

The aggregator itself exposes metrics that can be scraped by Prometheus. Configure Prometheus to scrape the aggregator's `/metrics` endpoint.

## License

This project is licensed under the GNU General Public License v3.0 - see the [LICENSE](LICENSE) file for details.