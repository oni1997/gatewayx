# Changelog

All notable changes to GatewayX will be documented in this file.

## [Unreleased]

### Added
- ML analysis engine: security scanner, bottleneck finder, rate limit recommender
- Container deployment section in README with Docker and Podman instructions

### Changed
- Package `internal/ai` renamed to `internal/ml` (machine learning, not AI)
- Optimized README images: logo (73KB), architecture (495KB)

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

[Unreleased]: https://github.com/oni1997/gatewayx/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/oni1997/gatewayx/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/oni1997/gatewayx/releases/tag/v0.1.0
