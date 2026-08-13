# Dashboard

The dashboard is served alongside metrics on port 9090:

```
http://localhost:9090/
```

## Pages

- Home — overview and top routes
- Services — per-route statistics
- Metrics — request/latency charts
- History — recent requests with trace IDs
- API Keys — create/revoke API keys (persisted in SQLite)
- Certificates — manage certificates (persisted in SQLite)
- Health — service health checks
- Settings — configuration editor

## API endpoints

| Endpoint | Description |
|----------|-------------|
| `:9090/api/keys` | API key CRUD |
| `:9090/api/certs` | Certificate CRUD |
| `:9090/security` | ML security scan |
| `:9090/bottlenecks` | Bottleneck analysis |
| `:9090/recommendations` | Recommendations |
