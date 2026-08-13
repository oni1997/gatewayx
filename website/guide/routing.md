# Routing

## Basic route

```yaml
routes:
  - name: "my-service"
    listen_path: "/api"
    upstream_urls:
      - "http://backend:3000"
```

## Path stripping

```yaml
routes:
  - name: "api"
    listen_path: "/api"
    strip_path: true
    upstream_urls: ["http://backend:3000"]
```

A request to `/api/users` is forwarded as `/users`.

## Host-based routing

```yaml
routes:
  - name: "api-v1"
    listen_path: "/api"
    hosts: ["api.example.com"]
    upstream_urls: ["http://api-v1:3000"]
```

## Load balancing

```yaml
routes:
  - name: "balanced"
    listen_path: "/api"
    upstream_urls:
      - "http://backend-1:3000"
      - "http://backend-2:3000"
    load_balancing: "round_robin"
```

See [Routing](/guide/routing) for the full reference.
