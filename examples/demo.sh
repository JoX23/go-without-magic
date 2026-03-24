#!/usr/bin/env bash
# =============================================================================
# demo.sh — Demostración completa de go-without-magic
#
# Escenario: tienda online básica con usuarios y catálogo de productos.
#
# Requisitos: curl, jq
# Uso:
#   chmod +x examples/demo.sh
#   ./examples/demo.sh                      # contra localhost:8080
#   BASE_URL=http://mi-servidor ./examples/demo.sh
# =============================================================================

set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"
PASS=0
FAIL=0

# ── Colores ──────────────────────────────────────────────────────────────────
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
BOLD='\033[1m'
RESET='\033[0m'

# ── Helpers ───────────────────────────────────────────────────────────────────
header() {
  echo ""
  echo -e "${BLUE}${BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
  echo -e "${BLUE}${BOLD}  $1${RESET}"
  echo -e "${BLUE}${BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
}

step() {
  echo -e "\n${YELLOW}▶ $1${RESET}"
}

ok() {
  echo -e "${GREEN}  ✓ $1${RESET}"
  PASS=$((PASS + 1))
}

fail() {
  echo -e "${RED}  ✗ $1${RESET}"
  FAIL=$((FAIL + 1))
}

assert_status() {
  local expected="$1"
  local actual="$2"
  local label="$3"
  if [ "$actual" -eq "$expected" ]; then
    ok "$label (HTTP $actual)"
  else
    fail "$label — esperado HTTP $expected, recibido HTTP $actual"
  fi
}

assert_field() {
  local json="$1"
  local field="$2"
  local expected="$3"
  local actual
  actual=$(echo "$json" | jq -r "$field" 2>/dev/null || echo "")
  if [ "$actual" = "$expected" ]; then
    ok "campo $field = \"$expected\""
  else
    fail "campo $field — esperado \"$expected\", recibido \"$actual\""
  fi
}

# ── Inicio ────────────────────────────────────────────────────────────────────
echo ""
echo -e "${BOLD}go-without-magic — demo completo${RESET}"
echo -e "Servidor: ${BASE_URL}"
echo -e "$(date)"

# =============================================================================
header "1. HEALTH CHECK"
# =============================================================================

step "GET /healthz — servicio listo"
RESPONSE=$(curl -s -w "\n%{http_code}" "${BASE_URL}/healthz")
BODY=$(echo "$RESPONSE" | sed '$d')
STATUS=$(echo "$RESPONSE" | tail -n 1)

assert_status 200 "$STATUS" "healthz responde"
echo "  $BODY"

# =============================================================================
header "2. GESTIÓN DE USUARIOS"
# =============================================================================

step "POST /users — crear usuario Alice"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "${BASE_URL}/users" \
  -H "Content-Type: application/json" \
  -d '{"email": "alice@example.com", "name": "Alice Smith"}')
BODY=$(echo "$RESPONSE" | sed '$d')
STATUS=$(echo "$RESPONSE" | tail -n 1)

assert_status 201 "$STATUS" "crear usuario"
assert_field "$BODY" ".email" "alice@example.com"
assert_field "$BODY" ".name" "Alice Smith"

ALICE_ID=$(echo "$BODY" | jq -r ".id")
ok "ID asignado: $ALICE_ID"

step "POST /users — crear segundo usuario Bob"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "${BASE_URL}/users" \
  -H "Content-Type: application/json" \
  -d '{"email": "bob@example.com", "name": "Bob Johnson"}')
BODY=$(echo "$RESPONSE" | sed '$d')
STATUS=$(echo "$RESPONSE" | tail -n 1)

assert_status 201 "$STATUS" "crear segundo usuario"
BOB_ID=$(echo "$BODY" | jq -r ".id")
ok "ID asignado: $BOB_ID"

step "POST /users — email duplicado (debe rechazarse)"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "${BASE_URL}/users" \
  -H "Content-Type: application/json" \
  -d '{"email": "alice@example.com", "name": "Alice Duplicada"}')
BODY=$(echo "$RESPONSE" | sed '$d')
STATUS=$(echo "$RESPONSE" | tail -n 1)

assert_status 409 "$STATUS" "duplicado rechazado con 409 Conflict"

step "POST /users — datos inválidos (email vacío)"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "${BASE_URL}/users" \
  -H "Content-Type: application/json" \
  -d '{"email": "", "name": "Sin email"}')
BODY=$(echo "$RESPONSE" | sed '$d')
STATUS=$(echo "$RESPONSE" | tail -n 1)

assert_status 400 "$STATUS" "email vacío rechazado con 400 Bad Request"

step "GET /users/${ALICE_ID} — buscar Alice por ID"
RESPONSE=$(curl -s -w "\n%{http_code}" "${BASE_URL}/users/${ALICE_ID}")
BODY=$(echo "$RESPONSE" | sed '$d')
STATUS=$(echo "$RESPONSE" | tail -n 1)

assert_status 200 "$STATUS" "buscar por ID"
assert_field "$BODY" ".id" "$ALICE_ID"
assert_field "$BODY" ".email" "alice@example.com"

