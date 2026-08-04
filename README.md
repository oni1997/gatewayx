<p align="center">
  <img src="logo.png" alt="GatewayX" width="200">
</p>

<h1 align="center">GatewayX</h1>
<p align="center"><strong>Developer Infrastructure Platform</strong></p>

<p align="center">
  <a href="https://github.com/oni1997/gatewayx/actions"><img src="https://img.shields.io/github/actions/workflow/status/oni1997/gatewayx/ci.yml" alt="CI"></a>
  <a href="https://github.com/oni1997/gatewayx/blob/main/LICENSE"><img src="https://img.shields.io/github/license/oni1997/gatewayx" alt="License"></a>
  <a href="https://go.dev"><img src="https://img.shields.io/github/go-mod/go-version/oni1997/gatewayx" alt="Go Version"></a>
</p>

<p align="center">
  <img src="Architecture.png" alt="GatewayX Architecture" width="800">
</p>

---

GatewayX is a high-performance, extensible API gateway built in Go. It serves as the foundation for a suite of developer infrastructure tools -- think Traefik and Kong, but built around the developer experience first.

## Features

- **Reverse Proxy** -- HTTP/HTTPS forwarding with load balancing
- **Host & Path Routing** -- Route traffic by hostname and URL path
- **Load Balancing** -- Round-robin and weighted round-robin
- **Authentication** -- JWT (HS256/RS256/ES256), API keys, Basic Auth, HMAC, OAuth-ready
- **RBAC** -- Role-based access control with path glob matching
- **Session Management** -- In-memory TTL sessions with cookie/header support
- **Rate Limiting** -- Token bucket and sliding window, per-IP, per-user, per-key
- **Health Checks** -- Active upstream health monitoring with JSON endpoint
- **TLS Support** -- HTTPS with certificate file or auto-cert (Let's Encrypt)
- **Metrics** -- Prometheus-format metrics endpoint
- **Structured Logging** -- JSON or text logging via slog with configurable levels
- **Configuration File** -- YAML-based declarative configuration with env var overrides
- **CLI Tool** -- Start, validate, and manage from the command line (Cobra)
- **Docker Ready** -- Multi-stage Dockerfile and docker-compose included
- **Token Caching** -- In-memory JWT cache to reduce validation overhead
- **Extensible** -- Plugin system (Phase 6)

## Quick Start

```bash
git clone https://github.com/oni1997/gatewayx.git
cd gatewayx

go build -o bin/gatewayx ./apps/gateway
go build -o bin/gatewayx-cli ./apps/cli

cp gatewayx.example.yaml gatewayx.yaml

./bin/gatewayx
```

## Example Config

```yaml
server:
  host: "0.0.0.0"
  port: 8080

routes:
  - name: "public-api"
    listen_path: "/api"
    upstream_urls: ["http://backend:3000"]
    rate_limit:
      rate: 100
      burst: 200
      per_ip: true

  - name: "admin-panel"
    listen_path: "/admin"
    upstream_urls: ["http://admin-svc:4000"]
    authentication:
      type: "jwt"
      options:
        secret: "${JWT_SECRET}"
        algorithm: "HS256"
    rate_limit:
      rate: 10
      per_user: true

logging:
  level: "info"
  format: "json"

metrics:
  enabled: true
  port: 9090

health:
  enabled: true
  path: "/health"
```

## Documentation

| Doc | |
|-----|---------|
| [Vision](docs/Vision.md) | Project goals and philosophy |
| [Architecture](docs/Architecture.md) | System design and data flow |
| [Getting Started](docs/GettingStarted.md) | Build and run in 5 minutes |
| [Installation](docs/Installation.md) | Docker, binary, source |
| [Configuration](docs/Configuration.md) | Full YAML reference |
| [Authentication](docs/Authentication.md) | JWT, API keys, Basic, HMAC, RBAC |
| [Routing](docs/Routing.md) | Host, path, methods, load balancing |
| [Plugins](docs/Plugins.md) | Plugin system design |
| [Dashboard](docs/Dashboard.md) | React dashboard (Phase 5) |
| [Security](docs/Security.md) | TLS, mTLS, headers, best practices |
| [API Reference](docs/APIReference.md) | Admin REST API |
| [Contributing](docs/Contributing.md) | How to contribute |
| [Roadmap](docs/Roadmap.md) | Phases 0-9 |
| [ADR](docs/ADR/) | Architecture decisions |

## Tech Stack

| Component | Technology |
|-----------|-----------|
| Language | Go |
| HTTP Engine | net/http + httputil.ReverseProxy |
| CLI | Cobra |
| Config | Viper (YAML + env vars) |
| Logging | slog |
| Auth | golang-jwt/jwt |
| Metrics | Prometheus |
| Tracing | OpenTelemetry (Planned) |
| Cache | Redis (optional) |
| Dashboard | React + TypeScript + Tailwind (Planned) |
| Build | GoReleaser |
| Container | Docker |
| CI/CD | GitHub Actions |

## Project Structure

```
apps/       -- Application entry points (gateway, cli, dashboard)
internal/   -- Internal packages (auth, config, proxy, ratelimit, middleware, health, logger)
pkg/        -- Shared packages (loadbalancer, compression)
plugins/    -- Plugin system
examples/   -- Example configurations
docs/       -- Full documentation suite
deploy/     -- Deployment manifests
sdk/        -- Plugin SDK
tests/      -- Unit and integration tests
website/    -- Project website
```

## Running on a Raspberry Pi

```bash
GOOS=linux GOARCH=arm64 go build -o bin/gatewayx ./apps/gateway
./bin/gatewayx
```

Runs comfortably on a Pi 4 with 50-100MB RAM.

## License

MIT
