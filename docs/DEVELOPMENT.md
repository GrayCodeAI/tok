# Development Guide

## Prerequisites

- **Go 1.21+** (Go 1.24+ recommended for latest features)
- **Git** (version control)
- **Make** (build automation)
- **SQLite** (not required - modernc.org/sqlite is pure Go)

### Recommended Tools

- **gopls** - Go language server (IDE support)
- **golangci-lint** - Meta linter
- **goimports** - Import management
- **staticcheck** - Static analysis

```bash
# Install development tools
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install honnef.co/go/tools/cmd/staticcheck@latest
```

## Project Structure

```
tokman/
├── cmd/tokman/              # Entry point
├── internal/                # Private packages
│   ├── commands/            # CLI commands (Cobra)
│   ├── filter/              # Compression pipeline (31-stage core + experimental extension)
│   ├── core/                # Core functionality
│   ├── config/              # Configuration
│   ├── tracking/            # SQLite tracking
│   └── utils/               # Utilities
├── tests/                   # Integration tests
├── benchmarks/              # Performance benchmarks
└── docs/                    # Documentation
```

## Getting Started

### 1. Clone and Build

```bash
# Fork on GitHub, then clone
git clone https://github.com/YOUR_USERNAME/tokman.git
cd tokman

# Build
make build

# Run
./tokman --version
```

### 2. Run Tests

```bash
# All tests with race detector
make test

# Tests with coverage
make test-cover
open coverage.html

# Specific package
go test ./internal/filter/ -v

# Run with verbose
go test ./... -v
```

### 3. Run Linters

```bash
# Format code
make fmt

# Run go vet
make vet

# Type check
make typecheck

# Run golangci-lint
make lint

# Run all checks
make check
```

## Development Workflow

### Feature Development

```bash
# 1. Create feature branch
git checkout -b feat/my-new-feature

# 2. Make changes
# ... edit code ...

# 3. Run checks
make check

# 4. Commit
git commit -m "feat: add my new feature"

# 5. Push
git push origin feat/my-new-feature
```

### Bug Fix Workflow

```bash
# 1. Create bug branch from main
git checkout -b fix/bug-description

# 2. Add test that reproduces bug
# ... write failing test ...

# 3. Fix code
# ... implement fix ...

# 4. Verify test passes
make test

# 5. Run full check suite
make check

# 6. Commit and push
git add .
git commit -m "fix: resolve bug description"
git push origin fix/bug-description
```

## Code Style

### Formatting

Follow standard Go formatting:

```bash
# Auto-format
go fmt ./...

# Use goimports for imports
goimports -w .
```

### Import Organization

```go
// Standard library first
import (
    "context"
    "fmt"
    "os"
    "time"
    
    // Third-party second
    "github.com/spf13/cobra"
    "github.com/spf13/viper"
    
    // Internal packages last
    "github.com/GrayCodeAI/tokman/internal/config"
    "github.com/GrayCodeAI/tokman/internal/filter"
)
```

### Naming Conventions

**Packages:** Lowercase, short, no underscores
```go
package filter       // ✓
package filter_manager  // ✗ (too long)
package filters      // ✗ (plural)
```

**Types:** PascalCase
```go
type PipelineCoordinator struct { }    // ✓
type pipeline_coordinator struct { }   // ✗
```

**Interfaces:** -er suffix for action interfaces
```go
type Runner interface { }        // ✓
type Reader interface { }        // ✓
type Filterable interface { }    // ✓
```

**Constants:** PascalCase or ALL_CAPS for special values
```go
const DefaultTimeout = 30 * time.Second   // ✓
const MAX_BUFFER_SIZE = 65536             // ✓ (special values)
```

**Variables:** camelCase
```go
var filteredOutput string   // ✓
var OutputFiltered string   // ✗
```

### Function Design

**Single responsibility:**
```go
// ✓ Good: Does one thing well
func FilterOutput(input string, mode filter.Mode) (string, int) {
    // Filter implementation
}

// ✗ Bad: Does too many things
func ProcessAndFilterAndSave(input string) {
    // ...
}
```

**Return errors properly:**
```go
// ✓ Good: Named errors for checking
var ErrConfigNotFound = errors.New("config file not found")

func LoadConfig(path string) (*Config, error) {
    if _, err := os.Stat(path); os.IsNotExist(err) {
        return nil, fmt.Errorf("load config: %w", ErrConfigNotFound)
    }
    // ...
}
```

**Use context for cancellable operations:**
```go
// ✓ Good: Context for cancellation
func RunCommand(ctx context.Context, cmd string) (string, error) {
    c := exec.CommandContext(ctx, cmd)
    // ...
}
```

