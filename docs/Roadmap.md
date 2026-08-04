# Roadmap

## Phase 0 -- Research ✅

- [x] Study major API gateways (Traefik, Kong, NGINX, Caddy, Envoy, HAProxy, Krakend)
- [x] Architecture document
- [x] Requirements
- [x] Vision document
- [x] Roadmap

## Phase 1 -- Core Reverse Proxy 🚧

- [x] Go project structure
- [x] Configuration loader (YAML + Viper)
- [x] Reverse proxy engine (net/http + httputil.ReverseProxy)
- [x] Host and path routing
- [x] Round-robin load balancing
- [x] Health check endpoint
- [x] Structured logging (slog)
- [x] HTTPS/TLS support
- [x] Middleware system (recovery, CORS, max body)
- [ ] Comprehensive test suite
- [ ] Docker image
- [ ] Binary releases via GoReleaser

## Phase 2 -- Authentication

- [ ] JWT authentication
- [ ] API key authentication
- [ ] Basic auth
- [ ] OAuth 2.0 integration
- [ ] mTLS support
- [ ] HMAC signing
- [ ] RBAC / permissions
- [ ] Session management
- [ ] Token caching

## Phase 3 -- Rate Limiting

- [ ] Per-user rate limiting
- [ ] Per-IP rate limiting
- [ ] Per-API key rate limiting
- [ ] Burst support
- [ ] Sliding window algorithm
- [ ] Token bucket algorithm
- [ ] Redis-backed distributed rate limiting
- [ ] In-memory rate limiting

## Phase 4 -- Observability

- [ ] Prometheus metrics
- [ ] OpenTelemetry tracing
- [ ] Structured log aggregation
- [ ] Grafana dashboard templates
- [ ] Request history
- [ ] Error analytics

## Phase 5 -- Dashboard

- [ ] React + TypeScript frontend
- [ ] Tailwind CSS styling
- [ ] Admin REST API
- [ ] Service management UI
- [ ] User management
- [ ] API key management
- [ ] Live metrics
- [ ] Log viewer
- [ ] Plugin management
- [ ] Settings panel

## Phase 6 -- Plugin SDK

- [ ] Plugin interface definition
- [ ] Lifecycle hooks
- [ ] Event system
- [ ] Plugin configuration
- [ ] Plugin marketplace
- [ ] Authentication plugins
- [ ] Rate limiting plugins
- [ ] Monitoring plugins
- [ ] Webhook plugins
- [ ] AI integration plugins

## Phase 7 -- Kubernetes

- [ ] Helm chart
- [ ] Kubernetes operator
- [ ] Ingress controller
- [ ] Auto-scaling
- [ ] Secret management
- [ ] Certificate management (cert-manager integration)

## Phase 8 -- AI Assistant

- [ ] Attack detection
- [ ] Bottleneck analysis
- [ ] Log summarization
- [ ] Cache recommendations
- [ ] Rate limit recommendations
- [ ] Configuration generation
- [ ] Error explanation

## Phase 9 -- Enterprise Features

- [ ] LDAP integration
- [ ] SAML support
- [ ] Audit logging
- [ ] High-availability clustering
- [ ] Backup and restore
- [ ] Multi-tenant support
- [ ] Billing integration
- [ ] License management
