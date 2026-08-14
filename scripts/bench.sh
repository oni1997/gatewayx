#!/bin/bash
set -euo pipefail

# GatewayX benchmark script
# Spins up a test backend, runs the gateway, and load-tests with hey or wrk.

CONCURRENCY=${CONCURRENCY:-50}
DURATION=${DURATION:-10s}
REQUESTS=${REQUESTS:-100000}

BENCH_TOOL=""
if command -v hey &>/dev/null; then
  BENCH_TOOL="hey"
elif command -v wrk &>/dev/null; then
  BENCH_TOOL="wrk"
else
  echo "Error: no load testing tool found."
  echo "Install one of:"
  echo "  hey:  go install github.com/rakyll/hey@latest"
  echo "  wrk:  apt install wrk / brew install wrk"
  exit 1
fi

cleanup() {
  kill $GW_PID $BACKEND_PID 2>/dev/null || true
  wait $GW_PID $BACKEND_PID 2>/dev/null || true
}
trap cleanup EXIT

echo "=== GatewayX Benchmark ==="
echo "Tool: $BENCH_TOOL | Concurrency: $CONCURRENCY | Duration: $DURATION"

echo ""
echo "1. Building gateway..."
go build -o /tmp/gatewayx-bench ./apps/gateway

echo "2. Starting test backend (Go net/http)..."
cat > /tmp/gatewayx-backend.go <<'EOF'
package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"message":"hello from backend","path":"%s"}`, r.URL.Path)
	})
	http.ListenAndServe(":18080", nil)
}
EOF
go run /tmp/gatewayx-backend.go &
BACKEND_PID=$!
sleep 1

echo "3. Starting gateway..."
cat > /tmp/gatewayx-bench.yaml <<EOF
server:
  host: "0.0.0.0"
  port: 18081
routes:
  - name: "bench"
    listen_path: "/"
    upstream_urls: ["http://localhost:18080"]
logging:
  level: "error"
metrics:
  enabled: false
health:
  enabled: false
EOF
GATEWAYX_CONFIG=/tmp/gatewayx-bench.yaml /tmp/gatewayx-bench &
GW_PID=$!
sleep 1

echo "4. Warmup..."
curl -sf http://localhost:18081/ >/dev/null

echo "5. Running load test against gateway (port 18081)..."
if [ "$BENCH_TOOL" = "hey" ]; then
  hey -z "$DURATION" -c "$CONCURRENCY" http://localhost:18081/api/bench
else
  wrk -t4 -c "$CONCURRENCY" -d "$DURATION" http://localhost:18081/api/bench
fi

echo ""
echo "6. Direct-to-backend baseline (port 18080)..."
if [ "$BENCH_TOOL" = "hey" ]; then
  hey -z "$DURATION" -c "$CONCURRENCY" http://localhost:18080/api/bench
else
  wrk -t4 -c "$CONCURRENCY" -d "$DURATION" http://localhost:18080/api/bench
fi

echo ""
echo "=== Benchmark complete ==="
