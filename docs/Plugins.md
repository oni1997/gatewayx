# Plugins

GatewayX supports a plugin system to extend its functionality. Plugins can intercept requests at various lifecycle points.

## Plugin Lifecycle

```
Initialize → Configure → Start → [Serve Requests] → Stop
```

## Plugin Interface

```go
type Plugin interface {
    Name() string
    Version() string
    Init(config map[string]interface{}) error
    OnRequest(ctx context.Context, req *http.Request) error
    OnResponse(ctx context.Context, res *http.Response) error
    Close() error
}
```

## Lifecycle Hooks

| Hook | Description |
|------|-------------|
| `OnRequest` | Called before the request reaches the upstream |
| `OnResponse` | Called after receiving the upstream response |
| `OnError` | Called when an error occurs during proxying |

## Plugin Types

### Authentication Plugins

```yaml
plugins:
  enabled:
    - auth-ldap
    - auth-saml
    - auth-oauth
```

### Rate Limiting Plugins

```yaml
plugins:
  enabled:
    - ratelimit-redis
    - ratelimit-memory
```

### Monitoring Plugins

```yaml
plugins:
  enabled:
    - monitoring-datadog
    - monitoring-newrelic
```

### Storage Plugins

```yaml
plugins:
  enabled:
    - storage-s3
    - storage-gcs
```

### AI Plugins

```yaml
plugins:
  enabled:
    - ai-threat-detection
    - ai-log-summary
```

## Plugin Configuration

Each plugin can have its own configuration block:

```yaml
plugins:
  enabled:
    - auth-ldap
  config:
    auth-ldap:
      server: "ldap://ldap.example.com"
      base_dn: "dc=example,dc=com"
      user_filter: "(uid=%s)"
```

## Writing a Plugin

1. Create a Go package
2. Implement the `Plugin` interface
3. Build as a Go plugin: `go build -buildmode=plugin -o myplugin.so`
4. Place in the plugins directory

## SDK (Coming in Phase 6)

The GatewayX SDK will provide:
- Type-safe plugin interface
- Configuration helpers
- Event system
- Testing utilities
- Plugin registry and marketplace
