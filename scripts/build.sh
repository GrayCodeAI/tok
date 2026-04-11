#!/bin/bash
set -e

# TokMan Build Script
# Usage: ./scripts/build.sh [target]

PROJECT_NAME="tokman"
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS="-X main.version=${VERSION} -X main.buildTime=${BUILD_TIME}"

echo "Building ${PROJECT_NAME}..."
echo "Version: ${VERSION}"
echo "Build Time: ${BUILD_TIME}"

# Create build directory
mkdir -p build/dist

# Build main binary
echo "Building main binary..."
go build -ldflags "${LDFLAGS}" -o build/dist/tokman ./cmd/tokman

# Build microservices if requested
if [ "$1" == "all" ] || [ "$1" == "services" ]; then
    echo "Building microservices..."

    # API Gateway
    echo "  -> api-gateway"
    go build -ldflags "${LDFLAGS}" -o build/dist/api-gateway ./services/api-gateway/cmd

    # Compression Service
    echo "  -> compression-service"
    go build -ldflags "${LDFLAGS}" -o build/dist/compression-service ./services/compression-service/cmd

    # Analytics Service
    echo "  -> analytics-service"
    go build -ldflags "${LDFLAGS}" -o build/dist/analytics-service ./services/analytics-service/cmd

    # Security Service
    echo "  -> security-service"
    go build -ldflags "${LDFLAGS}" -o build/dist/security-service ./services/security-service/cmd

    # Config Service
    echo "  -> config-service"
    go build -ldflags "${LDFLAGS}" -o build/dist/config-service ./services/config-service/cmd
fi

echo "Build complete! Binaries in build/dist/"
