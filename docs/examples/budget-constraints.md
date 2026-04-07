# Budget Constraints Examples

## Setting Token Budgets

TokMan can enforce strict token budgets on output:

### CLI Flag

```bash
# Limit output to 500 tokens
tokman --budget 500 git diff

# Limit output to 2000 tokens
tokman --budget 2000 cat large_file.py
```

### Environment Variable

```bash
# Set default budget
export TOKMAN_BUDGET=1000

# All commands now respect 1000 token limit
tokman git log -n 50
tokman git diff HEAD~10
```

### Config File

```toml
# ~/.config/tokman/config.toml
[pipeline]
max_context_tokens = 2000

[filter]
mode = "minimal"
```

## Budget Strategy

### How Budget Enforcement Works

1. **Pre-processing:** TokMan estimates the uncompressed token count
2. **Pipeline:** Filters reduce output progressively through 31 layers
3. **Budget check:** After each layer, checks if budget is met
4. **Early exit:** Pipeline stops when output fits within budget
5. **Hard cap:** If still over budget, truncation is applied

### Example: Large Git Diff

```bash
# Without budget - might produce 10,000+ tokens
tokman git diff HEAD~20

# With 500 token budget - only most relevant changes
tokman --budget 500 git diff HEAD~20

# Output:
# 15 files changed, +342 -128
# Key changes:
#   internal/filter/pipeline.go: +45 -12 (pipeline optimization)
#   internal/core/runner.go: +23 -8 (timeout handling)
#   README.md: +120 -45 (documentation update)
# [12 files with minor changes omitted]
```

### Per-Command Budgets

```bash
# Tight budget for status checks
tokman --budget 100 git status

# Generous budget for diffs
tokman --budget 5000 git diff

# Minimal budget for test results
tokman --budget 50 go test ./...
```

## Cost-Aware Budgets

```bash
# View cost analysis
tokman cost --model claude-sonnet

# Output:
# Current session: 45,000 tokens used
# Cost so far: $0.34
# Projected (1 hour): $2.04
# With TokMan: $0.41 (80% savings)

# Set budget based on cost
tokman --budget 2000 --query "find the bug" git diff
```
