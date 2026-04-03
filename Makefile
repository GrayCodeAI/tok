.PHONY: build build-small build-all build-simd build-tiny test test-cover race bench lint typecheck vet fmt clean benchmark coverage tidy check check-quick ci

BINARY_NAME := tokman
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_DIR := bin

# Go 1.25+ includes SIMD by default; GOEXPERIMENT is no longer needed.
GO ?= go

# Aggressive optimization flags for smaller binary
LDFLAGS := -s -w -X github.com/GrayCodeAI/tokman/internal/commands.Version=$(VERSION)

# Standard build (stripped symbols)
build:
	$(GO) build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/tokman

# Optimized small binary (strip + compress)
build-small:
	$(GO) build -ldflags="$(LDFLAGS)" -trimpath -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/tokman
	@echo "Binary size: $$(du -h $(BUILD_DIR)/$(BINARY_NAME) | cut -f1)"

# Tiny binary with maximum optimization
build-tiny:
	CGO_ENABLED=0 $(GO) build -ldflags="$(LDFLAGS) -extldflags '-static'" -trimpath -tags netgo,osusergo -o $(BUILD_DIR)/$(BINARY_NAME)-tiny ./cmd/tokman
	@echo "Tiny binary size: $$(du -h $(BUILD_DIR)/$(BINARY_NAME)-tiny | cut -f1)"
	@if command -v upx >/dev/null 2>&1; then \
		upx --best $(BUILD_DIR)/$(BINARY_NAME)-tiny -o $(BUILD_DIR)/$(BINARY_NAME)-upx 2>/dev/null || true; \
		echo "UPX compressed size: $$(du -h $(BUILD_DIR)/$(BINARY_NAME)-upx 2>/dev/null | cut -f1 || echo 'N/A')"; \
	fi

# SIMD-optimized build (SIMD is default in Go 1.25+, target kept for compatibility)
build-simd:
	$(GO) build -ldflags="$(LDFLAGS)" -trimpath -o $(BUILD_DIR)/$(BINARY_NAME)-simd ./cmd/tokman
	@echo "SIMD binary size: $$(du -h $(BUILD_DIR)/$(BINARY_NAME)-simd | cut -f1)"

# Multi-platform build
build-all:
	GOOS=linux GOARCH=amd64 $(GO) build -ldflags="$(LDFLAGS)" -trimpath -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/tokman
	GOOS=linux GOARCH=arm64 $(GO) build -ldflags="$(LDFLAGS)" -trimpath -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/tokman
	GOOS=darwin GOARCH=amd64 $(GO) build -ldflags="$(LDFLAGS)" -trimpath -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/tokman
	GOOS=darwin GOARCH=arm64 $(GO) build -ldflags="$(LDFLAGS)" -trimpath -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/tokman
	GOOS=windows GOARCH=amd64 $(GO) build -ldflags="$(LDFLAGS)" -trimpath -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/tokman

test:
	$(GO) test -race -count=1 ./...

test-short:
	$(GO) test -short -count=1 ./...

test-cover:
	$(GO) test -race -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html

# race: run tests with the race detector
race:
	$(GO) test -race ./...

# bench: run benchmarks with memory profiling
bench:
	$(GO) test -bench=. -benchmem ./...

# coverage: generate HTML coverage report
coverage:
	$(GO) test -race -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report written to coverage.html"

# tidy: tidy go module dependencies
tidy:
	$(GO) mod tidy

lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not found, falling back to go vet"; \
		$(GO) vet ./...; \
	fi

typecheck:
	$(GO) vet ./...

vet:
	$(GO) vet ./...

fmt:
	gofmt -s -w .
	goimports -w .

clean:
	rm -rf $(BUILD_DIR)/ coverage.out coverage.html

# Run all checks
check: vet typecheck lint test fmt

# Quick check (skip slow tests)
check-quick: fmt vet typecheck lint test-short

# CI check (what CI runs)
ci: test lint bench