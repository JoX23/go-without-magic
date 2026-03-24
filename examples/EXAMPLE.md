# Ejemplo real: Tienda online básica

Este walkthrough muestra cómo usar **go-without-magic** para levantar una API
funcional con usuarios y catálogo de productos, observabilidad incluida.

---

## Escenario

Construimos el backend de una tienda online mínima:

- **Usuarios** — registro con email único
- **Productos** — catálogo con SKU único, precio y estados (`draft` / `published` / `archived`)
- **Observabilidad** — trazas distribuidas + métricas Prometheus en tiempo real

```
Cliente HTTP
     │
     ▼
 Middleware (RequestID → Logging → Tracing → Metrics → RecoveryPanic)
     │
     ▼
 Handler HTTP  ─────────────►  Service  ──────────►  Repository
 POST /users                  UserService            memory.UserRepository
 GET  /users/:id              ProductService         memory.ProductRepository
 POST /products
 GET  /products/:id
 GET  /healthz
 GET  /metrics
```

---

## Opción A: Levantar con Docker Compose (recomendado)

Levanta el servicio + PostgreSQL con un solo comando:

```bash
docker compose up --build
```

Espera hasta ver:

```
app  | {"level":"info","msg":"HTTP server listening","addr":":8080"}
```

Listo. El servicio está en `http://localhost:8080`.

---

## Opción B: Levantar en local (sin Docker)

```bash
# Instalar dependencias
go mod download

# Ejecutar (usa repositorio en memoria por defecto)
make run
```

---

## Demo automatizada

Con el servicio corriendo, ejecuta el script que valida todos los endpoints:

```bash
chmod +x examples/demo.sh
./examples/demo.sh
```

Salida esperada:

```
go-without-magic — demo completo
Servidor: http://localhost:8080

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  1. HEALTH CHECK
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

▶ GET /healthz — servicio listo
  ✓ healthz responde (HTTP 200)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  2. GESTIÓN DE USUARIOS
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

...
  ✓ Todos los checks pasaron (18/18)
```

---

## Walkthrough manual paso a paso

### Health check

```bash
curl http://localhost:8080/healthz
```

```json
{"status":"ok"}
```

---

### Usuarios

#### Crear usuario

```bash
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"email": "alice@example.com", "name": "Alice Smith"}'
```

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "alice@example.com",
  "name": "Alice Smith",
  "created_at": "2026-03-24T18:00:00Z"
}
```

El `id` es un UUID v4 generado en el momento de creación, nunca reutilizado.

#### Intentar email duplicado

```bash
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"email": "alice@example.com", "name": "Otra Alice"}'
```

```json
{"error": "user already exists"}
```

HTTP `409 Conflict` — el dominio rechaza duplicados **atómicamente** (sin race conditions, incluso bajo carga).

#### Buscar por ID

```bash
curl http://localhost:8080/users/550e8400-e29b-41d4-a716-446655440000
```

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "alice@example.com",
  "name": "Alice Smith",
  "created_at": "2026-03-24T18:00:00Z"
}
```

#### ID inexistente

```bash
curl http://localhost:8080/users/00000000-0000-0000-0000-000000000000
```

```json
{"error": "user not found"}
```

HTTP `404 Not Found`.

#### Listar todos

```bash
curl http://localhost:8080/users
```

```json
[
  {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "alice@example.com",
    "name": "Alice Smith",
    "created_at": "2026-03-24T18:00:00Z"
  }
]
```

---

### Productos

#### Crear producto

```bash
curl -X POST http://localhost:8080/products \
  -H "Content-Type: application/json" \
  -d '{
    "sku": "LAPTOP-001",
    "name": "MacBook Pro 14",
    "price": 1999.99
  }'
```

```json
{
  "id": "7c9e6679-7425-40de-944b-e07fc1f90ae7",
  "sku": "LAPTOP-001",
  "name": "MacBook Pro 14",
  "price": 1999.99,
  "status": "draft",
  "created_at": "2026-03-24T18:01:00Z"
}
```

El campo `status` es `"draft"` por defecto (definido en el schema del generador como primer valor del enum).

#### SKU duplicado

```bash
curl -X POST http://localhost:8080/products \
  -H "Content-Type: application/json" \
  -d '{"sku": "LAPTOP-001", "name": "Copia", "price": 500}'
```

```json
{"error": "product already exists"}
```

HTTP `409 Conflict`.

#### Buscar producto por ID

```bash
curl http://localhost:8080/products/7c9e6679-7425-40de-944b-e07fc1f90ae7
```

#### Listar productos

```bash
curl http://localhost:8080/products
```