step "GET /users — listar todos los usuarios"
RESPONSE=$(curl -s -w "\n%{http_code}" "${BASE_URL}/users")
BODY=$(echo "$RESPONSE" | sed '$d')
STATUS=$(echo "$RESPONSE" | tail -n 1)

assert_status 200 "$STATUS" "listar usuarios"
COUNT=$(echo "$BODY" | jq '. | length')
ok "total usuarios en sistema: $COUNT"

step "GET /users/id-inexistente — usuario no encontrado"
RESPONSE=$(curl -s -w "\n%{http_code}" "${BASE_URL}/users/00000000-0000-0000-0000-000000000000")
STATUS=$(echo "$RESPONSE" | tail -n 1)

assert_status 404 "$STATUS" "ID inexistente retorna 404"

# =============================================================================
header "3. CATÁLOGO DE PRODUCTOS"
# =============================================================================

step "POST /products — crear laptop"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "${BASE_URL}/products" \
  -H "Content-Type: application/json" \
  -d '{
    "sku": "LAPTOP-001",
    "name": "MacBook Pro 14",
    "price": 1999.99
  }')
BODY=$(echo "$RESPONSE" | sed '$d')
STATUS=$(echo "$RESPONSE" | tail -n 1)

assert_status 201 "$STATUS" "crear producto"
assert_field "$BODY" ".sku" "LAPTOP-001"
assert_field "$BODY" ".name" "MacBook Pro 14"

LAPTOP_ID=$(echo "$BODY" | jq -r ".id")
ok "ID asignado: $LAPTOP_ID"

step "POST /products — crear teclado"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "${BASE_URL}/products" \
  -H "Content-Type: application/json" \
  -d '{
    "sku": "KEYBOARD-001",
    "name": "Keychron K2 Pro",
    "price": 119.99
  }')
BODY=$(echo "$RESPONSE" | sed '$d')
STATUS=$(echo "$RESPONSE" | tail -n 1)

assert_status 201 "$STATUS" "crear segundo producto"
KEYBOARD_ID=$(echo "$BODY" | jq -r ".id")
ok "ID asignado: $KEYBOARD_ID"

step "POST /products — SKU duplicado (debe rechazarse)"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "${BASE_URL}/products" \
  -H "Content-Type: application/json" \
  -d '{
    "sku": "LAPTOP-001",
    "name": "Otro laptop con mismo SKU",
    "price": 999.00
  }')
STATUS=$(echo "$RESPONSE" | tail -n 1)

assert_status 409 "$STATUS" "SKU duplicado rechazado con 409 Conflict"

step "GET /products/${LAPTOP_ID} — buscar laptop por ID"
RESPONSE=$(curl -s -w "\n%{http_code}" "${BASE_URL}/products/${LAPTOP_ID}")
BODY=$(echo "$RESPONSE" | sed '$d')
STATUS=$(echo "$RESPONSE" | tail -n 1)

assert_status 200 "$STATUS" "buscar producto por ID"
assert_field "$BODY" ".sku" "LAPTOP-001"

step "GET /products — listar todos los productos"
RESPONSE=$(curl -s -w "\n%{http_code}" "${BASE_URL}/products")
BODY=$(echo "$RESPONSE" | sed '$d')
STATUS=$(echo "$RESPONSE" | tail -n 1)

assert_status 200 "$STATUS" "listar productos"
PROD_COUNT=$(echo "$BODY" | jq '. | length')
ok "total productos en catálogo: $PROD_COUNT"

# =============================================================================
header "4. OBSERVABILIDAD"
# =============================================================================

step "GET /metrics — métricas Prometheus"
RESPONSE=$(curl -s -w "\n%{http_code}" "${BASE_URL}/metrics")
BODY=$(echo "$RESPONSE" | sed '$d')
STATUS=$(echo "$RESPONSE" | tail -n 1)

assert_status 200 "$STATUS" "endpoint de métricas disponible"

if echo "$BODY" | grep -q "http_requests_total"; then
  ok "métrica http_requests_total presente"
else
  fail "métrica http_requests_total no encontrada"
fi

if echo "$BODY" | grep -q "http_request_duration_seconds"; then
  ok "métrica http_request_duration_seconds presente"
else
  fail "métrica http_request_duration_seconds no encontrada"
fi

# =============================================================================
header "5. RESUMEN"
# =============================================================================

echo ""
echo -e "  Usuarios creados:  Alice ($ALICE_ID)"
echo -e "                     Bob   ($BOB_ID)"
echo -e "  Productos creados: LAPTOP-001   ($LAPTOP_ID)"
echo -e "                     KEYBOARD-001 ($KEYBOARD_ID)"
echo ""

TOTAL=$((PASS + FAIL))
if [ "$FAIL" -eq 0 ]; then
  echo -e "${GREEN}${BOLD}  ✓ Todos los checks pasaron ($PASS/$TOTAL)${RESET}"
  echo ""
  exit 0
else
  echo -e "${RED}${BOLD}  ✗ $FAIL checks fallaron ($PASS/$TOTAL pasaron)${RESET}"
  echo ""
  exit 1
fi
