# CLI Reference

GatewayX ships two binaries:

- `gatewayx` — the gateway server
- `gatewayx-cli` — the command-line tool (validate, init, version)

## gatewayx

The main server binary. Reads `gatewayx.yaml` (or the path in `GATEWAYX_CONFIG`).

```bash
# Run with default config (./gatewayx.yaml)
./gatewayx

# Run with a specific config
GATEWAYX_CONFIG=/etc/gatewayx/prod.yaml ./gatewayx
```

## gatewayx-cli

### init

Generate a configuration interactively.

```bash
gatewayx-cli init                # writes gatewayx.yaml
gatewayx-cli init -o prod.yaml   # writes to a specific file
```

Walks through questions: port, routes, upstreams, auth, rate limits, cache, websocket.

### validate

Validate a config file and warn on missing recommended settings.

```bash
gatewayx-cli validate
gatewayx-cli validate -c prod.yaml
```

Example output:

```
Configuration is valid (2 routes configured)

3 warning(s):
  !  route "api" has no authentication configured
  !  route "api" has no rate limiting configured
  !  route "web" has a single upstream with no health check
```

### serve

Start the gateway server.

```bash
gatewayx-cli serve
gatewayx-cli serve -c prod.yaml
```

### version

```bash
gatewayx-cli version
```

Example output:

```
GatewayX v0.4.2
  commit:    abc1234
  build date: 2026-08-14T10:00:00Z
```

## Docker / Podman

```bash
# Pull and run
podman pull ghcr.io/oni1997/gatewayx:latest
podman run -d -p 8080:8080 -p 9090:9090 \
  -v $(pwd)/gatewayx.yaml:/etc/gatewayx/gatewayx.yaml \
  ghcr.io/oni1997/gatewayx:latest

# Run the CLI inside the container
podman exec -it $(podman ps -q --filter ancestor=ghcr.io/oni1997/gatewayx) gatewayx-cli version
```

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `:8080/health` | Gateway health check |
| GET | `:9090/metrics` | Prometheus metrics |
| GET | `:9090/history` | Request history (JSON) |
| GET | `:9090/security` | ML security scan |
| GET | `:9090/bottlenecks` | Bottleneck analysis |
| GET | `:9090/recommendations` | Rate limit / cache recommendations |
| GET | `:9090/analysis` | Full ML report |
| GET | `:9090/api/keys` | List API keys |
| POST | `:9090/api/keys` | Create API key |
| DELETE | `:9090/api/keys/{id}` | Revoke API key |
| GET | `:9090/api/certs` | List certificates |
| POST | `:9090/api/certs` | Add certificate |
| GET | `:9090/api/config` | Gateway configuration summary |
| GET | `:9090/api/audit` | Audit log |
| GET | `:9090/oauth/login` | OAuth login (if configured) |
| GET | `:9090/oauth/callback` | OAuth callback |
| GET | `:9090/oauth/logout` | OAuth logout |

## Admin API Authentication

If `admin.token` is configured, all `/api/*` endpoints require authentication:

```bash
# Without token (rejected)
curl localhost:9090/api/keys
# → 401

# With Bearer token
curl -H "Authorization: Bearer my-token" localhost:9090/api/keys

# With query param
curl "localhost:9090/api/keys?token=my-token"
```
