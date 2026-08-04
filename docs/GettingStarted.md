# Getting Started

## Prerequisites

- Go 1.21 or later
- Docker (optional, for containerized deployment)

## Build from Source

```bash
git clone https://github.com/gatewayx/gatewayx.git
cd gatewayx

go build -o bin/gatewayx ./apps/gateway
go build -o bin/gatewayx-cli ./apps/cli
```

## Create a Configuration

```bash
cp gatewayx.example.yaml gatewayx.yaml
```

Edit `gatewayx.yaml` to define your routes:

```yaml
server:
  host: "0.0.0.0"
  port: 8080

routes:
  - name: "api"
    listen_path: "/api"
    upstream_urls:
      - "http://localhost:3000"
    methods:
      - GET
      - POST
    strip_path: false

logging:
  level: "info"
  format: "json"
```

## Run the Gateway

```bash
./bin/gatewayx
```

The gateway starts on port 8080 and proxies requests to your configured upstreams.

## Test It

```bash
curl http://localhost:8080/api/health
curl http://localhost:8080/health
```

## Docker

```bash
docker build -t gatewayx .
docker run -p 8080:8080 -v $(pwd)/gatewayx.yaml:/etc/gatewayx/gatewayx.yaml gatewayx
```

## Next Steps

- Read the [Configuration](Configuration.md) guide
- Set up [Authentication](Authentication.md)
- Configure [Routing](Routing.md) rules
- Explore the [Plugin](Plugins.md) system
