# Deployment Guide

## Overview

TokMan can be deployed in several ways depending on your use case:

1. **Single User** - Local installation on developer machine
2. **Team** - Shared configuration with individual installations
3. **CI/CD** - Automated pipeline integration
4. **Docker** - Containerized deployment
5. **Enterprise** - Managed deployment across organization

---

## Single User Deployment

### Quick Install

```bash
# Option 1: Go install
go install github.com/GrayCodeAI/tokman/cmd/tokman@latest

# Option 2: From source
git clone https://github.com/GrayCodeAI/tokman.git
cd tokman && make build
sudo mv tokman /usr/local/bin/

# Option 3: Pre-built binary
curl -fsSL https://github.com/GrayCodeAI/tokman/releases/latest/download/tokman_$(uname -s)_$(uname -m).tar.gz | tar xz
sudo mv tokman /usr/local/bin/
```

### Post-Install Setup

```bash
# 1. Verify installation
tokman --version
tokman doctor

# 2. Initialize for your AI tool
tokman init -g                    # Claude Code
tokman init -g --cursor           # Cursor
tokman init -g --copilot          # GitHub Copilot
tokman init --all                 # All detected tools

# 3. Configure (optional)
mkdir -p ~/.config/tokman
tokman config init

# 4. Test
tokman git status
tokman ls .
```

### Configuration

Default config location: `~/.config/tokman/config.toml`

```toml
[tracking]
enabled = true
database_path = "~/.local/share/tokman/tokman.db"

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
export TOKMAN_MODE=minimal
export TOKMAN_BUDGET=2000
export TOKMAN_PRESET=balanced
```

---

## Team Deployment

### Shared Configuration

Create a team config file and distribute:

```bash
# 1. Create team config
cat > tokman-team.toml << 'EOF'
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
# Each developer copies to ~/.config/tokman/config.toml
```

### Team Setup Script

```bash
#!/bin/bash
# team-setup.sh - Run on each developer machine

set -e

echo "Setting up TokMan for team..."

# Install
go install github.com/GrayCodeAI/tokman/cmd/tokman@latest

# Copy team config
mkdir -p ~/.config/tokman
cp tokman-team.toml ~/.config/tokman/config.toml

# Initialize for detected AI tools
tokman init --all

# Verify
tokman doctor

echo "TokMan setup complete!"
```

---

## CI/CD Integration

### GitHub Actions

```yaml
# .github/workflows/tokman.yml
name: TokMan CI Integration

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
      
      - name: Install TokMan
        run: go install github.com/GrayCodeAI/tokman/cmd/tokman@latest
      
      - name: Run tests with TokMan
        run: tokman go test ./... 2>&1 | head -50
      
      - name: Check token usage
        run: tokman stats --json > tokman-report.json
      
      - name: Upload report
        uses: actions/upload-artifact@v4
        with:
          name: tokman-report
          path: tokman-report.json
```

### GitLab CI

```yaml
# .gitlab-ci.yml
tokman:
  image: golang:1.24
  stage: test
  script:
    - go install github.com/GrayCodeAI/tokman/cmd/tokman@latest
    - tokman go test ./...
    - tokman stats
  artifacts:
    reports:
      metrics: tokman-metrics.txt
```

### Pre-Commit Hook

```bash
#!/bin/bash
# .git/hooks/pre-commit

# Run TokMan doctor before each commit
if command -v tokman &> /dev/null; then
    tokman doctor --quiet
    if [ $? -ne 0 ]; then
        echo "TokMan: Hook integrity check failed!"
        echo "Run 'tokman doctor' for details."
        exit 1
    fi
fi
```

---

## Docker Deployment

### Dockerfile

```dockerfile
# Multi-stage build
FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /tokman ./cmd/tokman

# Runtime image
FROM alpine:3.19

RUN apk --no-cache add git bash

COPY --from=builder /tokman /usr/local/bin/tokman

# Create default config
RUN mkdir -p /root/.config/tokman
COPY config/default.toml /root/.config/tokman/config.toml

ENTRYPOINT ["tokman"]
CMD ["--help"]
```

### Docker Compose

