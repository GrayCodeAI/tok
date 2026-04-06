# TokMan Competitive Analysis & Improvement Roadmap

## Top 10 Competitors

| # | Competitor | Category | Approach | Language | Stars |
|---|-----------|----------|----------|----------|-------|
| 1 | **LLMLingua / LongLLMLingua** (Microsoft) | Prompt Compression Library | Model-based perplexity scoring | Python | 4.5K+ |
| 2 | **Aider** | AI Coding CLI | Smart context selection + repo-map | Python | 25K+ |
| 3 | **Cursor** | AI Code Editor | Proprietary context engine | TypeScript | N/A (closed) |
| 4 | **Continue.dev** | Open IDE Extension | RAG-based context retrieval | TypeScript | 20K+ |
| 5 | **Cody (Sourcegraph)** | Code AI Platform | Graph-based code intelligence | Go/TS | 2K+ |
| 6 | **Letta (ex-MemGPT)** | Memory Management Framework | Tiered memory + compaction | Python | 13K+ |
| 7 | **OpenHands (ex-OpenDevin)** | AI Agent Platform | Container sandboxing + context | Python | 45K+ |
| 8 | **SWE-agent** (Princeton) | SWE Benchmark Agent | ACI interface + smart context | Python | 14K+ |
| 9 | **Sweep AI** | Automated PR Agent | Planning + chunked file reading | Python | 7K+ |
| 10 | **Cline/Roo Code** | VS Code AI Agent | Diff-based context, sliding window | TypeScript | 18K+ |

---

## Feature-by-Feature Comparison

### 1. Compression / Context Reduction

| Feature | TokMan | LLMLingua | Aider | Cursor | Letta | Cline |
|---------|--------|-----------|-------|--------|-------|-------|
| Multi-layer pipeline | ✅ 37 layers | ✅ 3 stages | ❌ | ❌ | ❌ | ❌ |
| Entropy filtering | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ |
| Perplexity pruning | ✅ (heuristic) | ✅ (model-based) | ❌ | ❌ | ❌ | ❌ |
| AST-aware compression | ✅ | ❌ | ❌ | ✅ | ❌ | ❌ |
| Semantic chunking | ✅ | ❌ | ❌ | ✅ | ✅ | ❌ |
| Budget enforcement | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| N-gram dedup | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| Query-aware filtering | ✅ | ✅ (LongLLMLingua) | ❌ | ✅ | ❌ | ❌ |
| Claimed reduction | 60-90% | 50-80% | 30-50% | Unknown | 40-60% | 20-40% |

**TokMan advantage:** Broadest compression pipeline — 37 layers vs 3 for LLMLingua.
**TokMan gap:** LLMLingua uses **actual LLM perplexity scores** while TokMan uses heuristic approximation. This gives LLMLingua more precise pruning on ambiguous content.

### 2. Context Management

| Feature | TokMan | Aider | Cursor | Continue | Cody | Cline |
|---------|--------|-------|--------|----------|------|-------|
| Repo-map / code graph | ❌ | ✅ (tree-sitter) | ✅ | ✅ (RAG) | ✅ (SCIP) | ❌ |
| Auto file discovery | ❌ | ✅ | ✅ | ✅ | ✅ | Partial |
| Smart @-mentions | ❌ | ❌ | ✅ | ✅ | ✅ | ❌ |
| Cross-file reference | ❌ | ✅ | ✅ | ✅ | ✅ (graph) | ❌ |
| Conversation compaction | ✅ | ❌ | ✅ | ❌ | ❌ | ✅ |
| Session memory | ✅ | ❌ | ✅ | ❌ | ❌ | ✅ |
| Read modes (map/sig/diff) | ✅ | Partial | ✅ | ❌ | ✅ | Partial |

**TokMan gap:** No repository-level code intelligence. Aider's repo-map and Cody's SCIP graph understand cross-file relationships — TokMan only compresses what's piped to it.

### 3. Agent Integration

| Feature | TokMan | Aider | Cursor | OpenHands | SWE-agent | Cline |
|---------|--------|-------|--------|-----------|-----------|-------|
| Claude Code hooks | ✅ | ❌ | N/A | ❌ | ❌ | ❌ |
| Cursor integration | ✅ | ❌ | Native | ❌ | ❌ | ❌ |
| Multi-agent support | ✅ (10 agents) | Self-only | Self-only | Self-only | Self-only | Self-only |
| MCP server | ✅ | ❌ | ❌ | ❌ | ❌ | ✅ |
| Transparent proxy | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| HTTP proxy for APIs | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| Zero-config setup | ✅ (`init -g`) | ✅ | ✅ | Docker | Manual | ✅ |

**TokMan advantage:** Only tool that works as transparent middleware across 10+ agents. Competitors are self-contained.

