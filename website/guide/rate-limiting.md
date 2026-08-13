# Rate Limiting

## Global

```yaml
rate_limit:
  rate: 100
  burst: 200
  strategy: "token_bucket"
```

## Per-IP

```yaml
rate_limit:
  rate: 10
  per_ip: true
```

## Per-user

```yaml
rate_limit:
  rate: 50
  per_user: true
```

## Per-API-key

```yaml
rate_limit:
  rate: 100
  per_key: true
```

## Distributed (Redis)

```yaml
rate_limit:
  rate: 1000
  redis_addr: "localhost:6379"
  strategy: "sliding_window"
```
