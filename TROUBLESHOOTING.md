# 🆘 Troubleshooting Guide

## Installation Issues

### "tokman: command not found"

**Cause:** TokMan binary is not in your PATH.

**Solution:**

```bash
# If you used Homebrew (macOS/Linux)
brew link tokman
echo 'eval "$(/opt/homebrew/bin/brew shellenv)"' >> ~/.zshrc
source ~/.zshrc

# If you used install script
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc

# Verify installation
which tokman
tokman --version
```

### Install script fails

**Cause:** Network issue or missing dependencies.

**Solution:**

```bash
# Check internet connection
curl -I https://github.com

# Check if curl/wget is installed
which curl || which wget

# Manual installation
cd /tmp
curl -fsSL https://github.com/GrayCodeAI/tokman/releases/download/v0.1.0/tokman-darwin-arm64.tar.gz
tar -xzf tokman-darwin-arm64.tar.gz
sudo mv tokman /usr/local/bin/
```

### Homebrew installation fails

```bash
# Update Homebrew
brew update

# Doctor check
brew doctor

# Try again
brew install GrayCodeAI/tokman/tokman
```

---

## Hook Issues

### Hooks not working

**Check installation:**

```bash
# Verify hook files exist
ls -la ~/.config/claudecode/hooks/

# Check permissions
chmod +x ~/.config/claudecode/hooks/*.sh

# Verify dependencies
which jq
which tokman
tokman --version
```

**Restart your AI assistant:**

Close and reopen Claude Code, Cursor, or your AI coding tool after running `tokman init`.

### "tokman rewrite" not found

```bash
# Check tokman version (rewrite requires >= 0.1.0)
tokman --version

# If outdated, upgrade
brew upgrade tokman
# or
curl -fsSL https://raw.githubusercontent.com/GrayCodeAI/tokman/main/install.sh | sh
```

### Commands not being rewritten

**Test manually:**

```bash
# Should output: tokman git status
tokman rewrite "git status"

# Should output nothing (exit code 1)
tokman rewrite "echo hello"

# Should output nothing (exit code 2 - denied)
tokman rewrite "rm -rf /"
```

---

## Filter Issues

### Filter syntax errors

```bash
# Validate all filters
tokman validate

# Validate specific filter
tokman validate path/to/your-filter.toml
```

### Filter tests failing

```bash
# Run all tests
tokman tests

# Run tests for specific filter (verbose)
tokman tests git_status -v

# Common causes:
# 1. Expected output doesn't match actual
# 2. Filter configuration changed
# 3. New edge case in input
```

### Custom filter not working

```bash
# Check filter is in correct location
ls -la ~/.config/tokman/filters/

# Validate filter syntax
tokman validate ~/.config/tokman/filters/my-filter.toml

# Test filter manually
echo "sample input" | tokman filter-test my-filter --command "my-command"

# Check logs (verbose mode)
tokman -vv git status
```

---

## Performance Issues

### Slow filtering

```bash
# Check pipeline performance
tokman profile

# Identify slow layers
tokman --profile=verbose git status

# Try fast preset
tokman --preset=fast git status

# Disable specific layers
tokman --disable-layer=perplexity,h2o git status
```

### High memory usage

```bash
# Check memory profile
tokman profile

# Use streaming for large files
tokman --stream large-file.txt

# Reduce budget
tokman --budget 1000 git log
```

---

## Output Issues

### No output shown

```bash
# Check if filtering too aggressively
tokman --mode=none git status

# Increase budget
tokman --budget 5000 git status

# Use full output mode
tokman --output-mode=full git status
```

### Output looks wrong

```bash
# See what was removed
tokman --verbose git status

# Compare original vs filtered
tokman diff git status
```

---

## Quality Issues

### Low quality score

```bash
# Analyze current quality
tokman quality file.txt

# See detailed breakdown
tokman quality --verbose file.txt

# Compare modes
tokman compare file.txt

# Try different preset
tokman --preset=balanced file.txt
```

---

## Configuration Issues

### Reset configuration

```bash
# Backup current config
cp ~/.config/tokman/config.toml ~/.config/tokman/config.toml.bak

# Generate fresh config
tokman config --generate

# Or remove and reinitialize
rm ~/.config/tokman/config.toml
tokman init
```

---

## Common Error Messages

| Error | Cause | Solution |
|-------|-------|----------|
| `no matching filter` | Command not recognized | Add TOML filter or use `--mode=none` |
| `budget exceeded` | Output too large for budget | Increase `--budget` or use aggressive mode |
| `hook version mismatch` | Outdated tokman | Run `brew upgrade tokman` |
| `invalid TOML filter` | Syntax error in filter | Run `tokman validate` |
| `tracking database locked` | Another process using DB | Wait or force unlock: `rm ~/.local/share/tokman/*.db.lock` |

---

## Getting Help

If your issue isn't listed here:

1. **Check our documentation**
   ```bash
   tokman --help
   tokman <command> --help
   ```

2. **Search existing issues**
   https://github.com/GrayCodeAI/tokman/issues

3. **Enable verbose logging**
   ```bash
   tokman -vvv git status 2>&1 | tee /tmp/tokman-debug.log
   ```

4. **Open an issue**
   https://github.com/GrayCodeAI/tokman/issues/new

Include:
   - Tokman version: `tokman --version`
   - OS and architecture: `uname -a`
   - Command that failed
   - Debug log (if applicable)

---

<div align="center">

**Found a bug?** [Open an issue](https://github.com/GrayCodeAI/tokman/issues/new)

**Have a question?** [Start a discussion](https://github.com/GrayCodeAI/tokman/discussions)

</div>