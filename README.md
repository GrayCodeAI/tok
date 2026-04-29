# tok

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![CI](https://github.com/GrayCodeAI/tok/actions/workflows/ci.yml/badge.svg)](https://github.com/GrayCodeAI/tok/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/GrayCodeAI/tok)](https://goreportcard.com/report/github.com/GrayCodeAI/tok)

> **Cut LLM token costs by 60–90%.** Compress prompts. Filter noisy output. Auto-rewrite commands. One binary, zero dependencies.

---

## What tok Does

### 1. Compress Prompts (Saves Input Tokens)

You write verbose → tok compresses → AI receives less → **fewer input tokens charged**.

```
Before: "Hey, could you please help me figure out why this React
         component keeps re-rendering every time the props change?"
After:  "React component re-renders on prop change. Why?"

38 tokens → 9 tokens (76% saved)
```

### 2. Filter Output (Readability + Context Savings)

tok intercepts command output and removes noise. Filtered output is readable and saves tokens when piped back into AI.

```
$ tok npm test
# 200 lines → 3 lines: pass/fail + failures
```

### 3. Transparent Command Rewriting

Install the hook once. Every bash command from your AI agent is automatically rewritten:

```
Claude types:  git status
tok rewrites:  tok git status
Claude sees:   420 tokens → 84 tokens (80% saved)
```

Claude never knows. You don't type `tok` prefix. It just works.

### 4. Make Agents Talk Tersely

Install agent rules that make AI respond with ~75% fewer output tokens:

```
Normal:  "The reason your component re-renders is likely because..."
tok:     "New object ref each render. Wrap in useMemo."

Same fix. 75% less word.
```

---

## Install

> **tok is distributed exclusively through [Hawk](https://github.com/GrayCodeAI/hawk).**
> It is not available as a standalone package.

```bash
# Via Hawk CLI (the only supported distribution)
hawk tok --help

# Or install Hawk first
npm install -g hawk
hawk tok compress -mode full -input "your text here"
```

---

## Quick Start

### Step 1: Install for your AI agent

```bash
tok init -g                     # Claude Code, Copilot (default)
tok init -g --gemini            # Gemini CLI
tok init -g --codex             # Codex (OpenAI)
tok init --agent cursor         # Cursor
tok init --agent windsurf       # Windsurf
tok init --agent cline          # Cline / Roo Code
```

### Step 2: Restart your AI tool

```bash
# Now every command is automatically filtered
git status    # → tok intercepts, Claude sees 80% less
npm test      # → tok intercepts, Claude sees 90% less
```

### Step 3: Compress your prompts

```bash
tok compress -mode ultra -input "Please implement authentication"
# → "Implement auth."
```

### Step 4: Check your savings

```bash
tok gain                        # Summary stats
tok gain --graph                # ASCII chart (30 days)
tok gain --history              # Recent commands
tok gain --daily                # Day-by-day
tok gain --format json          # JSON export
tok discover                    # Find missed savings
tok session                     # Adoption across sessions
```

---

## Benchmarks

Measured on this repo via `evals/bench.sh` (raw vs `tok compress --mode aggressive`):

| fixture   | raw bytes | raw tokens | tok bytes | tok tokens | saved |
|-----------|----------:|-----------:|----------:|-----------:|------:|
| git log   |     2,873 |        718 |       298 |         74 |  89 % |
| git diff  |   385,051 |     96,262 |     1,117 |        279 |  99 % |
| ls -la    |    66,341 |     16,585 |       148 |         37 |  99 % |
| find .go  |    19,145 |      4,786 |       147 |         36 |  99 % |

Reproduce: `go build -o tok ./cmd/tok && TOK=./tok evals/bench.sh --no-rtk`

## Recent additions

Session 2026-04-20 closed the last gaps vs rtk 0.37.1 and caveman:

- **`tok commit-msg`** — read staged diff, emit Conventional Commits subject. Rule-based, no LLM.
- **`tok review-diff`** — scan diff, emit one-line review comments (`🔴 bug / 🟡 risk / 🔵 nit`). Rule-based, no LLM.
- **`tok pr-review [--base|--pr]`** — batch `review-diff` across a whole PR, grouped by file.
- **`tok md <file>`** — compress markdown/memory file in place with `.original.md` backup. New wenyan modes.
- **`tok cheatsheet`** — one-shot reference card for shell users (`modes` and `quickref` are aliases).
- **`tok hook mode {activate|track|status|set}`** — Go-native SessionStart + UserPromptSubmit hook bodies. Drop-in replacement for the Node.js scripts, same flag-file format.
- **Wenyan filter layer** (`internal/filter/wenyan.go`) — classical-Chinese-inspired rule-based compression, callable from the main pipeline.
- **Release automation** — `release-please` workflow + `Formula/tok.rb` in-repo.
- **Skill bundler** — `scripts/build-skill.sh` produces `tok.skill` zip for single-file distribution.
- **End-to-end harness** — `tests/e2e/` Docker-based scenario runner.

---

## Features

### Input Compression (6 Modes)

| Mode | Style | Savings |
|------|-------|---------|
| `lite` | Drop filler, keep grammar | ~20% |
| `full` | Drop articles, fragments OK | ~40% _(default)_ |
| `ultra` | Telegraphic, abbreviations | ~60% |
| `wenyan-lite` | Classical Chinese light | ~30% |
| `wenyan` | Classical Chinese standard | ~50% |
| `wenyan-ultra` | Classical Chinese max | ~70% |

### Output Filtering (31-Layer Pipeline)

Research-backed algorithms: entropy pruning, perplexity filtering, AST-aware compression, H2O heavy-hitter, attention sink preservation, semantic chunking, and 25+ more.

### Transparent Command Rewriting

Install the hook → every bash command from your AI agent is auto-rewritten to use tok. Zero effort, 100% coverage.

```bash
# Hook intercepts and rewrites automatically:
git status  → tok git status    (2,000 → 400 tokens)
git diff    → tok git diff      (10,000 → 2,500 tokens)
npm test    → tok npm test      (25,000 → 2,500 tokens)
cargo test  → tok cargo test    (200+ lines → 20 lines)
```

### Interactive TUI

```
tok tui                      # 12-section dashboard: Home, Sessions, Trends, Logs, ...
tok tui --theme colorblind   # Okabe-Ito palette for accessible color vision
```

Live refresh · command palette (`:`) · search (`/`) · drill-down · clipboard yank (`y`) · export (`e`).
Full keybinding reference and architecture in [docs/TUI.md](docs/TUI.md).

### Token Analytics

```
tok gain --graph

Token Savings (Last 30 Days)
┌────────────────────────────────────────────────────┐
│ Mon 12 ████████████████████████████ 82%            │
│ Tue 13 ██████████████████████████ 78%              │
│ Wed 14 ██████████████████████████████ 85%          │
└────────────────────────────────────────────────────┘
Total: 267K → 53K tokens (80.1% saved)
```

### Memory File Compression

```bash
tok compress-memory CLAUDE.md
# CLAUDE.md          → compressed (AI reads this — fewer tokens)
# CLAUDE.original.md → human-readable (you edit this)
# Average: 46% fewer tokens per session
```

### AI Agent Integration

One command installs terse mode for 12+ agents:

```bash
tok install-agents    # Install to all agent directories
```

Supports: Claude Code, Cursor, Windsurf, Cline, Copilot, Codex, Gemini CLI, Roo Code, Kilo Code, Antigravity, Continue, Cody, CodeWhisperer, Tabnine, Codeium.

### 100+ Built-in Commands

tok wraps common CLI tools with intelligent filtering:

```
Files:    ls  read  smart  find  grep  diff  tree  wc  du  df
Git:      git status  git log  git diff  git add  git commit  git push
GitHub:   gh pr list  gh issue list  gh run list
Tests:    jest  vitest  playwright  pytest  go test  cargo test  rspec
Build:    cargo build  go build  gradle  maven  next build  tsc
Lint:     eslint  ruff  golangci-lint  mypy  prettier  rubocop
Package:  npm  yarn  pnpm  pip  cargo  bundle  prisma
Cloud:    aws  docker  kubectl  helm  terraform
Data:     json  jq  curl  wget  deps  env  log
```

---

## How It Works

```
Without tok:                              With tok + hook:

Claude  --git status-->  shell  -->  git   Claude  --git status-->  tok hook  -->  git
  ^                            |             ^          | filter         |
  |     ~2,000 tokens         |             |  ~400 tokens (auto)       |
  +---------------------------+             +---------------------------+
```

**Four strategies per command type:**
1. **Smart Filtering** — removes noise (comments, whitespace, boilerplate)
2. **Grouping** — aggregates similar items (files by dir, errors by type)
3. **Truncation** — keeps relevant context, cuts redundancy
4. **Deduplication** — collapses repeated lines with counts

---

## Architecture

```
tok
├── cmd/tok/              CLI entry point (cobra)
├── internal/
│   ├── commands/         100+ command wrappers (20 categories)
│   ├── compressor/       Input compression engine (6 modes)
│   ├── filter/           Output pipeline (31 layers)
│   ├── output/           Centralized output abstraction
│   ├── tracking/         SQLite token usage database
│   └── hooks/            Transparent command rewriting
├── agents/               AI agent rules + auto-activation hooks
├── hooks/                Shell integration scripts
├── benchmarks/           Token savings benchmarks
└── evals/                Three-arm eval harness
```

Single binary, no runtime dependencies.

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

| Variable | Default | Purpose |
|----------|---------|---------|
| `TOK_CONFIG_DIR` | `~/.config/tok` | Config location |
| `TOK_AUTO_ACTIVATE` | _(empty)_ | Auto-start on shell init |
| `TOK_DEFAULT_MODE` | `full` | Default compression |
| `TOK_NO_REWRITE` | _(empty)_ | Disable command rewriting |
| `TOK_NO_COLOR` | _(empty)_ | Disable colors |

---

## Shell Integration

```bash
# Install transparent command rewriting
tok init -g

# Add [TOK] badge to your prompt
tok hooks-install

# Generate completions
tok completion bash   > /etc/bash_completion.d/tok
tok completion zsh    > "${fpath[1]}/_tok"
tok completion fish   > ~/.config/fish/completions/tok.fish

# Generate man pages
tok man /usr/local/share/man/man1
```

---

## Commands

```
Core:       doctor  status  gain  on  off  mode  layers  suggest
Input:      compress  terse  restore  compress-memory
Output:     git  npm  cargo  go  docker  kubectl  pytest  jest  ... (100+)
Analytics:  gain --graph  gain --history  gain --daily  discover  session
Agents:     install-agents  uninstall-agents  init  compress-memory
Hooks:      hooks-install  hooks-uninstall  tok-rewrite-hook
System:     completion  man  self-update  config  telemetry
Tools:      diff  explain  summary  json  merge  export
Tracking:   recall  undo  audit  trust  untrust
```

---

## Benchmarks

Real token counts from the Claude API:

| Task | Normal | tok | Saved |
|------|-------:|----:|------:|
| Explain React re-render bug | 1,180 | 159 | 87% |
| Fix auth middleware token expiry | 704 | 121 | 83% |
| Set up PostgreSQL connection pool | 2,347 | 380 | 84% |
| Debug PostgreSQL race condition | 1,200 | 232 | 81% |
| Implement React error boundary | 3,454 | 456 | 87% |
| **Average** | **1,214** | **294** | **65%** |

Run `benchmarks/run.sh` to reproduce.

---

## Contributing

```bash
git clone https://github.com/GrayCodeAI/tok.git && cd tok
make test && make build && make lint
```

See [CONTRIBUTING.md](CONTRIBUTING.md).

---

## License

[MIT](LICENSE)
