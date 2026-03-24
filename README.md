<div align="center">

# Go Without Magic

**Production-ready Go microservice template — every line of code is yours to read, own, and modify.**

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Tests](https://img.shields.io/badge/tests-passing-brightgreen)](#testing)
[![Race Detector](https://img.shields.io/badge/race%20detector-clean-brightgreen)](#concurrency-safety)
[![Architecture](https://img.shields.io/badge/architecture-clean-blue)](#architecture)

[Quick Start](#-quick-start) · [Code Generator](#-code-generator-new) · [Features](#-whats-inside) · [Comparison](#-vs-other-frameworks)

</div>

---

> **The problem with most Go frameworks:** they do too much for you. When something breaks in production, you're debugging framework internals instead of your own business logic.
>
> **Go Without Magic** gives you a complete, battle-tested microservice foundation where every pattern is explicit, every layer is yours, and the code looks like *you* wrote it — because you understand all of it.

---

## Why This Exists

Most Go microservice starters fall into one of two traps:

- **Too minimal** — just a `main.go` and a "good luck" comment
- **Too magical** — a framework that hides everything and locks you in

This template hits the middle: a **complete, production-grade architecture** that you can clone, read top-to-bottom in an afternoon, and ship to production by Friday.

```
Clone → Understand → Ship
```

No DSLs. No annotations. No hidden layers. Just Go.

---

## What's Inside

| Layer | What it does | Key decisions |
|---|---|---|
| **Domain** | Business entities & contracts | Zero external imports |
| **Service** | Orchestrates use cases | Depends only on domain interfaces |
| **Handler** | HTTP & gRPC translation | Thin layer, maps transport ↔ domain |
| **Repository** | Data persistence | Memory (tests) + PostgreSQL (production) |
| **Middleware** | Cross-cutting concerns | Composable, GoKit-inspired |
| **Observability** | Traces, metrics, logs | OpenTelemetry + Prometheus + Zap |
| **Resilience** | Circuit breaker + rate limiter | Go-Zero inspired, zero dependencies |
| **Code Generator** | Scaffolds new entities | YAML → 9 files, 0 boilerplate |

### Full feature checklist

- [x] Clean Architecture (Domain → Service → Handler → Repository)
- [x] HTTP API with Go 1.22 routing (no external router)
- [x] PostgreSQL via `pgx/v5` with connection pooling
- [x] Structured logging (Uber Zap)
- [x] Distributed tracing (OpenTelemetry)
- [x] Metrics (Prometheus — HTTP, business, uptime)
- [x] Composable middleware (RequestID, Logging, RecoveryPanic, Tracing, BusinessMetrics)
- [x] Circuit Breaker (Closed → Open → HalfOpen)
- [x] Rate Limiting (token bucket, thread-safe)
- [x] Input validation with consistent error mapping
- [x] OpenAPI spec generation
- [x] Optional gRPC support with error code mapping
- [x] Graceful shutdown with cleanup
- [x] Docker + Docker Compose ready
- [x] GitHub Actions CI/CD
- [x] Race-detector clean (`go test -race ./...` ✅)
- [x] **Code generator** — new entity in seconds

---

## Quick Start

```bash
git clone https://github.com/JoX23/go-without-magic.git
cd go-without-magic
go mod download
make run
```

Server starts at `http://localhost:8080`.

```bash
# Create a user
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"email": "alice@example.com", "name": "Alice"}'

# List users
curl http://localhost:8080/users
```

That's it. No config files, no daemon setup, no surprises.

---

## Code Generator (NEW)

> Add a new entity to your service in seconds — not hours.

**The problem:** adding a second entity (`Product`, `Order`, `Invoice`) means touching ~9 files and writing ~685 lines of nearly-identical boilerplate.

**The solution:** describe your entity in YAML, generate everything.

```bash
# 1. Define your entity
cat > product.yaml << 'EOF'
version: "1"
name: Product
fields:
  - name: Sku
    type: string
    unique: true
    validate: "required"
  - name: Name
    type: string
    validate: "required,min=3"
  - name: Price
    type: float64
    validate: "required,min=0"
  - name: Status
    type: enum
    values: ["draft", "published", "archived"]
lookup_keys:
  - field: Sku
EOF

# 2. Generate
go run ./tools/codegen/ generate --schema product.yaml
```

**Output — 9 files, 0 compilation errors:**

```
  WRITE  internal/domain/product_entity.go
  WRITE  internal/domain/product_errors.go
  WRITE  internal/domain/product_repository.go
  WRITE  internal/service/product_service.go
  WRITE  internal/handler/http/product_handler.go
  WRITE  internal/repository/memory/product_repository.go
  WRITE  internal/repository/postgres/product_repository.go
  WRITE  internal/grpc/proto/product.proto
  WRITE  internal/grpc/product_service.go

[codegen] Generated: 9 files
```

The generated code **looks like you wrote it** — no annotations, no magic tags, no framework coupling. Every file passes `gofmt` and is immediately readable.

### Generator commands

```bash
# Preview without writing anything
go run ./tools/codegen/ generate --schema product.yaml --dry-run

# Overwrite with automatic backup
go run ./tools/codegen/ generate --schema product.yaml --force --backup

# Validate schema only
go run ./tools/codegen/ validate --schema product.yaml

# List supported types and profiles
go run ./tools/codegen/ list
```

### Supported field types

| YAML | Go | PostgreSQL | Proto |
|---|---|---|---|
| `string` | `string` | `TEXT` | `string` |
| `int` / `int64` | `int` / `int64` | `INTEGER` / `BIGINT` | `int32` / `int64` |
| `float64` | `float64` | `NUMERIC(12,4)` | `double` |
| `bool` | `bool` | `BOOLEAN` | `bool` |
| `uuid` | `uuid.UUID` | `UUID` | `string` |
| `time` | `time.Time` | `TIMESTAMPTZ` | `string` |
| `enum` | `type T string` + `const` | `TEXT` | `string` |

### Generation profiles

| Profile | Generates | Use when |
|---|---|---|
| `full` *(default)* | All 9 layers | New entity from scratch |
| `api` | Domain + service + HTTP + memory | HTTP-only, no gRPC/Postgres |
| `domain-only` | Domain (entity + errors + interface) | Design the model first |
| `no-grpc` | Everything except gRPC | Pure HTTP service |

---

## Architecture

### The layers

```
┌─────────────────────────────────────────┐
│              HTTP / gRPC                │  ← Transport layer
├─────────────────────────────────────────┤
│               Middleware                │  ← Cross-cutting concerns
├─────────────────────────────────────────┤
│               Handler                  │  ← Translates transport ↔ domain
├─────────────────────────────────────────┤
│               Service                  │  ← Business logic (pure Go)
├─────────────────────────────────────────┤
│               Domain                   │  ← Entities, errors, interfaces
├─────────────────────────────────────────┤
│            Repository                  │  ← Memory | PostgreSQL
└─────────────────────────────────────────┘
```

The dependency rule: **inner layers never import outer layers**. Domain knows nothing about HTTP. Service knows nothing about PostgreSQL. This is what makes the code testable, refactorable, and long-lived.

### Code you can actually read

```go
// Every dependency is explicit. No magic injection.
userRepo    := memory.NewUserRepository()
userService := service.NewUserService(userRepo, logger)
userHandler := handler.NewUserHandler(userService, logger)
userHandler.RegisterRoutes(mux)
```

Compare that to frameworks where the startup code is a wall of decorators, registrations, and reflection calls. Here, you see the graph.

### Project structure

```
.
├── cmd/server/           # Entrypoint + dependency wiring
├── internal/
│   ├── domain/           # Entities, errors, repository interfaces
│   ├── service/          # Business logic
│   ├── handler/http/     # HTTP handlers
│   ├── repository/
│   │   ├── memory/       # In-memory (tests + local dev)
│   │   └── postgres/     # Production persistence
│   ├── middleware/       # Composable middleware chain
│   ├── observability/    # Tracing, metrics, logging
│   ├── resilience/       # Circuit breaker, rate limiter
│   ├── validator/        # Input validation + error mapping
│   ├── grpc/             # Optional gRPC server
│   └── openapi/          # OpenAPI spec generation
├── tools/codegen/        # Code generator
│   ├── templates/        # Go templates for each layer
│   ├── schema/           # YAML schema parser + validator
│   └── examples/         # Example entity schemas
└── deployments/docker/   # Dockerfile + Compose
```

---

## Middleware System

Compose exactly what you need, in the order you want:

```go
handler := middleware.Chain(
    middleware.RequestID(),
    middleware.Logging(logger),
    middleware.Tracing(tracer, spanProcessor, metrics),
    middleware.BusinessMetrics(metrics, spanProcessor),
    middleware.RecoveryPanic(logger),
)(userHandler)
```

Each middleware does one thing. You can add, remove, or reorder them without touching business logic.

---

## Resilience Patterns

```go
// Circuit Breaker — fail fast when a dependency degrades
cb := resilience.NewCircuitBreaker(5, 30*time.Second)
err := cb.Call(func() error {
    return callExternalService()
})
// After 5 consecutive failures: opens for 30s, then probes

// Rate Limiter — smooth traffic distribution
rl := resilience.NewRateLimiter(100) // 100 req/sec
if !rl.Allow() {
    http.Error(w, "too many requests", http.StatusTooManyRequests)
}
```

Both are HTTP-middleware compatible and zero-dependency (no Hystrix, no Sentinel).

---

## Observability

Three signals, fully wired:

```go
// Distributed tracing — every request gets a span
tp, _ := observability.NewTracerProvider("my-service", "1.0.0")
tracer := tp.Tracer("handler")

ctx, span := observability.StartSpan(ctx, tracer, "create-user")
defer span.End()

// Prometheus metrics — HTTP + business + uptime
metrics := observability.NewMetrics()
metrics.RecordHTTPRequest("POST", "/users", 201, duration)
metrics.RecordUserOperation("create", "success")

// Structured logs — machine-readable from day one
logger.Info("user created",
    zap.String("id", user.ID.String()),
    zap.Duration("latency", duration),
)
```

---

## Concurrency Safety

Verified clean under load:

| Test | Result |
|---|---|
| `go test -race ./...` | 0 race conditions |
| Load test (100 concurrent) | 7,307+ req/sec sustained |
| Atomic repository operations | Verified — no lost writes |
| Graceful shutdown | Idempotent, drains in-flight requests |

---

## Configuration

```env
DATABASE_URL=postgres://user:password@localhost:5432/mydb
APP_ENV=development
LOG_LEVEL=info
PORT=8080
```

All config via environment variables (12-factor). Managed by Viper — supports `.env` files, env vars, and config files interchangeably.

---

## Available Commands

```bash
make run          # Run in development mode
make build        # Compile production binary → bin/
make test         # Run tests with race detector
make test-cover   # Generate coverage.html
make lint         # golangci-lint
make docker-build # Build Docker image
make tidy         # Clean and verify dependencies
```

---

## vs. Other Frameworks

| | Go Without Magic | GoKit | Kratos | Go-Zero |
|---|---|---|---|---|
| **Philosophy** | Explicit, own every line | Flexible toolkit | Full framework | Code generation first |
| **Setup** | Clone & run (5 min) | Integrate (15 min) | CLI setup (10 min) | goctl install (5 min) |
| **Code you read** | All of it | Most of it | Framework internals | Generated + DSL |
| **Lock-in** | None | Low | Medium | High (goctl DSL) |
| **Code generation** | ✅ (your architecture) | ❌ | ❌ | ✅ (framework's architecture) |
| **Observability** | ✅ Full | ⚠️ Minimal | ✅ Full | ✅ Full |
| **gRPC** | ✅ Optional | ✅ Pluggable | ✅ Built-in | ✅ Built-in |
| **Resilience** | ✅ Built-in | ✅ Pluggable | ✅ Built-in | ✅ Built-in |
| **Learning curve** | Low | Medium | Medium-High | Low→High |
| **Debuggability** | Very high | High | Medium | Low (generated code) |

### Pick this if you

- Want to understand every line of your service in production
- Are learning Go architecture patterns and want working reference code
- Value debuggability over convenience
- Want code generation that respects *your* architecture, not a framework's
- Are building an MVP and need to move fast without accumulating tech debt

### Pick something else if you

- **GoKit** — need maximum composability across a large microservices mesh
- **Kratos** — need protobuf-first development with enterprise support
- **Go-Zero** — need to ship dozens of CRUD services as fast as possible

---

## Dependencies

Carefully chosen — every dependency earns its place:

| Package | Why |
|---|---|
| `go.uber.org/zap` | Structured, performant logging |
| `github.com/spf13/viper` | 12-factor config management |
| `github.com/jackc/pgx/v5` | Best-in-class PostgreSQL driver |
| `github.com/google/uuid` | UUID generation |
| `github.com/stretchr/testify` | Test assertions |
| `go.opentelemetry.io/otel` | Distributed tracing standard |
| `github.com/prometheus/client_golang` | Metrics |
| `google.golang.org/grpc` | Optional gRPC transport |
| `gopkg.in/yaml.v3` | Schema parsing (codegen only) |

No ORMs. No DI containers. No reflection magic.

---

## CI/CD

GitHub Actions workflows (`.github/workflows/`) run on every push and PR:

- `go test -race ./...` — tests with race detector
- `golangci-lint` — static analysis
- `go build ./...` — compilation check

---

## Contributing

Contributions are welcome. The bar is: **every addition must be readable, explicit, and serve a clear purpose**.

1. Fork the repo
2. Create a branch (`git checkout -b feat/my-feature`)
3. Make your changes — keep them focused
4. Run `make test && make lint`
5. Open a PR with a clear description

If you're adding a new pattern, explain *why* it belongs here and not in a separate library.

---

## License

MIT — use it, fork it, build on it.

---

<div align="center">

**If this saved you hours of boilerplate, consider giving it a ⭐**

Built with the philosophy that the best framework is one you fully understand.

</div>
