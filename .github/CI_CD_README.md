# CI/CD Pipeline Documentation

This document describes the continuous integration and deployment pipeline for Tok.

## Overview

The CI/CD pipeline ensures code quality, security, and automated releases through GitHub Actions.

## Workflows

### 1. CI Workflow (`ci.yml`)

Runs on every push and pull request to `main` and `develop` branches.

**Jobs:**
- **Test**: Runs tests across Go 1.24, 1.25, and 1.26
- **Lint**: Runs `go vet`, formatting checks, and `golangci-lint`
- **Build**: Cross-compiles binaries for Linux, macOS, and Windows (amd64/arm64)
- **Coverage**: Enforces minimum 50% code coverage threshold
- **Integration**: Runs integration tests and CLI smoke tests

### 2. Security Workflow (`security.yml`)

Runs on every push, PR, and daily via cron schedule.

**Jobs:**
- **gosec**: Security vulnerability scanner
- **govulncheck**: Go vulnerability database check
- **CodeQL**: GitHub's semantic code analysis
- **Dependency Review**: Checks for known vulnerabilities in dependencies
- **Trivy**: Container and filesystem vulnerability scanner

### 3. Release Workflow (`release.yml`)

Triggered on version tags (e.g., `v1.2.3`).

**Jobs:**
- **Test**: Pre-release validation
- **Build**: Creates binaries for all supported platforms
- **Release**: Creates GitHub release with artifacts and changelog
- **Docker**: Builds and publishes multi-arch Docker images to GHCR

## Automated Dependency Management

### Dependabot

Configured in `.github/dependabot.yml`:

- **Go modules**: Weekly updates, grouped by minor/patch
- **GitHub Actions**: Weekly updates
- **Docker**: Monthly updates

## Coverage Reporting

### Codecov Integration

Configured in `codecov.yml`:

- **Project threshold**: 50% minimum coverage
- **Patch threshold**: 60% minimum for new code
- **Component tracking**: Separate coverage for core modules
- **PR comments**: Automated coverage reports on pull requests

## Local Development

### Pre-commit Hooks

Install pre-commit hooks to catch issues before pushing:

```bash
# Install pre-commit
pip install pre-commit

# Install hooks
pre-commit install

# Run manually
pre-commit run --all-files
```

### Required Tools

```bash
# Install required tools
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install golang.org/x/vuln/cmd/govulncheck@latest
go install github.com/securego/gosec/v2/cmd/gosec@latest
```

## Environment Variables

### Required Secrets

- `GITHUB_TOKEN`: Automatically provided by GitHub Actions
- `CODECOV_TOKEN`: For Codecov coverage uploads (optional for public repos)

## Release Process

1. **Create a tag**:
   ```bash
   git tag -a v1.2.3 -m "Release version 1.2.3"
   git push origin v1.2.3
   ```

2. **Automated Actions**:
   - Tests run automatically
   - Binaries built for all platforms
   - Docker images published to GHCR
   - GitHub release created with changelog

3. **Version Format**:
   - Stable: `v1.2.3`
   - Prerelease: `v1.2.3-alpha.1`, `v1.2.3-beta.1`, `v1.2.3-rc.1`

## Troubleshooting

### Common Issues

**Lint failures:**
```bash
make fmt
make lint
```

**Test failures:**
```bash
make test
make test-race
```

**Pre-commit hook failures:**
```bash
pre-commit run --all-files
```

## Monitoring

- Check workflow status in the **Actions** tab
- Review security findings in the **Security** tab
- Monitor coverage trends on **Codecov**
- View release artifacts in **Releases**

## Contributing

When contributing:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Ensure pre-commit hooks pass
5. Open a pull request
6. Wait for CI to pass
7. Request review from maintainers

## Badge Status

Add these badges to your README:

```markdown
![CI](https://github.com/lakshmanpatel/tok/workflows/CI/badge.svg)
![Security](https://github.com/lakshmanpatel/tok/workflows/Security/badge.svg)
[![codecov](https://codecov.io/gh/lakshmanpatel/tok/branch/main/graph/badge.svg)](https://codecov.io/gh/lakshmanpatel/tok)
[![Go Report Card](https://goreportcard.com/badge/github.com/lakshmanpatel/tok)](https://goreportcard.com/report/github.com/lakshmanpatel/tok)
```
