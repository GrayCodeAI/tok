# TokMan Example Use Cases

**Version:** 1.0.0  
**Last Updated:** 2026-03-28

## Overview

This document provides practical examples of using TokMan in various development scenarios.

---

## Basic Usage

### 1. Git Status Compression

Reduce git status output for AI context:

```bash
# Without TokMan
git status
# ~500 tokens of verbose output

# With TokMan
tokman git status
# ~150 tokens, essential info preserved
```

**Output Comparison:**
```
# Original: 47 lines, 1,247 chars
On branch main
Changes not staged for commit:
  modified:   src/app.go
  modified:   src/handler.go
...

# Compressed: 12 lines, 312 chars
branch: main
modified: src/app.go, src/handler.go
untracked: test/new_test.go
```

### 2. Docker Logs Filtering

```bash
# Compress verbose Docker logs
tokman docker logs my-container --tail 1000

# Filter to errors only
tokman docker logs my-container --tail 1000 --filter error
```

### 3. NPM Install Output

```bash
# Compress npm install output
tokman npm install

# Original: 200+ lines of dependency tree
# Compressed: Summary of added/updated packages
```

---

## AI Agent Integration

### 4. Claude Code Integration

```bash
# Install hooks for Claude Code
tokman init -g

# Now all CLI commands in Claude Code are compressed
# git diff → 60% smaller context
```

### 5. Cursor Integration

```bash
# Setup for Cursor IDE
tokman init --cursor

# Compress terminal output in Cursor's AI chat
```

### 6. Multiple Agent Setup

```bash
# Install for all detected AI agents
tokman init --all

# Shows which agents were configured:
# ✓ Claude Code configured
# ✓ Cursor configured  
# ✓ Copilot configured
```

---

## Pipeline Configuration

### 7. Fast Mode for Quick Operations

```bash
# Use fast preset for minimal latency
tokman --preset fast git log --oneline -20

# Good for: quick checks, frequent commands
```

### 8. Full Mode for Large Outputs

```bash
# Use full preset for maximum compression
tokman --preset full kubectl logs deployment/app --tail 5000

# Good for: large logs, verbose commands
```

### 9. Budget Mode

```bash
# Limit output to specific token budget
tokman --budget 500 cat large-file.log

# Ensures output fits within token limit
```

### 10. Query-Aware Compression

```bash
# Focus compression on relevant content
tokman --query "find authentication errors" docker logs auth-service

# Highlights lines matching query intent
```

---

## Advanced Scenarios

### 11. CI/CD Pipeline Integration

```yaml
# .github/workflows/ci.yml
steps:
  - name: Run tests with compressed output
    run: |
      tokman npm test
      tokman npm run lint
```

### 12. Git Diff for Code Review

```bash
# Compress large diffs while preserving structure
tokman git diff main...feature-branch

# Preserves:
# - Function signatures
# - Changed lines
# - File structure
# Removes:
# - Unchanged lines
# - Verbose context
```

### 13. Log Analysis

```bash
# Analyze production logs efficiently
tokman --mode aggressive cat /var/log/app.log | grep ERROR

# Combine with grep for powerful filtering
```

### 14. Database Query Output

```bash
# Compress SQL query results
tokman psql -c "SELECT * FROM users LIMIT 100"

# Preserves column headers, compresses data rows
```

### 15. Kubernetes Diagnostics

```bash
# Get pod status across namespace
tokman kubectl get pods -n production

# Describe specific pod
tokman kubectl describe pod/api-server-123
```

---

## Session Management

### 16. Persistent Sessions

```bash
# Start a session for continued work
tokman session start --name "feature-review"

# All subsequent commands are tracked
tokman git log -10
tokman git diff HEAD~5

# View session stats
tokman session stats
# Output: 15 commands, 12,500 tokens saved

# End session
tokman session end
```

### 17. Session Restore

```bash
# Restore previous session context
tokman session restore feature-review

# Continue from where you left off
```

---

## Custom Filters

### 18. Creating Custom TOML Filter

```bash
# Create filter for custom tool
cat > ~/.config/tokman/filters/mytool.toml << 'EOF'
[mytool]
match = "^my-tool (build|test|deploy)"
output_patterns = ["^Building...", "^Testing..."]
strip_lines_matching = ["^DEBUG:", "^TRACE:"]
max_lines = 100
EOF

# Now 'my-tool' commands are compressed
tokman my-tool build
```

