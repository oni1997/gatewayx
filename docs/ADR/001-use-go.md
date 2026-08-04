# ADR-001: Use Go as the Primary Language

## Status

Accepted

## Context

We need to choose a primary language for the GatewayX core. The gateway must be high-performance, have low resource overhead, produce a single static binary, and be easy to deploy.

Candidates considered:
- **Go** -- Fast compilation, excellent concurrency, single binary, strong standard library
- **Rust** -- Maximum performance, but steeper learning curve and slower iteration
- **Node.js** -- Large ecosystem, but higher resource usage and complex deployment
- **Java** -- Mature ecosystem, but high memory overhead and slow startup

## Decision

Use Go for the gateway core and CLI. Go provides:
- `net/http` and `httputil.ReverseProxy` in the standard library
- Goroutines for highly concurrent connection handling
- Cross-compilation to all major platforms from a single codebase
- Low memory footprint and fast startup
- A vibrant ecosystem of infrastructure tools (Docker, Kubernetes, Terraform, etc.)

## Consequences

- **Positive:** Single static binary deployment. Excellent performance. Battle-tested HTTP stack.
- **Positive:** Large pool of Go developers familiar with infrastructure patterns.
- **Positive:** Cobra/Viper are industry-standard for CLIs and configuration.
- **Negative:** Less expressive type system than Rust; no null safety.
- **Negative:** Plugin system will require either Go plugins (limited) or a process-based approach.
