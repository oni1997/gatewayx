# Quick Start

## Prerequisites

- Go 1.25+ or Docker/Podman

## Build from source

```bash
git clone https://github.com/oni1997/gatewayx.git
cd gatewayx

go build -o bin/gatewayx ./apps/gateway
go build -o bin/gatewayx-cli ./apps/cli

cp gatewayx.example.yaml gatewayx.yaml

./bin/gatewayx
```

## Run with Docker/Podman

```bash
podman pull ghcr.io/oni1997/gatewayx:latest
podman run -d -p 8080:8080 -p 9090:9090 \
  -v $(pwd)/gatewayx.yaml:/etc/gatewayx/gatewayx.yaml \
  ghcr.io/oni1997/gatewayx:latest
```

Then visit:

- `http://localhost:8080/health` — health check
- `http://localhost:9090/metrics` — Prometheus metrics
- `http://localhost:9090/` — Dashboard

## Interactive config generation

```bash
gatewayx-cli init
```
