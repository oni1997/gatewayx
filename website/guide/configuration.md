# Configuration

GatewayX is configured via YAML. The default location is `./gatewayx.yaml`.

```yaml
server:
  host: "0.0.0.0"
  port: 8080

routes:
  - name: "api"
    listen_path: "/api"
    upstream_urls: ["http://backend:3000"]

logging:
  level: "info"
  format: "json"

metrics:
  enabled: true
  port: 9090
```

## Environment variables

| Variable | Description |
|----------|-------------|
| `GATEWAYX_CONFIG` | Config file path |
| `GATEWAYX_DB_PATH` | SQLite database path |
| `GATEWAYX_WEBHOOK_URL` | Slack/Discord webhook |

All config values can be overridden with `GATEWAYX_` prefixed env vars.

## Hot reload

Edit the config file — GatewayX auto-reloads via file watching. Or send `SIGHUP`:

```bash
kill -HUP $(pgrep gatewayx)
```

See the full [Configuration](/guide/configuration) reference in `docs/Configuration.md`.
