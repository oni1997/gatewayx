# Configuration

GatewayX is configured via a YAML file. The default location is `./gatewayx.yaml`, or you can specify a path with `-c` (CLI) or `GATEWAYX_CONFIG` (env).

## Full Configuration Reference

```yaml
server:
  host: "0.0.0.0"           # Listen address
  port: 8080                # Listen port
  read_timeout: 30s         # Maximum duration for reading the entire request
  write_timeout: 30s        # Maximum duration before timing out writes
  idle_timeout: 120s        # Maximum idle connection duration
  max_header_bytes: 1048576 # Maximum header size (1MB)
  shutdown_timeout: 10s     # Graceful shutdown timeout

admin:
  token: "${ADMIN_TOKEN}"   # If set, /api/* endpoints require this token

oauth:
  provider: "github"        # github or google
  client_id: "${OAUTH_CLIENT_ID}"
  client_secret: "${OAUTH_CLIENT_SECRET}"
  redirect_url: "https://gateway.example.com:9090/oauth/callback"

routes:
  - name: "my-api"          # Route identifier
    listen_path: "/api"     # Path prefix to match
    upstream_urls:          # Backend service URLs
      - "http://localhost:3000"
    methods:                # Allowed HTTP methods (empty = all)
      - GET
      - POST
    hosts:                  # Hostnames to match (empty = all)
      - "api.example.com"
    strip_path: false       # Strip listen_path before forwarding
    preserve_host: false    # Preserve original Host header
    headers:                # Headers to add to responses
      X-Frame-Options: "DENY"
    timeout: 30s            # Per-route timeout
    retry_count: 0          # Number of retries on failure
    load_balancing: "round_robin"  # round_robin or weighted
    compression: false      # Enable gzip compression
    websocket: false        # Enable WebSocket upgrade support
    cache:                  # Response caching (GET/HEAD only)
      ttl: 30s
      max_size: 1000        # Max cached entries
    health_check:           # Active health checks + backend draining
      path: "/health"
      interval: 10s
      timeout: 5s
      healthy: 3
      unhealthy: 3
    authentication:         # Authentication configuration
      type: "jwt"           # jwt, api_key, basic, hmac, oauth, mtls, rbac, session
      options:
        secret: "${JWT_SECRET}"
        algorithm: "HS256"
        cache_ttl: 5m       # JWT token cache
    rate_limit:             # Rate limiting
      rate: 100
      burst: 200
      strategy: "token_bucket"  # token_bucket or sliding_window
      per_ip: false
      per_user: false
      per_key: false
      redis_addr: ""        # Redis for distributed rate limiting

logging:
  level: "info"            # debug, info, warn, error
  format: "json"           # json or text
  output: "stdout"         # stdout or file
  file: "gatewayx.log"     # File path when output is "file"

metrics:
  enabled: true
  port: 9090
  path: "/metrics"
  tracing: false           # Distributed tracing
  history: 1000            # Request history buffer size

plugins:
  dir: "./plugins"
  enabled: []

tls:
  enabled: false
  cert_file: "/path/to/cert.pem"
  key_file: "/path/to/key.pem"
  min_version: "1.2"
  auto_cert: false         # Let's Encrypt
  auto_cert_dir: "/var/lib/gatewayx/certs"

health:
  enabled: true
  path: "/health"

security:
  max_body_size: 10485760   # 10MB
  allowed_hosts: []
  trusted_proxies: []
```

## Authentication Types

| Type | Description | Key options |
|------|-------------|-------------|
| `jwt` | JWT bearer tokens | `secret`, `secret_file`, `public_key_file`, `algorithm`, `cache_ttl` |
| `api_key` | API key header | `keys_file`, `header` |
| `basic` | HTTP Basic Auth | `htpasswd_file`, `realm` |
| `hmac` | HMAC signatures | `secret`, `algorithm`, `header`, `clock_skew` |
| `oauth` | OAuth 2.0 (GitHub/Google) | `provider`, `client_id`, `client_secret` |
| `mtls` | Client certificates | `ca_cert`, `verify_depth` |
| `rbac` | Role-based access (wraps another auth) | `delegate`, `roles_claim`, `perm_*` |
| `session` | Session tokens | `ttl`, `max_sessions` |

## Environment Variables

All configuration values can be overridden with environment variables using the `GATEWAYX_` prefix:

```bash
export GATEWAYX_SERVER_PORT=9090
export GATEWAYX_LOGGING_LEVEL=debug
export GATEWAYX_TLS_ENABLED=true
```

Additional environment variables:

| Variable | Description |
|----------|-------------|
| `GATEWAYX_CONFIG` | Config file path |
| `GATEWAYX_DB_PATH` | SQLite database path for persistence |
| `GATEWAYX_WEBHOOK_URL` | Slack/Discord webhook for alerts |
| `GATEWAYX_DASHBOARD_PATH` | Dashboard static files path |

## Hot Reload

Edit the config file — GatewayX auto-reloads via file watching. Or send `SIGHUP`:

```bash
kill -HUP $(pgrep gatewayx)
```

## Validation

```bash
gatewayx-cli validate -c gatewayx.yaml
```

Shows errors and warnings (missing auth, rate limiting, timeouts, health checks).

## Generate a Config Interactively

```bash
gatewayx-cli init
```
