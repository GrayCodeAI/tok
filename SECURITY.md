# Security Policy

## Supported Versions

We actively support the following versions of Tok with security updates:

| Version | Supported          |
| ------- | ------------------ |
| latest  | :white_check_mark: |
| 0.28.x  | :white_check_mark: |
| < 0.28  | :x:                |

## Reporting a Vulnerability

**Please do not report security vulnerabilities through public GitHub issues.**

We take the security of Tok seriously. If you discover a security vulnerability, please follow these steps:

### 1. Private Disclosure

Send an email to **security@graycode.ai** (or the maintainer's email) with:

- **Subject:** `[SECURITY] Brief description of the vulnerability`
- **Description:** Detailed description of the vulnerability
- **Impact:** What could an attacker do with this vulnerability?
- **Steps to Reproduce:** Detailed steps to reproduce the issue
- **Proof of Concept:** If possible, include a minimal PoC (code, commands, etc.)
- **Suggested Fix:** If you have ideas for how to fix it
- **Your Contact Info:** How we can reach you for follow-up

### 2. What to Expect

- **Acknowledgment:** We will acknowledge receipt within **24 hours**
- **Initial Assessment:** We will provide an initial assessment within **72 hours**
- **Status Updates:** We will keep you updated on our progress
- **Fix Timeline:** Critical vulnerabilities will be addressed within **7 days**
- **Public Disclosure:** We will coordinate disclosure timing with you
- **Credit:** We will credit you in the security advisory (unless you prefer to remain anonymous)

### 3. Disclosure Policy

- We follow **coordinated disclosure**
- We will not publicly disclose the vulnerability until a fix is available
- We ask that you do the same
- Once fixed, we will publish a security advisory
- We will notify affected users
- We will credit the reporter (with permission)

## Security Features

Tok implements several security best practices:

### Hook Integrity Verification

- Hooks are verified using SHA-256 checksums
- `tok doctor` detects tampered hooks
- `tok hook-audit` provides detailed integrity reports
- Hooks can be re-verified with `tok verify`

### Input Validation

- All user inputs are validated and sanitized
- Command injection prevention via allowlist
- Path traversal protection
- SQL injection prevention (parameterized queries)

### Data Protection

- Sensitive data is not logged
- Secrets are not included in telemetry
- Config files should not contain credentials
- Use environment variables for sensitive values

### Dependency Security

- Dependencies are regularly scanned with `go mod tidy` and security scanners
- Dependabot automatically updates vulnerable dependencies
- We use minimal dependencies to reduce attack surface

### Code Security

- Static analysis with `gosec`
- Code scanning with GitHub CodeQL
- Regular security audits of critical paths
- Fuzzing for parser and filter code

## Security Best Practices for Users

### Installation

- **Verify releases:** Check GPG signatures (coming soon)
- **Use official sources:** Install from official releases, Homebrew, or verified package managers
- **Check checksums:** Verify SHA-256 checksums before installing

### Configuration

- **Protect config files:** Set appropriate permissions (0600)
- **No secrets in configs:** Use environment variables or secret managers
- **Review hooks:** Understand what hooks do before installing
- **Regular updates:** Keep Tok updated to get security fixes

### Hook Usage

- **Verify hooks:** Run `tok hook-audit` after installation
- **Trusted sources only:** Only install hooks from trusted sources
- **Review before install:** Check hook scripts before running `tok init`
- **Monitor changes:** `tok doctor` detects unauthorized modifications

### Telemetry

- **Opt-in only:** Telemetry is opt-in (disabled by default)
- **No sensitive data:** We never collect secrets, API keys, or file contents
- **Anonymous:** Telemetry data is anonymized
- **Transparent:** See what's collected in our privacy policy

### Network Security

- **HTTPS only:** MCP server (when enabled) should use TLS
- **Local only:** By default, MCP server binds to localhost
- **Authentication:** Enable authentication for MCP server
- **Firewall:** Configure firewall rules appropriately

## Known Security Considerations

### Hook Execution

**Risk:** Hooks run arbitrary code on your system

**Mitigation:**
- Hooks are only installed with explicit user consent
- Hooks are auditable (plain text shell scripts)
- Hooks can be uninstalled anytime (`tok init --uninstall`)
- Hook integrity is verified

### Command Interception

**Risk:** Tok intercepts shell commands

**Mitigation:**
- Only intercepts commands explicitly configured
- Transparent operation (hooks are visible)
- Can be bypassed with full command paths
- No sensitive data is captured

### SQLite Database

**Risk:** Command history stored in local database

**Mitigation:**
- Database is local-only (never transmitted)
- Permissions set to user-only (0600)
- Can be cleared with `tok sessions clear`
- Does not store secrets or credentials

### MCP Server (Optional)

**Risk:** Network-accessible service

**Mitigation:**
- Disabled by default
- Binds to localhost only
- Authentication available
- Rate limiting enabled
- Can be firewalled

## Security Roadmap

Future security enhancements:

- [ ] GPG signature verification for releases
- [ ] Code signing for binaries (macOS, Windows)
- [ ] Reproducible builds
- [ ] SBOM (Software Bill of Materials) generation
- [ ] Security audit by third party
- [ ] OpenSSF Best Practices Badge
- [ ] CVE monitoring and automatic alerts
- [ ] Sandboxed filter execution
- [ ] Encrypted database option
- [ ] Zero-knowledge telemetry

## Bug Bounty Program

We currently do not have a bug bounty program, but we deeply appreciate security researchers who responsibly disclose vulnerabilities. We will:

- Publicly acknowledge and credit reporters
- Fast-track fixes for reported issues
- Consider rewards for critical findings (on a case-by-case basis)

## Security Hall of Fame

We thank the following researchers for responsibly disclosing security issues:

<!-- Will be updated as reports come in -->

*No reports yet - you could be first!*

## Contact

- **Security Issues:** security@graycode.ai
- **General Issues:** Open a GitHub issue (non-security only)
- **Urgent Security:** Tag issue as [URGENT] in email subject

## Legal

We believe in responsible disclosure and will work with you in good faith. We will not pursue legal action against security researchers who:

- Make a good faith effort to avoid privacy violations and data destruction
- Do not exploit a vulnerability beyond what's necessary to demonstrate it
- Give us reasonable time to fix the issue before public disclosure
- Do not access, modify, or delete others' data

Thank you for helping keep Tok and our users safe!

---

**Last Updated:** April 7, 2026  
**Next Review:** July 7, 2026
