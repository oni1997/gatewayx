#!/bin/bash
set -euo pipefail

echo "=== GatewayX E2E Test Suite ==="
echo ""

cleanup() {
    echo "Cleaning up..."
    podman stop gatewayx-e2e 2>/dev/null || true
    podman rm gatewayx-e2e 2>/dev/null || true
    podman network rm gatewayx-test 2>/dev/null || true
}

trap cleanup EXIT

echo "BUILD: Building Go binary..."
go build -o bin/gatewayx-e2e ./apps/gateway

echo "PREP: Creating test network..."
podman network create gatewayx-test

echo "PREP: Starting test backend (nginx)..."
podman run -d --name test-backend --network gatewayx-test nginx:alpine
sleep 2

BACKEND_IP=$(podman inspect -f '{{.NetworkSettings.Networks.gatewayx-test.IPAddress}}' test-backend)
echo "  Backend IP: $BACKEND_IP"

echo "PREP: Creating test config..."
cat > /tmp/gatewayx-e2e.yaml <<EOF
server:
  host: "0.0.0.0"
  port: 8080
routes:
  - name: "test-route"
    listen_path: "/"
    upstream_urls:
      - "http://${BACKEND_IP}:80"
    methods:
      - GET
    timeout: 5s
logging:
  level: "error"
  format: "text"
metrics:
  enabled: false
health:
  enabled: true
  path: "/health"
plugins:
  enabled: []
EOF

echo "START: Running GatewayX..."
GATEWAYX_CONFIG=/tmp/gatewayx-e2e.yaml ./bin/gatewayx-e2e &
GW_PID=$!
sleep 2

echo ""
echo "--- TESTS ---"

echo "TEST: Health check..."
HEALTH=$(curl -sf http://localhost:8080/health)
if echo "$HEALTH" | grep -q "healthy"; then
    echo "  PASS: Health endpoint returns healthy"
else
    echo "  FAIL: Health endpoint failed"
    kill $GW_PID 2>/dev/null || true
    exit 1
fi

echo "TEST: Proxy to backend..."
RESPONSE=$(curl -sf http://localhost:8080/ -o /dev/null -w "%{http_code}")
if [ "$RESPONSE" = "200" ]; then
    echo "  PASS: Proxy returned 200 from backend"
else
    echo "  FAIL: Expected 200, got $RESPONSE"
    kill $GW_PID 2>/dev/null || true
    exit 1
fi

echo "TEST: Backend content..."
BODY=$(curl -sf http://localhost:8080/)
if echo "$BODY" | grep -q "nginx"; then
    echo "  PASS: Backend returned nginx welcome page"
else
    echo "  FAIL: Unexpected backend response"
    kill $GW_PID 2>/dev/null || true
    exit 1
fi

echo "TEST: 404 on unknown path..."
STATUS=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/nonexistent/path)
if [ "$STATUS" = "404" ]; then
    echo "  PASS: Unknown path returns 404"
else
    echo "  FAIL: Expected 404, got $STATUS"
    kill $GW_PID 2>/dev/null || true
    exit 1
fi

echo ""
echo "=== All E2E tests passed ==="

kill $GW_PID 2>/dev/null || true
wait $GW_PID 2>/dev/null || true
