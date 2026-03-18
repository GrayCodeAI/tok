# RTK vs TokMan: Detailed Comparison

**Date**: 2026-03-18  
**Goal**: Comprehensive analysis of both token-saving CLI proxies

---

## Executive Summary

| Aspect | RTK (Rust) | TokMan (Go) | Winner |
|--------|------------|-------------|--------|
| **Language** | Rust 2021 | Go 1.24 | Tie |
| **Lines of Code** | 43,595 | 28,513 | TokMan (smaller) |
| **Test Count** | 961 tests | 543 tests | RTK |
| **Overhead** | <10ms | ~15ms | RTK |
| **Extensibility** | TOML DSL | Go code required | RTK |
| **Token Counting** | Heuristic only | tiktoken + heuristic | TokMan |
| **Unique Features** | verify, tee, session, trust | smart, discover, economics | Tie |

---

## 1. Architecture & Design

### RTK Architecture
```
┌─────────────────────────────────────────────┐
│              RTK Binary (Rust)              │
├─────────────────────────────────────────────┤
│  clap CLI parser                           │
│  ├── Built-in handlers (git, test, etc.)   │
│  └── TOML filter engine (8-stage pipeline) │
├─────────────────────────────────────────────┤
│  SQLite tracking + config (serde/toml)     │
│  Hooks: PreToolUse, OpenCode plugin        │
└─────────────────────────────────────────────┘
```

**Key Design Decisions**:
- **Zero-dependency binary**: Statically compiled, no runtime
- **Two-tier filtering**: Built-in Rust handlers + TOML DSL
- **8-stage pipeline**: strip_ansi → replace → match_output → strip/keep_lines → truncate → head/tail → max_lines → on_empty

### TokMan Architecture
```
┌─────────────────────────────────────────────┐
│            TokMan Binary (Go)               │
├─────────────────────────────────────────────┤
│  cobra CLI + viper config                  │
│  ├── Command handlers (internal/commands/) │
│  └── Filter engine (internal/filter/)      │
├─────────────────────────────────────────────┤
│  SQLite tracking + tiktoken integration    │
│  Hooks: PreToolUse, shell aliases          │
│  Dashboard: Web UI + REST API              │
└─────────────────────────────────────────────┘
```

**Key Design Decisions**:
- **Modular handlers**: Each command has dedicated Go file
- **Fixed filter set**: ANSI, comments, imports, aggregation
- **tiktoken integration**: Actual OpenAI tokenizer available

---

## 2. Codebase Statistics

| Metric | RTK | TokMan |
|--------|-----|--------|
| **Total Lines** | 43,595 | 28,513 |
| **Test Files** | 62 | 22 |
| **Test Cases** | 961 | 543 |
| **Source Files** | ~120 .rs | ~69 .go |
| **Dependencies** | 20 crates | 30+ modules |
| **Binary Size** | ~3MB (stripped) | ~15MB (with SQLite) |

### Language Breakdown
- **RTK**: 100% Rust
- **TokMan**: 100% Go

---

## 3. Token Counting

### RTK Token Counting
```rust
// Heuristic only (fast, no overhead)
fn estimate_tokens(text: &str) -> usize {
    (text.len() as f64 / 4.0).ceil() as usize
}
```

**Pros**: Zero overhead, consistent  
**Cons**: Less accurate for non-English text

### TokMan Token Counting
```go
// Primary: Heuristic (same as RTK)
func EstimateTokens(text string) int {
    return (len(text) + 3) / 4
}

// Optional: tiktoken (accurate)
func CountTokens(text string, model string) int {
    // Uses actual BPE tokenizer
}
```

**Pros**: Optional accurate counting with tiktoken  
**Cons**: Slightly more overhead when enabled

---

## 4. Performance Comparison

