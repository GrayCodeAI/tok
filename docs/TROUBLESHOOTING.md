# TokMan Troubleshooting Guide

**Solutions to common issues and error messages**

---

## Table of Contents

1. [Installation Issues](#installation-issues)
2. [Shell Integration Issues](#shell-integration-issues)
3. [Command Not Found Errors](#command-not-found-errors)
4. [Output Issues](#output-issues)
5. [Performance Issues](#performance-issues)
6. [Database Issues](#database-issues)
7. [Hook Integrity Issues](#hook-integrity-issues)
8. [Plugin Issues](#plugin-issues)
9. [Environment Variables](#environment-variables)
10. [Debug Mode](#debug-mode)

---

## Installation Issues

### `command not found: tokman`

**Cause**: TokMan is not in your PATH.

**Solutions**:

1. Verify installation:
   ```bash
   which tokman
   # Should output: /path/to/tokman
   ```

2. If missing, reinstall:
   ```bash
   go install github.com/GrayCodeAI/tokman/cmd/tokman@latest
   ```

3. Ensure Go bin is in PATH:
   ```bash
   echo $PATH | grep -q "$(go env GOPATH)/bin" || echo 'export PATH="$PATH:$(go env GOPATH)/bin"' >> ~/.bashrc
   source ~/.bashrc
   ```

---

### `permission denied` errors

**Cause**: Insufficient permissions.

**Solutions**:

1. For system-wide install:
   ```bash
   sudo mv tokman /usr/local/bin/
   ```

2. For user install:
   ```bash
   mkdir -p ~/bin
   mv tokman ~/bin/
   echo 'export PATH="$HOME/bin:$PATH"' >> ~/.bashrc
   source ~/.bashrc
   ```

---

### Build errors from source

**Cause**: Missing dependencies or Go version.

**Solutions**:

1. Ensure Go 1.21+:
   ```bash
   go version
   ```

2. Update dependencies:
   ```bash
   cd tokman
   go mod tidy
   go mod download
   ```

3. Clean build:
   ```bash
   go clean -cache
   go build -o tokman ./cmd/tokman
   ```

---

## Shell Integration Issues

### Commands not being rewritten automatically

**Cause**: Shell hook not installed or not loaded.

**Solutions**:

1. Verify hook exists:
   ```bash
   ls -la ~/.claude/hooks/tokman-rewrite.sh
   ```

2. Check shell config:
   ```bash
   grep -q "tokman" ~/.bashrc || echo "Hook not in config"
   ```

3. Reinstall hook:
   ```bash
   tokman init
   source ~/.bashrc  # or ~/.zshrc
   ```

4. Manual verification:
   ```bash
   type git
   # Should show: git is a function
   ```

---

### `tokman init` fails

**Cause**: Permission issues or missing directories.

**Solutions**:

1. Create required directories:
   ```bash
   mkdir -p ~/.claude/hooks
   mkdir -p ~/.config/tokman
   mkdir -p ~/.local/share/tokman
   ```

2. Check permissions:
   ```bash
   ls -la ~/.claude/hooks/
   chmod 755 ~/.claude/hooks/
   ```

3. Run with verbose output:
   ```bash
   tokman -v init
   ```

---

### Hook conflicts with other tools

**Cause**: Other tools modifying the same shell functions.

**Solutions**:

1. Check for conflicts:
   ```bash
   type git | head -5
   ```

2. Ensure TokMan loads last:
   ```bash
   # Move TokMan init to end of ~/.bashrc
   # Other tool inits should come before
   ```

3. Use explicit tokman prefix:
   ```bash
   tokman git status  # Always works
   ```

---

## Command Not Found Errors

### `tokman: unknown command "xyz"`

**Cause**: Command not yet supported by TokMan.

**Solutions**:

1. Use proxy mode (still tracked, no filtering):
   ```bash
   tokman proxy xyz args...
   ```

2. Check supported commands:
   ```bash
   tokman rewrite list
   ```

3. Request support on GitHub Issues.

---

### `original command not found`

**Cause**: The underlying command isn't installed.

**Solutions**:

1. Verify original command exists:
   ```bash
   which docker
   which kubectl
   ```

2. Install missing tools:
   ```bash
   # Example: Docker
   brew install docker

   # Example: kubectl
   brew install kubectl
   ```

---

## Output Issues

### Output is too verbose

**Cause**: Standard mode may not be aggressive enough.

**Solutions**:

1. Use ultra-compact mode:
   ```bash
   tokman -u git status
   ```

2. Create custom plugin:
   ```bash
   tokman plugin create my-filter
   # Edit ~/.config/tokman/plugins/my-filter.json
   ```

3. Use aggressive config:
   ```toml
   # ~/.config/tokman/config.toml
   [filter]
   mode = "aggressive"
   ```

---

### Important output is being filtered

**Cause**: Aggressive filtering removing useful info.

**Solutions**:

1. Use verbose mode:
   ```bash
   tokman -v git status
   ```

2. Check tee output for failed commands:
   ```bash
   ls ~/.local/share/tokman/tee/
   cat ~/.local/share/tokman/tee/<latest_file>
   ```

3. Use proxy mode to see full output:
   ```bash
   tokman proxy git status
   ```

4. Adjust filter settings:
   ```toml
   # ~/.config/tokman/config.toml
   [filter]
   mode = "minimal"
   ```

---

### ANSI color codes showing in output

**Cause**: Terminal doesn't support ANSI or filtering is off.

**Solutions**:

1. Enable ANSI stripping:
   ```toml
   # ~/.config/tokman/config.toml
   [filter]
   strip_ansi = true
   ```

2. Force color off in original command:
   ```bash
   tokman git -c color.ui=never status
   ```

---

## Performance Issues

### Slow command execution

**Cause**: Large output being processed.

**Solutions**:

1. Use ultra-compact mode for faster processing:
   ```bash
   tokman -u command
   ```

2. Limit output size:
   ```toml
   # ~/.config/tokman/config.toml
   [filter]
   max_output_lines = 500
   ```

3. Use streaming mode (if available):
   ```bash
   tokman command --stream
   ```

---

### High memory usage

**Cause**: Very large outputs being buffered.

**Solutions**:

1. Process in chunks:
   ```bash
   command | head -1000 | tokman proxy cat
   ```

2. Use native command with less:
   ```bash
   TOKMAN_DISABLED=1 command | less
   ```

3. Profile memory:
   ```bash
   tokman -v command 2>&1 | grep -i memory
   ```

---

## Database Issues

### `database is locked` error

**Cause**: Multiple TokMan processes accessing the same DB.

**Solutions**:

1. Check for running processes:
   ```bash
   ps aux | grep tokman
   ```

2. Kill zombie processes:
   ```bash
   pkill -9 tokman
   ```

3. Move database:
   ```bash
   mv ~/.local/share/tokman/tokman.db ~/.local/share/tokman/tokman.db.bak
   # Database will be recreated on next run
   ```

---

### Database corruption

**Symptoms**: Strange errors, missing data, crashes.

**Solutions**:

1. Backup and recreate:
   ```bash
   cp ~/.local/share/tokman/tokman.db ~/.local/share/tokman/tokman.db.backup
   rm ~/.local/share/tokman/tokman.db
   # Database will be recreated
   ```

2. Check integrity:
   ```bash
   sqlite3 ~/.local/share/tokman/tokman.db "PRAGMA integrity_check;"
   ```

3. Repair if possible:
   ```bash
   sqlite3 ~/.local/share/tokman/tokman.db ".recover" > recover.sql
   sqlite3 ~/.local/share/tokman/tokman_new.db < recover.sql
   mv ~/.local/share/tokman/tokman_new.db ~/.local/share/tokman/tokman.db
   ```

---

### `no such table` error

**Cause**: Database schema not initialized.

**Solutions**:

1. Delete and let TokMan recreate:
   ```bash
   rm ~/.local/share/tokman/tokman.db
   tokman status  # Triggers recreation
   ```

2. Manual schema creation:
   ```bash
   sqlite3 ~/.local/share/tokman/tokman.db < schemas/commands.sql
   ```

---

## Hook Integrity Issues

### `hook integrity check failed`

**Cause**: Hook file was modified or corrupted.

**Solutions**:

1. Reinstall hook:
   ```bash
   tokman init --force
   ```

2. Verify hash:
   ```bash
   tokman verify
   cat ~/.claude/hooks/tokman-rewrite.sh.sha256
   ```

3. Check for tampering:
   ```bash
   sha256sum ~/.claude/hooks/tokman-rewrite.sh
   # Compare with stored hash
   ```

---

### Hook keeps getting corrupted

**Cause**: System updates or other tools modifying files.

**Solutions**:

1. Make hook read-only:
   ```bash
   chmod 444 ~/.claude/hooks/tokman-rewrite.sh
   ```

2. Use system-level hook (if available):
   ```bash
   sudo tokman init --system
   ```

3. Add to version control:
   ```bash
   # Backup hook to a repo
   cp ~/.claude/hooks/tokman-rewrite.sh ~/dotfiles/
   ```

---

## Plugin Issues

### `plugin load failed`

**Cause**: Invalid JSON or missing file.

**Solutions**:

1. Validate JSON:
   ```bash
   python3 -m json.tool ~/.config/tokman/plugins/myplugin.json
   ```

2. Check plugin directory:
   ```bash
   ls -la ~/.config/tokman/plugins/
   ```

3. Use verbose mode:
   ```bash
   tokman -v plugin list
   ```

---

### Plugin not applying to commands

**Cause**: Command mismatch in plugin config.

**Solutions**:

1. Check plugin commands field:
   ```json
   {
     "name": "my-plugin",
     "commands": ["git", "npm"],  // Must match command name
     ...
   }
   ```

2. Verify plugin is enabled:
   ```bash
   tokman plugin list
   tokman plugin enable my-plugin
   ```

---

### Regex pattern errors

**Cause**: Invalid regex in plugin patterns.

**Solutions**:

1. Test regex separately:
   ```bash
   echo "test line" | grep -P "your-pattern"
   ```

2. Use Go regex syntax (not PCRE):
   ```json
   {
     "patterns": [
       {
         "match": "^\\s*$",  // Go syntax: \\s not \s
         "replace": ""
       }
     ]
   }
   ```

---

## Environment Variables

### `TOKMAN_DISABLED` not working

**Cause**: Variable not exported or wrong scope.

**Solutions**:

1. Export the variable:
   ```bash
   export TOKMAN_DISABLED=1
   git status  # Now runs natively
   ```

2. Use inline:
   ```bash
   TOKMAN_DISABLED=1 git status
   ```

---

### Custom database path ignored

**Cause**: Variable not set before initialization.

**Solutions**:

1. Set before running TokMan:
   ```bash
   export TOKMAN_DATABASE_PATH=/custom/path/tokman.db
   tokman status
   ```

2. Use config file:
   ```toml
   # ~/.config/tokman/config.toml
   [tracking]
   database_path = "/custom/path/tokman.db"
   ```

---

## Debug Mode

### Enable verbose logging

```bash
# Single command
tokman -v git status

# More verbose
tokman -vv git status

# Maximum verbosity
tokman -vvv git status
```

### Check logs

```bash
# View log file
cat ~/.local/share/tokman/tokman.log

# Tail logs
tail -f ~/.local/share/tokman/tokman.log

# Search for errors
grep -i error ~/.local/share/tokman/tokman.log
```

### Debug shell integration

```bash
# Show what the hook does
cat ~/.claude/hooks/tokman-rewrite.sh

# Test hook manually
source ~/.claude/hooks/tokman-rewrite.sh
type git
```

### Database inspection

```bash
# Show all tables
sqlite3 ~/.local/share/tokman/tokman.db ".tables"

# Show recent commands
sqlite3 ~/.local/share/tokman/tokman.db "SELECT * FROM commands ORDER BY timestamp DESC LIMIT 10"

# Show savings summary
sqlite3 ~/.local/share/tokman/tokman.db "SELECT command, SUM(saved_tokens) as total FROM commands GROUP BY command ORDER BY total DESC LIMIT 10"
```

---

## Still Having Issues?

1. **Check the logs**:
   ```bash
   tokman -vvv command 2>&1 | tee debug.log
   ```

2. **Verify environment**:
   ```bash
   tokman config
   tokman verify
   which tokman
   echo $PATH
   ```

3. **Clean reinstall**:
   ```bash
   # Backup data
   cp -r ~/.config/tokman ~/tokman-backup/
   cp -r ~/.local/share/tokman ~/tokman-backup/

   # Remove everything
   rm -rf ~/.config/tokman ~/.local/share/tokman ~/.claude/hooks/tokman-rewrite.sh*

   # Reinstall
   go install github.com/GrayCodeAI/tokman/cmd/tokman@latest
   tokman init
   ```

4. **Report the issue**:
   - GitHub Issues: [github.com/GrayCodeAI/tokman/issues](https://github.com/GrayCodeAI/tokman/issues)
   - Include: debug log, `tokman --version`, OS, shell

---

## Quick Reference

| Issue | Quick Fix |
|-------|-----------|
| Command not found | `go install github.com/GrayCodeAI/tokman/cmd/tokman@latest` |
| Hook not working | `tokman init && source ~/.bashrc` |
| Too verbose output | `tokman -u command` |
| Missing output | `tokman proxy command` |
| Database locked | `pkill -9 tokman` |
| Integrity failed | `tokman init --force` |
| Plugin not loading | `tokman plugin list && tokman plugin enable name` |
| Disable temporarily | `TOKMAN_DISABLED=1 command` |
