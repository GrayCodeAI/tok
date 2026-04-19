# Deployment Guide

## Overview

Tok can be deployed in several ways depending on your use case:

1. **Single User** - Local installation on developer machine
2. **Team** - Shared configuration with individual installations
3. **CI/CD** - Automated pipeline integration
4. **Enterprise** - Managed deployment across organization

---

## Single User Deployment

### Quick Install

```bash
# Option 1: Go install
go install github.com/lakshmanpatel/tok/cmd/tok@latest

# Option 2: From source
git clone https://github.com/lakshmanpatel/tok.git
cd tok && make build
sudo mv tok /usr/local/bin/

# Option 3: Pre-built binary
curl -fsSL https://github.com/lakshmanpatel/tok/releases/latest/download/tok_$(uname -s)_$(uname -m).tar.gz | tar xz
sudo mv tok /usr/local/bin/
```

### Post-Install Setup

```bash
# 1. Verify installation
tok --version
tok doctor

# 2. Initialize for your AI tool
tok init -g                    # Claude Code
tok init -g --cursor           # Cursor
tok init -g --copilot          # GitHub Copilot
tok init --all                 # All detected tools

# 3. Configure (optional)
mkdir -p ~/.config/tok
tok config init

# 4. Test
tok git status
tok ls .
```

### Configuration

Default config location: `~/.config/tok/config.toml`

```toml
[tracking]
enabled = true
database_path = "~/.local/share/tok/tok.db"

[filter]
mode = "minimal"

[pipeline]
max_context_tokens = 2000000
enable_entropy = true
enable_compaction = true
enable_h2o = true
enable_attention_sink = true

[hooks]
excluded_commands = []

[dashboard]
port = 8080
enabled = false
```

### Environment Variables

```bash
# Add to ~/.bashrc or ~/.zshrc
export TOK_MODE=minimal
export TOK_BUDGET=2000
export TOK_PRESET=balanced
```

---

## Team Deployment

### Shared Configuration

Create a team config file and distribute:

```bash
# 1. Create team config
cat > tok-team.toml << 'EOF'
[filter]
mode = "minimal"

[pipeline]
max_context_tokens = 2000000
enable_entropy = true
enable_compaction = true

[tracking]
enabled = true
EOF

# 2. Distribute to team
# Each developer copies to ~/.config/tok/config.toml
```

### Team Setup Script

```bash
#!/bin/bash
# team-setup.sh - Run on each developer machine

set -e

echo "Setting up Tok for team..."

# Install
go install github.com/lakshmanpatel/tok/cmd/tok@latest

# Copy team config
mkdir -p ~/.config/tok
cp tok-team.toml ~/.config/tok/config.toml

# Initialize for detected AI tools
tok init --all

# Verify
tok doctor

echo "Tok setup complete!"
```

---

## CI/CD Integration

### GitHub Actions

```yaml
# .github/workflows/tok.yml
name: Tok CI Integration

on: [push, pull_request]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.24'
      
      - name: Install Tok
        run: go install github.com/lakshmanpatel/tok/cmd/tok@latest
      
      - name: Run tests with Tok
        run: tok go test ./... 2>&1 | head -50
      
      - name: Check token usage
        run: tok stats --json > tok-report.json
      
      - name: Upload report
        uses: actions/upload-artifact@v4
        with:
          name: tok-report
          path: tok-report.json
```

### GitLab CI

```yaml
# .gitlab-ci.yml
tok:
  image: golang:1.24
  stage: test
  script:
    - go install github.com/lakshmanpatel/tok/cmd/tok@latest
    - tok go test ./...
    - tok stats
  artifacts:
    reports:
      metrics: tok-metrics.txt
```

### Pre-Commit Hook

```bash
#!/bin/bash
# .git/hooks/pre-commit

# Run Tok doctor before each commit
if command -v tok &> /dev/null; then
    tok doctor --quiet
    if [ $? -ne 0 ]; then
        echo "Tok: Hook integrity check failed!"
        echo "Run 'tok doctor' for details."
        exit 1
    fi
fi
```

---

## Binary Distribution

### GoReleaser Configuration

