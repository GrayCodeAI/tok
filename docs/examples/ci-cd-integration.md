# CI/CD Integration Examples

## GitHub Actions

### Basic Integration

```yaml
name: Build with TokMan
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

      - name: Run tests (compressed output)
        run: tokman go test ./... -v

      - name: Check token budget
        run: |
          tokman stats --json > token-report.json
          echo "Token savings this run:"
          tokman gain
```

### Full CI Pipeline

```yaml
name: TokMan CI Pipeline
on:
  push:
    branches: [main]
  pull_request:

jobs:
  test:
    runs-on: ${{ matrix.os }}
    strategy:
      matrix:
        os: [ubuntu-latest, macos-latest, windows-latest]
        go: ['1.21', '1.24']
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: ${{ matrix.go }}

      - name: Install TokMan
        run: go install github.com/GrayCodeAI/tokman/cmd/tokman@latest

      - name: Test
        run: tokman --mode minimal go test ./... -race -coverprofile=coverage.out

      - name: Lint
        run: tokman --mode aggressive golangci-lint run

      - name: Upload coverage
        uses: codecov/codecov-action@v4
        with:
          file: coverage.out

  report:
    needs: test
    runs-on: ubuntu-latest
    steps:
      - name: Token savings report
        run: |
          tokman stats --json
          tokman cost --model gpt-4 --json
```

## GitLab CI

```yaml
stages:
  - test
  - report

test:
  stage: test
  image: golang:1.24
  before_script:
    - go install github.com/GrayCodeAI/tokman/cmd/tokman@latest
  script:
    - tokman go test ./... -v
    - tokman go vet ./...
  after_script:
    - tokman gain

token-report:
  stage: report
  image: golang:1.24
  script:
    - tokman stats --json > token-report.json
  artifacts:
    reports:
      metrics: token-report.json
```

## Docker-based CI

```dockerfile
# ci.Dockerfile
FROM golang:1.24-alpine

RUN go install github.com/GrayCodeAI/tokman/cmd/tokman@latest

WORKDIR /app
COPY . .

RUN tokman go test ./...
RUN tokman go build ./cmd/...
```

```bash
# Build and test
docker build -f ci.Dockerfile .
```

## Pre-commit Hooks

```bash
#!/bin/bash
# .git/hooks/pre-commit

# Verify TokMan hooks integrity before each commit
if command -v tokman &> /dev/null; then
    echo "Running TokMan checks..."

    # Check hook integrity
    tokman doctor --quiet || {
        echo "TokMan hook integrity check failed!"
        exit 1
    }

    # Run compressed tests
    tokman go test ./... || {
        echo "Tests failed!"
        exit 1
    }

    echo "TokMan checks passed ✓"
fi
```

## Makefile Integration

```makefile
# Makefile

.PHONY: test lint build report

test:
	tokman go test ./... -race -coverprofile=coverage.out

lint:
	tokman --mode aggressive golangci-lint run

build:
	tokman go build -o bin/myapp ./cmd/myapp

report:
	@echo "=== Token Savings Report ==="
	tokman gain
	@echo "=== Cost Analysis ==="
	tokman cost --model claude-sonnet
```
