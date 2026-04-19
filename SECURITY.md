# Security Policy

## Supported Versions

| Version | Supported |
|---------|-----------|
| 0.29.x  | ✅ |
| < 0.29  | ❌ |

## Reporting a Vulnerability

We take security seriously. If you discover a vulnerability:

1. **Do NOT** open a public issue
2. Email us at [security@graycode.ai](mailto:security@graycode.ai) or use [GitHub Security Advisories](https://github.com/GrayCodeAI/tok/security/advisories)
3. Include:
   - Description of the vulnerability
   - Steps to reproduce
   - Potential impact
   - Suggested fix (if any)

We will respond within **48 hours** and aim to release a fix within **14 days**.

## Security Considerations

### Input Handling

- tok processes untrusted input from terminal output and user input
- All input is validated for size (50MB limit) and encoding
- Shell scripts executed by hook commands are validated for path traversal and permissions

### Data Storage

- Tracking data is stored locally in SQLite (`~/.local/share/tok/tracking.db`)
- No data is transmitted externally
- Telemetry is anonymous and opt-out

### Dependencies

- We monitor dependencies via Dependabot (`.github/dependabot.yml`)
- All dependencies are pinned in `go.sum`
- SBOM is generated for each release (CycloneDX format)

### Best Practices

- All secrets should be environment variables, never hardcoded
- No credentials in source code, commits, or logs
- Use `go vet` and `golangci-lint` for static analysis
- Race detector enabled in CI (`make test-race`)
