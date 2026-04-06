# 🤝 Contributing to TokMan

First off, **thank you** for considering contributing to TokMan! 🎉

We're building the most advanced token compression system for AI coding assistants, and we need your help to make it even better.

## 📋 Table of Contents

- [Code of Conduct](#code-of-conduct)
- [How Can I Contribute?](#how-can-i-contribute)
- [Development Setup](#development-setup)
- [Making Changes](#making-changes)
- [Coding Standards](#coding-standards)
- [Testing](#testing)
- [Submitting Changes](#submitting-changes)
- [Community](#community)

## 📜 Code of Conduct

This project adheres to a Code of Conduct. By participating, you are expected to uphold this code. Please be respectful and constructive in all interactions.

**Key principles:**
- 🤝 Be welcoming and inclusive
- 🙏 Be respectful of differing viewpoints
- 💪 Focus on what's best for the community
- 🎯 Show empathy towards other community members

## 🚀 How Can I Contribute?

### 🐛 Reporting Bugs

Before creating a bug report, please check the [existing issues](https://github.com/GrayCodeAI/tokman/issues) to avoid duplicates.

**Good bug reports include:**

```markdown
**Environment:**
- OS: macOS 14.2 / Ubuntu 22.04 / Windows 11
- Go version: 1.26.1
- TokMan version: v2.0.0

**Steps to Reproduce:**
1. Run `tokman git status`
2. Observe error message

**Expected Behavior:**
Should compress and display git status

**Actual Behavior:**
Error: "pipeline failed"

**Additional Context:**
- Repository size: 500 files
- Git version: 2.42.0
- Logs: [attach relevant logs]
```

### 💡 Suggesting Features

We love new ideas! Before suggesting a feature:

1. **Check existing issues** for similar suggestions
2. **Describe the problem** you're trying to solve
3. **Explain your proposed solution** with examples
4. **Consider alternatives** you've thought about

**Feature request template:**

```markdown
**Problem Statement:**
As a [type of user], I want [goal] so that [benefit].

**Proposed Solution:**
Add a new flag `--smart-mode` that...

**Alternatives Considered:**
- Option A: ...
- Option B: ...

**Additional Context:**
Related to issue #123
```

### 🎨 Improving Documentation

Documentation improvements are always welcome!

- Fix typos or unclear wording
- Add examples and use cases
- Improve code comments
- Create tutorials or guides

### 🔧 Code Contributions

We welcome code contributions of all sizes!

**Good first issues:**
- Look for issues labeled `good first issue`
- Check the [project board](https://github.com/GrayCodeAI/tokman/projects)
- Ask in [Discord](https://discord.gg/HrVA7ePyV) if unsure where to start

## 💻 Development Setup

### Prerequisites

- **Go 1.26+** - [Install Go](https://go.dev/doc/install)
- **Git** - Version control
- **Make** - Build automation
- **golangci-lint** - Code linting (optional)

### Initial Setup

```bash
# 1. Fork the repository on GitHub

# 2. Clone your fork
git clone https://github.com/YOUR_USERNAME/tokman.git
cd tokman

# 3. Add upstream remote
git remote add upstream https://github.com/GrayCodeAI/tokman.git

# 4. Install dependencies
go mod download

# 5. Build the project
make build

# 6. Run tests
make test

# 7. Verify installation
./bin/tokman --version
```

### Development Workflow

```bash
# 1. Sync with upstream
git checkout main
git pull upstream main

# 2. Create a feature branch
git checkout -b feature/my-awesome-feature

# 3. Make your changes
# ... code code code ...

# 4. Test your changes
make test
make lint

# 5. Commit your changes
git add .
git commit -m "feat: add awesome feature"

# 6. Push to your fork
git push origin feature/my-awesome-feature

# 7. Open a Pull Request on GitHub
```

## 🛠️ Making Changes

### Project Structure

```
tokman/
├── cmd/tokman/          # Main CLI entry point
├── internal/            # Internal packages
│   ├── commands/        # CLI commands
│   ├── filter/          # 31-layer compression pipeline
│   ├── toml/            # TOML filter system
│   ├── simd/            # SIMD optimizations
│   ├── plugin/          # WASM plugin system
│   └── ...              # Other packages
├── docs/                # Documentation
├── tests/               # Integration tests
└── benchmarks/          # Performance benchmarks
```

### Adding a New Feature

#### 1. Adding a New Command

```bash
# Create command file
touch internal/commands/mycategory/mycommand.go
```

```go
package mycategory

import (
    "github.com/spf13/cobra"
    "github.com/GrayCodeAI/tokman/internal/commands/registry"
)

var myCmd = &cobra.Command{
    Use:   "mycommand",
    Short: "Brief description",
    Long:  "Detailed description...",
    RunE:  runMyCommand,
}

func init() {
    // Register command
    registry.Add(func() { registry.Register(myCmd) })
}

func runMyCommand(cmd *cobra.Command, args []string) error {
    // Implementation
    return nil
}
```

#### 2. Adding a New Filter Layer

```bash
# Create layer file
touch internal/filter/my_layer.go
```

```go
package filter

type MyLayer struct {
    enabled bool
}

func NewMyLayer(enabled bool) *MyLayer {
    return &MyLayer{enabled: enabled}
}

func (l *MyLayer) Apply(input string, mode Mode) (string, int) {
    if !l.enabled || len(input) == 0 {
        return input, 0
    }
    
    // Your compression logic here
    output := compress(input)
    saved := len(input) - len(output)
    
    return output, saved
}
```

#### 3. Adding a TOML Filter

```bash
# Create filter file
touch internal/toml/builtin/mytool.toml
```

```toml
# MyTool - Description
[mytool]
match = "^mytool (build|test)"
mode = "aggressive"
description = "MyTool with compact output"

strip_lines_matching = [
  "^Building...",
  "^Loading..."
]

output_patterns = [
  "^Success:",
  "^Error:",
  "^Warning:"
]

compact_repeated_lines = true
max_repeated_context = 3
```

## 📏 Coding Standards

### Go Style Guide

We follow the [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md) with some additions:

**General Principles:**
- ✅ Write clear, idiomatic Go code
- ✅ Prefer simplicity over cleverness
- ✅ Use meaningful variable names
- ✅ Add comments for non-obvious code
- ✅ Keep functions small and focused

**Formatting:**
```bash
# Format code
go fmt ./...

# Or use goimports (preferred)
goimports -w .
```

**Naming Conventions:**
```go
// ✅ Good
func CompressText(input string) string
type PipelineConfig struct
const MaxTokens = 1000000

// ❌ Bad
func compress_text(input string) string
type pipeline_config struct
const max_tokens = 1000000
```

**Error Handling:**
```go
// ✅ Good - wrap errors with context
if err != nil {
    return fmt.Errorf("compress text: %w", err)
}

// ❌ Bad - lose error context
if err != nil {
    return err
}
```

**Comments:**
```go
// ✅ Good - explain WHY
// Use aggressive mode because user data is already sanitized
mode := ModeAggressive

// ❌ Bad - state the obvious
// Set mode to aggressive
mode := ModeAggressive
```

### Code Organization

**Package Structure:**
- One package per directory
- Keep related code together
- Minimize inter-package dependencies
- Use internal/ for private packages

**File Naming:**
- `snake_case.go` for file names
- `_test.go` suffix for tests
- Group related files by prefix (e.g., `filter_*.go`)

## 🧪 Testing

### Writing Tests

```go
func TestMyFunction(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {"empty input", "", ""},
        {"simple case", "hello", "HELLO"},
        {"with numbers", "test123", "TEST123"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := MyFunction(tt.input)
            if result != tt.expected {
                t.Errorf("MyFunction(%q) = %q, want %q", 
                    tt.input, result, tt.expected)
            }
        })
    }
}
```

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-cover

# Run specific package tests
go test ./internal/filter/...

# Run tests with verbose output
go test -v ./...

# Run tests with race detector
go test -race ./...

# Run benchmarks
make bench
```

### Test Coverage Goals

- **Aim for 70%+** coverage on new code
- **100% coverage** for critical paths
- **Edge cases** must be tested
- **Error conditions** must be tested

## 📝 Commit Messages

We follow [Conventional Commits](https://www.conventionalcommits.org/):

### Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `perf`: Performance improvements
- `test`: Adding or updating tests
- `chore`: Maintenance tasks
- `ci`: CI/CD changes

### Examples

```bash
# Feature
feat(filter): add SIMD acceleration for ANSI stripping

# Bug fix
fix(cli): resolve panic on empty input

# Documentation
docs(readme): update installation instructions

# Multiple changes
feat(filter): add semantic chunking layer

- Implement semantic boundary detection
- Add configurable chunk sizes
- Include tests and benchmarks

Closes #123
```

## 🎯 Pull Request Process

### Before Submitting

**Checklist:**
- [ ] Code follows style guidelines
- [ ] Tests pass locally (`make test`)
- [ ] New tests added for new features
- [ ] Documentation updated
- [ ] Commit messages follow convention
- [ ] No unnecessary changes (formatting, etc.)
- [ ] Branch is up to date with main

### PR Template

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Unit tests added/updated
- [ ] Integration tests added/updated
- [ ] Manual testing completed

## Screenshots (if applicable)
[Add screenshots for UI changes]

## Checklist
- [ ] Code follows style guidelines
- [ ] Self-review completed
- [ ] Comments added for complex code
- [ ] Documentation updated
- [ ] No new warnings generated
- [ ] Tests pass locally
```

### Review Process

1. **Automated checks** run (tests, linting)
2. **Maintainer review** (usually within 48 hours)
3. **Address feedback** if requested
4. **Approval** from maintainer
5. **Merge** into main branch

### After Merge

- Your PR will be included in the next release
- You'll be added to CONTRIBUTORS.md
- Feel free to share your contribution! 🎉

## 🏗️ Development Tools

### Available Make Targets

```bash
make build          # Build binary
make build-all      # Build for all platforms
make build-simd     # Build with SIMD optimizations
make test           # Run tests
make test-cover     # Run tests with coverage
make bench          # Run benchmarks
make lint           # Run linters
make fmt            # Format code
make clean          # Clean build artifacts
make check          # Run all checks (fmt, vet, lint, test)
```

### Recommended Tools

- **VS Code** with Go extension
- **GoLand** (JetBrains IDE)
- **golangci-lint** - Comprehensive linting
- **delve** - Go debugger
- **goimports** - Import management

## 💬 Community

### Get Help

- 💬 [Discord Server](https://discord.gg/HrVA7ePyV) - Real-time chat
- 🐛 [Issue Tracker](https://github.com/GrayCodeAI/tokman/issues) - Bug reports
- 📧 [Email](mailto:hello@tokman.dev) - Direct contact

### Stay Updated

- ⭐ Star the repo on GitHub
- 👀 Watch for releases
- 🐦 Follow [@tokman_dev](https://twitter.com/tokman_dev) on Twitter
- 📰 Read the [CHANGELOG](./CHANGELOG.md)

## 🙏 Thank You!

Every contribution, no matter how small, makes a difference. Whether you're:
- 🐛 Fixing a typo
- 📝 Improving documentation  
- 🔧 Adding a feature
- 🧪 Writing tests
- 💡 Suggesting ideas

**You're making TokMan better for everyone!** 🎉

---

<div align="center">

**Questions? Join our [Discord](https://discord.gg/HrVA7ePyV)!**

Made with ❤️ by the TokMan community

</div>
