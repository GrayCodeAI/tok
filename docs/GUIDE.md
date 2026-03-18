# TokMan Getting Started Guide

**Your complete guide to reducing token usage in LLM interactions**

---

## Table of Contents

1. [Quick Start](#quick-start)
2. [Installation](#installation)
3. [Shell Integration](#shell-integration)
4. [Basic Usage](#basic-usage)
5. [Common Workflows](#common-workflows)
6. [Advanced Features](#advanced-features)
7. [Best Practices](#best-practices)
8. [FAQ](#faq)

---

## Quick Start

Get up and running in 30 seconds:

```bash
# Install
go install github.com/GrayCodeAI/tokman/cmd/tokman@latest

# Initialize (installs shell hook)
tokman init

# Reload shell
source ~/.bashrc  # or ~/.zshrc

# Start saving tokens!
git status  # Automatically rewritten to: tokman git status
```

---

## Installation

### Option 1: Go Install (Recommended)

```bash
go install github.com/GrayCodeAI/tokman/cmd/tokman@latest
```

### Option 2: Homebrew (macOS/Linux)

```bash
brew install GrayCodeAI/tap/tokman
```

### Option 3: Docker

```bash
# Pull from registry
docker pull ghcr.io/graycodeai/tokman:latest

# Or build locally
git clone https://github.com/GrayCodeAI/tokman.git
cd tokman
docker build -f docker/Dockerfile -t tokman:latest .
```

### Option 4: Build from Source

```bash
git clone https://github.com/GrayCodeAI/tokman.git
cd tokman
go build -o tokman ./cmd/tokman
sudo mv tokman /usr/local/bin/
```

---

## Shell Integration

### Automatic Setup

```bash
tokman init
```

This will:
1. Create the hook script at `~/.claude/hooks/tokman-rewrite.sh`
2. Store SHA-256 hash for integrity verification
3. Add `source` line to your shell config

### Manual Setup

Add to `~/.bashrc` or `~/.zshrc`:

```bash
# TokMan shell integration
if command -v tokman &> /dev/null; then
    eval "$(tokman init --hook-only)"
fi
```

### Verify Installation

```bash
tokman verify
# ✓ Hook integrity verified
```

### Shell Completions

```bash
# Bash
source <(tokman completion bash)

# Zsh (add to ~/.zshrc)
source <(tokman completion zsh)

# Fish
tokman completion fish | source
```

---

## Basic Usage

### Running Commands

TokMan wraps CLI commands to filter their output:

```bash
# Direct usage
tokman git status
tokman npm test
tokman docker ps

# Or use automatic rewriting (after init)
git status  # Automatically becomes: tokman git status
```

### Checking Savings

```bash
# Quick status
tokman status
# 🌸 TokMan Status
# Commands: 1,234 | Tokens Saved: 89,234 (71%)

# Detailed report
tokman report

# Comprehensive analysis with graphs
tokman gain --graph --history
```

### Token Counting

```bash
# Count tokens in text
tokman count "Hello, world!"
# 4 tokens

# Count tokens in file
tokman count main.go
# 156 tokens

# Use specific model encoding
tokman count --model gpt-4o "Hello, world!"
tokman count --model claude-3-sonnet "Hello, world!"

# Compare heuristic vs actual
tokman count --compare "Your text here"
# Heuristic: 5 tokens
# Actual (cl100k_base): 4 tokens
```

---

## Common Workflows

### Development Workflow

```bash
# Morning: Check status
tokman status

# Code changes
git status              # Filtered output
git diff                # Compact diff
tokman go test ./...    # Aggregated test results

# Build & verify
tokman go build ./...
tokman golangci-lint run

# Commit
git add .
git commit -m "feat: add new feature"

# Check daily savings
tokman gain --daily
```

### CI/CD Integration

```bash
# In GitHub Actions
tokman report --format json > tokman-report.json

# Generate summary for PR
tokman summary --format markdown >> $GITHUB_STEP_SUMMARY
```

### Multi-Project Analysis

```bash
# Filter to current project
tokman gain --project

# View all projects
tokman report --projects
```

### Debugging Failed Commands

When a command fails, TokMan automatically saves the full unfiltered output:

```bash
tokman go test ./...
# ❌ 1 test failed
# [full output saved: ~/.local/share/tokman/tee/1707753600_go_test.log]

# Read the full output
cat ~/.local/share/tokman/tee/1707753600_go_test.log
```

---

## Advanced Features

### Ultra-Compact Mode

Maximum token savings with ASCII-only output:

```bash
tokman -u git status
# M:3 S:2 U:5
# M src/main.go
```

### Custom Filter Plugins

Create plugins in `~/.config/tokman/plugins/`:

```json
{
  "name": "hide-npm-warnings",
  "description": "Hide npm deprecation warnings",
  "enabled": true,
  "patterns": ["npm WARN deprecated"],
  "mode": "hide"
}
```

Plugin commands:

```bash
tokman plugin list           # List loaded plugins
tokman plugin create myfilter # Create new plugin template
tokman plugin enable myfilter # Enable a plugin
tokman plugin disable myfilter # Disable a plugin
tokman plugin examples       # Generate example plugins
```

### Web Dashboard

Launch an interactive dashboard:

```bash
# Start on default port (8080)
tokman dashboard

# Custom port
tokman dashboard --port 3000

# Open browser automatically
tokman dashboard --open
```

API endpoints:
- `/api/stats` — Overall statistics
- `/api/history` — Command history
- `/api/projects` — List projects
- `/api/savings` — Savings breakdown

### Discovering Savings Opportunities

Find commands that could benefit from TokMan:

```bash
tokman discover
# 💡 Discovery Report
# Potentially rewritable commands:
#   - git log (run 234 times) → tokman git log
#   - npm test (run 156 times) → tokman npm test
# Estimated additional savings: 12,345 tokens
```

### Smart Code Summary

Get a 2-line heuristic summary of any code file:

```bash
tokman smart main.go
# Go module (5 fn, 2 struct) - 156 lines
# uses: fmt, os | patterns: error handling, tests
```

---

## Best Practices

### 1. Always Initialize

```bash
tokman init  # Sets up automatic rewriting
```

This ensures all supported commands are automatically filtered.

### 2. Use Ultra-Compact for Maximum Savings

```bash
tokman -u git status
tokman -u npm test
```

Additional 15-25% savings over standard mode.

### 3. Check Savings Regularly

```bash
# Daily habit
tokman status

# Weekly review
tokman gain --weekly --graph
```

### 4. Create Project-Specific Plugins

For project-specific noise, create a plugin:

```bash
tokman plugin create project-specific
# Edit ~/.config/tokman/plugins/project-specific.json
```

### 5. Use Tee for Debugging

Failed commands automatically save full output:

```bash
tokman npm test
# ❌ 2 tests failed
# [full output saved: ~/.local/share/tokman/tee/...]
```

### 6. Track Economics

```bash
tokman economics
# Shows cost savings in dollars
```

### 7. Exclude Sensitive Commands

If needed, exclude specific commands from rewriting:

```toml
# ~/.config/tokman/config.toml
[hooks]
excluded_commands = ["vault", "secret-tool"]
```

---

## FAQ

### Q: How does token counting work?

TokMan uses OpenAI's tiktoken library for accurate counting:
- **Default encoding**: `cl100k_base` (GPT-4, GPT-3.5-turbo, Claude)
- **GPT-4o**: `o200k_base`
- **GPT-3**: `p50k_base`

You can also use the simple heuristic: `(length + 3) / 4` tokens.

### Q: Does TokMan work with all LLMs?

Yes! While designed for Claude Code, TokMan works with any LLM that uses tokens:
- Claude (Anthropic)
- GPT-4, GPT-3.5 (OpenAI)
- Gemini (Google)
- Llama, Mistral, etc.

### Q: What if a command isn't supported?

Use `tokman proxy` to run without filtering (still tracked):

```bash
tokman proxy some-unsupported-command
```

Or suggest a new command on GitHub Issues.

### Q: How do I disable rewriting temporarily?

```bash
TOKMAN_DISABLED=1 git status  # Runs native git status
```

### Q: Where is my data stored?

Following XDG Base Directory Specification:
- **Config**: `~/.config/tokman/config.toml`
- **Database**: `~/.local/share/tokman/tokman.db`
- **Logs**: `~/.local/share/tokman/tokman.log`
- **Hooks**: `~/.claude/hooks/tokman-rewrite.sh`

### Q: Is there performance overhead?

TokMan is designed for minimal overhead:
- Target: <10ms per command
- Memory: <1MB for typical outputs
- Most overhead comes from the original command, not TokMan

### Q: Can I use TokMan in CI/CD?

Yes! TokMan has built-in CI integration:

```yaml
# GitHub Actions
- run: tokman summary --format markdown >> $GITHUB_STEP_SUMMARY
```

See `templates/` for complete examples.

### Q: How do I update TokMan?

```bash
# Go install
go install github.com/GrayCodeAI/tokman/cmd/tokman@latest

# Homebrew
brew upgrade tokman

# Docker
docker pull ghcr.io/graycodeai/tokman:latest
```

### Q: How do I uninstall?

```bash
# Remove binary
rm $(which tokman)

# Remove hook from shell config
# Edit ~/.bashrc or ~/.zshrc and remove the TokMan source line

# Remove data (optional)
rm -rf ~/.config/tokman ~/.local/share/tokman ~/.claude/hooks/tokman-rewrite.sh*
```

---

## Getting Help

- **Documentation**: [docs/](https://github.com/GrayCodeAI/tokman/tree/main/docs)
- **Issues**: [GitHub Issues](https://github.com/GrayCodeAI/tokman/issues)
- **Discussions**: [GitHub Discussions](https://github.com/GrayCodeAI/tokman/discussions)

---

## Next Steps

1. Run `tokman init` to set up shell integration
2. Try common commands: `git status`, `npm test`, `docker ps`
3. Check savings: `tokman status`
4. Explore advanced features: `tokman dashboard`, `tokman discover`
5. Create custom plugins for your workflow

🌸 **Happy token saving!**
