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

### JWT

```yaml
authentication:
  type: "jwt"
  options:
    secret: "${JWT_SECRET}"
    algorithm: "HS256"
    claims:
      - "sub"
      - "exp"
```

### Basic Auth

```yaml
authentication:
  type: "basic"
  options:
    users_file: "/etc/gatewayx/users.htpasswd"
```

### OAuth 2.0

```yaml
authentication:
  type: "oauth2"
  options:
    provider: "github"
    client_id: "${OAUTH_CLIENT_ID}"
    client_secret: "${OAUTH_CLIENT_SECRET}"
    redirect_url: "https://gateway.example.com/oauth/callback"
```

### mTLS

```yaml
authentication:
  type: "mtls"
  options:
    ca_cert: "/etc/gatewayx/ca.pem"
    verify_depth: 2
```

### HMAC

```yaml
authentication:
  type: "hmac"
  options:
    algorithm: "sha256"
    secret: "${HMAC_SECRET}"
    header: "X-Signature"
```

## RBAC & Permissions

Role-based access control can be layered on top of authentication:

```yaml
authentication:
  type: "jwt"
  options:
    secret: "${JWT_SECRET}"
    roles_claim: "roles"
    permissions:
      - path: "/admin/**"
        roles: ["admin"]
      - path: "/api/**"
        roles: ["admin", "developer"]
```

## Token Caching

For JWT and OAuth, GatewayX can cache validated tokens in Redis to reduce upstream authentication calls:

```yaml
authentication:
  type: "jwt"
  options:
    cache:
      enabled: true
      ttl: 300s
      backend: "redis"
      redis:
        addr: "localhost:6379"
```

## Upcoming

- LDAP support (Phase 9)
- SAML support (Phase 9)
- Multi-tenant authentication (Phase 9)
