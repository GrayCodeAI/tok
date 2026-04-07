# Aggressive Mode Examples

## When to Use Aggressive Mode

Aggressive mode enables all 31 compression layers and maximizes token reduction. Use it when:

- Working with very large outputs (>10K tokens)
- On a tight token budget
- Processing verbose build/test output
- You need maximum savings and can tolerate some information loss

## Usage

```bash
# Via flag
tokman --mode aggressive git diff HEAD~50

# Via environment variable
export TOKMAN_MODE=aggressive
tokman git log -n 100

# Via config
# ~/.config/tokman/config.toml
# [filter]
# mode = "aggressive"
```

## Comparison: Minimal vs Aggressive

### Git Diff (large change)

```bash
# Minimal mode - 75% reduction
$ tokman --mode minimal git diff
15 files changed
internal/filter/pipeline.go:
  + func (p *Pipeline) ProcessStream(ctx context.Context, ...) {
  -   // old implementation
  ... (42 lines)

# Aggressive mode - 92% reduction
$ tokman --mode aggressive git diff
15 files: pipeline.go(+45/-12), runner.go(+23/-8), 13 minor
Key: Added ProcessStream, refactored timeout handling
```

### Build Output

```bash
# Minimal mode - keeps structure
$ tokman --mode minimal cargo build
Compiling tokman v0.28.2
  23 warnings (12 unused imports, 11 dead code)
  Build succeeded in 45.2s

# Aggressive mode - just the essentials
$ tokman --mode aggressive cargo build
ok 45.2s, 23 warnings
```

### Test Output

```bash
# Minimal mode - shows summary
$ tokman --mode minimal go test ./...
144 packages, 891 tests
2 FAILED:
  internal/filter: TestEntropy (expected 0.85, got 0.82)
  internal/core: TestTimeout (deadline exceeded)
142 passed

# Aggressive mode - failures only
$ tokman --mode aggressive go test ./...
FAIL: TestEntropy(0.85!=0.82), TestTimeout(deadline)
```

## Combining with Budget

```bash
# Maximum compression within strict budget
tokman --mode aggressive --budget 200 git diff HEAD~100

# Result: Even a 100-commit diff fits in 200 tokens
# 100 commits, 342 files, +15420 -8931
# Top: api/handler.go(+892), pkg/auth/oauth.go(+456)
```

## Quality Impact

Aggressive mode trades some quality for compression:

```bash
# Check quality in aggressive mode
tokman --mode aggressive --quality git diff

# Typical grades:
# Minimal mode:    A  (0.92 overall)
# Aggressive mode: B+ (0.84 overall)
#   Semantic: 0.88  (some detail lost)
#   Signal: 0.91    (noise well removed)
#   Complete: 0.75  (some content omitted)
```

## Best Practices

1. **Start with minimal** - Switch to aggressive only when needed
2. **Check quality** - Use `--quality` to verify important info preserved
3. **Use with budget** - Combine `--mode aggressive --budget N`
4. **Avoid for debugging** - Use minimal mode when debugging issues
5. **Great for CI/CD** - Aggressive works well for automated pipelines
