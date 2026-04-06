# What TokMan Does vs What Competitors Do — Honest Comparison

## TokMan: What We Actually Ship

### Core: CLI Proxy (142 commands)
```
Agent runs: git status
Hook rewrites to: tokman git status
TokMan runs git, compresses output, returns to agent
```

**What happens inside:**
1. Execute the real command
2. Run output through 37-layer compression pipeline (heuristic, no ML)
3. Track tokens saved in SQLite
4. Return compressed output

### Our 37 Filter Layers
All heuristic/rule-based. No ML models. No LLM calls (unless `--llm` flag).

| Layer | What It Actually Does |
|-------|----------------------|
| Entropy | Remove low-information lines by word frequency |
| Perplexity | Remove repetitive lines (heuristic, NOT real perplexity) |
| Goal-driven | Score lines by relevance to query intent |
| AST preserve | Keep code structure markers (fn, class, if, return) |
| Contrastive | Rank lines by query similarity (keyword matching) |
| N-gram | Replace repeated phrases with abbreviations |
| Evaluator heads | Score importance by position + keywords |
| Gist | Extract key sentences per section |
| Hierarchical | Multi-level summary (headings → subheadings → content) |
| Budget | Hard-cut at token limit |
| Compaction | Compress conversation-style content |
| Attribution | Remove low-importance lines (positional + frequency scoring) |
| H2O | Keep first/last/frequent lines, drop middle |
| Attention sink | Keep first N + last N lines |
| Meta-token | LZ77-style pattern dedup |
| Semantic chunk | Split by topic boundaries, drop low-importance chunks |
| Sketch store | Budget-aware compression with reversibility |
| Lazy pruner | Progressive pruning with decay |
| Semantic anchor | Keep gradient-detected important lines |
| Agent memory | Extract knowledge graph from content |
| + 17 more | Various compression techniques from research papers |

### Analytics (our strongest feature)
- SQLite tracking database
- 35-endpoint web dashboard
- Daily/weekly/monthly reports
- Per-layer effectiveness stats
- Cost projections and team allocation
- CSV/JSON export
- Contribution graphs
- Anomaly detection

### Agent Integration
- Claude Code (PreToolUse hook)
- Cursor, Copilot, Windsurf, Cline, Gemini, Codex, OpenCode, Kiro
- MCP server
- HTTP proxy for LLM APIs
- gRPC API

### Other
- TOML-based custom filter rules
- Reversible compression (undo)
- 7 read modes (full, map, signatures, diff, aggressive, entropy, lines)
- Streaming for >500K tokens
- Result caching (SHA-based)
- Hook integrity verification (SHA-256)
- PII detection
- Encryption at rest

---

## Competitor 1: Mycelium (Rust) — 0 stars

### What They Actually Ship

**Core: CLI Proxy (50+ commands) — same concept as TokMan**
```
Agent runs: git status
Hook rewrites to: mycelium git status
Mycelium runs git, compresses output, returns to agent
```

**Their 5 filtering strategies (vs our 37 layers):**

| Strategy | What It Does |
|----------|-------------|
| Smart filtering | Remove comments, whitespace, boilerplate |
| Grouping | Aggregate by directory, error type, rule |
| Truncation | Head + tail, drop middle |
| Deduplication | Collapse repeated lines with counts |
| Adaptive sizing | Small pass-through, medium filter, large full compress |

**What they have that we DON'T:**

| Feature | Mycelium | TokMan |
|---------|----------|--------|
| `peek` command | ✅ 2-line file summary, ~95% reduction | ❌ Nothing like it |
| Ecosystem: Rhizome | ✅ Tree-sitter code intelligence | ❌ |
| Ecosystem: Hyphae | ✅ RAG memory + vector search | ❌ |
| Ecosystem: Cap | ✅ Web dashboard for memory browsing | ❌ (we have token dashboard only) |
| Plugin system | ✅ Custom filter plugins (shipped) | ❌ Planned but not built |
| Code-aware aggressive read | ✅ Folds function bodies >30 lines | ✅ Similar in read modes |

**What WE have that they don't:**

| Feature | TokMan | Mycelium |
|---------|--------|----------|
| 37-layer research pipeline | ✅ | ❌ 5 strategies only |
| Analytics dashboard (35 endpoints) | ✅ | ❌ |
| Cost tracking + projections | ✅ | ❌ |
| Team cost allocation | ✅ | ❌ |
| Per-layer stats | ✅ | ❌ |
| HTTP/API proxy | ✅ | ❌ |
| gRPC API | ✅ | ❌ |
| Reversible compression | ✅ | ❌ |
| Anomaly detection | ✅ | ❌ |
| PII detection | ✅ | ❌ |
| Encryption at rest | ✅ | ❌ |
| 142 commands | ✅ | ❌ 50+ commands |
| 10+ agent integrations | ✅ | Fewer agents |

