# Changelog

All notable changes to GatewayX will be documented in this file.

## [0.4.4] — 2026-08-14

### Fixed
- `gatewayx-cli serve` now actually starts the gateway (was a stub)
- Extracted shared server logic into `internal/server` package

### Added
- Admin API authentication (`admin.token` config)

## [0.4.0] — 2026-08-05

### Added
- Response caching middleware (in-memory TTL, X-Cache headers)
- WebSocket proxy support (connection upgrade)
- `gatewayx init` interactive CLI config generator
- SQLite persistence for API keys and certificates (survives restart)
- OAuth 2.0 end-to-end flow (login, callback, logout) for GitHub/Google
- Health-check-driven backend draining (health-aware load balancer)
- Config hot-reload via file watching (auto-reload on change)
- Compression middleware wired into proxy
- `gatewayx-cli` binary included in Docker image
- Per-key rate limiting integrated with admin API keys

### Changed
- Load balancer: added health-aware round robin
- README: comprehensive feature list, API endpoints, env vars

## [0.3.1] — 2026-08-05

### Fixed
- Dashboard SPA routing (refresh on sub-pages no longer 404s)
- Dashboard API Keys page now fetches real data from backend
- Dashboard Certificates page now fetches real data from backend
- Dashboard version badge shows v0.3.0

## [0.3.0] — 2026-08-05

### Added
- Hot reload: edit gatewayx.yaml and send SIGHUP to reload without restart
- Circuit breaker: automatic failure detection (5 failures → 30s open → half-open → 3 successes → closed)
- Grafana dashboard template: 9-panel JSON dashboard for Prometheus
- Webhook alerts: Slack/Discord alerts for rate limit spikes, security threats, backend failures

### Fixed
- Double mutex lock in proxy route builder causing deadlock

## [0.2.0] — 2026-08-05

### Added
- OAuth 2.0 authentication (GitHub and Google providers)
- mTLS authentication with client certificate validation
- Admin REST API for dashboard (API keys CRUD, certificate management, config)
- E2E integration test with Podman (build, health, proxy, 404)
- Rate limit + cache recommendations via ML analysis engine

### Changed
- CI: bumped all action versions to kill Node 20 deprecation warnings
- Dashboard API Keys, Certs, Settings pages now wired to real backend

## [0.1.0] — 2026-08-05

### Added
- Core reverse proxy with host/path routing and load balancing
- JWT authentication (HS256, RS256, ES256) with token caching
- API key authentication (inline and file-based)
- Basic Auth with htpasswd support and SHA hashing
- HMAC signature validation (SHA256/SHA512) with anti-replay
- RBAC engine with permission matching and path glob support
- Session management with in-memory TTL store
- Token bucket and sliding window rate limiting
- Per-IP, per-user, per-key rate limit modes
- Prometheus metrics exporter with route-level labels
- OpenTelemetry-compatible distributed tracing
- Request history ring buffer with JSON endpoint
- React dashboard (8 pages: Home, Services, Metrics, History, API Keys, Certs, Health, Settings)
- Plugin SDK with lifecycle hooks and event system
- Example authentication plugin
- Helm chart with deployment, service, ingress, HPA, configmap, secrets
- Raw Kubernetes manifests
- GitHub Actions CI pipeline (lint, test, build, Docker push to GHCR)
- 99 tests across all packages
- Full documentation suite with ADRs
- Docker and Podman deployment support

[Unreleased]: https://github.com/oni1997/gatewayx/compare/v0.4.4...HEAD
[0.4.4]: https://github.com/oni1997/gatewayx/compare/v0.4.3...v0.4.4
[0.4.3]: https://github.com/oni1997/gatewayx/compare/v0.4.2...v0.4.3
[0.4.2]: https://github.com/oni1997/gatewayx/compare/v0.4.1...v0.4.2
[0.4.1]: https://github.com/oni1997/gatewayx/compare/v0.4.0...v0.4.1
[0.4.0]: https://github.com/oni1997/gatewayx/compare/v0.3.1...v0.4.0
[0.3.1]: https://github.com/oni1997/gatewayx/compare/v0.3.0...v0.3.1
[0.3.0]: https://github.com/oni1997/gatewayx/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/oni1997/gatewayx/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/oni1997/gatewayx/releases/tag/v0.1.0
