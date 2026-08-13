# Metrics

GatewayX exposes Prometheus-format metrics on port 9090.

```bash
curl http://localhost:9090/metrics
```

## Available metrics

- `gatewayx_requests_total` — total requests
- `gatewayx_active_connections` — current connections
- `gatewayx_bytes_sent_total` — bytes sent
- `gatewayx_bytes_received_total` — bytes received
- `gatewayx_route_requests_total` — per-route request count
- `gatewayx_route_latency_avg_ms` — per-route latency
- `gatewayx_memory_alloc_bytes` — memory allocation
- `go_goroutines` — goroutine count

## Grafana

Import `deploy/grafana/gatewayx-dashboard.json` into Grafana for a pre-built 9-panel dashboard.
