# Telemetry Sink

A WAL-based telemetry ingestion sink with batching, rate-limiting, and gRPC transport. Designed to work together with a **Telemetry Node** that produces and streams telemetry data.

---

## Components

- **Telemetry Node** – produces telemetry and streams it to the sink
- **Telemetry Sink** – receives, rate-limits, batches, and persists telemetry

---

## Telemetry Node

The **Telemetry Node** is a client-side component responsible for:

- Collecting telemetry data (metrics, events, logs)
- Buffering telemetry in memory
- Sending telemetry to the sink using gRPC or HTTP
- Handling backpressure and retries

### Node Responsibilities

- **Non-blocking ingestion** – telemetry producers never block on network I/O
- **Transport abstraction** – supports multiple transports (gRPC, HTTP)
- **Queue-based buffering** – drops or retries messages when overloaded
- **Graceful shutdown** – flushes buffered telemetry before exit

### Data Flow (Node → Sink)

```
Producers  
↓  
Node Queue  
↓  
Transport (gRPC / HTTP)  
↓  
Telemetry Sink
```

### Configuration

#### Node

| Flag               | Default   | Description                   |
| ------------------ | --------- | ----------------------------- |
| `-node.queue-size` | `100`     | Telemetry queue buffer size   |
| `-node.rate`       | `100`     | Telemetry messages per second |
| `-node.sensor`     | `default` | Sensor name (used in metrics) |

#### Retry

| Flag                | Default | Description                 |
| ------------------- | ------- | --------------------------- |
| `-retry.base-delay` | `200ms` | Initial retry backoff delay |
| `-retry.max`        | `5`     | Maximum retry attempts      |
| `-retry.max-delay`  | `5s`    | Maximum retry backoff delay |

#### Transport

| Flag                      | Default                 | Description                   |
| ------------------------- | ----------------------- | ----------------------------- |
| `-transport.type`         | `grpc`                  | Transport type (http or grpc) |
| `-transport.sink-address` | `http://localhost:8080` | Telemetry sink address        |
| `-transport.timeout`      | `5s`                    | Transport request timeout     |

#### Transport TLS / mTLS

| Flag                         | Default               | Description                         |
| ---------------------------- | --------------------- | ----------------------------------- |
| `-transport.tls.enabled`     | `true`                | Enable TLS/mTLS for transport       |
| `-transport.tls.insecure`    | `false`               | Skip TLS verification (DEV ONLY)    |
| `-transport.tls.ca`          | `certs/ca/ca.pem`     | Path to CA certificate (PEM)        |
| `-transport.tls.cert`        | `certs/node/node.pem` | Path to client certificate (PEM)    |
| `-transport.tls.key`         | `certs/node/node.key` | Path to client private key (PEM)    |
| `-transport.tls.server-name` | `telemetry-sink`      | TLS server name override (optional) |

---

## Telemetry Sink

The sink is a server-side application that reliably stores telemetry.

### Features

- **WAL-style TelemetryLog** – append-only persistent storage
- **Batching** – flush by count, size, or time
- **Rate Limiting** – per-message and per-byte limits
- **gRPC Server** – streaming ingestion API
- **Graceful Shutdown** – flushes in-flight data before exit

### Configuration

#### Batch

| Flag                    | Default | Description                      |
| ----------------------- | ------- | -------------------------------- |
| `-batch.flush-interval` | `1s`    | Max time before batch is flushed |
| `-batch.max-bytes`      | `65536` | Max batch size in bytes          |
| `-batch.max-count`      | `100`   | Max telemetry messages per batch |

#### Rate Limit

| Flag                       | Default | Description                                 |
| -------------------------- | ------- | ------------------------------------------- |
| `-ratelimit.bytes-burst`   | `0`     | Burst size for byte rate limiter            |
| `-ratelimit.bytes-per-sec` | `0`     | Bytes per second rate limit (0 = unlimited) |
| `-ratelimit.msgs-burst`    | `0`     | Burst size for message rate limiter         |
| `-ratelimit.msgs-per-sec`  | `0`     | Max messages per second (0 = unlimited)     |

