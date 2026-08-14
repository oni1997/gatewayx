# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 0.4.x   | :white_check_mark: |
| 0.3.x   | :white_check_mark: |
| < 0.3.0 | :x:                |

## Reporting a Vulnerability

GatewayX takes security seriously. If you discover a security vulnerability, please report it privately — **do not open a public issue**.

**Email:** dzidzaimaenza@gmail.com

Please include:

- A description of the vulnerability
- Steps to reproduce
- Affected version(s)
- Any proof-of-concept code or logs

### What to expect

1. **Acknowledgment** within 48 hours
2. **Assessment** of the vulnerability's severity and impact
3. **Fix** released as a patch version
4. **Public disclosure** after the fix is released (credit given if desired)

## Security Best Practices

When deploying GatewayX in production:

1. **Always enable TLS** — never expose plain HTTP publicly
2. **Use environment variables for secrets** — never hardcode JWT secrets or API keys in `gatewayx.yaml`
3. **Set `max_body_size`** — prevent large payload attacks
4. **Configure `allowed_hosts`** — prevent host header injection
5. **Rotate API keys regularly** — keys are revoked via the admin API
6. **Use rate limiting** — protect against abuse and DoS
7. **Keep GatewayX updated** — security fixes ship in patch releases

## Security Features

- JWT (HS256/RS256/ES256), API keys, Basic Auth, HMAC, OAuth 2.0, mTLS
- RBAC with path-glob permission matching
- Token bucket and sliding window rate limiting
- ML-based attack detection (SQL injection, XSS, path traversal, shell injection)
- Circuit breaker to prevent cascading failures
- TLS with configurable minimum version