### Error Handling Pattern

```go
// 1. Define error variables
var (
    ErrFilterNotFound   = errors.New("filter not found")
    ErrBudgetExceeded   = errors.New("token budget exceeded")
    ErrInvalidInput     = errors.New("invalid input")
)

// 2. Wrap errors with context
func Process(input string) error {
    if err := validate(input); err != nil {
        return fmt.Errorf("process input: %w", err)
    }
    // ...
}

// 3. Check specific errors
func HandleError(err error) {
    if errors.Is(err, ErrFilterNotFound) {
        // Handle specific case
        return
    }
    // ...
}
```

### Testing Standards

#### Table-Driven Tests

```go
func TestFilterApply(t *testing.T) {
    tests := []struct {
        name          string
        input         string
        mode          filter.Mode
        expected      string
        expectError   bool
    }{
        {
            name:        "empty input",
            input:       "",
            mode:        filter.ModeMinimal,
            expected:    "",
            expectError: false,
        },
        {
            name:        "simple case",
            input:       "hello world\nline 2",
            mode:        filter.ModeMinimal,
            expected:    "hello world\nline 2",
            expectError: false,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            f := NewFilter()
            result, err := f.Apply(tt.input, tt.mode)
            
            if tt.expectError {
                assert.Error(t, err)
                return
            }
            
            assert.NoError(t, err)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

#### Golden File Tests

```go
func TestFilterGoldenOutput(t *testing.T) {
    input := loadFile(t, "testdata/input.txt")
    expected := loadFile(t, "testdata/golden_output.txt")
    
    filter := NewFilter()
    result, _ := filter.Apply(input, filter.ModeMinimal)
    
    if result != expected {
        // Update golden file with -update flag
        if *update {
            os.WriteFile("testdata/golden_output.txt", []byte(result), 0644)
            return
        }
        t.Errorf("output mismatch")
    }
}
```

#### Benchmark Tests

```go
func BenchmarkFilterApply(b *testing.B) {
    input := generateLargeInput(10000)  // 10K tokens
    filter := NewFilter()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        filter.Apply(input, filter.ModeMinimal)
    }
}

func BenchmarkFilterApply_Parallel(b *testing.B) {
    input := generateLargeInput(10000)
    filter := NewFilter()
    
    b.ResetTimer()
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            filter.Apply(input, filter.ModeMinimal)
        }
    })
}
```

### Logging

Use the utils logger package:

```go
import "github.com/GrayCodeAI/tokman/internal/utils"

// Different log levels
utils.Logger.Debug("Debug message with values", "key", value)
utils.Logger.Info("Operation successful")
utils.Logger.Warn("Potential issue detected")
utils.Logger.Error("Something went wrong", "error", err)

// Structured logging
utils.Logger.Info("Pipeline processed",
    "tokens_in", inputTokens,
    "tokens_out", outputTokens,
    "saved", saved,
    "duration", duration,
)
```

### Command Development

Adding a new command follows the registry pattern:

```go
// internal/commands/mycategory/mycmd.go

package mycategory

import (
    "github.com/spf13/cobra"
    "github.com/GrayCodeAI/tokman/internal/commands/registry"
)

var myCmd = &cobra.Command{
    Use:   "mycmd",
    Short: "Brief description of my command",
    Long: `Longer description that can span multiple lines.

Examples:
  tokman mycmd --flag value
  tokman mycmd arg1 arg2`,
    Args: cobra.MinimumNArgs(1),
    RunE: myCmdRun,
}

func init() {
    // Register via the registry
    registry.Add(func() { registry.Register(myCmd) })
    
    // Add flags
    myCmd.Flags().String("flag", "", "Flag description")
}

func myCmdRun(cmd *cobra.Command, args []string) error {
    // Implementation
    return nil
}
```

Then add the import to `root.go`:

```go
import (
    // ... other imports
    _ "github.com/GrayCodeAI/tokman/internal/commands/mycategory"
)
```

### Filter Layer Development

Adding a new filter layer:

```go
// internal/filter/my_layer.go

package filter

// MyLayer implements the Layer interface.
type MyLayer struct {
    threshold float64
}

// NewMyLayer creates a new MyLayer.
func NewMyLayer() *MyLayer {
    return &MyLayer{
        threshold: 0.5,
    }
}

// Name returns the layer name.
func (l *MyLayer) Name() string {
    return "my_layer"
}

