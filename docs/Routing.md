# Routing

GatewayX routes requests to upstream services based on hostname, path, and HTTP method.

## Basic Route

```yaml
routes:
  - name: "my-service"
    listen_path: "/api"
    upstream_urls:
      - "http://backend:3000"
```

All requests to `/api/*` will be forwarded to `http://backend:3000`.

## Path Stripping

```yaml
routes:
  - name: "api"
    listen_path: "/api"
    strip_path: true
    upstream_urls:
      - "http://backend:3000"
```

A request to `/api/users` is forwarded as `/users`.

## Host-Based Routing

```yaml
routes:
  - name: "api-v1"
    listen_path: "/api"
    hosts:
      - "api.example.com"
    upstream_urls:
      - "http://api-v1:3000"

  - name: "api-v2"
    listen_path: "/api"
    hosts:
      - "api-v2.example.com"
    upstream_urls:
      - "http://api-v2:3000"
```

## Method Filtering

```yaml
routes:
  - name: "read-only"
    listen_path: "/api"
    methods:
      - GET
      - HEAD
    upstream_urls:
      - "http://read-replica:3000"

  - name: "write"
    listen_path: "/api"
    methods:
      - POST
      - PUT
      - DELETE
    upstream_urls:
      - "http://primary:3000"
```

## Load Balancing

```yaml
routes:
  - name: "balanced"
    listen_path: "/api"
    upstream_urls:
      - "http://backend-1:3000"
      - "http://backend-2:3000"
      - "http://backend-3:3000"
    load_balancing: "round_robin"

  - name: "weighted"
    listen_path: "/api"
    upstream_urls:
      - "http://large-instance:3000"
      - "http://small-instance:3000"
    load_balancing: "weighted"
    weights: [3, 1]
```

## Preserving Host Header

```yaml
routes:
  - name: "preserve"
    listen_path: "/api"
    preserve_host: true
    upstream_urls:
      - "http://backend:3000"
```

The original `Host` header is forwarded to the upstream.

## Retry on Failure

```yaml
routes:
  - name: "retry"
    listen_path: "/api"
    retry_count: 3
    upstream_urls:
      - "http://backend:3000"
```

## Timeouts

```yaml
routes:
  - name: "slow-service"
    listen_path: "/slow"
    timeout: 60s
    upstream_urls:
      - "http://slow-backend:3000"
```

## Full Route Configuration

```yaml
routes:
  - name: "production-api"
    listen_path: "/api/v1"
    hosts:
      - "api.example.com"
    methods:
      - GET
      - POST
      - PUT
      - DELETE
    upstream_urls:
      - "http://backend-1:3000"
      - "http://backend-2:3000"
    strip_path: false
    preserve_host: true
    load_balancing: "round_robin"
    timeout: 30s
    retry_count: 2
    compression: true
    headers:
      X-Frame-Options: "DENY"
      X-Content-Type-Options: "nosniff"
    health_check:
      path: "/health"
      interval: 10s
      timeout: 5s
      healthy: 3
      unhealthy: 3
    authentication:
      type: "jwt"
      options:
        secret: "${JWT_SECRET}"
    rate_limit:
      rate: 1000
      burst: 2000
      strategy: "token_bucket"
```
