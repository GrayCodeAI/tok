# 🏗️ TokMan Architecture

This document explains how TokMan is structured and how its components work together.

---

## 📐 High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        User / AI Assistant                    │
└───────────────────────┬─────────────────────────────────────┘
                        │
                        ▼
┌─────────────────────────────────────────────────────────────┐
│                      CLI Commands (Cobra)                     │
│  ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐      │
│  │  git  │ │ docker│ │ npm  │ │cargo │ │ pytest│ │ ...  │      │
│  └──┬───┘ └──┬───┘ └──┬───┘ └──┬───┘ └──┬───┘ └──┬───┘      │
└─────┼────────┼────────┼────────┼────────┼─────────┼──────────┘
      │        │        │        │        │         │
      ▼        ▼        ▼        ▼        ▼         ▼
┌─────────────────────────────────────────────────────────────┐
│                    Filter Pipeline (31 layers)                │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐           │
│  │Entropy  │ │Perplexity│ │  H2O    │ │  Gist   │  ...      │
│  └─────────┘ └─────────┘ └─────────┘ └─────────┘           │
└───────────────────────┬─────────────────────────────────────┘
                        │
                        ▼
┌─────────────────────────────────────────────────────────────┐
│                     Output Processing                        │
│  ┌───────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐       │
│  │  Quality  │ │ Vis Diff  │ │  Merge   │ │  Export  │       │
│  └───────────┘ └──────────┘ └──────────┘ └──────────┘       │
└───────────────────────┬─────────────────────────────────────┘
                        │
                        ▼
┌─────────────────────────────────────────────────────────────┐
│                 Tracking & Analytics (SQLite)                 │
└─────────────────────────────────────────────────────────────┘
```

---

## 📁 Directory Structure

```
tokman/
├── cmd/tokman/main.go              # Entry point
├── internal/
│   ├── commands/                   # CLI commands (Cobra)
│   │   ├── registry/               # Command registration
│   │   ├── shared/                 # Global state & flags
│   │   ├── core/                   # Core commands (doctor, etc.)
│   │   ├── analysis/               # Stats, quality, benchmark
│   │   ├── output/                 # Rewrite, diff, merge
│   │   ├── system/                 # ls, grep, find, tree
│   │   ├── vcs/                    # git, gh, gt
│   │   ├── container/              # docker, kubectl
│   │   ├── pkgmgr/                 # npm, cargo, pip
│   │   ├── filtercmd/              # filter, tests, validate
│   │   └── init/                   # tokman init (hooks)
│   ├── filter/                     # 31-layer pipeline
│   │   ├── pipeline.go             # Pipeline coordinator
│   │   ├── entropy.go              # Layer 1
│   │   ├── perplexity.go           # Layer 2
│   │   ├── ...                     # Layers 3-31
│   │   └── presets.go              # Fast/balanced/full
│   ├── toml/                       # TOML filter system
│   │   ├── parser.go               # Parse TOML filters
│   │   ├── loader.go               # Discover & load filters
│   │   ├── test.go                 # Inline test framework
│   │   └── builtin/                # 97+ builtin filters
│   ├── tracking/                   # SQLite tracking
│   ├── quality/                    # Quality scoring (6 metrics)
│   ├── visual/                     # Visual diff tool
│   ├── core/                       # Command runner, estimator
│   └── config/                     # Configuration (Viper + TOML)
├── hooks/                          # Delegating hook scripts
├── Formula/                        # Homebrew formula
├── homebrew-tokman/                # Homebrew tap repo
├── docs/                           # Documentation
└── OSS-REF/                        # Competitor reference repos
```

---

## 🔄 Command Flow

### 1. User invokes command
```bash
tokman git status
```

### 2. Command resolution
- Root command receives `git status`
- `registry` package finds matching handler
- Command runner executes shell command

### 3. Output capture
- Command runs, output captured via `os/exec`
- Output piped to token estimator
- Output stored in tracking database

### 4. Filter pipeline
```
Raw Output → Layer 1 (Entropy) → Layer 2 (Perplexity) → ...
                                                              → Layer 31 (Agent Memory)
```

Each layer:
- Receives text input
- Applies transformation
- Returns filtered text + tokens saved
- Early exit if budget met

### 5. Quality analysis (optional)
- 6 metrics calculated
- Grade assigned (A+ to F)
- Recommendations generated

### 6. Output delivery
- Filtered output to stdout
- Stats to stderr
- Results tracked in SQLite

---

## 🧩 Key Components

### Filter Pipeline (31 Layers)

| Layer | Component | Purpose |
|-------|-----------|---------|
| 1 | Entropy | Remove low-information content |
| 2 | Perplexity | Iterative token removal |
| 3 | Goal-Driven | Select lines relevant to query |
| 4 | AST Preserve | Syntax-aware code compression |
| 5 | Contrastive | Question-relevance scoring |
| 6-20 | Various | N-grams, gist, attention, etc. |
| 21-31 | Advanced | LLM, memory, agent-specific |

### TOML Filter System

```toml
[git_status]
match_command = "^git(\\s+status)?(\\s+.*)?$"
strip_lines_matching = ["^\\s*$", "^On branch"]
max_lines = 50

[[tests.git_status]]
name = "clean working tree"
input = "On branch main\nnothing to commit"
expected = "nothing to commit"
```

### Hook System (Delegating Pattern)

```
AI Assistant → Hook Script → tokman rewrite → Exit Code (0-7) → Hook Script → AI
```

Exit Codes:
- 0: Rewrite found, auto-allow
- 1: No equivalent, pass-through
- 2: Deny rule matched
- 3: Rewrite found, ask user
- 4-7: Other conditions

---

## 🧪 Testing Strategy

### Unit Tests
- Each package has `_test.go` files
- Filter layers tested in isolation
- Registry and commands tested

### Inline Filter Tests
- TOML-based test declarations
- Run via `tokman tests`
- 41 tests across 17 filters

### Integration Tests
- End-to-end command testing
- Full pipeline validation
- Cross-platform verification

---

## ⚡ Performance

### Build Optimizations
- `CGO_ENABLED=0` for static binaries
- ldflags strip (`-s -w`) for smaller size
- `-gcflags="-trimpath"` for reproducible builds
- `UPX` compression available

### Runtime Optimizations
- Stage gates skip unnecessary layers
- Early exit when budget met
- SIMD support (AVX2/AVX-512/NEON)
- Streaming for large inputs (>500K tokens)
- Fingerprint caching

---

## 🔒 Security

### Hook Integrity
- SHA-256 hash verification
- Runtime checks prevent tampering
- Trust model for project directories

### Safety Checks
- Dangerous commands denied (rm, dd)
- Unsafe operations flagged (curl | bash)
- User confirmation required for sudo

---

## 📊 Tracking

### SQLite Database
- Command history
- Token savings per command
- Per-project analytics
- 24h and total statistics

### Schema
```sql
CREATE TABLE command_records (
  id INTEGER PRIMARY KEY,
  command TEXT,
  input_tokens INTEGER,
  output_tokens INTEGER,
  saved_tokens INTEGER,
  project_path TEXT,
  timestamp DATETIME
);
```

---

<div align="center">

**TokMan Architecture v0.1.0**

*Production-ready token-aware CLI proxy*

</div>