**Honest assessment:**
- Same product concept, nearly identical UX
- They have ecosystem (code intel + memory) — we don't
- We have deeper compression + much better analytics
- They're Rust (slightly faster), we're Go (157K lines, more mature)

---

## Competitor 2: gsqz (Rust) — 1 star

### What They Actually Ship

**Core: YAML-configured CLI wrapper**
```bash
gsqz -- cargo test
# Matches "cargo test" pipeline → filter → group → truncate → dedup
```

**Their pipeline (5 steps per command):**
1. Match — regex matches command to pipeline
2. Filter — remove lines matching patterns
3. Group — aggregate (git_status, lint_by_rule, errors_warnings, by_file)
4. Truncate — head + tail with omission marker
5. Dedup — collapse consecutive similar lines

**9 built-in pipelines:**
git-status, git-diff, git-log, pytest, cargo-test, generic-test, python-lint, js-lint, cargo-build

**What they have that we DON'T:**

| Feature | gsqz | TokMan |
|---------|------|--------|
| YAML config | ✅ More familiar to devs | ❌ TOML only |
| Per-pipeline grouping modes | ✅ (git_status, lint_by_rule, by_file) | Partial (TOML rules) |
| Simpler mental model | ✅ 5 clear steps | ❌ 37 layers is confusing |

**What WE have that they don't:**

| Feature | TokMan | gsqz |
|---------|--------|------|
| Everything beyond basic compression | ✅ | ❌ |
| Analytics | ✅ Full dashboard | ❌ `--stats` flag only |
| Agent hooks | ✅ 10+ agents | ❌ Standalone wrapper only |
| Tracking DB | ✅ | ❌ |
| MCP server | ✅ | ❌ |
| HTTP proxy | ✅ | ❌ |
| 142 commands | ✅ | ❌ 9 pipelines |

**Honest assessment:**
- Much simpler than TokMan — that's both weakness and strength
- YAML config is more accessible than TOML
- Part of Gobby platform (could grow)
- We crush them on features, but they're easier to understand

---

## Competitor 3: claude-context-optimizer (TypeScript) — 23 stars ⭐ MOST STARS

### What They Actually Ship

**Core: MCP server with 6 specialized tools (NOT a CLI proxy)**

Claude calls these tools directly instead of running bash commands:

| Tool | What It Does | Their Numbers |
|------|-------------|--------------|
| `smart_read` | Reads file, returns only lines relevant to query | 4,980 → 57 tokens (99%) |
| `compress_logs` | Deduplicates + summarizes log files | 50K → 597 tokens (99%) |
| `project_map` | Generates compact project structure | 95K → 815 tokens (99%) |
| `function_extractor` | Extracts specific function by name | 1,245 → 249 tokens (80%) |
| `bulk_search` | Semantic search across codebase | 50K → 2,284 tokens (95%) |
| `task_checkpoint` | Saves/restores task state across sessions | N/A |

**What they have that we DON'T:**

| Feature | claude-context-optimizer | TokMan |
|---------|------------------------|--------|
| `project_map` | ✅ 95K→815 tokens | ❌ NOTHING like this |
| `smart_read` with query | ✅ Returns only query-relevant lines | Partial (query-aware layer exists but not as a tool) |
| `function_extractor` | ✅ Extract single function by name | ❌ |
| `bulk_search` | ✅ Search across files semantically | ❌ |
| `task_checkpoint` | ✅ Cross-session state | ❌ |
| Published benchmarks | ✅ Reproducible `node tests/benchmark.js` | ❌ No public proof |

**What WE have that they don't:**

| Feature | TokMan | claude-context-optimizer |
|---------|--------|------------------------|
| CLI proxy (transparent) | ✅ Works without changing agent behavior | ❌ Agent must call MCP tools |
| 37-layer pipeline | ✅ | ❌ Per-tool logic only |
| Multi-agent support | ✅ 10+ agents | ❌ Claude Code only |
| Analytics dashboard | ✅ | ❌ |
| Tracking DB | ✅ | ❌ |
| 142 commands | ✅ | ❌ 6 tools |
| HTTP proxy | ✅ | ❌ |

