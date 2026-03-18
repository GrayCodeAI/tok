# TokMan Features Documentation

**Complete guide to all 50+ commands, syntax, and token savings**

---

## Table of Contents

1. [Query-Aware Compression](#query-aware-compression) ⭐ NEW
2. [Hierarchical Summarization](#hierarchical-summarization) ⭐ NEW
3. [Local LLM Integration](#local-llm-integration) ⭐ NEW
4. [Multi-File Context Optimization](#multi-file-context-optimization) ⭐ NEW
5. [Custom LLM Prompt Templates](#custom-llm-prompt-templates) ⭐ NEW
6. [Core Commands](#core-commands)
7. [Git Commands](#git-commands)
8. [Infrastructure Commands](#infrastructure-commands)
9. [Build & Test Commands](#build--test-commands)
10. [Package Manager Commands](#package-manager-commands)
11. [Utility Commands](#utility-commands)
12. [Analysis Commands](#analysis-commands)
13. [Output Modes](#output-modes)
14. [Token Savings by Command](#token-savings-by-command)

---

## Query-Aware Compression ⭐ NEW

Tailor output filtering based on your agent's task intent for better context quality.

```bash
# CLI flag
tokman --query debug cargo test
tokman --query review git diff
tokman --query deploy docker ps

# Environment variable
TOKMAN_QUERY=debug tokman npm test
```

**Supported Intents:**

| Intent | Prioritizes | Use Case |
|--------|-------------|----------|
| `debug` | Errors, stack traces, failures | Finding bugs |
| `review` | Diffs, changes, file references | Code review |
| `deploy` | Status, versions, health | Deployments |
| `search` | File names, definitions | Finding code |
| `test` | Test results, coverage | Testing |
| `build` | Errors, warnings | Build status |

**Example - Debug Mode:**
```bash
$ tokman --query debug cargo test

# Output focuses on:
# ✗ test_auth_failed - assertion error at auth.rs:45
# ✗ test_connection_timeout - timed out after 30s
# Stack trace: ...
# 
# Skipped: passing test output, progress bars, verbose logs
```

**Token Savings:** +10-20% quality-weighted savings

See [Advanced Compression](./ADVANCED_COMPRESSION.md) for full documentation.

---

## Hierarchical Summarization ⭐ NEW

Multi-level summarization for very large outputs (500+ lines) - automatically compresses verbose output into structured summaries.

**Automatic activation**: Triggers when output exceeds 500 lines (~10K tokens)

**How it works:**
- Segments output into logical sections by detecting boundaries
- Scores sections by importance (errors > warnings > success)
- Preserves high-importance sections verbatim
- Compresses mid-importance sections into one-line summaries
- Drops low-importance sections entirely

**Example:**
```
[Hierarchical Summary: 1000 lines → 15 sections]

├─ [L1-200] Compiling dependencies... (200 lines, score: 0.25)
error[E0277]: the trait bound `String: Into<i32>` is not satisfied
  --> src/main.rs:10:5
   |
10 |     let x: i32 = String::new().into();
   |                 ^^^^^^^^^^^^^^^^^^^^^ the trait `From<String>` is not implemented
├─ [L250-300] warnings about unused variables (50 lines, score: 0.45)
```

**Token Savings:** Up to 10x compression for large outputs

---

## Local LLM Integration ⭐ NEW

Intelligent summarization using local LLMs (Ollama, LM Studio) for 40-60% better semantic preservation.

```bash
# CLI flag (requires Ollama or LM Studio running)
tokman --llm cargo test

# Environment variable
TOKMAN_LLM=true tokman npm test

# Configure LLM provider
TOKMAN_LLM_PROVIDER=ollama
TOKMAN_LLM_MODEL=llama3.2:3b
TOKMAN_LLM_BASE_URL=http://localhost:11434
```

**Supported Providers:**

| Provider | Default URL | Models |
|----------|-------------|--------|
| Ollama | http://localhost:11434 | llama3.2, mistral, phi3 |
| LM Studio | http://localhost:1234 | Any GGUF model |
| OpenAI-compatible | Configurable | Any compatible API |

**Features:**
- Auto-detects running local LLM
- Intent-aware prompts (debug/review/test/build)
- Automatic fallback to semantic filter if unavailable
- Summary caching for repeated content

**Performance:**
- Latency: 50-200ms (depends on hardware)
- Quality: 40-60% better semantic preservation
- Privacy: All processing is local

**Token Savings:** Higher quality preservation, variable compression

---

## Multi-File Context Optimization ⭐ NEW

Cross-file deduplication for projects with multiple related files - reduces redundancy while preserving relationships.

**Automatic activation**: Detects file markers (`=== File:`, `diff --git`, `---`)

**How it works:**
- Parses combined output to identify individual files
- Detects file relationships (imports, same-module, content similarity)
- Deduplicates shared imports across files
- Creates unified output with shared content extracted

**Example:**
```bash
# Before: 3 files with shared imports
=== File: main.go ===
import "fmt"
import "os"
...

=== File: utils.go ===
import "fmt"  # Duplicate!
import "strings"
...

# After: Shared imports consolidated
=== Shared Imports ===
import "fmt"

=== File: main.go ===
import "os"
...

=== File: utils.go ===
import "strings"
...
```

**Features:**
- Supports diff format (`diff --git a/ b/`)
- Configurable similarity threshold for deduplication
- Preserves file boundaries in output
- Aggressive mode shows signatures only

**Token Savings:** 10-30% for multi-file outputs

---

## Custom LLM Prompt Templates ⭐ NEW

Define custom prompts for different LLM summarization scenarios.

**Built-in Templates:**

| Template | Intent | Focus |
|----------|--------|-------|
| `debug` | Debugging | Errors, stack traces, failures |
| `review` | Code review | Changes, issues, API changes |
| `test` | Testing | Results, coverage, failures |
| `build` | Build | Status, errors, warnings |
| `deploy` | Deployment | Status, health, resources |
| `search` | Search | Files, definitions, imports |
| `concise` | General | Brief 3-5 sentence summary |
| `detailed` | General | Full technical details |

**Custom Templates:**
```bash
# Template stored in ~/.local/share/tokman/prompts/my_template.json
{
  "name": "security_review",
  "description": "Security-focused code review",
  "system_prompt": "You are a security auditor.",
  "user_prompt": "Focus on: vulnerabilities, auth issues, data exposure\n\n{{content}}",
  "intent": "review",
  "max_tokens": 400,
  "temperature": 0.2
}
```

**API Usage:**
```go
mgr := llm.NewDefaultPromptTemplateManager()
template, _ := mgr.GetTemplate("debug")
prompt := mgr.BuildPrompt(template, content, nil)
```

**Token Savings:** Variable (improves LLM summarization quality)

---

## Core Commands

### `tokman init`

Initialize TokMan and install shell hook for automatic command rewriting.

```bash
tokman init                    # Install hook to ~/.bashrc or ~/.zshrc
tokman init --hook-only        # Output hook content only
tokman init --shell bash       # Force specific shell
```

**What it does:**
- Creates `~/.claude/hooks/tokman-rewrite.sh`
- Stores SHA-256 hash for integrity verification
- Adds `source` line to shell config

---

### `tokman status`

Quick token savings summary.

```bash
tokman status
# Output:
# 🌸 TokMan Status
# Commands: 1,234 | Tokens Saved: 89,234 (71%)
```

---

### `tokman report`

Detailed usage analytics with breakdowns.

```bash
tokman report                  # Full report
tokman report --daily          # Daily breakdown
tokman report --weekly         # Weekly breakdown
tokman report --top 10         # Top 10 commands
```

---

### `tokman gain`

Comprehensive savings analysis with graphs, history, and quota estimates.

```bash
tokman gain                    # Basic summary
tokman gain --graph            # ASCII graph of daily savings
tokman gain --history          # Recent command history
tokman gain --quota --tier pro # Quota analysis (Pro tier)
tokman gain --format json      # Export as JSON
tokman gain --all              # All breakdowns
```

**Output includes:**
- Total commands run
- Tokens saved (original vs filtered)
- Savings percentage
- Top commands by savings
- Daily/weekly/monthly trends
- Estimated cost savings

---

### `tokman config`

Show or create configuration file.

```bash
tokman config                  # Show current config
tokman config --create         # Create default config
tokman config --edit           # Open in editor
```

---

### `tokman verify`

Verify hook integrity using SHA-256 hash.

```bash
tokman verify
# Output:
# ✓ Hook integrity verified
# SHA-256: a1b2c3d4...
```

---

### `tokman economics`

Show spending vs savings analysis.

```bash
tokman economics
# Output:
# 💰 Economics Analysis
# Estimated spent: $12.34
# Estimated saved:  $45.67
# Net benefit:      +$33.33
```

---

## Git Commands

All git commands support the same syntax as native git, with filtered output.

### `tokman git status`

Porcelain parsing with emoji formatting.

```bash
tokman git status
# Output:
# 📝 Modified: 3 | ➕ Staged: 2 | ❓ Untracked: 5
#   M src/main.go
#   M internal/filter/engine.go
#   A cmd/new.go
```

**Token savings:** 70-85%

---

### `tokman git diff`

Stats summary with compact hunks.

```bash
tokman git diff                # Working directory
tokman git diff HEAD~1         # Against commit
tokman git diff --cached       # Staged changes
tokman git diff main...feature # Branch comparison
```

**Output format:**
```
📊 Changes: +45 -12 in 3 files
File: src/main.go
  +15 -3
  @@ func main() {
  -   log.Println("old")
  +   log.Println("new")
```

**Token savings:** 60-80%

---

### `tokman git log`

Compact oneline format with smart limits.

```bash
tokman git log                 # Last 10 commits
tokman git log -20             # Last 20 commits
tokman git log --oneline       # Ultra compact
tokman git log --graph         # ASCII graph
tokman git log --author "Ada"  # Filter by author
```

**Output format:**
```
a1b2c3d Add token counting feature
d4e5f6 Fix filter edge case
g7h8i9 Update documentation
... +7 more commits
```

**Token savings:** 85-95%

---

### `tokman git add`

Compact confirmation output.

```bash
tokman git add .
tokman git add src/main.go
tokman git add -A
# Output: ✓ Staged 5 files
```

---

### `tokman git commit`

Show hash on success.

```bash
tokman git commit -m "message"
# Output: ✓ a1b2c3d message
```

---

### `tokman git push`

Show branch on success.

```bash
tokman git push origin main
# Output: ✓ Pushed to origin/main
```

---

### `tokman git pull`

Show stats summary.

```bash
tokman git pull origin main
# Output: ✓ Pulled 3 commits (+45 -12)
```

---

### Other Git Commands

| Command | Description | Savings |
|---------|-------------|---------|
| `git branch` | Compact listing | 50-70% |
| `git stash` | Compact list/apply/drop | 60-80% |
| `git show` | Commit summary + compact diff | 70-85% |
| `git fetch` | Show new refs count | 80-90% |
| `git worktree` | Compact listing | 70-85% |
| `git rebase` | Compact progress | 60-75% |
| `git cherry-pick` | Compact result | 60-75% |

---

## Infrastructure Commands

### `tokman docker`

Docker CLI with filtered output.

```bash
tokman docker ps               # Compact container list
tokman docker images           # Compact image list
tokman docker logs container   # Deduplicated logs
tokman docker build -t app .   # Build with compact output
```

**Token savings:** 60-80%

---

### `tokman kubectl`

Kubernetes CLI with filtered output.

```bash
tokman kubectl get pods        # Compact pod list
tokman kubectl get deployments # Compact deployment list
tokman kubectl logs pod-name   # Deduplicated logs
tokman kubectl describe pod    # Key info only
```

**Token savings:** 70-85%

---

### `tokman aws`

AWS CLI with filtered output.

```bash
tokman aws s3 ls               # Compact bucket list
tokman aws ec2 describe-instances # Key instance info
tokman aws lambda list-functions # Function summary
```

**Token savings:** 60-80%

---

### `tokman gh`

GitHub CLI with token-optimized output.

```bash
tokman gh run list             # Workflow runs (compact)
tokman gh release list         # Releases (compact)
tokman gh api repos/:owner/:repo # JSON structure output
tokman gh pr list              # PRs (compact)
tokman gh issue list           # Issues (compact)
```

**Token savings:** 75-90%

---

### `tokman gt`

Graphite stacked PR commands.

```bash
tokman gt stack                # Stack summary
tokman gt submit               # Submit with compact output
```

**Token savings:** 70-85%

---

## Build & Test Commands

### Go Commands

```bash
tokman go test ./...           # Aggregated results
tokman go build ./...          # Error-only output
tokman go vet ./...            # Compact vet output
tokman go run main.go          # Compact run output
```

**`go test` output:**
```
🧪 Go Test Results
✓ pkg/filter (12 tests)
✓ pkg/tracking (8 tests)
✗ pkg/commands (1 failed)
  - TestSmartCommand: expected X, got Y
Total: 20 passed, 1 failed
```

**Token savings:** 80-95%

---

### Rust/Cargo Commands

```bash
tokman cargo test              # Compact test output
tokman cargo build             # Error-only output
tokman cargo clippy            # Compact linter output
tokman cargo run               # Compact run output
```

**`cargo test` output:**
```
🧪 Cargo Test Results
✓ test_filter (12 tests)
✓ test_tracking (8 tests)
✗ test_commands (1 failed)
Total: 20 passed, 1 failed
```

**Token savings:** 85-95%

---

### JavaScript/TypeScript Commands

```bash
tokman npm test                # Compact test output (90% reduction)
tokman npm run build           # Error-only output
tokman vitest                  # Compact vitest output
tokman jest                    # Compact jest output (90% reduction)
tokman tsc                     # TypeScript compiler (errors only)
```

**Token savings:** 80-95%

---

### Python Commands

```bash
tokman pytest                  # Compact test output
tokman pytest tests/           # Test specific directory
tokman ruff check .            # Compact linter output
tokman mypy src/               # Compact type checker
```

**`pytest` output:**
```
🧪 Pytest Results
✓ tests/test_filter.py (8 tests)
✓ tests/test_tracking.py (5 tests)
✗ tests/test_commands.py (1 failed)
Total: 13 passed, 1 failed
```

**Token savings:** 85-95%

---

### Build Tools

```bash
tokman next build              # Next.js with route summary
tokman prettier --check .      # Prettier check (compact)
tokman prisma generate         # Prisma generate (compact)
tokman golangci-lint run       # Go linter (compact)
tokman playwright test         # E2E tests (compact)
```

**`next build` output:**
```
🚀 Next.js Build Summary:
   📄 15 static | 8 SSG | 3 SSR pages

Routes:
   ○ / (static)
   ● /blog (SSG)
   λ /api/users (SSR)
   ... +23 more
```

**Token savings:** 80-90%

---

## Package Manager Commands

### npm/pnpm/npx

```bash
tokman npm install             # Compact install
tokman npm list                # Compact list
tokman pnpm install            # Ultra-compact install
tokman pnpm list               # Ultra-compact list
tokman npx create-react-app    # Intelligent routing
```

**Token savings:** 70-85%

---

### pip

```bash
tokman pip install package     # Compact install
tokman pip list                # Compact list
tokman pip outdated            # Show outdated only
tokman pip freeze              # Compact freeze
```

**Token savings:** 60-80%

---

### cargo

```bash
tokman cargo add package       # Compact add
tokman cargo update            # Compact update
tokman cargo tree              # Compact dependency tree
```

**Token savings:** 70-85%

---

## Utility Commands

### File Operations

```bash
tokman ls                      # Hide noise dirs, human sizes
tokman ls -la                  # Detailed but compact
tokman tree                    # Compact tree output
tokman tree -L 2               # Limited depth
tokman find . -name "*.go"     # Find files (compact)
tokman grep -r "pattern"       # Compact grep (groups by file)
tokman diff file1 file2        # Ultra-condensed diff
```

**`ls` output:**
```
📁 src/
  📄 main.go (2.1KB)
  📄 config.go (1.5KB)
  📂 internal/
  📂 cmd/
```

**Token savings:** 70-90%

---

### Data Operations

```bash
tokman json config.json        # Show JSON structure
tokman json --depth 2 file.json # Limited depth
tokman env                     # Show env vars (sensitive masked)
tokman env -f AWS              # Filter to pattern
tokman deps                    # Summarize dependencies
tokman deps /path/to/project   # Specific project
tokman log app.log             # Filter/deduplicate logs
tokman wc file.txt             # Word/line/byte count compact
```

**`json` output:**
```
{
  "name": "tokman",
  "version": "1.0.0",
  "dependencies": { ... 5 keys },
  "scripts": { ... 8 keys }
}
```

**Token savings:** 80-95%

---

### Network Operations

```bash
tokman curl https://api.example.com  # Auto-JSON detection
tokman curl -I https://example.com   # Headers only
tokman wget https://example.com/file # Download with compact output
```

**Token savings:** 70-85%

---

### Analysis Utilities

```bash
tokman summary <command>       # Heuristic summary of long output
tokman count "text"            # Count tokens using tiktoken
tokman count file.go           # Count tokens in file
tokman count --model gpt-4o    # Use specific encoding
tokman count --compare "text"  # Heuristic vs actual
tokman err <command>           # Run command, show only errors/warnings
tokman proxy <command>         # Run without filtering (still tracked)
```

**Token savings:** Varies (summary: 90%+, count: N/A, err: 95%+)

---

## Analysis Commands

### `tokman discover`

Find missed savings in Claude Code history.

```bash
tokman discover               # Scan recent history
tokman discover --days 7      # Last 7 days
tokman discover --all         # Full history
```

**Output:**
```
💡 Discovery Report
Potentially rewritable commands:
  - git status (run 234 times) → tokman git status
  - npm test (run 156 times) → tokman npm test
Estimated additional savings: 12,345 tokens
```

---

### `tokman learn`

Generate CLI correction rules from errors.

```bash
tokman learn                  # Analyze error patterns
tokman learn --apply          # Auto-apply suggestions
```

---

### `tokman hook-audit`

Show hook rewrite metrics.

```bash
tokman hook-audit
# Output:
# Hook Rewrite Metrics
# Total rewrites: 1,234
# Most rewritten: git status (234 times)
# Success rate: 99.2%
```

---

### `tokman smart`

Generate 2-line technical summary using heuristic analysis.

```bash
tokman smart main.go
tokman smart src/index.ts
```

**Output:**
```
Go module (5 fn, 2 struct) - 156 lines
uses: fmt, os | patterns: error handling, tests
```

---

### `tokman rewrite`

Rewrite commands to use TokMan.

```bash
tokman rewrite "git status"   # Output: tokman git status
tokman rewrite list           # List all registered rewrites
```

---

## Output Modes

### Standard Mode (default)

Human-readable output with emoji icons and formatting.

```bash
tokman git status
# 📝 Modified: 3 | ➕ Staged: 2
#   M src/main.go
```

---

### Ultra-Compact Mode (`-u`)

ASCII icons, inline format, maximum token savings.

```bash
tokman -u git status
# M:3 S:2 U:5
# M src/main.go
```

**Token savings:** Additional 15-25% over standard mode

---

### JSON Output (`--format json`)

Structured output for scripting.

```bash
tokman status --format json
# {"commands": 1234, "saved": 89234, "percent": 71}
```

---

### Verbose Mode (`-v`)

Detailed debug information.

```bash
tokman -v git status
# [DEBUG] Running: git status --porcelain
# [DEBUG] Parsing output...
```

---

## Token Savings by Command

| Command Category | Avg Savings | Best Case | Worst Case |
|------------------|-------------|-----------|------------|
| Git log | 90-95% | 98% | 75% |
| Test runners | 85-95% | 98% | 70% |
| Build tools | 80-90% | 95% | 60% |
| Package managers | 70-85% | 90% | 50% |
| Docker/kubectl | 70-85% | 90% | 55% |
| File operations | 70-90% | 95% | 50% |
| Git status/diff | 70-85% | 90% | 55% |

### Factors Affecting Savings

1. **Output size** — Larger outputs have higher savings percentages
2. **Error presence** — Errors are preserved, reducing savings
3. **Verbosity level** — Ultra-compact mode adds 15-25% savings
4. **Custom plugins** — Additional filtering can increase savings

---

## Compound Operators

TokMan supports shell compound operators:

```bash
tokman git status && npm test              # AND (run second if first succeeds)
tokman go build ./... || go vet ./...      # OR (run second if first fails)
tokman echo "a"; echo "b"                  # Sequential
tokman npm test | grep "PASS"              # Pipe
tokman npm start &                         # Background (preserved)
```

---

## Environment Variables

| Variable | Description |
|----------|-------------|
| `TOKMAN_DISABLED=1` | Disable command rewriting |
| `TOKMAN_DATABASE_PATH` | Override database location |
| `XDG_CONFIG_HOME` | Override config directory |
| `XDG_DATA_HOME` | Override data directory |

---

## Shell Integration

### Aliases (auto-generated by `tokman init`)

```bash
ts  → tokman status          # Quick status check
tr  → tokman rewrite         # Rewrite command
```

### Completions

```bash
# Bash
source <(tokman completion bash)

# Zsh
source <(tokman completion zsh)

# Fish
tokman completion fish | source
```

---

## See Also

- [GUIDE.md](GUIDE.md) — Getting started guide
- [API.md](API.md) — Plugin development and dashboard API
- [TROUBLESHOOTING.md](TROUBLESHOOTING.md) — Common issues and fixes
