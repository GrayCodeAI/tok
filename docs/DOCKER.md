# Docker Guide for TokMan

## Quick Start

### Build the Docker Image

```bash
# Build production image
docker build -t tokman:latest .

# Build development image
docker build -f Dockerfile.dev -t tokman:dev .
```

### Run TokMan in Docker

```bash
# Show help
docker run --rm tokman:latest

# Run a command
docker run --rm tokman:latest version

# Process a file
docker run --rm -v $(pwd):/workspace tokman:latest summary /workspace/README.md
```

### Using Docker Compose

```bash
# Start services
docker-compose up -d

# Run TokMan commands
docker-compose exec tokman tokman git status

# Access dashboard
docker-compose up tokman-dashboard
# Then open http://localhost:8080

# Development environment
docker-compose up -d tokman-dev
docker-compose exec tokman-dev sh
```

## Production Deployment

### Minimal Image (< 15 MB)

The production Dockerfile uses multi-stage builds:
- **Builder stage**: Compiles with full Go toolchain
- **Final stage**: Scratch image with only the binary

```bash
# Build for production
docker build \
  --build-arg VERSION=$(git describe --tags --always) \
  -t tokman:prod \
  -f Dockerfile .

# Run production container
docker run --rm tokman:prod --help
```

### Image Sizes

| Image | Size | Use Case |
|-------|------|----------|
| `tokman:latest` | ~15 MB | Production |
| `tokman:dev` | ~500 MB | Development |

## Configuration

### Environment Variables

```bash
# Set config path
docker run -e TOKMAN_CONFIG=/config/config.toml tokman:latest

# Enable verbose mode
docker run -e TOKMAN_VERBOSE=1 tokman:latest

# Set data directory
docker run -e TOKMAN_DATA_DIR=/data tokman:latest
```

### Volume Mounts

```bash
# Mount config
docker run -v /host/config:/config/tokman:ro tokman:latest

# Mount data for persistence
docker run -v tokman-data:/data/tokman tokman:latest

# Mount workspace
docker run -v $(pwd):/workspace tokman:latest
```

## Use Cases

### 1. CI/CD Integration

```dockerfile
# In your CI pipeline
FROM tokman:latest as compressor
COPY . /workspace
RUN tokman summary /workspace/output.txt > /workspace/compressed.txt

FROM your-app:latest
COPY --from=compressor /workspace/compressed.txt /app/
```

### 2. Pre-commit Hook

```yaml
# .pre-commit-config.yaml
repos:
  - repo: local
    hooks:
      - id: tokman-compress
        name: Compress with TokMan
        entry: docker run --rm -v $(pwd):/workspace tokman:latest summary
        language: system
        files: '\.(log|txt)$'
```

### 3. GitHub Actions

```yaml
name: Compress Output
on: [push]
jobs:
  compress:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Compress logs
        run: |
          docker run --rm -v $(pwd):/workspace \
            tokman:latest summary /workspace/build.log
```

### 4. Local Development

```bash
# Start dev environment
docker-compose up -d tokman-dev

# Run tests
docker-compose exec tokman-dev go test ./...

# Build binary
docker-compose exec tokman-dev go build -o tokman ./cmd/tokman

# Lint
docker-compose exec tokman-dev golangci-lint run
```

## Advanced Usage

### Custom Config

Create `docker-config.toml`:

```toml
[filter]
mode = "aggressive"

[tracking]
enabled = true

[dashboard]
port = 8080
host = "0.0.0.0"
```

Run with custom config:

```bash
docker run \
  -v $(pwd)/docker-config.toml:/config/tokman/config.toml:ro \
  tokman:latest
```

### Multi-Stage Compression

```bash
# Stage 1: Initial compression
docker run --rm -v $(pwd):/workspace tokman:latest \
  summary --mode minimal /workspace/input.txt > /tmp/stage1.txt

# Stage 2: Aggressive compression
docker run --rm -v /tmp:/workspace tokman:latest \
  summary --mode aggressive /workspace/stage1.txt
```

### Network Access

```bash
# For MCP server
docker run -p 8080:8080 tokman:latest mcp-server

# For dashboard
docker run -p 8080:8080 tokman:latest dashboard
```

## Troubleshooting

### Permission Issues

```bash
# Run as current user
docker run --rm -u $(id -u):$(id -g) -v $(pwd):/workspace tokman:latest
```

### Config Not Found

```bash
# Check config path
docker run --rm -e TOKMAN_CONFIG=/config/tokman/config.toml \
  -v $(pwd)/config:/config/tokman:ro \
  tokman:latest config show
```

### Database Persistence

```bash
# Use named volume for database
docker run --rm \
  -v tokman-db:/data/tokman \
  tokman:latest archive list
```

## Best Practices

1. **Use specific tags** in production: `tokman:v0.28.2` instead of `tokman:latest`
2. **Mount read-only** when possible: `-v config:/config:ro`
3. **Use named volumes** for data persistence
4. **Multi-stage builds** for smaller images
5. **Non-root user** in production (already configured)

## Health Checks

The production image includes a health check:

```bash
# Check container health
docker ps --format "table {{.Names}}\t{{.Status}}"

# Manual health check
docker exec tokman tokman --version
```

## Building for Different Architectures

```bash
# Build for ARM64 (Apple Silicon)
docker build --platform linux/arm64 -t tokman:arm64 .

# Build for AMD64
docker build --platform linux/amd64 -t tokman:amd64 .

# Multi-arch build
docker buildx build --platform linux/amd64,linux/arm64 -t tokman:latest .
```

## Registry Push

```bash
# Tag for registry
docker tag tokman:latest ghcr.io/graycodeai/tokman:latest
docker tag tokman:latest ghcr.io/graycodeai/tokman:v0.28.2

# Push to registry
docker push ghcr.io/graycodeai/tokman:latest
docker push ghcr.io/graycodeai/tokman:v0.28.2
```

---

**Related:** See [INSTALLATION.md](./INSTALLATION.md) for other installation methods.
