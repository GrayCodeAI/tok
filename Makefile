.PHONY: build build-small build-all test test-race test-cover lint typecheck check install clean help

# Binary name
BINARY=tokman
BUILD_DIR=cmd/tokman

# Version from git tag (e.g., v0.1.0 -> 0.1.0) or "dev"
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null | sed 's/^v//' || echo "dev")

# Build flags with version injection
LDFLAGS=-ldflags="-s -w -X 'github.com/GrayCodeAI/tokman/internal/commands/shared.Version=$(VERSION)'"

# Go flags
GOFLAGS=CGO_ENABLED=0

## build: Build standard binary
build:
	$(GOFLAGS) go build -o $(BINARY) $(LDFLAGS) ./$(BUILD_DIR)

## build-small: Build optimized small binary (with UPX if available)
build-small:
	$(GOFLAGS) go build -o $(BINARY) $(LDFLAGS) -gcflags="-trimpath" ./$(BUILD_DIR)
	@command -v upx >/dev/null 2>&1 && upx --best --lzma $(BINARY) 2>/dev/null || echo "UPX not found, skipping compression"

## build-tiny: Build ultra-optimized binary
build-tiny:
	$(GOFLAGS) go build -tags netgo -o $(BINARY) $(LDFLAGS) \
		-gcflags="-trimpath" \
		-asmflags="-trimpath" \
		./$(BUILD_DIR)
	@command -v upx >/dev/null 2>&1 && upx --ultra-brute $(BINARY) 2>/dev/null || echo "UPX not found, skipping compression"

## build-all: Build for all platforms (Linux, macOS, Windows)
build-all:
	@echo "Building for all platforms..."
	GOOS=linux GOARCH=amd64 $(GOFLAGS) go build -o tokman-linux-amd64 $(LDFLAGS) ./$(BUILD_DIR)
	GOOS=linux GOARCH=arm64 $(GOFLAGS) go build -o tokman-linux-arm64 $(LDFLAGS) ./$(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 $(GOFLAGS) go build -o tokman-darwin-amd64 $(LDFLAGS) ./$(BUILD_DIR)
	GOOS=darwin GOARCH=arm64 $(GOFLAGS) go build -o tokman-darwin-arm64 $(LDFLAGS) ./$(BUILD_DIR)
	GOOS=windows GOARCH=amd64 $(GOFLAGS) go build -o tokman-windows-amd64.exe $(LDFLAGS) ./$(BUILD_DIR)
	@echo "Done! Created binaries:"
	@ls -lh tokman-*

## build-simd: Build with SIMD optimizations (Go 1.26+)
build-simd:
	$(GOFLAGS) go build -tags simd -o $(BINARY) $(LDFLAGS) ./$(BUILD_DIR)

## test: Run tests
test:
	go test -cover ./...

## test-race: Run tests with race detector
test-race:
	go test -race ./...

## test-cover: Run tests with coverage report
test-cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

## test-verbose: Run tests with verbose output
test-verbose:
	go test -v -cover ./...

## lint: Run linters
lint:
	go vet ./...
	@command -v golangci-lint >/dev/null 2>&1 && golangci-lint run || echo "golangci-lint not installed, skipping"

## typecheck: Run type checking
typecheck:
	go vet ./...

## check: Run all checks (fmt, vet, typecheck, lint, test)
check: typecheck lint test
	@echo "All checks passed!"

## install: Install binary to ~/.local/bin
install: build
	@mkdir -p $(HOME)/.local/bin
	@cp $(BINARY) $(HOME)/.local/bin/$(BINARY)
	@echo "Installed $(BINARY) to $(HOME)/.local/bin/$(BINARY)"
	@echo "Make sure $(HOME)/.local/bin is in your PATH"

## install-global: Install binary to /usr/local/bin (requires sudo)
install-global: build
	@sudo cp $(BINARY) /usr/local/bin/$(BINARY)
	@echo "Installed $(BINARY) to /usr/local/bin/$(BINARY)"

## clean: Clean build artifacts
clean:
	rm -f $(BINARY) tokman-* coverage.out coverage.html
	go clean -testcache
	@echo "Cleaned build artifacts"

## benchmark: Run benchmarks
benchmark:
	go test -bench=. -benchmem ./...

## version: Show version
version:
	@echo "Version: $(VERSION)"

## help: Show this help
help:
	@echo "TokMan Makefile"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'
