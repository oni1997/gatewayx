# Dashboard

The GatewayX dashboard provides a web-based interface for managing routes, users, API keys, and monitoring traffic.

## Technology

- **React** -- Component-based UI library
- **TypeScript** -- Type-safe JavaScript
- **Tailwind CSS** -- Utility-first CSS framework
- **Vite** -- Fast build tool and dev server

## Pages

| Page | Description |
|------|-------------|
| Home | Overview of gateway status, traffic, and health |
| Services | Manage routes, upstreams, and load balancing |
| Users | User management and permissions |
| API Keys | Generate and revoke API keys |
| Logs | Real-time and historical request logs |
| Metrics | Traffic graphs, latency, error rates |
| Plugins | Enable, disable, and configure plugins |
| Settings | Global gateway configuration |
| Health | Service health dashboard |
| Certificates | TLS certificate management |

## Running the Dashboard

```bash
cd apps/dashboard
npm install
npm run dev
```

## Building for Production

```bash
cd apps/dashboard
npm run build
```

The dashboard communicates with the GatewayX Admin REST API, which runs alongside the proxy on a separate port.

## Admin API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/routes` | List all routes |
| POST | `/api/routes` | Create a route |
| PUT | `/api/routes/:id` | Update a route |
| DELETE | `/api/routes/:id` | Delete a route |
| GET | `/api/health` | Gateway health |
| GET | `/api/metrics` | Gateway metrics |
| GET | `/api/plugins` | List plugins |
| POST | `/api/plugins/:id/toggle` | Enable/disable plugin |

The dashboard is scheduled for Phase 5 of development.