| Benchmark | RTK | TokMan | Notes |
|-----------|-----|--------|-------|
| **Git status (50 files)** | ~5ms | ~8ms | RTK faster |
| **Go test (100 tests)** | ~4ms | ~10ms | RTK faster |
| **Docker ps (100 containers)** | ~6ms | ~12ms | RTK faster |
| **Large output (1000 lines)** | ~8ms | ~15ms | RTK faster |
| **Typical overhead** | <10ms | ~15ms | Both acceptable |

### RTK Performance Optimizations
```toml
[profile.release]
opt-level = 3
lto = true
codegen-units = 1
panic = "abort"
strip = true
```

### TokMan Performance
- Uses `modernc.org/sqlite` (pure Go, no CGO)
- Streaming where possible
- Memory: <1MB for 100KB output

---

## 5. Extensibility

### RTK: TOML-Based DSL ✨
```toml
# ~/.config/rtk/filters/my-tool.toml
[[filters]]
name = "my-custom-tool"
command = "mytool"
strip_ansi = true
max_lines = 50
replace = [
  { from = "DEBUG:.*", to = "" },
]
on_empty = "ok"
```

**Benefits**:
- No recompilation needed
- Project-local filters (`.rtk/filters.toml`)
- Testable with `rtk verify`

### TokMan: Go Code Required
```go
// Must edit internal/commands/mytool.go
func runMyTool(cmd *cobra.Command, args []string) error {
    // ... implementation
}
```

**Benefits**:
- Full control and power
- Type safety
- IDE support

**Limitation**: Requires Go development for new commands

---

## 6. Unique Features

### RTK Unique Features
| Feature | Description |
|---------|-------------|
| `rtk verify` | Test TOML filter rules against sample inputs |
| `rtk tee` | Auto-save raw output on failure for recovery |
| `rtk session` | Track RTK adoption across Claude sessions |
| `rtk trust` | Security model for local filter files |
| `rtk smart` | 2-line heuristic code summary |
| Inline tests | Test cases embedded in TOML files |

### TokMan Unique Features
| Feature | Description |
|---------|-------------|
| `tokman smart` | 2-line heuristic code summary |
| `tokman discover` | Find missed savings in Claude Code history |
| `tokman economics` | Spending vs savings with quota estimates |
| `tokman count` | Actual tiktoken token counting |
| `tokman dashboard` | Web UI for analytics (8080) |
| Plugin system | JSON-based custom filters |
| Integrity verification | SHA-256 hook tampering protection |

---

## 7. Command Coverage

### Both Support (68+ commands)
```
Files:        ls, tree, read, find, grep, diff, wc
Git:          status, diff, log, add, commit, push, pull, branch, stash
GitHub:       gh pr, gh issue, gh run
Tests:        cargo test, go test, pytest, vitest, jest, npm test
Build:        cargo build, go build, next build
Lint:         eslint, biome, ruff, mypy, golangci-lint, clippy
Containers:   docker ps, docker logs, kubectl get, kubectl logs
Package:      pnpm, npm, pip, cargo
Network:      curl, wget
Data:         json, env, deps, log
```

### RTK-Only Commands
```
rtk verify     - Test TOML filters
rtk tee        - Failure recovery
rtk session    - Session tracking
rtk trust      - Local filter security
```

### TokMan-Only Commands
```
tokman discover   - Missed savings scan
tokman economics  - Cost analysis
tokman count      - tiktoken counting
tokman dashboard  - Web UI
tokman agents     - Agent registration
```

---

## 8. Configuration

### RTK Config (~/.config/rtk/config.toml)
```toml
[tracking]
database_path = "~/.local/share/rtk/history.db"

[hooks]
exclude_commands = ["curl", "playwright"]

[tee]
enabled = true
mode = "failures"  # "failures", "always", "never"
max_files = 20
```

### TokMan Config (~/.config/tokman/config.toml)
```toml
[tracking]
enabled = true
telemetry = false

[filter]
mode = "minimal"  # "minimal" or "aggressive"
noise_dirs = [".git", "node_modules", "target"]

[hooks]
excluded_commands = []
```

