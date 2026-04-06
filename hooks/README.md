# TokMan Hooks - Delegating Pattern

**Version:** 2.0  
**Pattern:** Delegating hooks (logic in binary, not shell)

---

## 🎯 Overview

TokMan hooks use a **delegating pattern** where all rewrite logic lives in the `tokman rewrite` command instead of in shell scripts. This makes hooks:

- ✅ **Easier to maintain** - Edit Go code, not shell scripts
- ✅ **More reliable** - Single source of truth
- ✅ **Testable** - Unit tests for rewrite logic
- ✅ **Version-safe** - Version guard prevents old binaries

---

## 🏗️ Architecture

```
┌─────────────────┐
│  AI Assistant   │ (Claude Code, Cursor, etc.)
└────────┬────────┘
         │
         │ Bash command
         ▼
┌─────────────────┐
│   Hook Script   │ (thin delegating shell script)
└────────┬────────┘
         │
         │ tokman rewrite "git status"
         ▼
┌─────────────────┐
│ tokman rewrite  │ (Go binary - single source of truth)
└────────┬────────┘
         │
         │ Exit code protocol
         ▼
┌─────────────────┐
│   Hook Script   │ (interprets exit code)
└────────┬────────┘
         │
         │ Updated command
         ▼
┌─────────────────┐
│  AI Assistant   │ (executes rewritten command)
└─────────────────┘
```

---

## 🔢 Exit Code Protocol

The `tokman rewrite` command uses exit codes to communicate with hook scripts:

| Code | Meaning | Action | Output |
|------|---------|--------|--------|
| 0 | Rewrite found, auto-allow | Update command, allow execution | Rewritten command |
| 1 | No tokman equivalent | Pass through unchanged | None |
| 2 | Deny rule matched | Pass through (let AI deny) | None |
| 3 | Ask rule matched | Update command, ask user | Rewritten command |
| 4 | Invalid input | Pass through | None |
| 5 | Command disabled | Pass through | None |
| 6 | Unsafe operation | Pass through | None |
| 7 | Resource-intensive | Pass through | None |

---

## 📝 Hook Script Template

```bash
#!/usr/bin/env bash
# tokman-hook-version: 2.0

# Read input
INPUT=$(cat)
CMD=$(echo "$INPUT" | jq -r '.tool_input.command // empty')

# Delegate to tokman rewrite
REWRITTEN=$(tokman rewrite "$CMD" 2>/dev/null)
EXIT_CODE=$?

case $EXIT_CODE in
  0)  # Auto-allow
    # Update command and approve
    ;;
  1)  # Pass through
    exit 0
    ;;
  2)  # Deny
    exit 0
    ;;
  3)  # Ask user
    # Update command but don't auto-allow
    ;;
  *)  # Unknown
    exit 0
    ;;
esac
```

---

## 🔧 Installation

### Claude Code

```bash
# Install hook
tokman init -g

# Or manually
mkdir -p ~/.config/claudecode/hooks
cp hooks/tokman-delegating-hook.sh ~/.config/claudecode/hooks/
chmod +x ~/.config/claudecode/hooks/tokman-delegating-hook.sh
```

### Cursor

```bash
tokman init --cursor
```

### Windsurf

```bash
tokman init --windsurf
```

### All Detected

```bash
tokman init --all
```

---

## ✅ Verify Installation

```bash
# Check hook is installed
ls -la ~/.config/claudecode/hooks/

# Check tokman rewrite works
tokman rewrite "git status"
# Should output: tokman git status (exit code 0)

tokman rewrite "echo hello"
# Should output nothing (exit code 1)

tokman rewrite "rm -rf /"
# Should output nothing (exit code 2)
```

---

## 🎨 Customization

### Add Custom Rules

Edit `internal/commands/core/rewrite.go`:

```go
func isSupportedCommand(baseCmd string) bool {
    supported := []string{
        "git",
        "docker",
        "my-custom-tool", // Add your tool here
    }
    // ...
}
```

Then rebuild:

```bash
make build
make install
```

### Disable Commands

Add to deny list in `rewrite.go`:

```go
func isDenied(baseCmd string, parts []string) bool {
    denyList := []string{
        "rm",
        "dd",
        "my-dangerous-tool", // Add here
    }
    // ...
}
```

### Add Ask Rules

Add to confirmation list in `rewrite.go`:

