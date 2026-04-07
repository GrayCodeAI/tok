# Basic Usage Examples

## Simple Command Filtering

### Git Status

```bash
# Without TokMan (112 tokens):
$ git status
On branch main
Your branch is up to date with 'origin/main'.

Changes to be committed:
  (use "git restore --staged <file>..." to unstage)
        modified:   README.md
        new file:   docs/API.md

Changes not staged for commit:
  (use "git add <file>..." to update what will be committed)
        modified:   internal/filter/pipeline.go

# With TokMan (16 tokens):
$ tokman git status
M README.md, docs/API.md (staged)
M internal/filter/pipeline.go (unstaged)
```

**Savings: 85.7%**

### Git Log

```bash
# Without TokMan (371 tokens):
$ git log --oneline -5
a1b2c3d feat: add entropy filtering layer
d4e5f6a fix: handle empty input in pipeline
7890abc docs: update architecture guide
bcd1234 refactor: simplify config loading
ef56789 test: add pipeline benchmarks

# With TokMan (53 tokens):
$ tokman git log -n 5
a1b2c3d feat: entropy filtering
d4e5f6a fix: empty input
7890abc docs: architecture
bcd1234 refactor: config
ef56789 test: benchmarks
```

**Savings: 85.7%**

### Go Test

```bash
# Without TokMan (689 tokens):
$ go test ./...
ok   github.com/GrayCodeAI/tokman/internal/filter    4.306s
ok   github.com/GrayCodeAI/tokman/internal/core       1.221s
ok   github.com/GrayCodeAI/tokman/internal/config      2.359s
ok   github.com/GrayCodeAI/tokman/internal/tracking    5.355s
ok   github.com/GrayCodeAI/tokman/internal/toml       2.336s
ok   github.com/GrayCodeAI/tokman/internal/utils       2.577s
FAIL github.com/GrayCodeAI/tokman/internal/commands    0.034s

# With TokMan (12 tokens):
$ tokman go test ./...
6 passed, 1 FAILED
FAIL internal/commands (0.034s)
```

**Savings: 98.3%** - Only failures highlighted!

### Docker PS

```bash
# Without TokMan (300 tokens):
$ docker ps
CONTAINER ID   IMAGE          COMMAND                  CREATED       STATUS       PORTS                    NAMES
abc123def456   postgres:15    "docker-entrypoint.s…"   2 hours ago   Up 2 hours   0.0.0.0:5432->5432/tcp   db
789ghi012jkl   redis:7        "docker-entrypoint.s…"   2 hours ago   Up 2 hours   0.0.0.0:6379->6379/tcp   cache
345mno678pqr   nginx:latest   "/docker-entrypoint.…"   2 hours ago   Up 2 hours   0.0.0.0:80->80/tcp       web

# With TokMan (60 tokens):
$ tokman docker ps
3 containers: db(postgres:15), cache(redis:7), web(nginx)
```

**Savings: 80%**

## Using Presets

```bash
# Fast preset - fewer layers, faster processing
tokman --preset fast git diff

# Balanced preset - default mix (recommended)
tokman --preset balanced git diff

# Full preset - all layers, maximum compression
tokman --preset full git diff
```

## Using Modes

```bash
# Minimal mode - light compression
tokman --mode minimal git status

# Aggressive mode - maximum compression
tokman --mode aggressive git log -n 20
```

## Verbose Output

```bash
# See what TokMan is doing
tokman -v git status

# Output includes:
# [TokMan] Intercepting: git status
# [TokMan] Applying 31-layer pipeline
# [TokMan] Layer 1 (entropy): skipped (too short)
# [TokMan] Layer 10 (budget): applied, saved 85 tokens
# [TokMan] Result: 112 → 16 tokens (85.7% saved)
```

## Quality Metrics

```bash
# See compression quality grades
tokman --quality git diff

# Output includes:
# Quality: A (0.92/1.0)
# Semantic: 0.95  Signal: 0.91  Context: 0.93
# Readable: 0.89  Complete: 0.90  Ratio: 0.85
```
