# Security

## TLS

```yaml
tls:
  enabled: true
  cert_file: "/etc/gatewayx/cert.pem"
  key_file: "/etc/gatewayx/key.pem"
  min_version: "1.2"
```

## mTLS

```yaml
tls:
  enabled: true
  cert_file: "/etc/gatewayx/cert.pem"
  key_file: "/etc/gatewayx/key.pem"
  client_auth: "require"
  client_ca_file: "/etc/gatewayx/ca.pem"
```

## ML security scanning

```
GET :9090/security
```

Detects SQL injection, XSS, path traversal, shell injection, and brute-force attempts.

## Best practices

1. Always enable TLS in production
2. Use environment variables for secrets
3. Set `max_body_size` to prevent large payload attacks
4. Report vulnerabilities to dzidzaimaenza@gmail.com
