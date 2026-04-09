# ADR-004: CI/CD Pipeline Design

**Date:** 2026-04-09
**Status:** Accepted

## Decision

GitHub Actions for CI/CD with the following jobs:

1. **Lint** - Code quality checks (gofmt, go vet, golangci-lint)
2. **Test** - Unit tests with coverage
3. **Security** - gosec, govulncheck
4. **Build** - Multi-platform binaries
5. **Docker** - Container image building

## Pipeline Stages

```yaml
on: [push, pull_request]
jobs:
  lint:
  test:
    needs: [lint]
  security:
    needs: [test]
  build:
    needs: [test]
  docker:
    needs: [test]
```

## Release Process

- Tag-triggered releases
- Multi-platform builds (Linux, macOS, Windows)
- Docker images to GHCR and Docker Hub

## Consequences

- **Positive:** Automated quality gates
- **Positive:** Consistent build process
- **Positive:** Security scanning in pipeline
- **Negative:** More complex configuration
