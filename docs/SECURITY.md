# TokMan Security Best Practices

**Version:** 1.0.0  
**Last Updated:** 2026-03-28

## Overview

This document outlines security best practices for deploying and operating TokMan in production environments.

## Authentication

### API Key Authentication

TokMan uses API keys for authenticating HTTP API requests:

```bash
# Generate a new API key
tokman config set api_key $(openssl rand -hex 32)

# Use in requests
curl -H "X-API-Key: your-api-key" http://localhost:8080/api/v1/compress
```

**Best Practices:**
- Rotate API keys every 90 days
- Use different keys for different environments
- Never commit API keys to version control
- Store keys in environment variables or secure vaults

### JWT Authentication (Optional)

For extended sessions, TokMan supports JWT tokens:

```bash
# Generate JWT secret
tokman config set jwt_secret $(openssl rand -base64 64)

# Token includes: user_id, exp, iat claims
```

## Transport Security

### TLS Configuration

Always use TLS in production:

```toml
# ~/.config/tokman/config.toml
[server]
tls_enabled = true
tls_cert = "/path/to/cert.pem"
tls_key = "/path/to/key.pem"

# Minimum TLS version
tls_min_version = "1.3"
```

### mTLS for Service-to-Service

For microservice deployments, enable mutual TLS:

```toml
[grpc]
mtls_enabled = true
client_ca = "/path/to/client-ca.pem"
server_cert = "/path/to/server-cert.pem"
server_key = "/path/to/server-key.pem"
```

## Input Validation

### Sanitization

TokMan sanitizes all inputs before processing:

1. **Command Injection Prevention**: All shell commands are escaped
2. **Path Traversal Protection**: Paths are validated against allowed directories
3. **Buffer Overflow Protection**: Input size limits enforced

```go
// Maximum input size: 100MB
const MaxInputSize = 100 * 1024 * 1024

// Maximum command length: 10KB
const MaxCommandLength = 10 * 1024
```

### Rate Limiting

Configure rate limits to prevent abuse:

```toml
[rate_limit]
enabled = true
requests_per_minute = 60
burst = 10
```

## Hook Integrity

TokMan verifies hook integrity to prevent tampering:

```bash
# Verify hooks
tokman verify

# Store integrity hash
tokman trust
```

### Integrity Check Process

1. On installation, TokMan stores SHA-256 hashes of hook files
2. Before each execution, hashes are verified
3. Mismatches trigger warnings and require re-trust

## Secrets Management

### Environment Variables

Recommended approach for secrets:

```bash
# Set via environment
export TOKMAN_API_KEY="your-secure-key"
export TOKMAN_JWT_SECRET="your-jwt-secret"
export TOKMAN_DB_ENCRYPTION_KEY="your-encryption-key"
```

### Configuration File Security

```bash
# Set restrictive permissions
chmod 600 ~/.config/tokman/config.toml

# Use encrypted storage when possible
```

## Database Security

### SQLite Encryption

Enable SQLCipher for encrypted database:

```toml
[database]
encryption_enabled = true
# Key loaded from environment: TOKMAN_DB_KEY
```

### Database Permissions

```bash
# Restrict database file permissions
chmod 600 ~/.local/share/tokman/tokman.db
```

## Network Security

### Firewall Rules

Restrict access to TokMan services:

```bash
# Allow only local access to gateway
iptables -A INPUT -p tcp --dport 8080 -s 127.0.0.1 -j ACCEPT

# Allow service mesh communication
iptables -A INPUT -p tcp --dport 50051:50053 -j ACCEPT
```

### Network Policies (Kubernetes)

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: tokman-policy
spec:
  podSelector:
    matchLabels:
      app: tokman
  ingress:
  - from:
    - podSelector:
        matchLabels:
          app: tokman-gateway
```

## Audit Logging

### Enable Audit Logs

```toml
[audit]
enabled = true
log_path = "/var/log/tokman/audit.log"
log_format = "json"
retention_days = 90
```

### Log Contents

Each audit entry includes:
- Timestamp
- User/API key identifier
- Command or API endpoint
- Source IP address
- Action result (success/failure)
- Token count changes

## Security Headers

For HTTP responses, TokMan adds security headers:

```
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
X-XSS-Protection: 1; mode=block
Content-Security-Policy: default-src 'self'
Strict-Transport-Security: max-age=31536000; includeSubDomains
```

## Dependency Security

### Vulnerability Scanning

```bash
# Run security scan
make security-scan

# Uses govulncheck and gosec
govulncheck ./...
gosec ./...
```

### Dependency Updates

```bash
# Check for vulnerable dependencies
go list -m -u all | grep -i security

# Update dependencies
go get -u ./...
go mod tidy
```

## Incident Response

### Security Incident Checklist

1. **Identify**: Detect the incident via audit logs or alerts
2. **Contain**: Disable affected API keys, block IPs
3. **Eradicate**: Remove malicious hooks, update credentials
4. **Recover**: Restore from trusted backup
5. **Report**: Document incident and notify stakeholders

### Emergency Commands

```bash
# Disable all external access
tokman config set mode offline

# Revoke all API keys
tokman config reset api_keys

# Rotate secrets
tokman config rotate-secrets
```

## Compliance Considerations

### Data Retention

```toml
[compliance]
data_retention_days = 30
anonymize_after_days = 7
delete_after_days = 90
```

### GDPR Compliance

- User data is stored locally by default
- No external data transmission without explicit consent
- Right to deletion supported via `tokman clean --all`

## Security Checklist

- [ ] TLS enabled for all endpoints
- [ ] API keys rotated in last 90 days
- [ ] Rate limiting configured
- [ ] Audit logging enabled
- [ ] Hook integrity verified
- [ ] Database encrypted
- [ ] Dependencies scanned for vulnerabilities
- [ ] Network policies in place
- [ ] Incident response plan documented

## Reporting Security Issues

If you discover a security vulnerability, please report it responsibly:

1. Email: security@tokman.dev
2. Do not disclose publicly until patched
3. Include steps to reproduce
4. We aim to respond within 48 hours

## Security Updates

Subscribe to security advisories:
- GitHub Security Advisories: https://github.com/GrayCodeAI/tokman/security/advisories
- Release notes include security fixes