```go
func requiresConfirmation(baseCmd string, parts []string) bool {
    askList := []string{
        "sudo",
        "systemctl",
        "my-sensitive-tool", // Add here
    }
    // ...
}
```

---

## 🧪 Testing

### Test Rewrite Logic

```bash
# Run tests
go test ./internal/commands/core -v -run TestRewrite

# Test specific scenarios
tokman rewrite "git status"        # Should rewrite
tokman rewrite "git push --force"  # Should rewrite (git is safe)
tokman rewrite "sudo apt upgrade"  # Should rewrite with ask (exit 3)
tokman rewrite "rm -rf /"          # Should deny (exit 2)
tokman rewrite "unknown-cmd"       # Should pass through (exit 1)
```

### Test Hook Script

```bash
# Test hook script directly
echo '{"tool_input":{"command":"git status"}}' | \
  ~/.config/claudecode/hooks/tokman-delegating-hook.sh

# Should output JSON with updated command
```

---

## 🔍 Debugging

### Enable Verbose Mode

```bash
# Add to hook script after shebang
set -x  # Enable tracing

# Then check logs
tail -f ~/.config/claudecode/logs/hook.log
```

### Check Hook Version

```bash
# Check hook file
head -n 3 ~/.config/claudecode/hooks/tokman-delegating-hook.sh

# Should show: tokman-hook-version: 2.0
```

### Verify Dependencies

```bash
# Check jq
command -v jq && echo "jq installed" || echo "jq missing"

# Check tokman
command -v tokman && tokman --version

# Check version >= 0.1.0
tokman --version | grep -oE '[0-9]+\.[0-9]+\.[0-9]+'
```

---

## 📚 Examples

### Example 1: Git Status (Auto-allow)

```bash
$ tokman rewrite "git status"
tokman git status
$ echo $?
0
```

Hook behavior: Updates command to `tokman git status` and auto-allows.

### Example 2: Unknown Command (Pass through)

```bash
$ tokman rewrite "echo hello"
$ echo $?
1
```

Hook behavior: No output, passes through unchanged.

### Example 3: Dangerous Command (Deny)

```bash
$ tokman rewrite "rm -rf /"
$ echo $?
2
```

Hook behavior: No output, lets AI assistant deny.

### Example 4: Sudo Command (Ask)

```bash
$ tokman rewrite "sudo apt upgrade"
tokman sudo apt upgrade
$ echo $?
3
```

Hook behavior: Updates command but asks user for confirmation.

---

## 🔄 Migration from Old Hooks

If you have old hooks (version 1.0), update them:

```bash
# Backup old hook
cp ~/.config/claudecode/hooks/tokman-hook.sh ~/.config/claudecode/hooks/tokman-hook.sh.old

# Install new hook
cp hooks/tokman-delegating-hook.sh ~/.config/claudecode/hooks/tokman-hook.sh
chmod +x ~/.config/claudecode/hooks/tokman-hook.sh

# Restart AI assistant
```

---

## 🆘 Troubleshooting

### Hook not working

**Check installation:**
```bash
ls -la ~/.config/claudecode/hooks/
```

**Check permissions:**
```bash
chmod +x ~/.config/claudecode/hooks/*.sh
```

**Check dependencies:**
```bash
command -v jq && command -v tokman
```

### Commands not being rewritten

**Check version:**
```bash
tokman --version
# Should be >= 0.1.0
```

**Test rewrite directly:**
```bash
tokman rewrite "git status"
# Should output: tokman git status
```

**Check hook script:**
```bash
cat ~/.config/claudecode/hooks/tokman-delegating-hook.sh | grep "tokman-hook-version"
# Should show: tokman-hook-version: 2.0
```

### Version too old

**Upgrade tokman:**
```bash
# Homebrew
brew upgrade tokman

# Install script
curl -fsSL https://raw.githubusercontent.com/GrayCodeAI/tokman/main/install.sh | sh

# From source
git pull && make build && make install
```

---

## 📖 Learn More

- **Rewrite Logic:** `internal/commands/core/rewrite.go`
- **Hook Script:** `hooks/tokman-delegating-hook.sh`
- **Tests:** `internal/commands/core/rewrite_test.go`
- **Installation:** `docs/INSTALLATION.md`
- **Documentation:** `README.md`

---

<div align="center">

**TokMan Hooks v2.0 - Delegating Pattern**

*Logic in binary, not shell scripts*

</div>
