# Security Best Practices

**Version:** 1.0.0
**Last Updated:** 2026-04-03

## Overview

This document outlines security best practices for operating Tok in production environments.

## Authentication

### API Key Authentication

Tok uses API keys for authenticating HTTP API requests:

```bash
# Generate a new API key
tok config set api_key $(openssl rand -hex 32)

# Use in requests
curl -H "X-API-Key: your-api-key" http://localhost:8080/api/v1/compress
```

**Best Practices:**
- Rotate API keys regularly
- Use different keys for different environments
- Never commit API keys to version control
- Store keys in environment variables or secure vaults

## Transport Security

### Local-Only Binding

By default, Tok binds to `localhost` only. Do not expose the HTTP API to untrusted networks without a reverse proxy and TLS termination.

For production deployments behind a reverse proxy (nginx, Caddy, etc.), terminate TLS at the proxy layer and forward to Tok over localhost.

## Input Validation

### Sanitization

Tok validates inputs before processing:

1. **Command Injection Prevention**: Shell commands are validated against allowed patterns
2. **Path Validation**: File paths are checked for safety

### Rate Limiting

The HTTP server supports configurable rate limiting via the `X-RateLimit-*` header conventions. Configure your reverse proxy to enforce rate limits in production.

## Secrets Management

### Environment Variables

```bash
# Set via environment
export TOK_API_KEY="your-secure-key"
```

### Configuration File Security

```bash
# Set restrictive permissions
chmod 600 ~/.config/tok/config.toml
```

## Network Security

### Firewall Rules

Restrict access to the Tok HTTP server:

```bash
# Allow only local access
iptables -A INPUT -p tcp --dport 8080 -s 127.0.0.1 -j ACCEPT
```

## Dependency Security

### Vulnerability Scanning

```bash
# Check for vulnerable dependencies
govulncheck ./...
gosec ./...
```

### Dependency Updates

```bash
# Update dependencies
go get -u ./...
go mod tidy
```

## Incident Response

### Security Incident Checklist

1. **Identify**: Detect the incident via logs or alerts
2. **Contain**: Disable affected API keys, block IPs
3. **Eradicate**: Update credentials, remove compromised configs
4. **Recover**: Restore from trusted backup
5. **Report**: Document incident and notify stakeholders

### Emergency Commands

```bash
# Revoke API key
tok config set api_key ""
```

## Security Checklist

- [ ] API keys rotated recently
- [ ] Configuration file permissions restricted (600)
- [ ] TLS termination configured at reverse proxy
- [ ] Firewall rules restrict external access
- [ ] Dependencies scanned for vulnerabilities
- [ ] Rate limiting configured at proxy layer

## Reporting Security Issues

If you discover a security vulnerability, please report it responsibly:

1. Email: security@tok.dev
2. Do not disclose publicly until patched
3. Include steps to reproduce
4. We aim to respond within 48 hours

## Security Updates

Subscribe to security advisories:
- GitHub Security Advisories: https://github.com/GrayCodeAI/tok/security/advisories
- Release notes include security fixes
