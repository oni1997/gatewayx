# Architecture Decision Records

This directory contains Architecture Decision Records (ADRs) for GatewayX.

ADRs document significant architectural decisions, the context in which they were made, and their consequences.

## Format

Each ADR follows this format:

```
# ADR-NNN: Title

## Status
[Proposed | Accepted | Deprecated | Superseded]

## Context
What is the issue we're seeing that's motivating this decision?

## Decision
What is the change we're proposing and/or doing?

## Consequences
What becomes easier or more difficult to do because of this change?
```

## Index

| ADR | Title | Status |
|-----|-------|--------|
| [ADR-001](ADR/001-use-go.md) | Use Go as the primary language | Accepted |
| [ADR-002](ADR/002-yaml-configuration.md) | Use YAML for configuration | Accepted |
| [ADR-003](ADR/003-standard-library-http.md) | Use net/http over frameworks | Accepted |