### 19. Filter Priority

```bash
# Higher priority filters take precedence
cat > ~/.config/tokman/filters/priority.toml << 'EOF'
[critical]
match = "^critical-.*"
priority = 100
strip_lines_matching = ["^INFO:"]
EOF
```

---

## Performance Tuning

### 20. Profile Pipeline Performance

```bash
# Enable profiling for analysis
tokman --profile git log -100

# View profile results
tokman profile view
```

### 21. Cache Management

```bash
# Clear compression cache
tokman cache clear

# View cache statistics
tokman cache stats
# Output: 1,234 entries, 45MB saved
```

### 22. Warmup Pipeline

```bash
# Pre-initialize pipeline for faster first run
tokman warmup

# Useful before time-sensitive operations
```

---

## API Usage

### 23. REST API Compression

```bash
# Start API server
tokman server start --port 8080

# Compress via API
curl -X POST http://localhost:8080/api/v1/compress \
  -H "Content-Type: application/json" \
  -d '{"content": "large text to compress", "preset": "balanced"}'
```

### 24. Streaming Compression

```bash
# Stream large files
cat large-log-file.log | tokman stream

# Or via API
curl -X POST http://localhost:8080/api/v1/compress/stream \
  -H "Content-Type: application/json" \
  -d '{"content": "very large content...", "preset": "full"}'
```

---

## Monitoring & Analytics

### 25. Usage Statistics

```bash
# View overall statistics
tokman stats

# Output:
# Total commands: 1,523
# Tokens saved: 847,321
# Compression ratio: 67%
```

### 26. Top Commands

```bash
# See most frequently compressed commands
tokman top

# Output:
# 1. git status (234 times)
# 2. docker logs (156 times)
# 3. npm test (98 times)
```

### 27. Cost Analysis

```bash
# Calculate token savings
tokman cost --model gpt-4

# Output:
# Tokens saved: 847,321
# Estimated cost saved: $12.71
```

---

## Troubleshooting Examples

### 28. Debug Mode

```bash
# Enable verbose logging
tokman --verbose git status

# See which layers are applied
# Layer 1 (Entropy): skipped (low content)
# Layer 2 (Perplexity): applied, saved 23 tokens
# ...
```

### 29. Dry Run

```bash
# Preview compression without applying
tokman --dry-run cat file.txt

# Shows original vs compressed token counts
```

### 30. Layer Inspection

```bash
# See which layers would apply
tokman --layers git diff

# Output:
# Applicable layers: entropy, ast_preserve, budget
# Skipped layers: perplexity (content too small)
```

---

## Integration Patterns

### 31. Shell Alias Setup

```bash
# Add to ~/.bashrc or ~/.zshrc
alias g='tokman git'
alias d='tokman docker'
alias n='tokman npm'

# Now use: g status, d logs, n install
```

### 32. Pre-commit Hook

```bash
# .git/hooks/pre-commit
#!/bin/bash
tokman git diff --cached --stat
```

### 33. Makefile Integration

```makefile
test:
	tokman go test ./... -v

build:
	tokman go build -o bin/app ./cmd/app
```

---

## Real-World Scenarios

### 34. Code Review Workflow

```bash
# Generate compressed diff for review
tokman git diff main...feature > review-context.txt

# AI assistant can now analyze 70% less tokens
```

### 35. Debugging Production Issue

```bash
# Get compressed pod logs
tokman kubectl logs deployment/api --since=1h --filter error

# Compress events
tokman kubectl get events --sort-by='.lastTimestamp'
```

### 36. Daily Standup Prep

```bash
# Get yesterday's work summary
tokman git log --since="yesterday" --oneline --author="me"

# Compressed format perfect for status updates
```

---

## Best Practices Summary

| Scenario | Command | Tokens Saved |
|----------|---------|--------------|
| Git status | `tokman git status` | 60% |
| Git diff | `tokman git diff` | 70% |
| Docker logs | `tokman docker logs --tail 1000` | 65% |
| NPM output | `tokman npm install` | 75% |
| Kubectl describe | `tokman kubectl describe pod` | 55% |
| Large files | `tokman cat large.log` | 60-80% |

---

## Getting Help

```bash
# General help
tokman --help

# Command-specific help
tokman git --help

# Doctor for diagnostics
tokman doctor
```
