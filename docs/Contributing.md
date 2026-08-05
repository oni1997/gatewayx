# Contributing

## Development Philosophy

Every contribution must satisfy:
- **Easy to understand** -- Clear code, clear intent
- **Easy to configure** -- Configuration-driven, not code-driven
- **Production ready** -- Tested, monitored, documented
- **Fast** -- Performance is a feature
- **Well documented** -- Code comments, docs, and examples

## Getting Started

```bash
git clone https://github.com/oni1997/gatewayx.git
cd gatewayx
make dev
```

## Project Structure

```
apps/       -- Application entry points
internal/   -- Internal packages (not importable externally)
pkg/        -- Shared, importable packages
plugins/    -- Plugin implementations
docs/       -- Documentation
tests/      -- Integration and e2e tests
```

## Development Workflow

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/my-feature`
3. Make your changes
4. Write tests: `make test`
5. Run linting: `make lint`
6. Submit a pull request

## Commit Convention

We follow [Conventional Commits](https://www.conventionalcommits.org/):

```
feat: add JWT authentication support
fix: resolve race condition in load balancer
docs: update configuration reference
test: add integration tests for rate limiting
refactor: extract middleware to shared package
chore: update dependencies
```

## Testing

```bash
# Unit tests
make test

# Integration tests
make test-integration

# End-to-end tests
make test-e2e

# Coverage
make coverage
```

## Code Style

- Follow standard Go conventions (`gofmt`, `go vet`, `golangci-lint`)
- Use `slog` for all logging
- Write table-driven tests
- Document all exported functions and types
- Keep functions small and focused

## Pull Request Process

1. Ensure all tests pass
2. Update documentation if needed
3. Add an ADR for significant architectural changes
4. Get review from the maintainer ([@oni1997](https://github.com/oni1997))
5. Squash merge to `main`

## Reporting Issues

- Use the GitHub issue tracker
- Include GatewayX version and configuration
- Provide steps to reproduce
- Attach relevant logs

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
