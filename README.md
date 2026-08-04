# GatewayX

**Developer Infrastructure Platform**

GatewayX is a high-performance, extensible API gateway built in Go. It serves as the foundation for a suite of developer infrastructure tools — think Traefik and Kong, but built around the developer experience first.

## Features

- **Reverse Proxy** — HTTP/HTTPS forwarding with load balancing
- **Host & Path Routing** — Route traffic by hostname and URL path
- **Load Balancing** — Round-robin and weighted round-robin
- **Health Checks** — Active upstream health monitoring
- **TLS Support** — HTTPS with certificate file or auto-cert
- **Metrics** — Prometheus-format metrics endpoint
- **Structured Logging** — JSON or text logging with configurable levels
- **Configuration File** — YAML-based declarative configuration
- **CLI Tool** — Validate, serve, and manage from the command line
- **Extensible** — Plugin system (coming in Phase 6)

## Quick Start

```bash
go build -o bin/gatewayx ./apps/gateway
go build -o bin/gatewayx-cli ./apps/cli

cp gatewayx.example.yaml gatewayx.yaml

./bin/gatewayx
```

## Documentation

- [Vision](docs/Vision.md)
- [Architecture](docs/Architecture.md)
- [Getting Started](docs/GettingStarted.md)
- [Installation](docs/Installation.md)
- [Configuration](docs/Configuration.md)
- [Authentication](docs/Authentication.md)
- [Routing](docs/Routing.md)
- [Plugins](docs/Plugins.md)
- [Dashboard](docs/Dashboard.md)
- [Security](docs/Security.md)
- [API Reference](docs/APIReference.md)
- [Contributing](docs/Contributing.md)
- [Roadmap](docs/Roadmap.md)
- [ADR](docs/ADR/)

## Tech Stack

| Component | Technology |
|-----------|-----------|
| Gateway | Go |
| HTTP | net/http + httputil.ReverseProxy |
| CLI | Cobra |
| Config | Viper |
| Logging | slog |
| Metrics | Prometheus |
| Tracing | OpenTelemetry |
| Cache | Redis |
| Dashboard | React + TypeScript + Tailwind |
| Build | GoReleaser |
| Container | Docker |
| Orchestration | Kubernetes |
| CI/CD | GitHub Actions |

## Project Structure

```
apps/       — Application entry points (gateway, cli, dashboard)
internal/   — Internal packages (config, proxy, router, middleware, health, logger)
pkg/        — Shared packages (loadbalancer, compression)
plugins/    — Plugin system
examples/   — Example configurations
docs/       — Documentation
deploy/     — Deployment manifests (Docker, Kubernetes, Helm)
sdk/        — Plugin SDK
tests/      — Test suites
website/    — Project website
```

## License

MIT
