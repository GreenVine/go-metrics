# go-metrics

## Overview

`go-metrics` is an end-to-end example of an IoT metrics ingestion gRPC server, written in Go. It includes a stressing tester library that simulates multiple devices emitting events at randomised intervals to test the server's performance and behaviour under load.

## Features

- **GORM Database Management**: Database operations are managed by GORM and supports automatic migrations.
- **Protobuf-Driven Design**: [Protocol Buffers](https://protobuf.dev) definitions are available in the [`proto`](/proto) directory. The codegen is managed by [Buf](https://buf.build).
- **Request Validation**: Implemented using [Protovalidate](https://buf.build/docs/protovalidate).
- **Rate Limiting**: Method level and device level rate limiting. Default configuration can be found at [`cmd/server/config.go`](/cmd/server/config.go).
- **Async Alert Generation**: Alerts are generated asynchronously based on device metrics and configured thresholds.

## Components

### Server Binary

The [`server`](/cmd/server/main.go) binary is the core metrics ingestion service that:
- Accepts configuration for IoT devices
- Receives and stores metrics in a SQLite database (using [gORM](https://gorm.io))
- Generates alerts when metric values breach configured thresholds
- Provides a healthz endpoint that shows all metrics, alerts, and configured devices since server startup

### Stresser Binary

The [`stresser`](/cmd/stresser/main.go) binary is a command-line tool designed to:
- Simulate a specified number of IoT devices
- Emit metric events for these devices at randomised intervals to simulate real-world behaviour
- Configure device thresholds randomly within specified ranges

### Envoy Gateway

[Envoy](https://www.envoyproxy.io) is used as an API gateway to transcode requests between gRPC and JSON. The config can be found in the [`envoy`](/envoy/api-gateway.yaml) directory.

## Getting Started

### Build Binaries

Go 1.24+ is required to build the binary.

```bash
go mod download

# Trigger Protobuf and descriptor codegen
go generate ./...

# Build server and stresser libraries
go build cmd/server
go build cmd/stresser
```

### Server Parameters

```
-alertGenInterval duration
      Interval between alert generation attempts (default 2s)
-dbPath string
      Path to SQLite database file (default "go-metrics.db")
-host string
      Binding address
-port int
      Listening port (default 3000)
```

### Stresser Parameters

```
-maxBatteryMetric int
      Maximum battery metric emitted for the device (default 100)
-maxBatteryThreshold int
      Maximum battery threshold for the device (default 70)
-maxMetricInterval duration
      Maximum interval before each metric emission attempt (default 5s)
-maxTemperatureMetric float
      Maximum temperature metric emitted for the device (default 70)
-maxTemperatureThreshold float
      Maximum temperature threshold for the device (default 50)
-metricsPerDevice int
      Number of metrics to be emitted per device (default 10)
-minBatteryMetric int
      Minimum battery metric emitted for the device (default 0)
-minBatteryThreshold int
      Minimum battery threshold for the device (default 30)
-minMetricInterval duration
      Minimum interval before each metric emission attempt (default 1s)
-minTemperatureMetric float
      Minimum temperature metric emitted for the device (default 0)
-minTemperatureThreshold float
      Minimum temperature threshold for the device (default 20)
-server string
      Address of the metrics server (default "127.0.0.1:8080")
-totalDevices int
      Number of devices to be simulated (default 1000)
```

## Running Locally

The entire stack can be built and ran using Docker Compose.

### Start Core Components

To start the `server` and `envoy`, run the following command:

```bash
docker compose -p go-metrics up --build
```

### Run with Stress Testing

To run the stress test against the server, you can activate the `tests` profile. This will start the `stresser` container in addition to the core components:

```bash
docker compose -p go-metrics --profile tests up --build
```

The gRPC and JSON API will be available by default at `127.0.0.1:8080`.

## Example Transactions

The following examples demonstrate how to interact with the API. A sample device ID `12345678-8888-C0DE-BEEF-123456789012` is used in these examples.

### greenvine.gometrics.device.v1.UpsertConfig

Creates or updates the configuration for a device, setting temperature and battery thresholds for alert generation.

**Request (textproto)**
```textproto
parent: "devices/12345678-8888-C0DE-BEEF-123456789012"
config: {
  temperature_threshold: 30.0
  battery_threshold: 20
}
```

**Request (JSON)**
```jsonc
// POST /v1/devices/12345678-8888-C0DE-BEEF-123456789012/config

{
  "temperatureThreshold": 30.0,
  "batteryThreshold": 20
}
```

**Response (JSON)**
```json
{
  "name": "devices/12345678-8888-C0DE-BEEF-123456789012/config",
  "create_time": "2025-07-01T00:11:22Z",
  "update_time": "2025-07-01T00:11:22Z",
  "temperature_threshold": 30.0,
  "battery_threshold": 20
}
```

### greenvine.gometrics.device.v1.CreateMetric

Sends a metric reading from a device to the server for storage and alert evaluation.

**Request (textproto)**
```textproto
parent: "devices/12345678-8888-C0DE-BEEF-123456789012"
metric: {
  temperature: 50.0
  battery: 15
}
```

**Request (JSON)**
```jsonc
// POST /v1/devices/12345678-8888-C0DE-BEEF-123456789012/metrics

{
  "temperature": 50.0,
  "battery": 15
}
```

**Response (JSON)**
```json
{
  "name": "devices/12345678-8888-C0DE-BEEF-123456789012/metrics/ABCDEF12-3456-7890-ABCD-EF1234567890",
  "create_time": "2025-07-01T00:11:22Z",
  "temperature": 50.0,
  "battery": 15
}
```

### greenvine.gometrics.device.v1.ListAlerts

Retrieves recent alerts for a device that were triggered by threshold breaches.

**Request (textproto)**
```textproto
parent: "devices/12345678-8888-C0DE-BEEF-123456789012"
page_size: 10
```

**Request (JSON)**
```
GET /v1/devices/12345678-8888-C0DE-BEEF-123456789012/alerts?pageSize=10
```

**Response (JSON)**
```json
{
  "alerts": [
    {
      "name": "devices/12345678-8888-C0DE-BEEF-123456789012/alerts/FEDCBA98-7654-3210-FEDC-BA9876543210",
      "create_time": "2025-07-01T11:22:33Z",
      "reason": "ALERT_REASON_TEMPERATURE",
      "value": 50.0,
      "threshold": 30.0
    },
    {
      "name": "devices/12345678-8888-C0DE-BEEF-123456789012/alerts/ABCDEF98-7654-3210-FEDC-BA9876543210",
      "create_time": "2025-07-01T00:11:22Z",
      "reason": "ALERT_REASON_BATTERY",
      "value": 15.0,
      "threshold": 20.0
    }
  ]
}
```

### greenvine.gometrics.core.v1.GetHealthInfo

Retrieves health and observability information from the service, including historical metrics, alerts, and configuration updates.

**Request**
```
GET /v1/healthz
```

**Response (JSON)**
```json
{
  "service_logs": {
    "metric_history": [
      {
        "name": "devices/12345678-8888-C0DE-BEEF-123456789012/metrics/ABCDEF12-3456-7890-ABCD-EF1234567890",
        "create_time": "2025-07-01T00:11:22Z",
        "temperature": 50.0,
        "battery": 15
      }
    ],
    "threshold_breaches": [
      {
        "name": "devices/12345678-8888-C0DE-BEEF-123456789012/alerts/FEDCBA98-7654-3210-FEDC-BA9876543210",
        "create_time": "2025-07-01T11:22:33Z",
        "reason": "ALERT_REASON_TEMPERATURE",
        "value": 50.0,
        "threshold": 30.0
      },
      {
        "name": "devices/12345678-8888-C0DE-BEEF-123456789012/alerts/ABCDEF98-7654-3210-FEDC-BA9876543210",
        "create_time": "2025-07-01T00:11:22Z",
        "reason": "ALERT_REASON_BATTERY",
        "value": 15.0,
        "threshold": 20.0
      }
    ],
    "config_updates": [
      {
        "name": "devices/12345678-8888-C0DE-BEEF-123456789012/config",
        "create_time": "2025-07-01T00:11:22Z",
        "update_time": "2025-07-01T00:11:22Z",
        "temperature_threshold": 30.0,
        "battery_threshold": 20
      }
    ]
  }
}
```
