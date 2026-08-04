# Architecture

## High-Level Overview

```
                    Dashboard
                         │
                         ▼
                  Admin REST API
                         │
         ┌───────────────┼───────────────┐
         │               │               │
         ▼               ▼               ▼
   Configuration      Plugin Host     Event Bus
         │               │               │
         └───────────────┼───────────────┘
                         ▼
                   Middleware Pipeline
     ┌──────────────┬──────────────┬──────────────┐
     ▼              ▼              ▼
Authentication  Rate Limiting   Logging/Metrics
     └──────────────┬──────────────┘
                    ▼
             Reverse Proxy Engine
                    ▼
             Backend Services
```

## Component Design

### Gateway Core (`apps/gateway`)

The main process that runs the reverse proxy server. It initializes configuration, sets up routes, starts the HTTP server, and handles graceful shutdown.

### Reverse Proxy Engine (`internal/proxy`)

Wraps Go's `net/http/httputil.ReverseProxy` with:
- Per-route proxy configuration
- Load balancer integration (round-robin, weighted)
- Request/response header manipulation
- Timeout management
- Structured request logging

### Router (`internal/router`)

Host- and path-based request routing. Matches incoming requests against configured routes and directs traffic to the appropriate upstream handlers.

### Middleware Pipeline (`internal/middleware`)

Composable middleware chain:
- **Recovery** — Panic recovery
- **CORS** — Cross-origin resource sharing
- **MaxBodySize** — Request body size limiting
- **Headers** — Response header injection

### Configuration (`internal/config`)

YAML-based declarative configuration loaded via Viper. Supports:
- File-based configuration
- Environment variable overrides (`GATEWAYX_` prefix)
- Default values for all settings
- Validation at startup

### Logger (`internal/logger`)

Structured logging via Go's `log/slog`. Supports JSON and text formats, configurable levels, and file or stdout output.

### Health Checks (`internal/health`)

Active health checking system. Each service can register a health check function. The `/health` endpoint returns JSON with per-service status.

### Load Balancer (`pkg/loadbalancer`)

Pluggable load balancing strategies:
- Round-robin
- Weighted round-robin

### CLI (`apps/cli`)

Command-line tool built with Cobra:
- `gatewayx serve` — Start the gateway
- `gatewayx validate` — Validate configuration
- `gatewayx version` — Print version info

## Data Flow

```
1. Request arrives at GatewayX listener
2. Router matches request to configured route
3. Middleware pipeline executes (recovery, CORS, body limit)
4. Route handler executes (load balancing, proxy, timeouts)
5. Request forwarded to upstream backend
6. Response returned through middleware chain
7. Structured log entry emitted
```

## Technology Choices

| Decision | Rationale |
|----------|-----------|
| Go | High performance, excellent standard library, simple deployment |
| net/http + ReverseProxy | Battle-tested standard library; no external HTTP framework needed |
| Cobra + Viper | Industry standard for Go CLIs and configuration |
| slog | Standard library structured logging (Go 1.21+) |
| YAML configuration | Human-readable, widely supported, easy to version |
| SQLite (future) | Zero-dependency storage for single-node deployments |
| Redis (future) | Distributed rate limiting and caching |