#### Sink

| Flag                     | Default           | Description                   |
| ------------------------ | ----------------- | ----------------------------- |
| `-sink.log-path`         | `./telemetry.wal` | Path to telemetry WAL file    |
| `-sink.queue-size`       | `1000`            | Telemetry channel buffer size |
| `-sink.shutdown-timeout` | `5s`              | Server shutdown timeout       |

#### Transport TLS / mTLS

| Flag                         | Default               | Description                         |
| ---------------------------- | --------------------- | ----------------------------------- |
| `-transport.tls.enabled`     | `true`                | Enable TLS/mTLS for transport       |
| `-transport.tls.insecure`    | `false`               | Skip TLS verification (DEV ONLY)    |
| `-transport.tls.ca`          | `certs/ca/ca.pem`     | Path to CA certificate (PEM)        |
| `-transport.tls.cert`        | `certs/sink/sink.pem` | Path to server certificate (PEM)    |
| `-transport.tls.key`         | `certs/sink/sink.key` | Path to server private key (PEM)    |
| `-transport.tls.server-name` | `telemetry-sink`      | TLS server name override (optional) |

#### Transport

| Flag                      | Default | Description          |
| ------------------------- | ------- | -------------------- |
| `-transport.sink-address` | `:9000` | Address to listen on |

---

## Usage

### Generate certificates

In project root folder execute 
```bash
./certs/generate.sh
```

### Node Usage Example

#### Run the telemetry node with default settings:

```bash
go run ./cmd/node
```

#### Run the node with custom settings:
```bash
go run ./cmd/node \
  -node.queue-size=200 \
  -node.rate=50 \
  -node.sensor="temperature" \
  -retry.base-delay=500ms \
  -retry.max=10 \
  -retry.max-delay=10s \
  -transport.type="grpc" \
  -transport.sink-address="localhost:50051" \
  -transport.timeout=5s \
  -transport.tls.enabled=true \
  -transport.tls.insecure=false \
  -transport.tls.ca="certs/ca/ca.pem" \
  -transport.tls.cert="certs/node/node.pem" \
  -transport.tls.key="certs/node/node.key" \
  -transport.tls.server-name="telemetry-sink"
```

#### Run Sink
```bash
go run ./cmd/sink \
  -sink.queue-size=1000 \
  -sink.log-path="/tmp/telemetry.log" \
  -sink.shutdown-timeout=5s \
  -batch.flush-interval=1s \
  -batch.max-bytes=65536 \
  -batch.max-count=100 \
  -ratelimit.msgs-per-sec=1000 \
  -ratelimit.msgs-burst=5000 \
  -ratelimit.bytes-per-sec=0 \
  -ratelimit.bytes-burst=0 \
  -transport.sink-address="localhost:50051" \
  -transport.tls.enabled=true \
  -transport.tls.insecure=false \
  -transport.tls.ca="certs/ca/ca.pem" \
  -transport.tls.cert="certs/sink/sink.pem" \
  -transport.tls.key="certs/sink/sink.key" \
  -transport.tls.server-name="telemetry-sink"
```

#### Notes: 
- TLS/mTLS is enabled by default. Use -transport.tls.insecure=true only for local testing.

- The node will automatically retry sending telemetry based on -retry.max and backoff settings.

- Sensor name (-node.sensor) can be any string identifying the source of telemetry.
---

## Architecture Overview
```
Telemetry Node
  ├─ Producers
  ├─ Buffered Queue
  └─ gRPC / HTTP Client
            ↓
Telemetry Sink
  ├─ RateLimitedIngestor
  ├─ ChannelIngestor
  ├─ TelemetryWorker
  └─ TelemetryLog (WAL)
```

## Graceful Shutdown

### Node
- Stops producers
- Flushes queued telemetry
- Closes transport connections

### Sink
- Stops accepting new connections
- Drains ingest channel
- Flushes remaining batches
- Closes WAL and exits cleanly
