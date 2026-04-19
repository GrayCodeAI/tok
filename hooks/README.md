# Tok Hooks

Optional shell helpers for users who want tok mode/status surfaced in shell prompts,
and transparent command rewriting for AI agent bash tool calls.

## Files

### Status Line
- `tok-statusline.sh` prints `[TOK]` badge based on active mode file.
- `tok-statusline.ps1` prints status badge for PowerShell.
- `install.sh` adds an optional shell snippet to `~/.zshrc` and `~/.bashrc`.
- `uninstall.sh` removes the snippet.
- `install.ps1` / `uninstall.ps1` do the same for PowerShell profile.

### Transparent Command Rewriting
- `tok-rewrite-hook.sh` - Main hook script that intercepts bash commands from AI
  agent tool calls and rewrites known commands to their tok equivalents.
- `tok-init.sh` - Installation script for setting up the rewrite hook across
  different AI agents.

## Install Status Line

```bash
bash hooks/install.sh
```

```powershell
powershell -ExecutionPolicy Bypass -File hooks\install.ps1
```

## Uninstall Status Line

```bash
bash hooks/uninstall.sh
```

```powershell
powershell -ExecutionPolicy Bypass -File hooks\uninstall.ps1
```

## Transparent Command Rewriting

The tok transparent rewriting hook works similarly to rtk's approach. It intercepts
bash commands from AI agent tool calls and rewrites known commands to their tok
equivalents (e.g., `git status` → `tok git status`). The AI agent never sees the
rewrite - it just gets compressed output.

### How It Works

1. The hook uses bash's `trap '...' DEBUG` mechanism to intercept commands before execution
2. Known commands are rewritten to use tok (e.g., `git status` → `tok git status`)
3. The rewritten command executes and produces compressed output
4. The AI agent receives the compressed output without knowing a rewrite occurred
5. If tok is not available or a command has no tok equivalent, the original command runs

### Scope Notes

- **Only affects bash tool calls** from AI agents
- Does NOT affect built-in agent tools (file operations, search, etc.)
- Only rewrites commands that have tok equivalents
- Falls through to original command if not a tok-wrapped command
- Adds <10ms overhead per command

### Installation

#### Global Installation (All Agents)

```bash
tok init -g
```

This installs the rewrite hook for all detected AI agents.

#### Specific Agent

```bash
tok init --agent claude-code
tok init --agent cursor
tok init --agent windsurf
tok init --agent cline
tok init --agent roo-code
tok init --agent codex
tok init --agent gemini
tok init --agent kilocode
tok init --agent antigravity
tok init --agent copilot
tok init --agent opencode
tok init --agent openclaw
```

#### Manual Installation

```bash
# Source the hook in your shell
source /path/to/tok/hooks/tok-rewrite-hook.sh

# Or install via script
bash /path/to/tok/hooks/tok-init.sh --agent <agent-name>
```

### Agent-Specific Setup

#### Claude Code
The hook is installed in `~/.claude/hooks/tok-rewrite-hook.sh` and referenced in
`~/.claude/settings.json` via the hooks configuration.

#### Cursor
The hook is installed in `~/.cursor/hooks/tok-rewrite-hook.sh` and referenced in
`~/.cursor/hooks.json`.

#### Windsurf
The hook is installed in `~/.windsurf/hooks/tok-rewrite-hook.sh` and added to
`.windsurfrules` in your project.

#### Cline / Roo Code
The hook is installed in `~/.cline/hooks/tok-rewrite-hook.sh` and added to
`.clinerules` in your project.

#### Gemini CLI
The hook is installed in `~/.gemini/hooks/tok-rewrite-hook.sh` and referenced in
`~/.gemini/settings.json`.

#### Codex
The hook is installed in `~/.codex/hooks/tok-rewrite-hook.sh` and referenced in
`AGENTS.md`.

#### GitHub Copilot
The hook config is installed in `.github/hooks/tok-rewrite.json` with instructions
in `.github/copilot-instructions.md`.

#### OpenCode
The hook is installed as a plugin in `~/.config/opencode/plugins/tok.ts`.

#### OpenClaw
The hook is installed in `~/.openclaw/hooks/tok-rewrite-hook.sh`.

#### Kilo Code
The hook config is installed in `.kilocode/rules/tok-rules.md`.

#### Google Antigravity
The hook config is installed in `.agents/rules/antigravity-tok-rules.md`.

### Environment Variables

- `TOK_NO_REWRITE=1` - Disable transparent rewriting entirely
- `TOK_ULTRA_COMPACT=1` - Enable ultra-compact output mode
- `TOK_REWRITE_LOG=/path/to/log` - Log all rewrites to a file for debugging/analytics

### Supported Commands

The hook rewrites the following commands to their tok equivalents:

- **Version Control**: `git`, `svn`, `hg`
- **Package Managers**: `npm`, `yarn`, `pnpm`, `cargo`, `go`, `pip`, `pip3`, `pipenv`, `poetry`, `uv`
- **Container/Orchestration**: `docker`, `docker-compose`, `kubectl`, `helm`
- **Infrastructure**: `terraform`, `ansible`
- **Testing**: `pytest`, `jest`, `mocha`, `vitest`
- **Linting/Formatting**: `ruff`, `black`, `isort`, `flake8`, `pylint`, `eslint`, `prettier`
- **Build Tools**: `tsc`, `webpack`, `vite`, `rollup`, `babel`, `make`, `cmake`, `gradle`, `mvn`
- **Rust**: `cargo`, `rustc`, `rustup`
- **JavaScript/Node**: `node`, `deno`, `bun`
- **Python**: `python`, `python3`
- **Network**: `curl`, `wget`, `ssh`, `scp`, `rsync`
- **System**: `systemctl`, `service`, `journalctl`, `dmesg`
- **And many more...**

### Troubleshooting

#### Hook not working
1. Check if tok is installed: `tok --version`
2. Check if hook is installed: `bash hooks/tok-rewrite-hook.sh --check`
3. Enable logging: `export TOK_REWRITE_LOG=/tmp/tok-rewrite.log`
4. Check the log file for rewrite activity

#### Commands not being rewritten
1. Verify the command has a tok equivalent
2. Check if `TOK_NO_REWRITE` is set to `1`
3. Ensure the hook is sourced in the correct shell environment
4. Check agent-specific configuration files

#### Performance issues
The hook adds <10ms overhead per command. If you experience significant slowdown:
1. Check if there are other DEBUG traps that might conflict
2. Verify tok binary is accessible in PATH
3. Consider disabling rewriting for specific commands

#### Uninstall
```bash
tok init --uninstall
```

Or remove manually:
```bash
rm ~/.claude/hooks/tok-rewrite-hook.sh
rm ~/.cursor/hooks/tok-rewrite-hook.sh
# etc.
```
