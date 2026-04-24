.PHONY: build build-small build-all test test-race test-cover lint typecheck check install clean help

# Binary name
BINARY=tok
BUILD_DIR=cmd/tok

# Version from git tag (e.g., v0.1.0 -> 0.1.0) or "dev"
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null | sed 's/^v//' || echo "dev")

# Build flags with version injection
LDFLAGS=-ldflags="-s -w -X 'github.com/GrayCodeAI/tok/internal/version.Version=$(VERSION)'"

# Go flags
GOFLAGS=CGO_ENABLED=0
GOBIN_DIR=$(shell go env GOPATH)/bin
GO_CMD=PATH="$(shell go env GOROOT)/bin:$$PATH" go

## build: Build standard binary
build:
	$(GOFLAGS) go build -o $(BINARY) $(LDFLAGS) ./$(BUILD_DIR)

## docker-build: Build Docker image
docker-build:
	docker build -t tok:latest .

## docker-build-dev: Build development Docker image
docker-build-dev:
	docker build -f Dockerfile.dev -t tok:dev .

## docker-run: Run Tok in Docker
docker-run:
	docker run --rm -v $(PWD):/workspace tok:latest

## docker-test: Run tests in Docker
docker-test:
	docker run --rm -v $(PWD):/app tok:dev go test ./...

## docker-push: Push Docker image to registry
docker-push:
	docker tag tok:latest ghcr.io/graycodeai/tok:latest
	docker tag tok:latest ghcr.io/graycodeai/tok:$(VERSION)
	docker push ghcr.io/graycodeai/tok:latest
	docker push ghcr.io/graycodeai/tok:$(VERSION)

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
	GOOS=linux GOARCH=amd64 $(GOFLAGS) go build -o tok-linux-amd64 $(LDFLAGS) ./$(BUILD_DIR)
	GOOS=linux GOARCH=arm64 $(GOFLAGS) go build -o tok-linux-arm64 $(LDFLAGS) ./$(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 $(GOFLAGS) go build -o tok-darwin-amd64 $(LDFLAGS) ./$(BUILD_DIR)
	GOOS=darwin GOARCH=arm64 $(GOFLAGS) go build -o tok-darwin-arm64 $(LDFLAGS) ./$(BUILD_DIR)
	GOOS=windows GOARCH=amd64 $(GOFLAGS) go build -o tok-windows-amd64.exe $(LDFLAGS) ./$(BUILD_DIR)
	@echo "Done! Created binaries:"
	@ls -lh tok-*

## build-simd: Build with SIMD optimizations (Go 1.26+)
build-simd:
	$(GOFLAGS) go build -tags simd -o $(BINARY) $(LDFLAGS) ./$(BUILD_DIR)

## test: Run tests
test:
	$(GO_CMD) test -cover ./...

## test-race: Run tests with race detector
test-race:
	$(GO_CMD) test -race ./...

## test-cover: Run tests with coverage report
test-cover:
	$(GO_CMD) test -coverprofile=coverage.out ./...
	$(GO_CMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

## test-verbose: Run tests with verbose output
test-verbose:
	$(GO_CMD) test -v -cover ./...

## lint: Run linters
lint:
	$(GO_CMD) vet ./...
	@command -v golangci-lint >/dev/null 2>&1 || (echo "golangci-lint not installed" && exit 1)
	@mkdir -p .cache/go-build .cache/golangci
	GOCACHE=$(CURDIR)/.cache/go-build GOLANGCI_LINT_CACHE=$(CURDIR)/.cache/golangci golangci-lint run

## typecheck: Run type checking
typecheck:
	$(GO_CMD) vet ./...

## fmt: Format Go code
fmt:
	gofmt -w .
	goimports -w .
	@echo "Formatted all Go files"

## check: Run all checks (fmt, vet, typecheck, lint, test)
check: fmt lint test
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
	rm -f $(BINARY) tok-* coverage.out coverage.html
	go clean -testcache
	@echo "Cleaned build artifacts"

## security: Run security scans
security:
	@echo "Running security scans..."
	@command -v $(GOBIN_DIR)/gosec >/dev/null 2>&1 || (echo "gosec not installed; run: go install github.com/securego/gosec/v2/cmd/gosec@latest" && exit 1)
	@command -v $(GOBIN_DIR)/govulncheck >/dev/null 2>&1 || (echo "govulncheck not installed; run: go install golang.org/x/vuln/cmd/govulncheck@latest" && exit 1)
	$(GOBIN_DIR)/gosec -severity high -confidence high -fmt json -out security-report.json ./...
	$(GOBIN_DIR)/govulncheck ./...
	@echo "Security scans complete"

## coverage: Generate and view coverage report
coverage: test-cover
	@echo "Coverage report: coverage.html"

## benchmark: Run benchmarks
benchmark:
	go test -bench=. -benchmem ./...

## benchmark-adaptive: Run adaptive benchmark compare and save report
benchmark-adaptive:
	@mkdir -p artifacts
	go test -run '^$$' -bench BenchmarkPipelineAdaptiveCompare -benchmem ./internal/filter | tee artifacts/benchmark-adaptive.txt

## benchmark-suite: Run scenario benchmark suite and save report
benchmark-suite:
	@mkdir -p artifacts
	go test -run TestBenchmarkSuiteScenarios -v ./internal/filter | tee artifacts/benchmark-suite.txt

## benchmark-tui: Run TUI performance benchmarks (chart, table, full frame)
benchmark-tui:
	./scripts/tui-bench.sh

## ablation: Run ablation baseline and save report
ablation:
	@mkdir -p artifacts
	go test -run TestLayerAblationBasic -v ./internal/filter | tee artifacts/ablation-baseline.txt

## eval-adaptive: Run local baseline vs adaptive evaluation report
eval-adaptive:
	./scripts/eval_adaptive.sh 5

## deps: Download and verify dependencies
deps:
	go mod download
	go mod verify
	go mod tidy

## outdated: Check for outdated dependencies
outdated:
	@command -v go-mod-outdated >/dev/null 2>&1 && go-mod-outdated || echo "go-mod-outdated not installed. Run: go install github.com/psampaz/go-mod-outdated@latest"

## generate: Run go generate
generate:
	go generate ./...

## ci: Run CI checks locally
ci: deps fmt lint test-race security
	@echo "All CI checks passed!"

## version: Show version
version:
	@echo "Version: $(VERSION)"

## help: Show this help
help:
	@echo "Tok Makefile"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'