---

## 9. Agent Integration

### Both Support
- **Claude Code**: PreToolUse hooks installed via `init -g`
- **Auto-rewrite**: Transparent command interception
- **Bash tool only**: Built-in tools (Read, Grep, Glob) not intercepted

### RTK Integration
- OpenCode plugin (`opencode-rtk.ts`)
- `rtk init -g --opencode`

### TokMan Integration
- Shell completions (bash/zsh/fish)
- Docker image with dashboard
- CI/CD templates (GitHub Actions, GitLab CI)

---

## 10. Documentation Quality

### RTK Documentation
| Document | Description |
|----------|-------------|
| README.md | Multi-language (6 languages) |
| ARCHITECTURE.md | Technical deep-dive |
| AUDIT_GUIDE.md | Token savings analytics |
| TROUBLESHOOTING.md | Common issues |
| INSTALL.md | Detailed installation |
| SECURITY.md | Security policy |

### TokMan Documentation
| Document | Description |
|----------|-------------|
| README.md | Comprehensive feature list |
| docs/FEATURES.md | All 68 commands documented |
| docs/GUIDE.md | Getting started, workflows |
| docs/TROUBLESHOOTING.md | Common issues |
| docs/PERFORMANCE.md | Benchmark results |
| IMPLEMENTATION_PLAN.md | Development roadmap |

---

## 11. Installation Methods

### RTK Installation
```bash
# Homebrew (recommended)
brew install rtk

# Quick install
curl -fsSL https://raw.githubusercontent.com/rtk-ai/rtk/master/install.sh | sh

# Cargo
cargo install --git https://github.com/rtk-ai/rtk

# Pre-built binaries (Linux, macOS, Windows)
```

### TokMan Installation
```bash
# Homebrew
brew install GrayCodeAI/tap/tokman

# Build from source
git clone https://github.com/GrayCodeAI/tokman
cd tokman && go build ./cmd/tokman

# Docker
docker pull ghcr.io/graycodeai/tokman:latest
```

---

## 12. Summary: When to Use Which

### Choose RTK When:
- ✅ You want **maximum performance** (<10ms overhead)
- ✅ You need **extensibility without coding** (TOML DSL)
- ✅ You want **massive test coverage** (961 tests)
- ✅ You prefer **zero-dependency binary**
- ✅ You use **OpenCode** (dedicated plugin)

### Choose TokMan When:
- ✅ You want **accurate token counting** (tiktoken)
- ✅ You prefer **Go ecosystem**
- ✅ You need **web dashboard** for analytics
- ✅ You want **economics analysis** (cost tracking)
- ✅ You want **smaller codebase** (28k vs 43k lines)
- ✅ You need **discover** feature (find missed savings)

---

## 13. Feature Parity Status

| Feature | RTK | TokMan | Status |
|---------|-----|--------|--------|
| Token filtering | ✅ | ✅ | Parity |
| Ultra-compact mode | ✅ | ✅ | Parity |
| Tee on failure | ✅ | ✅ | Parity |
| Smart summary | ✅ | ✅ | Parity |
| TOML extensibility | ✅ | ❌ | RTK advantage |
| tiktoken counting | ❌ | ✅ | TokMan advantage |
| Web dashboard | ❌ | ✅ | TokMan advantage |
| Economics analysis | ❌ | ✅ | TokMan advantage |
| Discover command | ❌ | ✅ | TokMan advantage |

---

## Conclusion

**Both tools are production-ready** and achieve the same core goal: reducing LLM token consumption by 60-90%.

**RTK** excels in:
- Performance (<10ms)
- Extensibility (TOML DSL)
- Test coverage (961 tests)

**TokMan** excels in:
- Accurate token counting (tiktoken)
- Analytics features (dashboard, economics, discover)
- Go ecosystem integration

**Recommendation**: Use **RTK** for maximum performance and extensibility. Use **TokMan** for accurate token counting and advanced analytics.
