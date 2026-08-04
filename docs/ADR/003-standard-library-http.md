# ADR-003: Use net/http over HTTP Frameworks

## Status

Accepted

## Context

The gateway core needs to handle HTTP requests. We must choose between Go's standard library and third-party frameworks.

Candidates considered:
- **net/http + httputil.ReverseProxy** -- Standard library, zero dependencies, well-tested
- **Fiber** -- fasthttp-based, very fast but incompatible with net/http middleware
- **Chi** -- Lightweight router on top of net/http
- **Gin** -- Full-featured framework with its own context type

## Decision

Use Go's `net/http` and `httputil.ReverseProxy` directly. The standard library provides everything needed for a reverse proxy: connection pooling, HTTP/2 support, timeouts, and TLS. We build our own lightweight router (see `internal/router`) to handle host and path matching, avoiding framework lock-in.

## Consequences

- **Positive:** Zero external HTTP dependencies. Fully compatible with any net/http middleware.
- **Positive:** Guaranteed long-term stability -- the standard library has a strong compatibility promise.
- **Positive:** Direct access to connection-level controls (timeouts, buffer sizes, TLS config).
- **Negative:** We need to implement our own routing logic instead of using a battle-tested router.
