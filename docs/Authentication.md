# Authentication

GatewayX supports multiple authentication strategies, configurable per route.

## Authentication Methods

### API Key

```yaml
authentication:
  type: "api_key"
  options:
    header: "X-API-Key"
    keys_file: "/etc/gatewayx/api_keys.txt"
```

Keys file format (`key:owner`, one per line):
```
sk-abc123:admin
sk-def456:readonly
```

### JWT

```yaml
authentication:
  type: "jwt"
  options:
    secret: "${JWT_SECRET}"       # HMAC secret (or secret_file)
    algorithm: "HS256"             # HS256, RS256, ES256
    cache_ttl: 5m                  # Optional token cache
```

For RS256/ES256, use `public_key_file` instead of `secret`:

```yaml
authentication:
  type: "jwt"
  options:
    public_key_file: "/etc/gatewayx/public.pem"
    algorithm: "RS256"
```

### Basic Auth

```yaml
authentication:
  type: "basic"
  options:
    realm: "GatewayX"
    htpasswd_file: "/etc/gatewayx/users.htpasswd"
```

Or inline users:

```yaml
authentication:
  type: "basic"
  options:
    admin: "password123"
```

### OAuth 2.0

```yaml
authentication:
  type: "oauth"
  options:
    provider: "github"             # github or google
    client_id: "${OAUTH_CLIENT_ID}"
    client_secret: "${OAUTH_CLIENT_SECRET}"
```

The full login flow (login, callback, logout) is enabled via the top-level `oauth` config block:

```yaml
oauth:
  provider: "github"
  client_id: "${OAUTH_CLIENT_ID}"
  client_secret: "${OAUTH_CLIENT_SECRET}"
  redirect_url: "https://gateway.example.com:9090/oauth/callback"
```

### mTLS

```yaml
authentication:
  type: "mtls"
  options:
    ca_cert: "/etc/gatewayx/ca.pem"
    verify_depth: 2
```

Validates the client certificate against the CA. Identity is extracted from the certificate CN.

### HMAC

```yaml
authentication:
  type: "hmac"
  options:
    algorithm: "sha256"            # sha256 or sha512
    secret: "${HMAC_SECRET}"
    header: "X-Signature"
    clock_skew: 5m
```

Signature format: `key_id|timestamp|signature`.

### Session

```yaml
authentication:
  type: "session"
  options:
    ttl: 30m
    max_sessions: 10000
```

## RBAC & Permissions

RBAC wraps another authenticator and adds role-based path/method checks:

```yaml
authentication:
  type: "rbac"
  options:
    delegate: "jwt"                # The underlying auth type
    secret: "${JWT_SECRET}"
    roles_claim: "roles"
    perm_1: "/admin/**:admin:GET,POST"
    perm_2: "/api/**:admin,developer"
    perm_3: "/public/*:guest:GET"
```

Permission format: `path:roles:methods` (methods optional). Supports `**` (recursive) and `*` (single segment) wildcards.

## Token Caching

JWT tokens can be cached in-memory to avoid re-validating signatures:

```yaml
authentication:
  type: "jwt"
  options:
    secret: "${JWT_SECRET}"
    cache_ttl: 5m
```

The cache respects the JWT `exp` claim — entries expire at the earlier of `cache_ttl` or token expiry.

## Admin API Authentication

The admin API (`/api/*`) is protected separately via the `admin.token` config:

```yaml
admin:
  token: "${ADMIN_TOKEN}"
```

Requests require `Authorization: Bearer <token>` or `?token=<token>`.

## Upcoming

- LDAP support (Phase 9)
- SAML support (Phase 9)
- Multi-tenant authentication (Phase 9)
