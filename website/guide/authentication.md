# Authentication

GatewayX supports multiple auth strategies per route.

## JWT

```yaml
authentication:
  type: "jwt"
  options:
    secret: "${JWT_SECRET}"
    algorithm: "HS256"
```

## API key

```yaml
authentication:
  type: "api_key"
  options:
    header: "X-API-Key"
    keys_file: "/etc/gatewayx/api_keys.txt"
```

## Basic auth

```yaml
authentication:
  type: "basic"
  options:
    admin: "password123"
```

## OAuth 2.0

```yaml
authentication:
  type: "oauth"
  options:
    provider: "github"
    client_id: "${GITHUB_CLIENT_ID}"
    client_secret: "${GITHUB_CLIENT_SECRET}"
```

## mTLS

```yaml
authentication:
  type: "mtls"
  options:
    ca_cert: "/etc/gatewayx/ca.pem"
```

## HMAC

```yaml
authentication:
  type: "hmac"
  options:
    secret: "${HMAC_SECRET}"
    algorithm: "sha256"
```

See [Authentication](/guide/authentication) for details on RBAC and sessions.