```json
[
  {
    "id": "7c9e6679-7425-40de-944b-e07fc1f90ae7",
    "sku": "LAPTOP-001",
    "name": "MacBook Pro 14",
    "price": 1999.99,
    "status": "draft",
    "created_at": "2026-03-24T18:01:00Z"
  },
  {
    "id": "a87ff679-a2f3-471d-b1a0-a3b1c8f6e3c4",
    "sku": "KEYBOARD-001",
    "name": "Keychron K2 Pro",
    "price": 119.99,
    "status": "draft",
    "created_at": "2026-03-24T18:02:00Z"
  }
]
```

---

### Observabilidad

#### Métricas Prometheus

```bash
curl http://localhost:8080/metrics
```

Fragmento relevante:

```
# HELP http_requests_total Total number of HTTP requests
# TYPE http_requests_total counter
http_requests_total{endpoint="/users",method="POST",status_code="201"} 2
http_requests_total{endpoint="/users",method="GET",status_code="200"} 5
http_requests_total{endpoint="/products",method="POST",status_code="201"} 2
http_requests_total{endpoint="/products/{id}",method="GET",status_code="200"} 1

# HELP http_request_duration_seconds HTTP request duration in seconds
# TYPE http_request_duration_seconds histogram
http_request_duration_seconds_bucket{endpoint="/users",method="POST",le="0.005"} 2
http_request_duration_seconds_sum{endpoint="/users",method="POST"} 0.000412

# HELP user_operations_total Total number of user operations
# TYPE user_operations_total counter
user_operations_total{operation="create",result="success"} 2
```

Conecta Prometheus + Grafana a `http://localhost:8080/metrics` para dashboards en tiempo real.

#### Trazas distribuidas (logs)

Cada request genera un span completo en los logs:

```json
{
  "Name": "http.request",
  "SpanContext": {
    "TraceID": "d1ff0aef0852d5c16102f8168bb6401d",
    "SpanID": "de17c85553c411b6"
  },
  "Attributes": [
    {"Key": "http.method",      "Value": {"Type": "STRING", "Value": "POST"}},
    {"Key": "http.url",         "Value": {"Type": "STRING", "Value": "/users"}},
    {"Key": "http.status_code", "Value": {"Type": "INT64",  "Value": 201}}
  ],
  "Status": {"Code": "Ok"}
}
```

El `TraceID` es consistente a través de todos los spans del mismo request. Útil para correlacionar logs en producción.

---

## Agregar una nueva entidad

¿Necesitas `Order`, `Invoice`, `Category`? El generador produce todo el boilerplate:

```bash
# 1. Define el schema
cat > order.yaml << 'EOF'
version: "1"
name: Order
fields:
  - name: UserId
    type: uuid
    validate: "required"
  - name: Total
    type: float64
    validate: "required,min=0"
  - name: Status
    type: enum
    values: ["pending", "confirmed", "shipped", "delivered", "cancelled"]
lookup_keys:
  - field: UserId
EOF

# 2. Genera los 9 archivos
go run ./tools/codegen/ generate --schema order.yaml

# 3. Agrega el wiring sugerido a cmd/server/main.go
# (el generador imprime exactamente qué agregar)
```

En menos de 2 minutos tienes:
- `GET /orders` · `POST /orders` · `GET /orders/{id}`
- Repositorio en memoria (listo para tests) y PostgreSQL
- Errores de dominio tipados y mapeados a HTTP
- Métricas automáticas por endpoint

---

## Estructura del código generado

Para entender cómo fluye una petición `POST /users`:

```
HTTP Request
    │
    ▼
middleware.Chain(RequestID → Logging → Tracing → RecoveryPanic)
    │
    ▼
UserHandler.CreateUser()          ← internal/handler/http/handler.go
    │  decode JSON
    │  validar campos requeridos
    ▼
UserService.CreateUser()          ← internal/service/service.go
    │  domain.NewUser() — valida invariantes
    │  repo.CreateIfNotExists() — atómico
    ▼
memory.UserRepository             ← internal/repository/memory/repository.go
    │  Lock() → check email → write → Unlock()
    ▼
domain.User{}                     ← internal/domain/entity.go
    │  retorna entidad
    ▼
UserHandler → toResponse() → JSON 201
```

Cada capa tiene **una sola responsabilidad** y no importa nada de las capas exteriores.

---

## Cambiar a PostgreSQL

Reemplaza la línea del repositorio en `cmd/server/main.go`:

```go
// Antes (memoria):
repo := memory.NewUserRepository()

// Después (PostgreSQL):
repo, err := postgres.New(cfg.Database)
if err != nil {
    return fmt.Errorf("connecting to database: %w", err)
}
defer repo.Close()
```

El resto del código no cambia. El servicio, los handlers y los tests son **completamente agnósticos** de la implementación de persistencia.
