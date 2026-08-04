# Configuration

GatewayX is configured via a YAML file. The default location is `./gatewayx.yaml`, or you can specify a path with `-c`.

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
    load_balancing: "round_robin"  # Load balancing strategy
    compression: false      # Enable gzip compression
    health_check:           # Active health checks
      path: "/health"
      interval: 10s
      timeout: 5s
      healthy: 3
      unhealthy: 3
    authentication:         # Authentication configuration
      type: "jwt"
      options:
        secret: "${JWT_SECRET}"
    rate_limit:             # Rate limiting
      rate: 100
      burst: 200
      strategy: "token_bucket"

logging:
  level: "info"            # debug, info, warn, error
  format: "json"           # json or text
  output: "stdout"         # stdout or file
  file: "gatewayx.log"     # File path when output is "file"

metrics:
  enabled: true
  port: 9090
  path: "/metrics"

plugins:
  dir: "./plugins"
  enabled: []

tls:
  enabled: false
  cert_file: "/path/to/cert.pem"
  key_file: "/path/to/key.pem"
  min_version: "1.2"
  auto_cert: false

health:
  enabled: true
  path: "/health"

security:
  max_body_size: 10485760   # 10MB
  allowed_hosts: []
  trusted_proxies: []
```

## Environment Variables

All configuration values can be overridden with environment variables using the `GATEWAYX_` prefix:

```bash
export GATEWAYX_SERVER_PORT=9090
export GATEWAYX_LOGGING_LEVEL=debug
export GATEWAYX_TLS_ENABLED=true
```

## Validation

```bash
gatewayx-cli validate -c gatewayx.yaml
```
