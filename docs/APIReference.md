# API Reference

The GatewayX Admin REST API provides programmatic access to manage routes, plugins, and monitor the gateway.

## Authentication

All admin API requests require an admin token:

```
Authorization: Bearer <admin-token>
```

## Endpoints

### Health

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
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### Routes

```
GET    /api/routes          List all routes
POST   /api/routes          Create a route
GET    /api/routes/:id      Get a route
PUT    /api/routes/:id      Update a route
DELETE /api/routes/:id      Delete a route
```

**Create Route:**
```json
{
  "name": "my-service",
  "listen_path": "/api",
  "upstream_urls": ["http://backend:3000"],
  "methods": ["GET", "POST"],
  "strip_path": false
}
```

### Plugins

```
GET    /api/plugins         List plugins
GET    /api/plugins/:id     Get plugin details
POST   /api/plugins/:id/enable   Enable plugin
POST   /api/plugins/:id/disable  Disable plugin
PUT    /api/plugins/:id/config    Update plugin configuration
```

### Metrics

```
GET /api/metrics
```

Returns current gateway metrics in JSON format.

### Configuration

```
GET    /api/config          Get current configuration
PUT    /api/config          Update configuration
POST   /api/config/reload   Reload configuration from file
```

## Error Responses

All errors follow this format:

```json
{
  "error": {
    "code": "ROUTE_NOT_FOUND",
    "message": "Route with id 'xyz' not found"
  }
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
| 409 | Conflict |
| 500 | Internal Server Error |
| 502 | Bad Gateway |
| 503 | Service Unavailable |

## Pagination

List endpoints support pagination:

```
GET /api/routes?page=1&limit=20
```

**Response:**
```json
{
  "data": [...],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 150,
    "pages": 8
  }
}
```
