# Contributing to TokMan

Thank you for your interest in contributing to TokMan! We welcome contributions of all kinds - from bug fixes and documentation improvements to new features and creative ideas.

## Quick Start

```bash
# 1. Fork and clone
git clone https://github.com/YOUR_USERNAME/tokman.git
cd tokman

# 2. Install dependencies
go mod download

# 3. Run tests
make test

# 4. Create a branch
git checkout -b my-great-feature
```

## Ways to Contribute

### 🐛 Report Bugs

- Check existing issues first
- Use our [bug report template](.github/ISSUE_TEMPLATE/bug_report.md)
- Include version, OS, and reproduction steps
- Attach logs if possible

### ✨ Suggest Features

- Open a [feature request](.github/ISSUE_TEMPLATE/feature_request.md)
- Describe the problem and your proposed solution
- Include examples and use cases

### 📝 Improve Documentation

- Fix typos, grammar, or clarity issues
- Add missing documentation
- Translate docs to other languages
- Create tutorials or examples

### 💻 Write Code

- Pick an issue labeled `good first issue`
- Comment on the issue to claim it
- Follow our coding standards (see below)
- Write tests for your changes
- Update documentation

### 🧪 Test and Review

- Test existing PRs on your system
- Review code for correctness and style
- Help with triaging issues
- Verify bug reports

## Development Setup

### Prerequisites

- **Go 1.21+** (1.24+ recommended)
- **Git**
- **Make** (for Makefile commands)
- **SQLite** (included via modernc.org/sqlite)

### Setup

```bash
# Clone
git clone https://github.com/GrayCodeAI/tokman.git
cd tokman

# Build
make build

# Run tests
make test

# Run linter
make lint
```

### Useful Commands

```bash
make build          # Build the binary
make test           # Run tests with race detector
make test-cover     # Tests with coverage
make lint           # Run golangci-lint
make fmt            # Format code
make vet            # Run go vet
make typecheck      # Type checking
make benchmark      # Run benchmarks
make check          fmt + vet + typecheck + lint + test
make clean          # Clean build artifacts
```

## Coding Standards

### Go Code Style

We follow standard Go conventions:

```bash
# Format all code
go fmt ./...

# Check for issues
go vet ./...

# Run linter
golangci-lint run
```

### Naming Conventions

- **Packages:** lowercase, short, concise names
- **Exported names:** PascalCase (`FilterPipeline`, `EstimateTokens`)
- **Unexported names:** camelCase (`filterPipeline`, `estimateTokens`)
- **Interfaces:** -er suffix (`Filter`, `Estimator`, `Runner`)
- **Constants:** PascalCase or ALL_CAPS for exported values

### Error Handling

```go
// Good: Named errors for checking
var ErrFilterNotFound = errors.New("filter not found")

// Good: Wrap errors with context
if err != nil {
    return fmt.Errorf("failed to filter output: %w", err)
}

// Good: Check specific errors
if errors.Is(err, ErrFilterNotFound) {
    // Handle specific case
}
```

### Function Design

```go
// Good: Single responsibility, clear name
func FilterOutput(input string, mode Mode) (string, int) {
    // ...
    return filtered, tokensSaved
}

// Good: Context for cancellable operations
func ProcessFile(ctx context.Context, path string) error {
    // ...
}
```

### Testing

```go
// Use table-driven tests
func TestFilterPipeline(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        mode     Mode
        expected string
    }{
        {
            name:     "empty input",
            input:    "",
            mode:     ModeMinimal,
            expected: "",
        },
        // ...
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test logic
        })
    }
}
```

### Commit Messages

