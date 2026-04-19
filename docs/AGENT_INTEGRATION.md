# AI Coding Agent CLI Integration Guide

## Top 10 AI Coding Agent CLIs (2025)

| Rank | Agent | Platform | Hook Support | Shell Integration | Config Location |
|------|-------|----------|--------------|-------------------|-----------------|
| 1 | **Claude Code** | CLI/Terminal | PreToolUse, PostToolUse, Notification | Child process via shell | `~/.claude/settings.json` |
| 2 | **Cursor** | VSCode/IDE | preToolUse | VSCode Terminal API | `~/.cursor/hooks.json` |
| 3 | **Aider** | CLI/Terminal | Git hooks only | Direct shell exec | `~/.aider.conf.yml` |
| 4 | **Cline** | VSCode Extension | Hooks (v3.36+) | VSCode Terminal API | `~/.vscode/extensions/...` |
| 5 | **OpenCode** | CLI/Terminal (Go) | Custom tools/middleware | Go exec.Command | `~/.config/opencode/config.toml` |
| 6 | **Kiro** | CLI/Terminal | Lifecycle hooks | Shell subprocess | `~/.kilorc` |
| 7 | **Kilo Code** | CLI/Terminal | Beta lifecycle hooks | Shell subprocess | `~/.kilorc` |
| 8 | **AdaL** | CLI/Terminal | MCP tools | Shell exec | `~/.adal/config` |
| 9 | **Continue** | VSCode/JetBrains | Limited | Terminal API | `~/.continue/config.json` |
| 10 | **AutoHand** | CLI/Terminal | Unknown | Shell exec | Project-based |

---

## Tok Integration Architecture

### Hook-Based Integration (Claude Code / Cursor)

Tok uses a thin delegator script plus native `tok hook <agent>` processors for Claude Code and Cursor:

```
┌─────────────────────────────────────────────────────────────┐
│                    Claude Code Flow                          │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  User: "Check git status"                                    │
│            │                                                 │
│            ▼                                                 │
│  Claude decides: bash("git status")                         │
│            │                                                 │
│            ▼                                                 │
│  ┌─────────────────────────────────────────────────────┐    │
│  │ PreToolUse Hook (tok-rewrite.sh)                 │    │
│  │                                                      │    │
│  │  INPUT: {"tool_input": {"command": "git status"}}   │    │
│  │            │                                         │    │
│  │            ▼                                         │    │
│  │  exec tok hook claude                            │    │
│  │            │                                         │    │
│  │            ▼                                         │    │
│  │  OUTPUT: {"updatedInput": {"command": "tok git status"}}│ │
│  └─────────────────────────────────────────────────────┘    │
│            │                                                 │
│            ▼                                                 │
│  Shell executes: tok git status                          │
│            │                                                 │
│            ▼                                                 │
│  Filtered output (~200 tokens instead of ~2000)              │
│            │                                                 │
│            ▼                                                 │
│  Claude receives compressed context                          │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

### Tok Hook Installation (`tok init -g`)

1. **Creates hook file**: `~/.claude/hooks/tok-rewrite.sh`
2. **Patches settings.json**: Adds PreToolUse hook entry
3. **Creates TOK.md**: Instructions for Claude to understand tok commands
4. **Patches CLAUDE.md**: Adds `@TOK.md` reference

### settings.json Structure

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Bash",
        "hooks": [
          {
            "type": "command",
            "command": "~/.claude/hooks/tok-rewrite.sh"
          }
        ]
      }
    ]
  }
}
```

---

## Tok Architecture

| Aspect | Tok (Go) |
|--------|-------------|
| **Binary Size** | ~12MB (dynamic) |
| **Startup Time** | ~15ms |
| **Registry** | 20 core patterns (regex) |
| **Hook Size** | 45 lines |
| **Rewrite Logic** | In binary (registry.go) |
| **Compound Commands** | &&, ||, ;, \|, & |
| **Env Prefixes** | sudo, VAR=val |
| **Exclusion** | config.toml |
| **Status** | TokStatus int |

### Command Coverage

| Category | Tok Commands |
|----------|-----------------|
| Git | git status/diff/log/add/commit/push/pull |
| GitHub | gh pr/issue/run/repo/api/release |
| Cargo | cargo build/test/clippy/check/fmt |
| Files | cat/head/tail, rg/grep, ls, find, tree, diff |
| Build | tsc, eslint, biome, prettier, next |
| Tests | vitest, jest, playwright, pytest |
| Go | go test/build/vet, golangci-lint |
| Python | ruff, pytest, pip, mypy |
| Containers | docker, kubectl |
| Cloud | aws, psql |
| Network | curl, wget |

---

## Integration Strategies by Agent

### 1. Claude Code (Full Support)

**Method**: PreToolUse hook
**Integration**: Automatic with `tok init -g`

