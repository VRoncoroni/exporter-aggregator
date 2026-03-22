# Exporter Aggregator (Go)

A lightweight, dependency-free metrics aggregator designed to collect and expose metrics from multiple Prometheus exporters through a single endpoint.

## 🚀 Features

* Single `/metrics` endpoint for multiple exporters
* Parallel requests (non-blocking)
* Built-in health metrics
* JSON configuration file
* No external dependencies
* Single static binary (easy deployment)

---

## 🧠 How It Works

The aggregator:

1. Fetches metrics from multiple targets in parallel
2. Concatenates their outputs
3. Adds health metrics for each target
4. Exposes everything in Prometheus format

---

## 📦 Requirements

* Go (for building only)
* Linux / Unix system (or compatible)

---

## 🛠 Installation

### 1. Clone or create project directory

```bash
mkdir -p /apps/exporter-aggregator
cd /apps/exporter-aggregator
```

### 2. Build the binary

```bash
go build -o ./build/exporter-aggregator
```

For cross-compilation (from Windows to Linux):

```bash
$env:GOOS="linux"
$env:GOARCH="amd64"
go build -o ./build/exporter-aggregator
```

---

## ⚙️ Configuration

Create a `config.json` file:

```json
{
  "port": 9999,
  "timeout_ms": 2000,
  "targets": [
    {
      "name": "svc_homepage_podman_exporter",
      "url": "http://localhost:9881/metrics"
    },
    {
      "name": "svc_homepage_systemd_exporter",
      "url": "http://localhost:9921/metrics"
    }
  ]
}
```

### Fields

| Field        | Description                |
| ------------ | -------------------------- |
| `port`       | HTTP port to expose        |
| `timeout_ms` | Timeout per target request |
| `targets`    | List of exporters          |

---

## ▶️ Run

```bash
./exporter-aggregator -config /apps/exporter-aggregator/config.json
```

---

## 📡 Endpoints

### `/metrics`

Prometheus-compatible endpoint containing:

* Aggregated metrics from all targets
* Health metrics per target
* Global health status

---

## 📊 Example Output

```
# HELP aggregator_target_up Health status of each target (1=up, 0=down)
# TYPE aggregator_target_up gauge

# TARGET svc_homepage_podman_exporter (http://localhost:9881/metrics)
<metrics...>
aggregator_target_up{name="svc_homepage_podman_exporter"} 1

# HELP aggregator_up Global health status (1=ok, 0=degraded)
# TYPE aggregator_up gauge
aggregator_up 1

# HELP aggregator_targets_total Number of targets
# TYPE aggregator_targets_total gauge
aggregator_targets_total 2
```

---

## 🔌 Prometheus Configuration

```yaml
scrape_configs:
  - job_name: "exporter-aggregator"
    static_configs:
      - targets: ["localhost:9999"]
```

---

## 📈 Available Metrics

| Metric                     | Description                      |
| -------------------------- | -------------------------------- |
| `aggregator_target_up`     | Target health (1 = up, 0 = down) |
| `aggregator_up`            | Global health status             |
| `aggregator_targets_total` | Number of configured targets     |

---

## ⚠️ Limitations

* Metrics are concatenated (may include duplicate `HELP` / `TYPE`)
* No deduplication or label rewriting
* Not intended to replace native Prometheus multi-target scraping

---

## 💡 Use Cases

* Reduce number of Prometheus scrape targets (multi user environments)
* Aggregate exporters behind firewalls or proxies
* Simple monitoring setups
* Edge environments

---

## 📄 License

GNU General Public License v3.0 (GPL-3.0)

---

## 👨‍💻 Author

Valentin RONCORONI