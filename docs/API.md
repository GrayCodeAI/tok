# TokMan API Documentation

## Plugin Development Guide

TokMan supports custom filter plugins for extending output processing capabilities.

### Plugin Architecture

Plugins are JSON-based configurations that define custom filtering rules.

```
~/.config/tokman/plugins/
├── my-plugin.json
├── another-plugin.json
└── ...
```

### Plugin Schema

```json
{
  "name": "plugin-name",
  "version": "1.0.0",
  "description": "Plugin description",
  "commands": ["git", "npm"],
  "patterns": [
    {
      "match": "regex-pattern",
      "replace": "replacement-text",
      "flags": ["i", "m"]
    }
  ],
  "hooks": {
    "pre": "pre-filter-script.sh",
    "post": "post-filter-script.sh"
  }
}
```

### Plugin Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | ✅ | Unique plugin identifier |
| `version` | string | ✅ | Semantic version |
| `description` | string | ❌ | Human-readable description |
| `commands` | []string | ✅ | Commands this plugin applies to |
| `patterns` | []Pattern | ❌ | Regex patterns for filtering |
| `hooks` | Hooks | ❌ | Pre/post filter scripts |

### Pattern Object

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `match` | string | ✅ | Go regex pattern |
| `replace` | string | ❌ | Replacement text (empty = remove) |
| `flags` | []string | ❌ | Regex flags: `i` (case-insensitive), `m` (multiline) |

### Example: Custom Git Plugin

```json
{
  "name": "git-compact",
  "version": "1.0.0",
  "description": "Ultra-compact git output",
  "commands": ["git"],
  "patterns": [
    {
      "match": "^\\s*$",
      "replace": ""
    },
    {
      "match": "(?m)^\\s*#.*$",
      "replace": ""
    }
  ]
}
```

### Example: Docker Log Deduplicator

```json
{
  "name": "docker-dedup",
  "version": "1.0.0",
  "description": "Deduplicate docker logs",
  "commands": ["docker", "kubectl"],
  "patterns": [
    {
      "match": "^(.+?)\\n(?:\\1\\n?)+",
      "replace": "$1 [repeated]\n",
      "flags": ["m"]
    }
  ]
}
```

---

## Dashboard API

TokMan includes a built-in web dashboard for visualizing token savings.

### Starting the Dashboard

```bash
tokman dashboard --port 8080 --host 0.0.0.0
```

### REST Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/stats` | GET | Get overall statistics |
| `/api/history` | GET | Get command history |
| `/api/projects` | GET | List all projects |
| `/api/savings` | GET | Get token savings breakdown |
| `/api/health` | GET | Health check |

### Response Examples

#### GET /api/stats

```json
{
  "total_commands": 1542,
  "total_tokens_saved": 89234,
  "total_tokens_original": 156789,
  "savings_percentage": 56.9,
  "top_commands": [
    {"command": "git status", "count": 234, "saved": 12340},
    {"command": "npm test", "count": 156, "saved": 8765}
  ]
}
```

#### GET /api/history

```json
{
  "commands": [
    {
      "id": 1,
      "command": "git status",
      "original_tokens": 156,
      "filtered_tokens": 42,
      "saved_tokens": 114,
      "timestamp": "2026-03-08T10:30:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "per_page": 50,
    "total": 1542
  }
}
```

---

## Tracking Database

TokMan stores command history in a SQLite database.

### Database Location

```
~/.local/share/tokman/tokman.db
```

### Schema

```sql
CREATE TABLE commands (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    command TEXT NOT NULL,
    original_output TEXT,
    filtered_output TEXT,
    original_tokens INTEGER,
    filtered_tokens INTEGER,
    saved_tokens INTEGER,
    project_path TEXT,
    exec_time_ms INTEGER,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    parse_success BOOLEAN
);

CREATE INDEX idx_commands_command ON commands(command);
CREATE INDEX idx_commands_project ON commands(project_path);
CREATE INDEX idx_commands_timestamp ON commands(timestamp);
```

### Querying Directly

```bash
sqlite3 ~/.local/share/tokman/tokman.db "
  SELECT command, SUM(saved_tokens) as total_saved
  FROM commands
  GROUP BY command
  ORDER BY total_saved DESC
  LIMIT 10
"
```

---

## Configuration API

### Config File Location

```
~/.config/tokman/config.toml
```

### Configuration Schema

```toml
[tracking]
enabled = true
database_path = "~/.local/share/tokman/tokman.db"

[filter]
max_output_lines = 500
strip_ansi = true
compact_json = true

[dashboard]
enabled = true
port = 8080
host = "localhost"

[plugins]
enabled = true
directory = "~/.config/tokman/plugins"
```

### Programmatic Access

```go
package main

import (
    "fmt"
    "github.com/GrayCodeAI/tokman/internal/config"
)

func main() {
    cfg, err := config.Load("")
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Tracking enabled: %v\n", cfg.Tracking.Enabled)
    fmt.Printf("Dashboard port: %d\n", cfg.Dashboard.Port)
}
```

---

## Hook Integration

TokMan can be integrated into shell hooks for automatic command filtering.

### Shell Hook Setup

Add to `~/.zshrc` or `~/.bashrc`:

```bash
# TokMan hook integration
if command -v tokman &> /dev/null; then
    eval "$(tokman init --hook-only)"
fi
```

### Hook Behavior

1. Intercepts commands before execution
2. Rewrites commands to use TokMan wrappers
3. Tracks token savings automatically
4. Falls back to original command if TokMan fails

### Supported Hook Rewrites

| Original | Rewritten |
|----------|-----------|
| `git status` | `tokman git status` |
| `npm test` | `tokman npm test` |
| `docker ps` | `tokman docker ps` |
| `cargo build` | `tokman cargo build` |

---

## Error Handling

### Error Codes

| Code | Description |
|------|-------------|
| `0` | Success |
| `1` | General error |
| `2` | Configuration error |
| `3` | Filter error |
| `4` | Tracking error |
| `5` | Plugin error |

### Error Response Format

```json
{
  "error": {
    "code": 3,
    "message": "Filter pattern compilation failed",
    "details": "invalid regex: unclosed group"
  }
}
```
