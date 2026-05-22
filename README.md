# FleetMetrics

HTTP service that tracks heartbeats and upload-time statistics for a fleet of devices.

## Prerequisites

- Go 1.22+
- A `devices.csv` file listing the registered device IDs (see [Configuration](#configuration))

## Quick Start

```bash
go run .
```

The server starts on port **8080** by default. A sample `devices.csv` is included in the repo.

## Configuration

All configuration is via environment variables:

| Variable     | Default        | Description                              |
|--------------|----------------|------------------------------------------|
| `PORT`       | `8080`         | TCP port the HTTP server listens on      |
| `DEVICES_CSV`| `devices.csv`  | Path to the CSV file of device IDs       |
| `LOG_LEVEL`  | `info`         | Log level: `debug`, `info`, `warn`, `error` |

The CSV must have a `device_id` column. Extra columns are ignored; duplicate IDs are de-duplicated.

```csv
device_id,name
aa-bb-cc-dd-ee-ff,sensor-1
11-22-33-44-55-66,sensor-2
```

## API

Base path: `/api/v1`

### POST `/devices/{deviceId}/heartbeat`

Record that a device is alive.

```json
{ "sent_at": "2024-01-01T12:00:00Z" }
```

Returns `204 No Content`.

### POST `/devices/{deviceId}/stats`

Record an upload-time sample (nanoseconds).

```json
{ "upload_time": 209226522788 }
```

Returns `204 No Content`.

### GET `/devices/{deviceId}/stats`

Retrieve uptime percentage and average upload time.

```json
{
  "uptime": 92.91667,
  "avg_upload_time": "3m29.226522788s"
}
```

`uptime` is `(active minutes / window minutes) * 100`. `avg_upload_time` uses Go's `time.Duration` string format.

### GET `/health`

```json
{ "status": "ok" }
```

## Running Tests

```bash
go test ./...
```

With coverage:

```bash
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

## Project Structure

```
.
├── main.go                   # Entry point, server wiring
├── devices.csv               # Registered device IDs
├── openapi.json              # OpenAPI spec (source of truth for the API)
└── internal/
    ├── api/
    │   ├── api.gen.go        # Generated from openapi.json (do not edit)
    │   └── server.go         # Handler implementations
    ├── model/
    │   └── device.go         # Per-device data and stats computation
    ├── service/
    │   ├── fleet.go          # Business logic
    │   └── csv.go            # Device ID loading from CSV
    └── store/
        └── memory.go         # In-memory device store
```
