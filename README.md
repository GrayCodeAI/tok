# tok

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![CI](https://github.com/GrayCodeAI/tok/actions/workflows/ci.yml/badge.svg)](https://github.com/GrayCodeAI/tok/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/GrayCodeAI/tok)](https://goreportcard.com/report/github.com/GrayCodeAI/tok)
[![codecov](https://codecov.io/gh/GrayCodeAI/tok/branch/main/graph/badge.svg)](https://codecov.io/gh/GrayCodeAI/tok)

> **Cut LLM token costs by up to 90%** — compress prompts before sending, filter noisy output after receiving. One binary, zero dependencies.

---

## The Problem

Every LLM interaction wastes tokens:

```
You type: "Hey, could you please help me figure out why
           this React component keeps re-rendering every
           time the props change? I'd really appreciate it."
                    ↓
        AI receives: 38 tokens (14 are filler)

AI responds with 200 lines of test output, dependency trees,
and stack traces. You only needed the 3 relevant lines.
                    ↓
        Context window: 87% consumed by noise
```

## The Solution

tok sits between you and the AI, trimming fat on both sides:

```
Your prompt ──compress──▶ AI (fewer tokens, same meaning)
AI output   ◀──filter──── Your terminal (signal only)
```

---

## Install

```bash
go install github.com/GrayCodeAI/tok/cmd/tok@latest
```

Or build from source:

```bash
git clone https://github.com/GrayCodeAI/tok.git
cd tok && make build
./tok --help
```

---

## Usage

### Compress Prompts

```bash
$ tok compress -mode ultra -input "Please implement a user authentication system with JWT tokens"
Implement user auth w/ JWT.
```

```bash
$ echo "Could you explain why this React component keeps re-rendering?" | tok compress
React component re-renders. Why?
```

**6 compression modes:**

| Mode | Example | Best For |
|------|---------|----------|
| `lite` | Drop filler, keep grammar | Professional emails |
| `full` | Drop articles, fragments OK | Everyday prompts _(default)_ |
| `ultra` | Telegraphic style | Code queries |
| `wenyan-lite` | Classical Chinese light | CJK prompts |
| `wenyan` | Classical Chinese standard | CJK prompts |
| `wenyan-ultra` | Classical Chinese max | CJK prompts |

### Filter Terminal Output

Just prefix any command with `tok`:

```bash
$ tok npm test
# Verbose test output → clean pass/fail summary

$ tok git diff
# 500-line diff → only changed lines

$ tok docker ps -a
# Container table → essential info only
```

**100+ commands wrapped** — git, npm, cargo, go, docker, kubectl, pytest, jest, and more. tok auto-detects the command and applies the right filter.

### Set AI Agent Tone

Tell coding agents to respond tersely:

```bash
$ tok on ultra     # Maximum brevity
$ tok on lite      # Professional but tight
$ tok status       # Current mode
$ tok gain         # Token savings dashboard
```

---

## Key Features

### 31-Layer Compression Pipeline

Research-backed algorithms from top labs:

| Layer | Technique | Source |
|-------|-----------|--------|
| L1 | Entropy pruning | Shannon |
| L2 | Perplexity filtering | LLMLingua (Microsoft) |
| L4 | AST-aware compression | LongCodeZip (NUS) |
| L8 | Gist tokens | Stanford/Berkeley |
| L13 | Heavy-hitter preservation | H2O |
| L14 | Attention sink | StreamingLLM |
| L16 | Semantic chunking | ChunkKV |
| +24 more | — | See `tok layers` |

### Token Tracking

```bash
$ tok gain

Session Savings
┌──────────┬─────────┬─────────┬──────────┐
│ Command  │ Original│ Filtered│ Saved    │
├──────────┼─────────┼─────────┼──────────┤
│ npm test │ 12,400  │ 1,860   │ 85.0%    │
│ git diff │ 8,200   │ 980     │ 88.0%    │
│ cargo b  │ 6,100   │ 1,220   │ 80.0%    │
└──────────┴─────────┴─────────┴──────────┘
Total: 26,700 → 4,060 tokens (84.8% saved)
```

### AI Agent Rules

One command installs terse mode for 12 coding agents:

```bash
$ tok install-agents
Installed: cursor/tok.mdc → ~/.cursor/rules/tok.mdc
Installed: claude-code/tok.md → ~/.claude/tok.md
... (12 agents total)
```

Supports: Cursor, Windsurf, Cline, Copilot, Claude Code, Aider, Continue, Roo Code, Cody, CodeWhisperer, Tabnine, Codeium.

---

## Architecture

```
tok
├── cmd/tok/              Cobra CLI entry point
├── internal/
│   ├── commands/         100+ command wrappers (20 categories)
│   ├── compressor/       Input compression engine (6 modes)
│   ├── filter/           Output pipeline (31 layers)
│   ├── output/           Centralized output abstraction
│   ├── tracking/         SQLite token usage database
│   └── telemetry/        Anonymous metrics (opt-out)
├── agents/               12 AI agent rule files
├── hooks/                Shell integration scripts
└── config/               TOML filters + examples
```

**Single binary, no runtime dependencies.** Built with Go, uses SQLite for local tracking, and embeds all agent rules and hook scripts.

---

## Configuration

```toml
# ~/.config/tok/config.toml
[core]
mode = "full"
auto_activate = true

[tracking]
database_path = "~/.local/share/tok/tracking.db"
```

All paths are configurable via environment variables:

| Variable | Default | Purpose |
|----------|---------|---------|
| `TOK_CONFIG_DIR` | `~/.config/tok` | Config location |
| `TOK_AUTO_ACTIVATE` | _(empty)_ | Auto-start on shell init |
| `TOK_DEFAULT_MODE` | `full` | Default compression |
| `TOK_DATABASE_PATH` | `~/.local/share/tok/tracking.db` | Tracking DB |
| `TOK_NO_COLOR` | _(empty)_ | Disable colors |

---

## Shell Integration

Add a `[TOK]` badge to your prompt:

```bash
tok hooks-install    # Adds to ~/.zshrc and ~/.bashrc
```

Generate completions:

```bash
tok completion bash   > /etc/bash_completion.d/tok
tok completion zsh    > "${fpath[1]}/_tok"
tok completion fish   > ~/.config/fish/completions/tok.fish
tok completion powershell > $PROFILE
```

Generate man pages:

```bash
tok man /usr/local/share/man/man1
```

---

## Commands

```
Core:       doctor  status  gain  on  off  mode  layers  suggest
Input:      compress  terse  restore  template
Output:     git  npm  cargo  go  docker  kubectl  pytest  jest  ... (100+)
Filter:     filter-create  filter-validate  filter-bench  tests
Agents:     install-agents  uninstall-agents  init
Hooks:      hooks-install  hooks-uninstall  hooks-install-pwsh
System:     completion  man  self-update  config  telemetry
Tools:      diff  explain  summary  json  merge  rewrite  export
Tracking:   recall  undo  audit  verify  trust  untrust
Session:    session  engram  learn  cache  clean
Build:      build  test  lint  format  benchmark  batch
```

---

## Contributing

```bash
git clone https://github.com/GrayCodeAI/tok.git && cd tok
make test    # Run tests
make build   # Build binary
make lint    # Run linters
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for the full guide.

---

## License

[MIT](LICENSE)
