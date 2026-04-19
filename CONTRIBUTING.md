# Contributing to tok

Thanks for contributing! This guide covers everything you need.

## Quick Start

```bash
git clone https://github.com/GrayCodeAI/tok.git && cd tok
go mod download
make test    # verify setup
make build   # build binary
```

## Development Workflow

### 1. Pick an Issue

Browse [open issues](https://github.com/GrayCodeAI/tok/issues) and look for:
- `good first issue` — great for newcomers
- `help wanted` — we'd love your help
- `bug` — fix something broken
- `enhancement` — add new features

Comment on the issue to claim it before starting work.

### 2. Create a Branch

```bash
git checkout -b feat/my-feature    # new feature
git checkout -b fix/my-bug-fix     # bug fix
git checkout -b docs/my-docs       # documentation
```

### 3. Make Changes

**Code style:**
- Run `make lint` before committing
- Follow `gofmt` formatting (enforced by CI)
- Write tests for new functionality
- Keep functions focused and under 50 lines when possible

**Commit messages:**
- Use [Conventional Commits](https://www.conventionalcommits.org/)
- Format: `type(scope): description`
- Types: `feat`, `fix`, `docs`, `refactor`, `test`, `chore`

```bash
feat(filter): add semantic chunk compression layer
fix(commands): handle nil args in compress command
docs(readme): update installation instructions
```

### 4. Run Tests

```bash
make test          # all tests
make test-race     # race detector
make test-cover    # coverage report
make lint          # golangci-lint
make typecheck     # go vet
```

### 5. Submit a PR

Push your branch and open a pull request at [github.com/GrayCodeAI/tok/pulls](https://github.com/GrayCodeAI/tok/pulls).

**PR checklist:**
- [ ] Tests pass (`make test`)
- [ ] Lint clean (`make lint`)
- [ ] Commit messages follow Conventional Commits
- [ ] Documentation updated (README, docs, etc.)
- [ ] CHANGELOG.md updated (for user-facing changes)

## Project Structure

```
tok/
├── cmd/tok/              # CLI entry point
├── internal/
│   ├── commands/         # Command implementations (cobra)
│   │   ├── core/         # Primary commands
│   │   ├── system/       # System utilities
│   │   ├── filtercmd/    # Filter pipeline
│   │   └── ...           # 17 more categories
│   ├── compressor/       # Input compression
│   ├── filter/           # Output filtering (31 layers)
│   └── output/           # Output abstraction layer
├── agents/               # AI agent rules
├── hooks/                # Shell scripts
└── config/               # TOML configs + filters
```

### Adding a New Command

1. Create a file in the appropriate `internal/commands/<category>/` directory
2. Register it with the command registry:

```go
package mycategory

import (
    "github.com/spf13/cobra"
    "github.com/GrayCodeAI/tok/internal/commands/registry"
)

func init() {
    registry.Register(&cobra.Command{
        Use:   "my-command",
        Short: "Brief description",
        RunE:  runMyCommand,
    })
}

func runMyCommand(cmd *cobra.Command, args []string) error {
    // Your implementation
    return nil
}
```

3. Add tests in `mycommand_test.go`
4. Update docs if needed

### Adding a Filter Layer

New compression layers go in `internal/filter/`. See [docs/LAYERS.md](docs/LAYERS.md) for the layer architecture and [internal/filter/AGENTS.md](internal/filter/AGENTS.md) for implementation guidelines.

## Testing

```bash
# Run all tests
make test

# Run specific package
go test ./internal/filter/... -v

# Run with coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Run fuzz tests
go test ./internal/filter/... -fuzz=FuzzPipelineProcess
go test ./internal/toml/... -fuzz=FuzzTOMLFilterParse
```

## Release Process

1. Create a tag: `git tag v0.XX.0`
2. Push tag: `git push origin v0.XX.0`
3. GitHub Actions builds binaries, generates SBOM, and creates release
4. Homebrew tap updates automatically

## Code of Conduct

See [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md). Be respectful and constructive.

## Questions?

- Open a [discussion](https://github.com/GrayCodeAI/tok/discussions)
- File an [issue](https://github.com/GrayCodeAI/tok/issues)
- Read the [docs](docs/)
