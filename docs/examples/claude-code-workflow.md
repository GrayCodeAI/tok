# Claude Code Workflow

## Setup

```bash
# Install TokMan hooks for Claude Code
tokman init -g

# Verify
tokman doctor
```

## How It Works

After setup, Claude Code's Bash tool calls are transparently intercepted:

```
Claude Code runs:  git status
Hook rewrites to:  tokman git status
Claude receives:   Compressed output (85% smaller)
```

Claude never sees the rewrite — it just receives less noise.

## Typical Session

### Before TokMan

In a typical 30-minute Claude Code session:

```
Operation          | Times | Tokens
-------------------|-------|--------
git status         | 10x   | 3,000
git diff           | 5x    | 10,000
git log            | 5x    | 2,500
cat/read files     | 20x   | 40,000
go test            | 5x    | 25,000
grep/search        | 8x    | 16,000
ls/tree            | 10x   | 2,000
Other              | 20x   | 20,000
TOTAL              |       | ~118,000
```

### After TokMan

```
Operation          | Times | Tokens | Savings
-------------------|-------|--------|--------
git status         | 10x   | 600    | -80%
git diff           | 5x    | 2,500  | -75%
git log            | 5x    | 500    | -80%
cat/read files     | 20x   | 12,000 | -70%
go test            | 5x    | 2,500  | -90%
grep/search        | 8x    | 3,200  | -80%
ls/tree            | 10x   | 400    | -80%
Other              | 20x   | 4,000  | -80%
TOTAL              |       | ~25,700| -78%
```

**Total savings: ~92,000 tokens (78%) per 30-minute session**

## Best Practices for Claude Code

### 1. Use Shell Commands for Filtered Output

```bash
# These go through TokMan hooks:
git status          # Filtered
cat file.py         # Filtered
go test ./...       # Filtered

# These bypass TokMan (Claude built-in tools):
# Read tool          → Not filtered
# Grep tool          → Not filtered
# Glob tool          → Not filtered
```

**Tip:** When TokMan is active, prefer shell commands for best savings.

### 2. Query Intent for Better Filtering

```bash
# Set query intent for smarter filtering
export TOKMAN_QUERY="find the authentication bug"

# Now TokMan prioritizes auth-related content
tokman git diff  # Highlights auth changes
tokman grep "auth" .  # Focused results
```

### 3. Monitor Savings

```bash
# Check how much you've saved
tokman gain

# Detailed stats
tokman stats

# Cost analysis
tokman cost --model claude-sonnet
```

### 4. Adjust Settings Mid-Session

```bash
# Switch to aggressive for large operations
export TOKMAN_MODE=aggressive
git log -n 100  # Compressed heavily

# Switch back to minimal for debugging
export TOKMAN_MODE=minimal
git diff  # More detail preserved
```

## Uninstalling

```bash
# Remove Claude Code hooks
tokman init --uninstall

# Verify hooks removed
tokman doctor
```