### 4. Analytics & Observability

| Feature | TokMan | Aider | Cursor | Continue | Cody | Letta |
|---------|--------|-------|--------|----------|------|-------|
| Token tracking DB | ✅ SQLite | ❌ | ❌ | ❌ | ❌ | ❌ |
| Cost dashboard | ✅ Web UI | ❌ | Basic | ❌ | ❌ | ❌ |
| Per-layer stats | ✅ | N/A | ❌ | ❌ | ❌ | ❌ |
| Daily/weekly reports | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| Model attribution | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| Cost projection | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| Export (CSV/JSON) | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| Team cost allocation | ✅ | ❌ | ❌ | ❌ | Enterprise | ❌ |

**TokMan advantage:** Most comprehensive analytics. No competitor tracks per-layer compression effectiveness.

### 5. Performance & Architecture

| Feature | TokMan | LLMLingua | Aider | Cursor | SWE-agent |
|---------|--------|-----------|-------|--------|-----------|
| Language | Go | Python | Python | TS/Rust | Python |
| Startup time | ~15ms | ~2s | ~1s | ~500ms | ~3s |
| Binary size | ~12MB | pip install | pip install | App bundle | pip install |
| Streaming support | ✅ (>500K) | ❌ | ❌ | ✅ | ❌ |
| SIMD optimization | Planned | ❌ | ❌ | ❌ | ❌ |
| Result caching | ✅ SHA-based | ❌ | Partial | ✅ | ❌ |
| Max context | 2M tokens | Model limit | Model limit | Model limit | Model limit |
| Offline capable | ✅ | ❌ (needs model) | Partial | ❌ | ❌ |

**TokMan advantage:** Go binary is 100x faster startup than Python tools. Fully offline.
**TokMan gap:** Python tools have richer ML ecosystem for advanced compression.

### 6. Extensibility

| Feature | TokMan | Aider | Cursor | Continue | Cline |
|---------|--------|-------|--------|----------|-------|
| TOML filter config | ✅ | ❌ | ❌ | ❌ | ❌ |
| Plugin system | Planned (WASM) | ❌ | ✅ Extensions | ✅ | ❌ |
| Custom layers | ✅ (Go) | ❌ | ❌ | ✅ (TS) | ❌ |
| DSL/rules engine | ✅ | ❌ | ❌ | ❌ | ❌ |
| Marketplace | In progress | ❌ | ✅ | ✅ | ❌ |
| API/SDK | ✅ gRPC | ❌ | ❌ | ✅ REST | ❌ |

---

## Gap Analysis — Where TokMan Falls Short

### GAP 1: No Repository-Level Code Intelligence (Critical)
**Who does it better:** Aider (repo-map), Cody (SCIP graph), Cursor (proprietary)

Aider builds a tree-sitter-based repository map showing all functions, classes, and their call relationships. When a user asks about a bug, Aider automatically includes relevant cross-file dependencies — not just the file being edited.

TokMan only compresses CLI output after it's generated. It has no understanding of which files are relevant to a task.

### GAP 2: No Semantic Code Search / RAG (High)
**Who does it better:** Continue.dev (embeddings), Cody (SCIP + embeddings), Cursor

Continue.dev indexes the entire codebase into a vector store and retrieves the most relevant snippets for each query. Cody uses Sourcegraph's code graph + embeddings for precise retrieval.

TokMan lacks any form of semantic search or embedding-based retrieval.

### GAP 3: Heuristic-Only Perplexity (Medium)
**Who does it better:** LLMLingua (GPT-2/LLaMA perplexity), CompactPrompt

LLMLingua uses actual language model perplexity to identify low-information tokens with high accuracy. TokMan's perplexity layer uses word frequency heuristics — effective but less precise.

### GAP 4: No Diff-Based Context Tracking (Medium)
**Who does it better:** Cline (diff tracking), Aider (git-aware), SWE-agent

Cline tracks which files have been modified during a session and prioritizes showing diffs over full file content. Aider automatically includes git diffs in context.

TokMan compresses output but doesn't track session-level file modifications to intelligently prioritize context.

### GAP 5: No Multi-Turn Planning (Medium)
**Who does it better:** OpenHands, SWE-agent, Sweep

OpenHands and SWE-agent maintain multi-step plans and only include context relevant to the current step. Sweep decomposes PRs into sub-tasks and manages context per task.

TokMan is stateless per-command — it doesn't understand multi-step workflows.

### GAP 6: WASM Plugin System Not Implemented (Low)
**Who does it better:** Continue.dev (TS plugins), Cursor (extensions)

The plugin system is documented and stubbed but not implemented.

