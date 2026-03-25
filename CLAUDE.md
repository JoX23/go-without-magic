# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
make run          # Start server (http://localhost:8080)
make build        # Compile Linux binary → bin/go-without-magic
make test         # Run tests with race detector (-race -count=1 -timeout=60s)
make test-cover   # Run tests and generate coverage.html
make lint         # golangci-lint (3m timeout)
make tidy         # go mod tidy + verify
make docker-build # Build multi-stage Docker image
docker-compose up # Full stack: app + PostgreSQL (ports 8080/9090/5432)
make kafka-up     # Full stack + Kafka KRaft + kafka-ui (port 8081)
make kafka-down   # Stop Kafka stack
```

**Single test:**
```bash
go test -race -run TestFunctionName ./internal/service/...
```

**Code generation (new entity from YAML schema):**
```bash
go run ./tools/codegen/ generate --schema product.yaml             # Generate 9 files (full profile)
go run ./tools/codegen/ generate --schema product.yaml --profile full-async  # + Kafka handler
go run ./tools/codegen/ validate --schema product.yaml             # Validate only
go run ./tools/codegen/ list                                        # Show supported types/profiles
```

**Integration tests (requires Docker or running Kafka):**
```bash
make kafka-up
go test -tags=integration -race ./internal/kafka/...
```

## Architecture

This project is a production-ready microservice template following **Clean Architecture** with an explicit dependency rule: outer layers import inner, never the reverse.

```
Transport (HTTP / gRPC / Kafka consumer)
    ↓
Middleware / Interceptors (cross-cutting concerns)
    ↓
Handler (transport ↔ domain translation)
    ↓
Service (business logic orchestration)
    ↓
Domain (entities, errors, repository interfaces) ← zero external imports
    ↓
Repository (persistence: memory | postgres)
```

**Dependency wiring** is entirely explicit in [cmd/server/main.go](cmd/server/main.go) — no DI containers, no reflection.

### Key packages

| Package | Role |
|---|---|
| `internal/domain` | Entities, domain errors, repository interfaces. No external imports. |
| `internal/service` | Business logic. Calls domain constructors, coordinates repos. |
| `internal/handler/http` | Decodes requests, calls service, maps errors to HTTP status codes. Uses Go 1.22 `net/http` routing (no external router). |
| `internal/repository/memory` | Thread-safe in-memory store (dev/tests). Uses `sync.RWMutex`; exposes `CreateIfNotExists()` for atomic check-and-write. |
| `internal/repository/postgres` | PostgreSQL via `pgx/v5` with connection pooling. |
| `internal/middleware` | Composable middleware: RequestID, Logging, RecoveryPanic, Tracing, BusinessMetrics. |
| `internal/observability` | Uber Zap logger, OpenTelemetry tracer, Prometheus metrics. All togglable via config. |
| `internal/resilience` | Zero-dependency circuit breaker and rate limiter, compatible with HTTP middleware. |
| `internal/config` | Viper-backed config: `config.yaml` + `APP_*` env var overrides. |
| `pkg/health` | `/healthz` handler. |
| `pkg/shutdown` | Graceful shutdown with signal handling, timeout, and `sync.Once` idempotency. |
| `internal/kafka` | Kafka transport (opt-in). Consumer loop, producer, interceptors (WithPanic, WithTracing, WithCircuitBreaker), DLT routing, Prometheus metrics. Activated when `kafka.brokers` is non-empty. |
| `internal/kafka/handler` | Business-layer Kafka handlers (`UserKafkaHandler`) and event producers (`UserEventProducer`). |
| `tools/codegen` | YAML schema → up to 10 generated files (domain, service, handler, memory repo, postgres repo, gRPC, kafka handler). Profiles: `full`, `full-async`, `api`, `domain-only`, `no-grpc`. |

### Error handling pattern

Domain errors are defined in `internal/domain/errors.go`. Handlers map them to HTTP status codes in their `handleError()` helper. Never return raw domain errors directly from HTTP handlers.

### Concurrency rules

- The memory repository's `CreateIfNotExists()` is the atomic primitive for avoiding check-then-act races. Do not split into separate `FindBy*` + `Save` calls under different locks.
- The shutdown manager (`pkg/shutdown`) uses `sync.Once` — calling `Wait()` multiple times is safe.

### Configuration

`config.yaml` at the project root drives all settings. Override any key via environment variable with the `APP_` prefix and underscores replacing dots (e.g., `APP_SERVER_HTTP_PORT=8080`).

### gRPC

Optional gRPC server lives in `internal/grpc/`. Proto files are in `internal/grpc/proto/`. The server uses a unary interceptor pattern mirroring the HTTP middleware chain.

### Kafka (opt-in transport)

Kafka lives in `internal/kafka/`. Activated by setting `kafka.brokers` in `config.yaml` or via `APP_KAFKA_BROKERS=host:9092`.

- **Consumer:** poll loop with manual offset commit (at-least-once), retry up to `kafka.max_retries`, then Dead Letter Topic (`{topic}{dlt_suffix}`).
- **Error disposition:** `DispositionFor(err)` maps domain errors → Retry / DLT / Skip. `ErrUserDuplicated` → Skip (idempotent). Parse errors (`ErrInvalidMessage`) → DLT immediately.
- **Interceptors:** `WithPanic` → `WithCircuitBreaker` → `WithTracing` wrap each `MessageHandler`. Reuses `resilience.CircuitBreaker` unchanged.
- **Shutdown:** `KafkaConsumerAdapter` implements `shutdown.Server`; registered LIFO before gRPC and HTTP so ingestion stops first.
- **Metrics:** 6 Prometheus metrics under `kafka_*` prefix (processed, retries, produced, errors, consumer lag, processing duration).
- **Local stack:** `make kafka-up` starts Kafka KRaft (no ZooKeeper) + kafka-ui on port 8081.

### Observability

All three signals are wired at startup in `internal/observability/`:
- **Logging:** structured Zap logger, propagated via context
- **Tracing:** OpenTelemetry SDK (stdout exporter by default; swap exporter to ship to Jaeger/OTLP)
- **Metrics:** Prometheus — HTTP request counters/histograms + business metrics + uptime gauge