```bash
tok init -g
# Restarts Claude Code
# All bash commands auto-rewritten
```

### 2. Cursor (Full Support)

**Method**: Native Cursor `preToolUse` hook
**Integration**: Automatic with `tok init --cursor`, which patches `~/.cursor/hooks.json`

```json
{
  "version": 1,
  "hooks": {
    "preToolUse": [
      {
        "matcher": "Shell",
        "command": "~/.cursor/hooks/tok-rewrite.sh"
      }
    ]
  }
}
```

### 3. Aider (Shell Wrapper)

**Method**: Shell aliases since Aider has no hook system
**Integration**: Source tok wrapper

```bash
# In ~/.bashrc or ~/.zshrc
alias git='tok git'
alias ls='tok ls'
alias cat='tok read'
alias rg='tok grep'
```

### 4. Cline (VSCode Terminal)

**Method**: Workspace rules file
**Integration**: `./.clinerules`

```md
<!-- tok:cline:start -->
# Tok Rules for Cline

Prefer `tok`-prefixed shell commands so large terminal output is reduced before it reaches the model.
<!-- tok:cline:end -->
```

### 5. OpenCode (Custom Tool)

**Method**: Global OpenCode plugin
**Integration**: `~/.config/opencode/plugins/tok.ts`

```ts
export const TokOpenCodePlugin = async ({ $ }) => ({
  "tool.execute.before": async (input, output) => {
    // rewrite shell commands through tok
  },
})
```

### 6. Kiro (Lifecycle Hooks)

**Method**: Kiro's hook system
**Integration**: `~/.kilorc`

```yaml
hooks:
  preToolUse:
    - matcher: "Bash"
      command: "tok rewrite"
```

### 7. Continue (Limited)

**Method**: Terminal environment
**Integration**: Set environment variable

```bash
export TOK_AUTO_REWRITE=1
```

---

## Recommended Integration Priority

| Priority | Agent | Method | Effort | Coverage |
|----------|-------|--------|--------|----------|
| **P0** | Claude Code | PreToolUse hook | Low | 100% |
| **P0** | Cursor | PreToolUse hook | Low | 100% |
| **P1** | Aider | Shell aliases | Medium | 80% |
| **P1** | Cline | Workspace rules | Low | 85% |
| **P1** | OpenCode | Global plugin | Low | 90% |
| **P2** | Kiro | Lifecycle hooks | Medium | 60% |
| **P3** | Continue | Environment | Low | 30% |
| **P3** | Others | Shell wrapper | Medium | Varies |

---

## Tok Integration TODO

1. Add explicit parity for secondary agents that still use generic hook installs instead of native config patching.
2. Track install-state and hook adoption per agent in the future dashboard.
3. Add richer per-agent telemetry rollups from the persisted local event history.
4. Keep agent docs aligned with battle-tested paths instead of placeholder examples.

---

## MCP Context Examples

Tok can also act as a context service instead of only a shell rewriter.

### Start the MCP server

```bash
tok mcp --port 8080
```

### Read one file under a token budget

```bash
curl -X POST http://localhost:8080/read \
  -H "Content-Type: application/json" \
  -d '{
    "path": "internal/server/server.go",
    "mode": "auto",
    "max_tokens": 350,
    "save_snapshot": true
  }'
```

### Request a graph-aware bundle

```bash
curl -X POST http://localhost:8080/bundle \
  -H "Content-Type: application/json" \
  -d '{
    "path": "internal/server/server.go",
    "mode": "graph",
    "related_files": 4,
    "max_tokens": 500
  }'
```

### Recommended agent usage

- Claude Code / Cursor:
  - use shell hooks for normal command rewriting
  - use `POST /read` or `POST /bundle` when the agent needs curated file context
- Codex / OpenCode:
  - keep shell wrapping for command noise reduction
  - use `POST /bundle` for target-file + related-file context delivery
- Any MCP-capable tool:
  - use `POST /read` for single-file refreshes
  - use `POST /bundle` for multi-file graph context

### Direct integration snippets

Claude Code / Cursor style bundle request:

```json
{
  "tool": "tok.read_bundle",
  "server": "http://localhost:8080",
  "method": "POST",
  "path": "/bundle",
  "body": {
    "path": "internal/server/server.go",
    "mode": "graph",
    "related_files": 4,
    "max_tokens": 500,
    "save_snapshot": true
  }
}
```

Codex / OpenCode style single-file refresh:

```json
{
  "tool": "tok.read_file",
  "server": "http://localhost:8080",
  "method": "POST",
  "path": "/read",
  "body": {
    "path": "internal/contextread/read.go",
    "mode": "auto",
    "max_tokens": 320,
    "save_snapshot": true
  }
}
```

Recommended pattern:
- use `/bundle` first for target file + neighbors
- switch to `/read` for focused refreshes
- use `mode=delta` after edits when the agent already saw the file earlier
