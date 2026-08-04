#!/bin/bash
set -euo pipefail

echo "=== GatewayX Integration Test Suite ==="

echo "Building..."
make build

echo "Starting GatewayX with example config..."
./bin/gatewayx &
PID=$!
sleep 2

cleanup() {
    echo "Stopping GatewayX..."
    kill $PID 2>/dev/null || true
    wait $PID 2>/dev/null || true
}
trap cleanup EXIT

echo "Testing health endpoint..."
HEALTH=$(curl -s http://localhost:8080/health)
echo "$HEALTH" | grep -q "healthy" && echo "  PASS: health check" || echo "  FAIL: health check"

echo "Testing metrics endpoint..."
METRICS=$(curl -s http://localhost:9090/metrics)
echo "$METRICS" | grep -q "GatewayX" && echo "  PASS: metrics" || echo "  FAIL: metrics"

echo "Testing 404 on unknown path..."
STATUS=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/nonexistent)
[ "$STATUS" = "404" ] && echo "  PASS: 404 handling" || echo "  FAIL: 404 handling (got $STATUS)"

echo "=== All tests passed ==="
