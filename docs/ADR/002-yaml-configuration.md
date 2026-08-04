# ADR-002: Use YAML for Configuration

## Status

Accepted

## Context

GatewayX needs a configuration format that is human-readable, easy to version control, and familiar to the target audience (DevOps engineers and backend developers).

Candidates considered:
- **YAML** — Human-friendly, supports comments, widely used in infrastructure
- **TOML** — Simpler than YAML, but less ecosystem support
- **JSON** — No comments support, verbose for humans
- **HCL** — Powerful but complex, steep learning curve

## Decision

Use YAML as the primary configuration format, loaded via Viper. Viper adds support for environment variable overrides with the `GATEWAYX_` prefix, JSON files, and TOML files as secondary options.

## Consequences

- **Positive:** Familiar format for Kubernetes users and DevOps engineers.
- **Positive:** Viper provides environment variable overrides, enabling 12-factor app patterns.
- **Positive:** Comments allowed in configuration files for documentation.
- **Negative:** YAML has edge cases (e.g., Norway problem with `no` being parsed as `false`). We mitigate this by validating config at startup.
