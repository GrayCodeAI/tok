# Tok Quick Start Guide

**Get started in under 30 seconds**

## One-Command Setup

```bash
tok quickstart
```

This automatically:
- Detects your AI agents (Claude Code, Cursor, Windsurf, etc.)
- Installs compression hooks
- Creates default configuration
- Verifies everything works

## What Happens Automatically

Tok uses a **4-tier adaptive system** that selects the right compression level based on content:

| Tier | Layers | When | Speed |
|------|--------|------|-------|
| 0 (Trivial) | 0 | Empty content | Instant |
| 1 (Simple) | 3 | <50 tokens, simple commands | <0.5ms |
| 2 (Medium) | 8 | Git diffs, tests, builds | <2ms |
| 3 (Full) | 20 | Large output, code, logs | <15ms |

**You don't need to configure anything** - Tok auto-detects content size and type.

## Presets (Optional)

If you want explicit control:

```bash
# Fast: minimal layer set, maximum speed
tok --preset=fast git status

# Balanced: default quality/latency tradeoff
tok --preset=balanced git diff

# Full: deepest available compression path
tok --preset=full cat large-file.log

# Auto: Let Tok decide (recommended)
tok --preset=auto make build
```

## Common Commands

```bash
# Check everything is working
tok doctor

# See your token savings
tok gain

# Find missed optimization opportunities
tok discover

# View current status
tok status

# Inspect or delete local telemetry data
tok telemetry --status
tok telemetry --forget
```

## How It Works

When you run commands through Tok:

```
git diff → tok → compressed output (60-90% fewer tokens)
```

The 20 compression layers are based on research from 120+ papers:
- **Entropy filtering** - Remove low-information tokens
- **Perplexity pruning** - Iterative token removal (Microsoft LLMLingua)
- **AST preservation** - Keep code structure intact
- **H2O filter** - Heavy-hitter detection (NeurIPS 2023)
- **And 16 more...**

## Manual Setup (if quickstart doesn't detect your agent)

```bash
# Claude Code
tok init --claude

# Cursor
tok init --cursor

# Remove an integration cleanly
tok init --claude --uninstall

# Windsurf
tok init --windsurf

# All detected agents
tok init --all
```

## Verification

```bash
tok doctor
```

Expected output:
```
tok doctor — diagnosing setup
================================
  ✓ Binary: /usr/local/bin/tok
  ✓ Config Dir: /home/user/.config/tok
  ✓ Database: /home/user/.local/share/tok/tracking.db
  ✓ Shell Hook: /home/user/.claude/hooks/tok-rewrite.sh
  ✓ PATH: /usr/local/bin/tok
  ✓ Platform: linux/amd64 Go go1.26.0
  ✓ Tokenizer: tiktoken-go (embedded)
  ✓ TOML Filters: 15 built-in filters
  ✓ Disk Space: database is 0.1MB
  ✓ Go: available (for development)
  ✓ Tier System: 4 tiers (0-3) with auto-detection

All checks passed!
```

## Need Help?

- `tok --help` - Command reference
- `tok <command> --help` - Command-specific help
- [Documentation](./docs/) - Full docs
- [GitHub Issues](https://github.com/GrayCodeAI/tok/issues) - Bug reports
