# Introduction

GatewayX is a high-performance, extensible API gateway built in Go. It serves as the foundation for a suite of developer infrastructure tools — think Traefik and Kong, but built around the developer experience first.

## Philosophy

- **Configuration over code** — Solve 95% of problems through configuration
- **Sane defaults** — Works in production without hours of tuning
- **Observability first** — Metrics, logs, and traces from day one
- **Performance** — Built in Go, optimized for high throughput
- **Security** — TLS, authentication, and rate limiting in the core

## Features

- Reverse proxy with host/path routing
- JWT, API keys, Basic Auth, HMAC, OAuth 2.0, mTLS
- Rate limiting (token bucket, sliding window)
- Response caching
- WebSocket support
- Circuit breaker
- Health-check-driven backend draining
- Hot config reload
- SQLite persistence
- Prometheus metrics
- ML analysis (attack detection, bottleneck finder)
- React dashboard
- Plugin SDK
