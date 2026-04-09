# Build stage
FROM golang:1.26-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w \
    -X github.com/GrayCodeAI/tokman/internal/commands/shared.Version=$(git describe --tags --always) \
    -X github.com/GrayCodeAI/tokman/internal/commands/shared.BuildDate=$(date -u +%Y-%m-%d) \
    -X github.com/GrayCodeAI/tokman/internal/commands/shared.GitCommit=$(git rev-parse --short HEAD)" \
    -o tokman cmd/tokman/main.go

# Runtime stage - minimal scratch image
FROM scratch

# Copy certificates for HTTPS
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy binary
COPY --from=builder /build/tokman /tokman

# Set environment variables
ENV TOKMAN_CONFIG=/config/config.toml

# Expose dashboard port
EXPOSE 8080

# Use non-root user (not applicable in scratch, but good practice)
# In scratch, we run as root but with minimal attack surface

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD ["/tokman", "status"] || exit 1

# Set entrypoint
ENTRYPOINT ["/tokman"]

# Default command
CMD ["--help"]
