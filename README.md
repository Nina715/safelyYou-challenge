# FleetMetrics

HTTP service that tracks heartbeats and upload-time statistics for a fleet of devices.

## Quick Start

```bash
go run .
```

The server starts on port **8080** by default. A sample `devices.csv` is included in the repo.

## Configuration

All configuration is via environment variables:

| Variable      | Default       | Description                                 |
| ------------- | ------------- | ------------------------------------------- |
| `PORT`        | `8080`        | TCP port the HTTP server listens on         |
| `DEVICES_CSV` | `devices.csv` | Path to the CSV file of device IDs          |
| `LOG_LEVEL`   | `info`        | Log level: `debug`, `info`, `warn`, `error` |

The CSV must have a `device_id` column. Extra columns are ignored; duplicate IDs are de-duplicated.

```csv
device_id,name
aa-bb-cc-dd-ee-ff,sensor-1
11-22-33-44-55-66,sensor-2
```

## Running the Simulator

```bash
./device-simulator -port 8080
```

Results are written to `results.txt` and printed to the console.

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
├── results.txt               # Simulator output (see Running the Simulator)
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

---

## Write-up

### Time Spent and Hardest Part

Initial work took 3-4 hours for a working product to be developed. It then took another 2-3 hours for debugging/unit test/performance update/readme introduction with the help of claude AI.

The trickiest bug was diagnosing why `POST /stats` returned `400` even though the JSON decoded successfully. The simulator sends `sent_at` as the zero time (`0001-01-01T00:00:00Z`) in stats requests, which our server was rejecting with a "sent_at is required" validation. Because the error wasn't logged verbosely enough, it looked like a JSON decode failure. Adding debug logging to dump the raw request body revealed the zero-time field immediately. The fix was to remove the `sent_at` validation from the stats endpoint — it is not used by the stats calculation and is effectively optional there.

### How Would You Extend the Data Model for More Metric Types?

The current design stores each metric type as explicit fields in `DeviceData` (heartbeat ring buffer + Welford's mean for upload time). Adding a new metric — say CPU load or memory usage — currently requires:

1. Adding fields to `DeviceData` in `internal/model/device.go`
2. Adding a `RecordXxx` method
3. Extending `DeviceStats` and `ComputeStats`
4. Adding a service method in `internal/service/fleet.go`
5. Adding a new API endpoint in `internal/api/server.go`

This explicit approach is clear and type-safe for a small number of metrics. For a larger set of heterogeneous metrics, a more flexible approach would be to introduce a generic metric type:

```go
// MetricSeries holds a running mean for any named scalar metric.
type MetricSeries struct {
    count int64
    mean  float64
}

type DeviceData struct {
    mu       sync.RWMutex
    metrics  map[string]*MetricSeries  // keyed by metric name, e.g. "upload_time", "cpu_load"
    // heartbeat ring buffer stays separate — it has different semantics
    ...
}
```

This makes it trivial to ingest any new scalar metric without touching the model layer. The trade-off is that compile-time type safety is replaced by runtime key lookups. The `Store` interface is already metric-agnostic, so no changes are needed there.

### Runtime Complexity

| Operation                 | Time                                                    | Space |
| ------------------------- | ------------------------------------------------------- | ----- |
| `POST /heartbeat`         | O(W) worst case; O(1) amortized for sequential arrivals | O(1)  |
| `POST /stats`             | O(1) — Welford's online algorithm                       | O(1)  |
| `GET /stats`              | O(W) — linear scan of the ring buffer                   | O(1)  |
| Server startup (CSV load) | O(D) where D = number of device IDs                     | O(D)  |

W = `WindowMinutes` = 1440 (24-hour window, a bounded constant).

**Space per device:** O(W) = 1440 bytes for the heartbeat ring buffer + O(1) for the upload mean. **Total:** O(D × W).

**Store lookups** are O(1) average via the sharded hash map (16 shards, FNV hash).

**Trade-off — rolling window vs. all-time uptime:** The spec defines uptime as `(sumHeartbeats / numMinutesBetweenFirstAndLastHeartbeat) * 100` using the device's entire history. Our implementation uses a 24-hour rolling window, which caps memory at 1.4 KB/device instead of growing indefinitely. For the simulator (which runs for ~8 hours), the results are identical. In production, "uptime over the last 24 hours" is also more operationally useful than a lifetime average. The window size is a named constant (`WindowMinutes`) and easy to adjust.

---

## AI Tool Usage

This solution was developed with the assistance of **Claude** (Anthropic), used as a pair-programming tool via the Claude Code CLI. Claude helped with:

- Diagnosing the `sent_at` zero-time bug by suggesting debug logging
- Implementing the ring buffer and Welford's algorithm
- Writing the unit test suite

All architectural decisions (layering, interface design, uptime formula) and code review were done by the author. Claude was directed by the author at each step and did not autonomously design the solution.

---

## Security, Testing, and Deployment

### Security

- **Input validation:** All request bodies are decoded with `encoding/json`; unknown fields are accepted gracefully (not rejected) to allow forward compatibility. `upload_time` is validated as non-negative. `sent_at` is validated as non-zero on the heartbeat endpoint.
- **No authentication:** This prototype has no auth. In production, device-to-cloud communication should use mTLS or short-lived JWT tokens per device.
- **No rate limiting:** A production deployment should add per-device rate limiting (e.g., via a middleware) to prevent a misbehaving device from flooding the store.
- **Read-only stats endpoint:** `GET /stats` is unauthenticated in this prototype. In production it should require an operator role.

### Testing

- Unit tests cover model logic, CSV parsing, fleet service, and HTTP handlers (18 tests, ~77% coverage of hand-written code).
- CI runs tests on every push to any branch and blocks merges to `master` if coverage drops below 70%.
- Integration/load tests are not included but would be the next step: spin up the server and run the device simulator as a black-box test.

### Deployment

For a production deployment:

1. **Container:** Package as a Docker image (`FROM golang:1.22-alpine` → build → `FROM scratch` with the binary).
2. **Persistence:** Replace `MemoryStore` with a time-series store (e.g., InfluxDB, TimescaleDB, or even Postgres with a heartbeats table). The `store.Store` interface makes this a one-file change.
3. **Horizontal scaling:** The in-memory store is single-node. To scale out, move device state to a shared store (Redis or a database) and make the API layer stateless.
4. **Observability:** The structured JSON logging (via `log/slog`) integrates directly with log aggregators (Datadog, Loki). Add Prometheus metrics for request latency and device counts as the next step.