### GAP 7: No Evaluation / Quality Benchmark Suite (Low)
**Who does it better:** LLMLingua (perplexity benchmarks), SWE-agent (SWE-bench)

No systematic way to measure whether compression degrades agent task performance.

---

## 25 Recommended Improvements (Prioritized)

### Tier 1: High Impact, Medium Effort (Do First)

#### 1. **Repository Map / Code Graph**
Build a tree-sitter-based repository map (like Aider's repo-map). When `tokman read main.go` is called, automatically include referenced symbols from other files.

```
tokman read --context main.go  →  main.go + imported function signatures
```

**Impact:** Agents get cross-file understanding without reading every file.
**Effort:** 2-3 weeks. Tree-sitter Go bindings exist.
**Reference:** Aider's `RepoMap` class.

#### 2. **Embedding-Based Semantic Search**
Add a local vector index (using Go embeddings or sqlite-vss) for semantic code search:

```
tokman search "where is the auth middleware?"  →  top-5 relevant snippets
```

**Impact:** Agents find relevant code in 1 call instead of grep + read cycles.
**Effort:** 2-3 weeks. Use pre-computed embeddings, store in SQLite.
**Reference:** Continue.dev's indexing pipeline.

#### 3. **Session-Aware Context Tracking**
Track files read/modified per session. Offer `tokman context` to return only changed files + their diffs:

```
tokman context  →  modified files since session start + relevant unchanged files
```

**Impact:** 50-70% reduction in redundant context reads.
**Effort:** 1-2 weeks. Extend existing session tracking.

#### 4. **Git-Aware Smart Diff**
Integrate with git to provide intelligent diffs:

```
tokman git diff --smart  →  only semantically meaningful changes (skip whitespace, imports)
```

**Impact:** Git diffs are one of the largest token consumers.
**Effort:** 1 week. Tree-sitter for language-aware diffing.

#### 5. **Model-Based Perplexity Option**
Add optional integration with local models (Ollama) for real perplexity scoring:

```
tokman --perplexity-model ollama:qwen2 git log
```

**Impact:** 10-20% better compression quality on ambiguous content.
**Effort:** 1 week. HTTP call to Ollama tokenize endpoint.

### Tier 2: High Impact, Higher Effort

#### 6. **Intelligent File Bundling**
When agent reads file A, automatically include related files (imports, tests, configs):

```
tokman read --bundle src/auth.go  →  auth.go + auth_test.go + middleware.go (signatures only)
```

**Impact:** Reduces multi-call overhead. Agents need fewer round trips.
**Effort:** 2-3 weeks. Requires import graph analysis.

#### 7. **Task-Aware Context Pruning**
Accept task description and only include context relevant to it:

```
TOKMAN_TASK="fix the login bug" tokman read src/
```

**Impact:** Massive reduction for large codebases.
**Effort:** 3-4 weeks. Combines embeddings + task decomposition.

#### 8. **Streaming Compression for Chat APIs**
Compress SSE/streaming API responses in real-time:

```
tokman http-proxy start --stream-compress
```

**Impact:** Reduces streaming response tokens (output costs).
**Effort:** 2 weeks. SSE parser + streaming pipeline already exist.

#### 9. **WASM Plugin Runtime**
Implement the planned WASM plugin system:

```
tokman plugin install my-custom-filter.wasm
```

**Impact:** Community-driven filter ecosystem.
**Effort:** 2-3 weeks. Wazero dependency already in go.mod.

#### 10. **Compression Quality Benchmark**
Build an evaluation suite that measures agent success rate with/without compression:

```
tokman benchmark --suite swe-bench-lite
```

**Impact:** Proves compression doesn't degrade quality. Marketing material.
**Effort:** 2-3 weeks.

### Tier 3: Medium Impact, Focused Effort

#### 11. **Multi-Step Plan Tracking**
Track agent's plan across commands and adjust context per step:

```
tokman plan start "refactor auth module"
tokman plan step 1  →  only auth-related context
```

**Effort:** 2 weeks.

#### 12. **Codebase Digest / Summary**
Generate a compressed project summary for new sessions:

```
tokman digest  →  project structure + key APIs + recent changes (500 tokens)
```

**Effort:** 1 week. Combine tree + signatures + recent git log.

#### 13. **Cost Alerts & Budget Guards**
Real-time alerts when session cost exceeds threshold:

```
tokman budget --daily-max $5 --alert slack
```

Existing alert infrastructure exists but lacks notification channels.
**Effort:** 1 week.

#### 14. **Differential Compression**
Only send diffs from last context read, not full content:

```
tokman read main.go  →  [first read: full file]
tokman read main.go  →  [second read: only changed lines since last read]
```

**Effort:** 1-2 weeks. Delta tracking infrastructure already exists.

#### 15. **A/B Testing for Layer Combinations**
Automatically test which layer combinations work best per command type:

```
tokman autotune --commands "git diff,go test" --duration 7d
```

A/B test infrastructure already exists in `internal/abtest/`.
**Effort:** 1-2 weeks to wire it up.

#### 16. **IDE Extension (VS Code)**
Create a VS Code extension that shows token savings in the status bar:

```
[TokMan: 847 tokens saved | $0.03 saved today]
```

**Effort:** 2-3 weeks.

#### 17. **Team Dashboard**
Extend dashboard with team-level views:
- Per-developer token usage
- Project-level cost breakdown
- Anomaly detection (already built)

Team cost allocation already exists in `internal/teamcosts/`.
**Effort:** 1-2 weeks.

#### 18. **Prompt Template Library**
Curated prompt templates that are pre-compressed:

```
tokman template apply "code-review" < diff.txt
```

Prompt management exists in `internal/llm/prompts.go`.
**Effort:** 1 week.

### Tier 4: Differentiators / Moats

#### 19. **Compression-as-a-Service API**
Public API for other tools to use TokMan's pipeline:

```
POST /v1/compress
{"text": "...", "mode": "extract", "budget": 2000}
```

gRPC server already exists. Add REST + rate limiting + auth.
**Effort:** 1-2 weeks.

#### 20. **Agent-Specific Compression Profiles**
Auto-detect which agent is calling and optimize compression for its behavior:
- Claude Code: aggressive on file reads, preserve error messages
- Cursor: preserve code blocks, compress explanations
- Aider: preserve git context, compress test output

**Effort:** 2 weeks. Agent detection already exists via TOKMAN_AGENT env.

#### 21. **Privacy-Preserving Compression**
PII detection before sending to cloud LLMs:

```
tokman --pii-filter read .env  →  [REDACTED: API_KEY=***]
```

PII detector already exists in `internal/pii/`.
**Effort:** 1 week to integrate into pipeline.

#### 22. **Token Usage Forecasting**
Predict daily/weekly token usage based on patterns:

```
tokman forecast --days 30
→ Projected: 2.4M tokens ($7.20) based on current trend
```

Projection handler already exists in dashboard.
**Effort:** 1 week.

#### 23. **Cross-Session Learning**
Learn which patterns produce the best compression per project:

```
tokman learn  →  auto-adjusts thresholds based on historical effectiveness
```

Autotune exists in `internal/autotune/`. Feedback loop exists in filter.
**Effort:** 2 weeks to connect and productize.

#### 24. **Output Verification / Guardrails**
Verify compressed output preserves critical information:

```
tokman verify --original full.txt --compressed compressed.txt
→ ✅ All error messages preserved
→ ✅ All file paths preserved
→ ⚠️ 2 stack frames removed (non-critical)
```

Guardrails engine exists in `internal/guardrails/`.
**Effort:** 1-2 weeks.

#### 25. **LLM Gateway with Automatic Prompt Optimization**
Full API gateway that transparently compresses all LLM API calls:

```
# Point your app at TokMan instead of OpenAI
OPENAI_BASE_URL=http://localhost:8080/v1 my-app
```

HTTP proxy exists but only handles simple cases. Extend to handle
function calling, tool use, image inputs, and streaming.
**Effort:** 3-4 weeks.

---

## Strategic Summary

```
┌─────────────────────────────────────────────────────────────┐
│                    TokMan's Position                         │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  STRONGEST:                                                  │
│  ✦ Multi-agent middleware (only player here)                │
│  ✦ Compression depth (37 layers, nobody close)              │
│  ✦ Analytics/observability (best in class)                  │
│  ✦ Go performance (100x faster than Python tools)           │
│  ✦ Offline/local (no cloud dependency)                      │
│                                                              │
│  WEAKEST:                                                    │
│  ✦ No code intelligence (Aider/Cody/Cursor gap)            │
│  ✦ No semantic search (Continue.dev gap)                    │
│  ✦ Heuristic-only ML (LLMLingua gap)                       │
│  ✦ No multi-step planning (OpenHands gap)                   │
│                                                              │
│  UNIQUE OPPORTUNITY:                                         │
│  → Combine TokMan's compression with code intelligence      │
│  → Become the "smart context layer" for ALL agents           │
│  → No competitor occupies this middleware + intelligence     │
│    intersection                                              │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

### Recommended Priority Order

1. **Repository map** (#1) — closes biggest gap, 2-3 weeks
2. **Session context tracking** (#3) — quick win, 1-2 weeks
3. **Git-aware smart diff** (#4) — high frequency command, 1 week
4. **Codebase digest** (#12) — instant value for new sessions, 1 week
5. **Compression benchmark** (#10) — proves quality, enables marketing
