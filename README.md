# tok

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![CI](https://github.com/GrayCodeAI/tok/actions/workflows/ci.yml/badge.svg)](https://github.com/GrayCodeAI/tok/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/GrayCodeAI/tok)](https://goreportcard.com/report/github.com/GrayCodeAI/tok)
[![codecov](https://codecov.io/gh/GrayCodeAI/tok/branch/main/graph/badge.svg)](https://codecov.io/gh/GrayCodeAI/tok)

**Unified token optimization CLI** — compress AI input text and filter verbose terminal output to reduce LLM token usage by up to 90%.

```
┌─────────────────┐     tok      ┌──────────────────┐
│  Your prompt    │ ──compress──▶│  AI (fewer tokens)│
│  (verbose text) │              │                  │
└─────────────────┘              └──────────────────┘
                                  ┌──────────────────┐
┌─────────────────┐     tok      │  Your terminal    │
│  AI output      │ ◀──filter─── │  (clean, concise) │
│  (verbose logs) │              │                  │
└─────────────────┘              └──────────────────┘
```

## Why tok?

LLM interactions waste tokens on both sides:

| Problem | Impact | tok Solution |
|---------|--------|-------------|
| Verbose prompts | High input cost | Compress text before sending (6 modes) |
| Noisy terminal output | Context window exhaustion | Filter terminal output (31-layer pipeline) |
| Repetitive logs/tests | Lost signal | Deduplicate, fold, summarize |
| Agent tool outputs | Token budget blowout | Smart compression per content type |

## Quick Start

### Install

```bash
# Go install
go install github.com/GrayCodeAI/tok/cmd/tok@latest

# Homebrew (coming soon)
# brew install GrayCodeAI/tap/tok

# Docker
docker run --rm ghcr.io/graycodeai/tok:latest --help

# Build from source
git clone https://github.com/GrayCodeAI/tok.git && cd tok && make build
```

### Compress Input

```bash
# Compress text via flag
tok compress -mode ultra -input "Please implement a user authentication system with JWT tokens"

# Compress from stdin
echo "Could you explain why this React component keeps re-rendering?" | tok compress -mode full

# Available modes: lite, full, ultra, wenyan-lite, wenyan, wenyan-ultra
```

### Filter Output

```bash
# Any command — tok intercepts and filters output
tok git status
tok npm test
tok cargo build
tok docker ps -a

# Works transparently — just prefix any command with `tok`
```

### Activate Terse Mode

```bash
# Tell AI agents to respond tersely
tok on full          # Classic terse style
tok on ultra         # Maximum compression
tok on lite          # Professional but tight

# Check status
tok doctor           # Diagnostics
tok status           # Current mode
tok gain             # Token savings analytics
```

## Features

### Input Compression (6 Modes)

| Mode | Style | Reduction |
|------|-------|-----------|
| `lite` | Drop filler, keep grammar | ~20% |
| `full` | Drop articles, fragments OK | ~40% |
| `ultra` | Telegraphic, abbreviations | ~60% |
| `wenyan-lite` | Classical Chinese light | ~30% |
| `wenyan` | Classical Chinese standard | ~50% |
| `wenyan-ultra` | Classical Chinese max | ~70% |

### Output Filtering (31-Layer Pipeline)

- **Entropy pruning** (L1) — Shannon entropy-based token scoring
- **Perplexity filtering** (L2) — LLMLingua-style iterative pruning
- **AST-aware compression** (L4) — Preserve function signatures, compress bodies
- **H2O heavy-hitter** (L13) — Heap-based important token preservation
- **Attention sink** (L14) — StreamingLLM-style boundary preservation
- **Semantic chunking** (L16) — ChunkKV-style relevance scoring
- **And 25+ more layers** — See `tok layers` for full architecture

### 100+ Built-in Commands

tok wraps common CLI tools with intelligent filtering:

```
git, npm, yarn, pnpm, cargo, go, make, gradle, maven, mvn
docker, kubectl, helm, terraform, aws, gh
pytest, jest, vitest, playwright, rspec
ruff, mypy, golangci-lint, prettier, tsc
... and many more
```

### AI Agent Integration

tok ships with rules for 12 AI coding agents:

```
cursor, windsurf, cline, copilot, claude-code, aider
continue, roo-code, cody, code-whisperer, tabnine, codeium
```

```bash
tok install-agents    # Install rules to all agent directories
tok uninstall-agents  # Remove all agent rules
```

## Architecture

```
tok/
├── cmd/tok/              # CLI entry point (cobra)
├── internal/
│   ├── commands/         # 100+ command implementations
│   │   ├── core/         # Primary commands (doctor, status, gain, etc.)
│   │   ├── system/       # System commands (ls, find, grep, etc.)
│   │   ├── filtercmd/    # Filter pipeline commands
│   │   ├── lang/         # Language-specific commands
│   │   ├── test/         # Test runner wrappers
│   │   ├── container/    # Docker, kubectl, etc.
│   │   └── ...           # 14 more command categories
│   ├── compressor/       # Input compression engine
│   ├── filter/           # Output filtering pipeline (31 layers)
│   ├── output/           # Centralized output abstraction
│   ├── hooks/            # Shell hook management
│   ├── tracking/         # Token usage tracking (SQLite)
│   └── telemetry/        # Anonymous usage metrics
├── agents/               # AI agent rule files (12 agents)
├── hooks/                # Shell integration scripts
├── config/               # Example configs + TOML filters
└── evals/                # Compression quality benchmarks
```

## Configuration

```toml
# ~/.config/tok/config.toml
[core]
mode = "full"
auto_activate = true

[tracking]
database_path = "~/.local/share/tok/tracking.db"

[telemetry]
enabled = true  # Anonymous usage data only
```

## Shell Integration

```bash
# Install shell hooks (adds [TOK] badge to prompt)
tok hooks-install

# Generate shell completions
tok completion bash | sudo tee /etc/bash_completion.d/tok
tok completion zsh > "${fpath[1]}/_tok"
tok completion fish > ~/.config/fish/completions/tok.fish
```

## Environment Variables

| Variable | Description |
|----------|-------------|
| `TOK_CONFIG_DIR` | Override config directory |
| `TOK_AUTO_ACTIVATE` | Auto-activate on startup (`1` = yes) |
| `TOK_DEFAULT_MODE` | Default compression mode |
| `TOK_DATABASE_PATH` | Override tracking database path |
| `TOK_NO_COLOR` | Disable colored output |
| `TOK_QUIET` | Suppress non-essential output |

## Contributing

We welcome contributions! See [CONTRIBUTING.md](CONTRIBUTING.md) for setup, coding standards, and PR guidelines.

Quick start for contributors:
```bash
git clone https://github.com/GrayCodeAI/tok.git && cd tok
make test        # Run all tests
make build       # Build binary
make lint        # Run linters
```

## License

[MIT](LICENSE) — Use freely, modify, and distribute.
