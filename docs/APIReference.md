# API Reference

GatewayX exposes a monitoring and admin API on the metrics port (default `9090`).

## Authentication

If `admin.token` is configured, all `/api/*` endpoints require authentication:

```
Authorization: Bearer <admin-token>
```

or as a query parameter:

```
GET /api/keys?token=<admin-token>
```

## Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Gateway health status |
| GET | `/metrics` | Prometheus-format metrics |
| GET | `/version` | Version, commit, build date |
| GET | `/history` | Recent request history (JSON) |
| GET | `/security` | ML security scan |
| GET | `/bottlenecks` | Bottleneck analysis |
| GET | `/recommendations` | Rate limit / cache recommendations |
| GET | `/analysis` | Full ML report (security + bottlenecks + recommendations) |
| GET | `/api/keys` | List API keys |
| POST | `/api/keys` | Create an API key |
| DELETE | `/api/keys/{id}` | Revoke an API key |
| GET | `/api/certs` | List certificates |
| POST | `/api/certs` | Add a certificate |
| GET | `/api/config` | Gateway configuration summary |
| GET | `/api/audit` | Audit log |
| GET | `/oauth/login` | OAuth login (if configured) |
| GET | `/oauth/callback` | OAuth callback |
| GET | `/oauth/logout` | OAuth logout |

## Health

```
GET /health
```

Returns gateway health status.

**Response:**
```json
{
  "status": "healthy",
  "uptime": "2h45m30s",
  "checks": {
    "gateway": "healthy"
  },
  "timestamp": "2026-01-01T00:00:00Z"
}
```

## Version

```
GET /version
```

**Response:**
```json
{
  "version": "0.4.4",
  "commit": "abc1234",
  "build_date": "2026-08-14T09:42:18Z"
}
```

## API Keys

```
GET    /api/keys           List keys
POST   /api/keys           Create key
DELETE /api/keys/{id}      Revoke key
```

**Create key:**
```json
POST /api/keys
{
  "name": "production",
  "owner": "team-a"
}
```

**Response:**
```json
{
  "id": "1785952293782-abc",
  "name": "production",
  "key": "sk-..."
}
```

## Certificates

```
GET    /api/certs          List certificates
POST   /api/certs          Add certificate
```

**Add certificate:**
```json
POST /api/certs
{
  "domain": "api.example.com",
  "issuer": "Lets Encrypt"
}
```

## ML Analysis

```
GET /security          Threat detection (SQL injection, XSS, brute force, scanners)
GET /bottlenecks       Slow route detection and latency spikes
GET /recommendations   Rate limit and cache suggestions
GET /analysis          Combined report
```

## Audit Log

```
GET /api/audit
```

Records admin actions: `key_created`, `key_revoked`, `cert_added`.

## Error Responses

```json
{
  "error": "unauthorized",
  "message": "invalid admin token"
}
```

## Status Codes

| Code | Description |
|------|-------------|
| 200 | Success |
| 201 | Created |
| 204 | No Content |
| 400 | Bad Request |
| 401 | Unauthorized |
| 403 | Forbidden |
| 404 | Not Found |
| 500 | Internal Server Error |
| 502 | Bad Gateway |
| 503 | Service Unavailable |
