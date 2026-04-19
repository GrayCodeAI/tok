# tok

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![CI](https://github.com/GrayCodeAI/tok/actions/workflows/ci.yml/badge.svg)](https://github.com/GrayCodeAI/tok/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/GrayCodeAI/tok)](https://goreportcard.com/report/github.com/GrayCodeAI/tok)

> **Write less, get more.** Compress your prompts before sending. Filter noisy output for readability.

---

## What tok Does

### 1. Compress Your Prompts (Saves Input Tokens)

You write a verbose prompt → tok compresses it → the compressed version is sent to the AI → **fewer input tokens charged**.

```
Before: "Hey, could you please help me figure out why this React
         component keeps re-rendering every time the props change?"
After:  "React component re-renders on prop change. Why?"

38 tokens → 9 tokens (76% saved on input)
```

### 2. Filter Terminal Output (Readability + Context Savings)

tok intercepts command output and removes noise. This doesn't save tokens on the AI's response (those are already generated), but it:
- Makes terminal output **readable** — shows only what matters
- Saves tokens when filtered output is **fed back** into another AI call

```
$ tok npm test
# 200 lines of test output → 3 lines: pass/fail + failures

$ tok git diff
# 500-line diff → only the changed lines
```

### 3. Set AI Agent Tone

Install rules that tell coding agents to respond tersely, saving tokens on their responses.

```bash
tok install-agents    # One command, 12 agents configured
```

---

## Install

```bash
go install github.com/GrayCodeAI/tok/cmd/tok@latest
```

Or build from source:

```bash
git clone https://github.com/GrayCodeAI/tok.git && cd tok && make build
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

**6 modes:**

| Mode | Style | Input Savings |
|------|-------|--------------|
| `lite` | Drop filler, keep grammar | ~20% |
| `full` | Drop articles, fragments OK | ~40% _(default)_ |
| `ultra` | Telegraphic, abbreviations | ~60% |
| `wenyan-lite` | Classical Chinese light | ~30% |
| `wenyan` | Classical Chinese standard | ~50% |
| `wenyan-ultra` | Classical Chinese max | ~70% |

### Filter Output

Prefix any command with `tok`:

```bash
tok npm test       # Clean test results
tok git diff       # Only changed lines
tok docker ps -a   # Essential container info
tok cargo build    # Build output, no noise
```

**100+ commands wrapped** — git, npm, cargo, go, docker, kubectl, pytest, jest, ruff, and more.

### Set Agent Tone

```bash
tok on ultra       # Tell agents: respond with maximum brevity
tok on lite        # Professional but tight
tok status         # Current mode
tok gain           # Token savings from input compression
```

---

## How It Works

### Input Compression

Your text goes through a compression engine that removes filler words, articles, and redundancy while preserving technical meaning. The compressed text is what gets sent to the AI.

```
Your prompt → tok compressor → compressed text → AI (charged for fewer tokens)
```

### Output Filtering

Command output passes through a 31-layer pipeline that removes noise, deduplicates, and highlights important lines. The filtered output is what you see in your terminal.

```
Command output → tok filter pipeline (31 layers) → clean output → your terminal
```

If you pipe filtered output into another AI call, you save tokens on that next call.

### Agent Rules

tok installs instruction files into AI coding agent directories (Cursor, Claude Code, Copilot, etc.) that tell the agent to respond tersely. The agent generates fewer tokens in its responses.

```
tok install-agents → agent rule files → agent responds tersely → fewer output tokens
```

---

## Key Features

| Feature | What It Does | Saves Tokens? |
|---------|-------------|---------------|
| Input compression | Compresses your prompts before sending | ✅ Input tokens |
| Output filtering | Cleans terminal output for readability | ❌ (but saves context on re-use) |
| Agent rules | Tells agents to respond tersely | ✅ Output tokens (agent-side) |
| Token tracking | Tracks your input compression savings | N/A (analytics) |

---

## Architecture

```
tok
├── cmd/tok/              CLI entry point
├── internal/
│   ├── commands/         100+ command wrappers
│   ├── compressor/       Input compression (6 modes)
│   ├── filter/           Output pipeline (31 layers)
│   └── tracking/         SQLite usage database
├── agents/               12 AI agent rule files
└── hooks/                Shell integration scripts
```

Single binary, no runtime dependencies.

---

## Configuration

```toml
# ~/.config/tok/config.toml
[core]
mode = "full"
auto_activate = true
```

| Variable | Default | Purpose |
|----------|---------|---------|
| `TOK_CONFIG_DIR` | `~/.config/tok` | Config location |
| `TOK_AUTO_ACTIVATE` | _(empty)_ | Auto-start on shell init |
| `TOK_DEFAULT_MODE` | `full` | Default compression |
| `TOK_NO_COLOR` | _(empty)_ | Disable colors |

---

## Shell Integration

```bash
tok hooks-install              # Add [TOK] badge to prompt
tok completion bash > /etc/bash_completion.d/tok
tok completion zsh > "${fpath[1]}/_tok"
tok man /usr/local/share/man/man1
```

---

## Commands

```
Core:       doctor  status  gain  on  off  mode  layers  suggest
Input:      compress  terse  restore
Output:     git  npm  cargo  go  docker  kubectl  pytest  jest  ... (100+)
Agents:     install-agents  uninstall-agents  init
Hooks:      hooks-install  hooks-uninstall
System:     completion  man  self-update  config  telemetry
Tools:      diff  explain  summary  json  merge  export
Tracking:   recall  undo  audit  trust  untrust
```

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