// Apply processes the input and returns filtered text and tokens saved.
func (l *MyLayer) Apply(input string, mode Mode) (string, int) {
    startTime := time.Now()
    
    // Filter implementation
    output := doFiltering(input, mode)
    
    tokensBefore := EstimateTokens(input)
    tokensAfter := EstimateTokens(output)
    tokensSaved := tokensBefore - tokensAfter
    
    Logger.Debug("my_layer applied",
        "tokens_before", tokensBefore,
        "tokens_after", tokensAfter,
        "saved", tokensSaved,
        "duration", time.Since(startTime),
    )
    
    return output, tokensSaved
}

// ShouldSkip checks if this layer would provide value.
func (l *MyLayer) ShouldSkip(input string, config PipelineConfig) bool {
    // Skip for short inputs
    if len(input) < 50 {
        return true
    }
    return false
}
```

Then update the pipeline:

1. Add config flag to `PipelineConfig`
2. Add field to `PipelineCoordinator` struct
3. Initialize in `NewPipelineCoordinator()`
4. Add `processLayer()` method
5. Add to `Process()` execution order

## Build Commands

```bash
# Standard build
make build

# Optimized small binary
make build-small

# Multi-platform build (via GoReleaser)
make build-all

# Run the binary
./tokman help

# Cross-compile for Linux
GOOS=linux GOARCH=amd64 go build -o tokman-linux ./cmd/tokman
```

## Performance Profiling

### CPU Profiling

```bash
# Generate CPU profile
go test -cpuprofile=cpu.prof -bench=BenchmarkFilter ./internal/filter/

# Analyze
go tool pprof cpu.prof

# Top functions
(pprof) top

# Web visualization
(pprof) web
```

### Memory Profiling

```bash
# Generate memory profile
go test -memprofile=mem.prof -bench=BenchmarkFilter ./internal/filter/

# Analyze
go tool pprof mem.prof

# Top allocations
(pprof) top
```

### Benchmarking

```bash
# Run benchmarks
make benchmark

# Detailed benchmarks
go test -bench=. -benchmem ./internal/filter/

# Compare benchmarks
benchstat old.txt new.txt
```

## Debugging

### Debug with Delve

```bash
# Install Delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug tests
dlv test ./internal/filter/

# Debug binary
dlv exec ./tokman -- help

# Breakpoints
(dlv) break internal/filter/pipeline.go:42
(dlv) continue
```

### Verbose Mode

```bash
# Enable verbose output
tokman -v [command]

# Enable debug logging
export TOKMAN_LOG=debug
tokman [command]
```

## Common Tasks

### Adding a New Dependency

```bash
# Get the package
go get github.com/example/package

# Tidy up
go mod tidy

# Verify
go mod verify
```

### Updating Dependencies

```bash
# Check for updates
go list -u -m all

# Update specific package
go get github.com/example/package@latest

# Update all
go get -u ./...
go mod tidy
```

### Creating a Release

```bash
# 1. Update CHANGELOG.md
# 2. Bump version in cmd/tokman/main.go
# 3. Run all checks
make check

# 4. Tag release
git tag v0.29.0
git push origin v0.29.0

# 5. Create GitHub release
# (Or use goreleaser)
goreleaser release --rm-dist
```

## Troubleshooting

### Build Fails

```bash
# Clean and rebuild
make clean
make build

# Check Go version
go version  # Must be 1.21+

# Download dependencies
go mod download
```

### Tests Fail

```bash
# Run with verbose
go test ./... -v

# Specific package
go test ./internal/filter/ -v

# Skip race detector (faster)
go test ./...

# Update golden files
go test ./... -update
```

### Linter Errors

```bash
# See all issues
golangci-lint run

# Auto-fix
golangci-lint run --fix

# Check specific linter
golangci-lint run --enable=staticcheck
```

## CI/CD Integration

### GitHub Actions

The project uses GitHub Actions for:

- Running tests on push/PR
- Linting and type checking
- Building for multiple platforms
- Releasing to GitHub Releases

### Local CI Simulation

```bash
# Simulate CI checks locally
make check

# Test on multiple platforms
GOOS=linux GOARCH=amd64 go build ./...
GOOS=darwin GOARCH=arm64 go build ./...
GOOS=windows GOARCH=amd64 go build ./...
```

## Best Practices

1. **Write tests** for new functionality, fix bugs
2. **Document changes** in CHANGELOG.md
3. **Follow Go conventions** (Effective Go)
4. **Keep PRs focused** - one change per PR
5. **Update docs** when behavior changes
6. **Use meaningful** - messages
7. **Comment the why** not the what
8. **Handle errors** properly don't ignore them
9. **Avoid globals** pass dependencies explicitly
10. **Optimize last** after correct, then clear, then fast

---

**Happy coding!** 🚀
