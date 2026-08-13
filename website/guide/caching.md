# Caching

Add response caching to a route with a TTL:

```yaml
routes:
  - name: "cached-api"
    listen_path: "/api"
    upstream_urls: ["http://backend:3000"]
    cache:
      ttl: 30s
      max_size: 1000
```

- Only `GET` and `HEAD` requests are cached
- Responses include `X-Cache: HIT` or `X-Cache: MISS` headers
- Cached entries expire after the TTL

## WebSocket support

```yaml
routes:
  - name: "realtime"
    listen_path: "/ws"
    upstream_urls: ["http://ws-backend:8080"]
    websocket: true
```

## Compression

```yaml
routes:
  - name: "compressed"
    listen_path: "/api"
    upstream_urls: ["http://backend:3000"]
    compression: true
```