```yaml
# docker-compose.yml
version: '3.8'

services:
  tokman:
    build: .
    volumes:
      - ./:/workspace
      - tokman-data:/root/.local/share/tokman
      - tokman-config:/root/.config/tokman
    working_dir: /workspace
    environment:
      - TOKMAN_MODE=minimal
      - TOKMAN_BUDGET=2000

  tokman-dashboard:
    build: .
    command: ["dashboard", "--port", "8080"]
    ports:
      - "8080:8080"
    volumes:
      - tokman-data:/root/.local/share/tokman

volumes:
  tokman-data:
  tokman-config:
```

### Docker Usage

```bash
# Build image
docker build -t tokman .

# Run command
docker run --rm -v $(pwd):/workspace tokman git status

# Interactive shell
docker run --rm -it -v $(pwd):/workspace tokman /bin/sh

# Dashboard
docker run -d -p 8080:8080 tokman dashboard --port 8080
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
  - main: ./cmd/tokman
    binary: tokman
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
      owner: GrayCodeAI
      name: homebrew-tokman
    homepage: "https://github.com/GrayCodeAI/tokman"
    description: "Token-aware CLI proxy with 31-stage core compression pipeline"
    license: "MIT"
    test: |
      system "#{bin}/tokman", "--version"

nfpms:
  - package_name: tokman
    vendor: GrayCode AI
    homepage: https://github.com/GrayCodeAI/tokman
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
# tokman_0.29.0_linux_amd64.tar.gz
# tokman_0.29.0_darwin_arm64.tar.gz
# tokman_0.29.0_windows_amd64.zip
# checksums.txt
```

---

## Upgrade Process

### Manual Upgrade

```bash
# Check current version
tokman --version

# Upgrade via Go
go install github.com/GrayCodeAI/tokman/cmd/tokman@latest

# Upgrade via Homebrew (when available)
brew upgrade tokman

# Verify
tokman --version
tokman doctor
```

### Automated Upgrade Check

```bash
# Check for updates
tokman version --check-update

# Auto-upgrade (when available)
tokman upgrade
```

---

## Rollback

### Manual Rollback

```bash
# Install specific version
go install github.com/GrayCodeAI/tokman/cmd/tokman@v0.28.0

# Or download specific release binary
curl -fsSL https://github.com/GrayCodeAI/tokman/releases/download/v0.28.0/tokman_0.28.0_$(uname -s)_$(uname -m).tar.gz | tar xz
sudo mv tokman /usr/local/bin/
```

### Hook Rollback

```bash
# Restore hooks to previous state
tokman init --uninstall
tokman init -g  # Re-install fresh hooks
```

---

## Uninstallation

```bash
# 1. Remove hooks
tokman init --uninstall

# 2. Remove binary
sudo rm /usr/local/bin/tokman
# Or: brew uninstall tokman

# 3. Remove config (optional)
rm -rf ~/.config/tokman

# 4. Remove data (optional)
rm -rf ~/.local/share/tokman

# 5. Remove Go cache
go clean -i github.com/GrayCodeAI/tokman/...
```

---

## Monitoring

### Health Checks

```bash
# Basic health check
tokman doctor

# Detailed audit
tokman hook-audit

# Check hook integrity
tokman verify
```

### Metrics

```bash
# View stats
tokman stats

# Export as JSON
tokman stats --json > metrics.json

# Token savings report
tokman gain
```

### Dashboard

```bash
# Start dashboard
tokman dashboard --port 8080

# Open in browser
open http://localhost:8080
```

---

## Troubleshooting

### Common Issues

**Binary not found:**
```bash
# Check PATH
which tokman
echo $PATH

# Add to PATH
export PATH="$HOME/go/bin:$PATH"
```

**Hooks not working:**
```bash
# Reinstall hooks
tokman init --uninstall
tokman init -g

# Check hook integrity
tokman doctor
```

**Database issues:**
```bash
# Reset database
rm ~/.local/share/tokman/tokman.db
tokman status  # Will recreate
```

**Permission issues:**
```bash
# Fix permissions
chmod 755 $(which tokman)
chmod 700 ~/.config/tokman
chmod 600 ~/.config/tokman/config.toml
```

---

## Security Considerations

1. **File Permissions:** Config files should be 0600 (user-only)
2. **Hook Scripts:** Review hooks before installing
3. **Database:** Local SQLite DB should not contain sensitive data
4. **Telemetry:** Opt-in only; no secrets transmitted
5. **Updates:** Keep TokMan updated for security fixes

See [SECURITY.md](../SECURITY.md) for full security policy.