**Honest assessment:**
- DIFFERENT approach (MCP tools vs CLI proxy). Not directly competing.
- Their `project_map` and `smart_read` are genuinely better for specific use cases
- But they require Claude to learn new tools. TokMan is transparent.
- They have the most stars (23) = best marketing/positioning
- **Their published benchmarks are a marketing weapon we lack**

---

## Competitor 4: claude-shorthand (Python) — 4 stars

### What They Actually Ship

**Core: LLMLingua-2 model as a Claude Code hook**
```
Agent runs: any command
Hook intercepts output
LLMLingua-2 model scores each token's importance
Low-importance tokens removed
~55% reduction
```

**What they have that we DON'T:**

| Feature | claude-shorthand | TokMan |
|---------|-----------------|--------|
| Model-based compression | ✅ LLMLingua-2 (real perplexity) | ❌ Heuristic only |
| Quality guarantee | ✅ Trained model knows what matters | ❌ Rule-based guessing |

**What WE have that they don't:**

| Feature | TokMan | claude-shorthand |
|---------|--------|-----------------|
| Speed | ✅ 15ms | ❌ 200ms+ (model inference) |
| Offline | ✅ | ❌ Needs model loaded |
| Savings | ✅ 60-90% | ❌ ~55% |
| Everything else | ✅ Analytics, agents, proxy, etc. | ❌ Just compression |

**Honest assessment:**
- They trade speed for quality. Model-based > heuristic for ambiguous content.
- But 55% vs our 60-90% means our heuristics + many layers beat one good model.
- Real threat: if they add more layers or agents ship LLMLingua natively.

---

## Competitor 5: claude-praetorian-mcp (TypeScript) — 14 stars

### What They Actually Ship

**Core: TOON format compaction for Claude Code conversations**

Not competing on CLI output — compacts conversation context:
- Incremental snapshots after research, subagent tasks
- TOON (Token-Oriented Object Notation) structured format
- Plugin/skill system for Claude Code
- Project-scoped storage in `.claude/praetorian/`

**Honest assessment:**
- Complementary, not competitive. They do conversation, we do CLI output.
- TOON format is interesting — TokMan's compaction layer does similar but less structured.

---

## The Honest Truth

### What competitors are doing that we're NOT:

| Gap | Who Has It | Impact |
|-----|-----------|--------|
| **Project map** (whole repo → 815 tokens) | claude-context-optimizer | 🔴 HUGE — saves thousands of tokens per session |
| **Function extraction** (pull single fn by name) | claude-context-optimizer | 🟡 Useful for targeted edits |
| **Ecosystem** (code intel + memory + dashboard) | Mycelium/Basidiocarp | 🔴 Full platform vs our standalone tool |
| **Model-based compression** (LLMLingua-2) | claude-shorthand | 🟡 Higher quality on ambiguous content |
| **YAML config** (more accessible) | gsqz | 🟡 Lower barrier to entry |
| **Published benchmarks** (reproducible proof) | claude-context-optimizer | 🔴 They can prove claims, we can't |
| **Plugin system** (shipped, not planned) | Mycelium | 🟡 Extensibility |
| **File peek** (2-line summary) | Mycelium | 🟡 Quick file understanding |

### What WE'RE doing that nobody else does:

| Our Advantage | Closest Competitor | Our Lead |
|--------------|-------------------|----------|
| **37-layer research pipeline** | Mycelium (5 strategies) | Massive — 7x more compression techniques |
| **Analytics dashboard** (35 endpoints) | Nobody | No competition |
| **10+ agent integrations** | Mycelium (2-3 agents) | 3-5x more agents |
| **Cost tracking + projections** | Nobody | No competition |
| **Team cost allocation** | Nobody | No competition |
| **142 CLI commands** | Mycelium (50+) | 3x coverage |
| **HTTP/gRPC proxy** | Nobody in this niche | Unique |
| **Reversible compression** | Nobody | Unique |
| **Hook integrity (SHA-256)** | Nobody | Unique |
| **PII detection** | Nobody | Unique |

### Bottom Line

```
THEM: Simpler products, but smarter features in specific areas
  - project_map, smart_read, function_extractor (context-optimizer)
  - Code intelligence ecosystem (Mycelium)  
  - Model-based quality (claude-shorthand)
  - Published proof (benchmarks)

US: Deepest product, but missing targeted intelligence features
  - 37 layers, 142 commands, 10+ agents
  - Best analytics in the category
  - Best security (integrity, PII, encryption)
  - No public benchmarks to prove claims
  - No project_map / smart_read equivalent
  - No code intelligence
  - Plugin system still "planned"
```
