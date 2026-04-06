# TokMan — Real OSS Competitors (April 2026)

## Actual Competitors Found on GitHub

There are **8 real competitors** — all emerged in late 2025 / early 2026, targeting the same niche:
**reducing token consumption for AI coding agents**.

| # | Project | Stars | Lang | Created | Approach |
|---|---------|:-----:|------|---------|----------|
| 1 | **[Mycelium](https://github.com/basidiocarp/mycelium)** | 0 | Rust | ~2026 | CLI proxy, 5 filtering strategies, part of Basidiocarp ecosystem |
| 2 | **[gsqz](https://github.com/GobbyAI/gsqz)** | 1 | Rust | Mar 2026 | YAML-configurable CLI compressor, 20+ pipelines, part of Gobby platform |
| 3 | **[claude-context-optimizer](https://github.com/AzozzALFiras/claude-context-optimizer)** | 23 | TypeScript | Mar 2026 | MCP server — smart_read, log compression, project_map, bulk_search |
| 4 | **[claude-praetorian-mcp](https://github.com/Vvkmnn/claude-praetorian-mcp)** | 14 | TypeScript | Dec 2025 | MCP server — TOON format compaction, incremental snapshots |
| 5 | **[Thicc](https://github.com/immapolar/Thicc)** | 9 | JavaScript | Dec 2025 | Conversation compressor — eliminates "Context low" warnings |
| 6 | **[claude-rolling-context](https://github.com/NodeNestor/claude-rolling-context)** | 9 | Python | Mar 2026 | Rolling context compression — auto-compresses old messages |
| 7 | **[claude-shorthand](https://github.com/gladehq/claude-shorthand)** | 4 | Python | Mar 2026 | LLMLingua-2 hook for Claude Code — ~55% reduction |
| 8 | **[PromptThin](https://github.com/theFoOl-oo-oo/promptthin)** | 0 | SaaS | Apr 2026 | API proxy — cache + compression + model routing + context pruning |

Also relevant but lower threat:
- **[logslimmer](https://github.com/aredesrafa/logslimmer)** (1★, JS) — log-only compression
- **[tinyprompt](https://github.com/sidedwards/tinyprompt)** (2★, Python) — generic prompt compression CLI
- **[prompt-compress](https://github.com/DevvGwardo/prompt-compress)** (2★, Rust) — token importance scoring
- **[prompt-compression-gateway](https://github.com/Kelpejol/prompt-compression-gateway)** (3★, Python) — LLMLingua API gateway

---

## Competitor Deep Dive

### 1. Mycelium (Basidiocarp) — MOST DIRECT COMPETITOR

**What it is:** Rust CLI proxy, nearly identical concept to TokMan.

**Similarities to TokMan (suspicious):**
- Same savings table (identical numbers, same format)
- Same command structure (`mycelium git status`, `mycelium -v`, `mycelium -u`)
- Same flags (`--verbose`, `--ultra-compact`, `--skip-env`)
- Same `init -g` pattern for Claude Code hooks
- Same 50+ commands across same categories
- Same adaptive filtering (small/medium/large thresholds)
- Same tee-on-failure, tracking DB, analytics

**Differences:**
| | TokMan | Mycelium |
|---|---|---|
| Language | Go | Rust |
| Pipeline | 37-layer research-based | 5 strategies (filter, group, truncate, dedup, adaptive) |
| Ecosystem | Standalone | Part of Basidiocarp (Hyphae memory, Rhizome code intel, Cap dashboard) |
| Code intel | ❌ | ✅ via Rhizome (tree-sitter) |
| RAG/memory | ❌ | ✅ via Hyphae (vector search) |
| Plugin system | Planned WASM | ✅ Custom filter plugins |
| MCP | ✅ | Via Stipe |

**Threat level: 🔴 HIGH** — Nearly identical product. Ecosystem approach (Rhizome + Hyphae) gives it code intelligence TokMan lacks.

---

### 2. gsqz (GobbyAI) — CLOSEST TECHNICAL COMPETITOR

**What it is:** Rust CLI compressor with YAML-configurable pipelines.

**Key features:**
- 20+ built-in pipelines (git, cargo, pytest, eslint...)
- YAML config (vs TokMan's TOML)
- Pattern matching → filter → group → truncate → dedup
- ~9ms overhead
- Claims >95% reduction
- Part of Gobby AI platform

**Differences:**
| | TokMan | gsqz |
|---|---|---|
| Config | TOML | YAML |
| Pipeline | 37 research layers | 5-step pipeline (match→filter→group→truncate→dedup) |
| Analytics | ✅ Full dashboard | Basic `--stats` flag |
| Agent integration | ✅ 10+ agents via hooks | Standalone wrapper |
| Tracking DB | ✅ SQLite | ❌ |
| HTTP proxy | ✅ | ❌ |
| MCP server | ✅ | ❌ |

**Threat level: 🟡 MEDIUM** — Simpler but YAML config is more accessible. Part of Gobby platform.

---

### 3. claude-context-optimizer — MCP APPROACH

**What it is:** MCP server providing 6 tools for Claude Code context reduction.

**Key features:**
- `smart_read` — reads files with semantic extraction (99% reduction on code files)
- `compress_logs` — deduplicates/summarizes log output
- `project_map` — generates compact project structure (95K → 815 tokens)
- `function_extractor` — extracts function signatures only
- `bulk_search` — semantic search across files
- `task_checkpoint` — saves/restores task state

**Benchmarked:** Claims 98% reduction with actual benchmark numbers.

**Differences:**
| | TokMan | claude-context-optimizer |
|---|---|---|
| Approach | CLI proxy (hook rewrite) | MCP server (tool calls) |
| Integration | Transparent — intercepts bash | Explicit — Claude calls MCP tools |
| Scope | All CLI output | File reads + logs + search |
| Compression | 37-layer pipeline | Per-tool specialized logic |
| Analytics | ✅ Full | ❌ |
| Multi-agent | ✅ 10+ agents | ❌ Claude Code only |

**Threat level: 🟡 MEDIUM** — Different approach (MCP vs proxy) but solves same problem. `smart_read` and `project_map` are features TokMan lacks.

---

### 4. claude-praetorian-mcp — CONTEXT COMPACTION

**What it is:** MCP server for aggressive context compaction using TOON format.

**Key features:**
- Incremental compaction snapshots
- TOON (Token-Oriented Object Notation) format
- Auto-compacts after web research, subagent tasks
- Project-scoped storage in `.claude/praetorian/`
- Plugin + skill system for Claude Code

**Differences:**
| | TokMan | Praetorian |
|---|---|---|
| Focus | CLI output compression | Conversation context compaction |
| Format | Plain text | TOON structured format |
| Trigger | Every command | High-value moments (research, subagents) |
| Storage | SQLite DB | Flat files |

**Threat level: 🟢 LOW** — Complementary. Compacts conversation context, not CLI output.

---

### 5. Thicc + claude-rolling-context — CONVERSATION COMPRESSORS

**What they are:** Tools to compress Claude Code conversation history.

**Key features:**
- Thicc: JSONL conversation compressor, Ollama integration
- Rolling-context: Auto-compresses old messages, keeps recent verbatim

**Threat level: 🟢 LOW** — Solves a different problem (conversation length, not CLI output).

---

### 6. claude-shorthand — LLMLingua HOOK

**What it is:** Python hook that runs LLMLingua-2 on CLI output before Claude sees it.

**Key features:**
- Uses actual LLMLingua-2 model for compression
- ~55% reduction
- PreToolUse hook (same pattern as TokMan)

**Differences:**
| | TokMan | claude-shorthand |
|---|---|---|
| Compression | 37 heuristic layers | LLMLingua-2 model |
| Speed | ~15ms | ~200ms+ (model inference) |
| Quality | Heuristic | Model-based (higher fidelity) |
| Offline | ✅ | Needs model loaded |
| Analytics | ✅ Full | ❌ |
| Coverage | 50+ commands | All output (generic) |

**Threat level: 🟡 MEDIUM** — Model-based compression is higher quality. If agents start preferring quality over speed, this approach wins.

---

### 7. PromptThin — API PROXY

**What it is:** SaaS proxy with 4 savings routes.

**Key features:**
- Semantic cache (100% savings on repeated queries)
- LLMLingua-2 prompt compression
- Model router (routes simple tasks to cheaper models)
- Context pruning (summarizes long conversations)

**Differences:**
| | TokMan | PromptThin |
|---|---|---|
| Deployment | Local binary | Cloud SaaS |
| Focus | CLI output | API request/response |
| Pricing | Free/OSS | Paid SaaS |
| Privacy | ✅ Local | ❌ Routes through cloud |

**Threat level: 🟢 LOW** — SaaS vs local, different deployment model. Privacy-sensitive users prefer TokMan.

---

## Competitive Matrix

| Feature | TokMan | Mycelium | gsqz | claude-context-opt | Praetorian | claude-shorthand |
|---------|:---:|:---:|:---:|:---:|:---:|:---:|
| **CLI proxy** | ✅ | ✅ | ✅ | ❌ MCP | ❌ MCP | ✅ Hook |
| **Multi-layer pipeline** | ✅ 37 | ❌ 5 strategies | ❌ 5 steps | ❌ per-tool | ❌ | ✅ LLMLingua |
| **Agent hook integration** | ✅ 10+ | ✅ Claude+Gemini | ❌ | ❌ | ✅ Claude | ✅ Claude |
| **Analytics dashboard** | ✅ Web | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Token tracking DB** | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ |
| **MCP server** | ✅ | Via Stipe | ❌ | ✅ | ✅ | ❌ |
| **HTTP/API proxy** | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Code intelligence** | ❌ | ✅ Rhizome | ❌ | ✅ smart_read | ❌ | ❌ |
| **RAG / memory** | ❌ | ✅ Hyphae | ❌ | ❌ | ✅ snapshots | ❌ |
| **Project map** | ❌ | ❌ | ❌ | ✅ | ❌ | ❌ |
| **Model-based compression** | ❌ heuristic | ❌ | ❌ | ❌ | ❌ | ✅ LLMLingua-2 |
| **Plugin system** | Planned | ✅ | ❌ | ❌ | ✅ skills | ❌ |
| **Custom filter config** | ✅ TOML | ✅ | ✅ YAML | ❌ | ❌ | ❌ |
| **Reversible compression** | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Cost tracking** | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Team features** | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Offline** | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ needs model |
| **Language** | Go | Rust | Rust | TypeScript | TypeScript | Python |
| **Startup** | ~15ms | ~5-15ms | ~9ms | N/A (MCP) | N/A (MCP) | ~200ms+ |

---

## What TokMan Should Do

### Immediate (Mycelium is copying your playbook)

| Priority | Action | Why |
|:--------:|--------|-----|
| 🔴 | **Ship WASM plugin system** | Mycelium already has plugins. gsqz has YAML extensibility. TokMan's is "planned". |
| 🔴 | **Add `project_map` command** | claude-context-optimizer's killer feature. 95K → 815 tokens. TokMan has nothing like it. |
| 🔴 | **Add `smart_read` with AST extraction** | claude-context-optimizer gets 99% reduction on code files by extracting signatures. TokMan's `read` modes exist but aren't exposed as a headline feature. |
| 🟡 | **Benchmark suite with published numbers** | claude-context-optimizer publishes reproducible benchmarks. TokMan claims 60-90% but has no public proof. |
| 🟡 | **YAML config option** | gsqz uses YAML which is more familiar to most devs than TOML. Consider supporting both. |

### Medium-term (Differentiate from the pack)

| Priority | Action | Why |
|:--------:|--------|-----|
| 🟡 | **Expose analytics as a competitive moat** | Nobody else has a dashboard, cost tracking, or team features. Make this the headline differentiator. |
| 🟡 | **LLMLingua integration as optional layer** | claude-shorthand proves model-based compression works. Add as opt-in layer via Ollama. |
| 🟡 | **Ecosystem story** | Mycelium has Rhizome (code intel) + Hyphae (memory) + Cap (dashboard). TokMan is standalone — position this as strength (simpler) or add partnerships. |
| 🟢 | **Conversation compaction marketing** | TokMan already has compaction (Layer 11). Thicc and rolling-context have stars for this. Market the existing feature harder. |
| 🟢 | **Speed benchmark** | TokMan (Go, 15ms) vs Mycelium (Rust, 5-15ms) vs gsqz (Rust, 9ms). Publish head-to-head numbers. |

### Do NOT do

| ❌ | Why not |
|----|---------|
| Build a SaaS version | PromptThin occupies this. TokMan's local-first is a strength. |
| Rewrite in Rust | Go is fast enough and the codebase is 157K lines. Stability > marginal speed. |
| Build a full IDE | Cursor/Continue territory. Stay middleware. |

---

## Threat Assessment Summary

```
THREAT LEVEL:

  Mycelium (Basidiocarp)     🔴🔴🔴🔴🔴  Nearly identical + ecosystem advantage
  gsqz (GobbyAI)             🟡🟡🟡       Simpler but growing, YAML config
  claude-context-optimizer    🟡🟡🟡       MCP approach, smart_read is strong
  claude-shorthand            🟡🟡         Model-based quality advantage
  Praetorian                  🟢🟢         Complementary, conversation focus
  Thicc / rolling-context     🟢           Narrow scope, conversation only
  PromptThin                  🟢           SaaS, different market

KEY INSIGHT:
  The market is forming RIGHT NOW (all competitors < 6 months old).
  TokMan has the deepest pipeline (37 layers) and best analytics.
  But Mycelium's ecosystem (code intel + memory) is a serious threat.
  
  Window to establish dominance: ~3-6 months.
```
