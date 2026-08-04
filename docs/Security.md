# Security

## Transport Security

### TLS

GatewayX supports TLS for all incoming connections:

```yaml
tls:
  enabled: true
  cert_file: "/etc/gatewayx/cert.pem"
  key_file: "/etc/gatewayx/key.pem"
  min_version: "1.2"
```

### mTLS

Mutual TLS for service-to-service authentication:

```yaml
tls:
  enabled: true
  cert_file: "/etc/gatewayx/cert.pem"
  key_file: "/etc/gatewayx/key.pem"
  client_auth: "require"
  client_ca_file: "/etc/gatewayx/ca.pem"
```

### Auto Cert (Let's Encrypt)

```yaml
tls:
  enabled: true
  auto_cert: true
  auto_cert_dir: "/var/lib/gatewayx/certs"
  domains:
    - "gateway.example.com"
```

## Request Security

### Body Size Limiting

```yaml
security:
  max_body_size: 10485760  # 10MB
```

### Host Header Validation

```yaml
security:
  allowed_hosts:
    - "api.example.com"
    - "admin.example.com"
```

### Trusted Proxies

When GatewayX is behind a load balancer:

```yaml
security:
  trusted_proxies:
    - "10.0.0.0/8"
    - "172.16.0.0/12"
```

## Authentication

See [Authentication](Authentication.md) for details on JWT, API keys, OAuth, mTLS, HMAC, and Basic Auth.

## Rate Limiting

See the rate limiting configuration to protect against abuse and DoS attacks.

## Security Headers

GatewayX adds security headers by default:

- `X-Content-Type-Options: nosniff`
- `X-Frame-Options: DENY`

Additional headers can be configured per-route:

```yaml
routes:
  - name: "secure"
    headers:
      Strict-Transport-Security: "max-age=31536000; includeSubDomains"
      Content-Security-Policy: "default-src 'self'"
      X-XSS-Protection: "1; mode=block"
```

## Best Practices

1. Always enable TLS in production
2. Use environment variables for secrets (never hardcode in YAML)
3. Set `max_body_size` to prevent large payload attacks
4. Configure `allowed_hosts` to prevent host header injection
5. Rotate API keys and JWT secrets regularly
6. Enable audit logging for compliance
7. Keep GatewayX updated to the latest version

## Reporting Vulnerabilities

Please report security vulnerabilities to security@gatewayx.dev. Do not open public issues.