```yaml
# .goreleaser.yml
version: 2

before:
  hooks:
    - go mod tidy
    - go generate ./...

builds:
  - main: ./cmd/tok
    binary: tok
    env:
      - CGO_ENABLED=0
    goos:
      - linux
      - darwin
      - windows
      - freebsd
    goarch:
      - amd64
      - arm64
      - "386"
    ldflags:
      - -s -w
      - -X main.version={{.Version}}
      - -X main.commit={{.Commit}}
      - -X main.date={{.Date}}
    ignore:
      - goos: windows
        goarch: arm64

archives:
  - formats: ['tar.gz']
    format_overrides:
      - goos: windows
        formats: ['zip']
    name_template: >-
      {{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}

checksum:
  name_template: 'checksums.txt'

snapshot:
  version_template: "{{ incpatch .Version }}-next"

changelog:
  sort: asc
  filters:
    exclude:
      - '^docs:'
      - '^test:'
      - '^chore:'

brews:
  - repository:
      owner: lakshmanpatel
      name: homebrew-tok
    homepage: "https://github.com/lakshmanpatel/tok"
    description: "Token-aware CLI proxy with practical 20-layer compression pipeline"
    license: "MIT"
    test: |
      system "#{bin}/tok", "--version"

nfpms:
  - package_name: tok
    vendor: GrayCode AI
    homepage: https://github.com/lakshmanpatel/tok
    maintainer: GrayCode AI <maintainers@graycode.ai>
    description: Token-aware CLI proxy for AI coding assistants
    license: MIT
    formats:
      - deb
      - rpm
```

### Release Process

```bash
# 1. Tag the release
git tag -a v0.29.0 -m "Release v0.29.0"
git push origin v0.29.0

# 2. Run GoReleaser
goreleaser release --clean

# 3. Verify artifacts
ls dist/
# tok_0.29.0_linux_amd64.tar.gz
# tok_0.29.0_darwin_arm64.tar.gz
# tok_0.29.0_windows_amd64.zip
# checksums.txt
```

---

## Upgrade Process

### Manual Upgrade

```bash
# Check current version
tok --version

# Upgrade via Go
go install github.com/lakshmanpatel/tok/cmd/tok@latest

# Upgrade via Homebrew (when available)
brew upgrade tok

# Verify
tok --version
tok doctor
```

### Automated Upgrade Check

```bash
# Check for updates
tok version --check-update

# Auto-upgrade (when available)
tok upgrade
```

---

## Rollback

### Manual Rollback

```bash
# Install specific version
go install github.com/lakshmanpatel/tok/cmd/tok@v0.28.0

# Or download specific release binary
curl -fsSL https://github.com/lakshmanpatel/tok/releases/download/v0.28.0/tok_0.28.0_$(uname -s)_$(uname -m).tar.gz | tar xz
sudo mv tok /usr/local/bin/
```

### Hook Rollback

```bash
# Restore hooks to previous state
tok init --uninstall
tok init -g  # Re-install fresh hooks
```

---

## Uninstallation

```bash
# 1. Remove hooks
tok init --uninstall

# 2. Remove binary
sudo rm /usr/local/bin/tok
# Or: brew uninstall tok

# 3. Remove config (optional)
rm -rf ~/.config/tok

# 4. Remove data (optional)
rm -rf ~/.local/share/tok

# 5. Remove Go cache
go clean -i github.com/lakshmanpatel/tok/...
```

---

## Monitoring

### Health Checks

```bash
# Basic health check
tok doctor

# Detailed audit
tok hook-audit

# Check hook integrity
tok verify
```

### Metrics

```bash
# View stats
tok stats

# Export as JSON
tok stats --json > metrics.json

# Token savings report
tok gain
```

### Dashboard

```bash
# Start dashboard
tok dashboard --port 8080

# Open in browser
open http://localhost:8080
```

---

## Troubleshooting

### Common Issues

**Binary not found:**
```bash
# Check PATH
which tok
echo $PATH

# Add to PATH
export PATH="$HOME/go/bin:$PATH"
```

**Hooks not working:**
```bash
# Reinstall hooks
tok init --uninstall
tok init -g

# Check hook integrity
tok doctor
```

**Database issues:**
```bash
# Reset database
rm ~/.local/share/tok/tok.db
tok status  # Will recreate
```

**Permission issues:**
```bash
# Fix permissions
chmod 755 $(which tok)
chmod 700 ~/.config/tok
chmod 600 ~/.config/tok/config.toml
```

---

## Security Considerations

1. **File Permissions:** Config files should be 0600 (user-only)
2. **Hook Scripts:** Review hooks before installing
3. **Database:** Local SQLite DB should not contain sensitive data
4. **Telemetry:** Opt-in only; no secrets transmitted
5. **Updates:** Keep Tok updated for security fixes

See [SECURITY.md](../SECURITY.md) for full security policy.