We follow [Conventional Commits](https://www.conventionalcommits.org/):

```
type(scope): description

[optional body]

[optional footer(s)]
```

**Types:**
- `feat:` New feature
- `fix:` Bug fix
- `docs:` Documentation changes
- `style:` Code style changes (formatting, semicolons, etc)
- `refactor:` Code refactoring
- `test:` Adding or updating tests
- `chore:` Build process, tooling, etc
- `perf:` Performance improvements
- `ci:` CI/CD changes

**Examples:**

```
feat(filter): add entropy-based filtering layer
fix(cli): handle empty config file gracefully
docs(readme): add quick start guide
test(pipeline): add table-driven tests for all layers
perf(core): reduce allocations in token estimator
```

## Pull Request Process

### Before Submitting

1. **Create an issue** first (except for typos/minor fixes)
2. **Discuss the approach** if your solution is non-trivial
3. **Write tests** for your changes
4. **Update docs** if applicable
5. **Run the checks:**
   ```bash
   make check  # fmt + vet + typecheck + lint + test
   ```

### PR Guidelines

1. **Keep it focused** - One feature/fix per PR
2. **Reference the issue** - Use "Fixes #123" to auto-close
3. **Explain your changes** - Fill out the PR template
4. **Add tests** - New code needs tests
5. **Update docs** - Update docs if behavior changes
6. **Stay updated** - Rebase on main if needed

### Review Process

1. **Code review** by maintainers
2. **Tests must pass** in CI
3. **Address feedback** promptly
4. **Merge** after approval

## Filter Development

Adding new filters? Follow this pattern:

```go
// internal/filter/my_filter.go

package filter

// MyFilter implements the [Layer] interface.
type MyFilter struct {
    // configuration
}

// NewMyFilter creates a new MyFilter.
func NewMyFilter() *MyFilter {
    return &MyFilter{}
}

// Apply processes the input and returns filtered text and tokens saved.
func (f *MyFilter) Apply(input string, mode Mode) (string, int) {
    // Implementation
    return filtered, 0
}

// shouldSkipMyFilter checks if this layer would provide value.
func shouldSkipMyFilter(input string) bool {
    return len(input) < 50
}
```

Don't forget to:

1. Add to `PipelineConfig` in `pipeline.go`
2. Add to `PipelineCoordinator` struct
3. Initialize in `NewPipelineCoordinator()`
4. Add `processLayer()` method with timing
5. Add to `Process()` pipeline execution

## Command Development

Adding new commands? Use the registry pattern:

```go
// internal/commands/mycategory/mycmd.go

package mycategory

import (
    "github.com/spf13/cobra"
    "github.com/GrayCodeAI/tokman/internal/commands/registry"
)

var myCmd = &cobra.Command{
    Use:   "mycmd",
    Short: "Brief description",
    Long:  "Long description",
    RunE: func(cmd *cobra.Command, args []string) error {
        // Implementation
        return nil
    },
}

func init() {
    registry.Add(func() { registry.Register(myCmd) })
}
```

Then add import to `root.go`:

```go
_ "github.com/GrayCodeAI/tokman/internal/commands/mycategory"
```

## Testing Guidelines

### Unit Tests

```bash
# Run all tests
make test

# Run specific package tests
go test ./internal/filter/...

# Run with coverage
make test-cover

# View coverage
go tool cover -html=coverage.html
```

### Integration Tests

```bash
# Run integration tests
go test ./tests/...

# Run with verbose output
go test ./tests/... -v
```

### Benchmarks

```bash
# Run benchmarks
make benchmark

# Profile
go test -bench=. -cpuprofile=cpu.prof -memprofile=mem.prof
```

### Fuzz Testing

```bash
# Run fuzz tests
go test ./internal/filter/ -fuzz=FuzzFilter
```

## Reporting Issues

### Bugs

Use our bug report template. Include:

- TokMan version
- Operating system
- AI tool (if using integration)
- Steps to reproduce
- Expected vs actual behavior
- Relevant logs (`tokman -v`)

### Features

Use the feature request template. Include:

- Problem statement
- Proposed solution
- Real-world use cases
- Prior art or comparable tools if relevant

## Style Guide

### Documentation

- Use clear, concise language
- Include code examples
- Explain the "why" not just the "what"
- Keep examples up-to-date
- Use markdown formatting consistently

### Comments

```go
// Good: Explain why, not what
// Entropy filtering removes low-information tokens that don't
// contribute to semantic understanding. Based on Selective Context
// (Mila 2023), this targets tokens with low self-information.
func (f *EntropyFilter) Apply(input string, mode Mode) string {
```

### Logging

```go
// Use the logger package
import "github.com/GrayCodeAI/tokman/internal/utils"

utils.Logger.Debug("Processing %d tokens", count)
utils.Logger.Info("Filter applied successfully", "layer", "entropy")
utils.Logger.Error("Failed to process", "error", err)
```

## Release Process

1. Update `CHANGELOG.md`
2. Bump version in `cmd/tokman/main.go`
3. Create release tag: `git tag v0.29.0`
4. Push tag: `git push origin v0.29.0`
5. Create GitHub release with notes

## Getting Help

- **GitHub Issues:** For bugs, features, questions
- **Discord:** (coming soon!) For community discussions
- **Email:** maintainers@graycode.ai

## Recognition

Contributors are recognized in:

- `AUTHORS.md` file
- Release notes
- Our website (coming soon!)
- Special thanks for significant contributions

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

---

**Thank you for contributing to TokMan!** 🚀

Every contribution matters, no matter how small. We're building something great together!
