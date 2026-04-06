# TokMan vs Competitors — Full Comparison Table

## Overview

| | TokMan | LLMLingua | Aider | Cursor | Continue.dev | Cody | Letta (MemGPT) | OpenHands | SWE-agent | Cline |
|---|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|
| **Type** | CLI Proxy | Library | CLI Agent | IDE | IDE Extension | IDE Extension | Framework | Agent Platform | Agent | VS Code Agent |
| **Language** | Go | Python | Python | TS/Rust | TypeScript | Go/TS | Python | Python | Python | TypeScript |
| **Open Source** | ✅ MIT | ✅ MIT | ✅ Apache 2 | ❌ Closed | ✅ Apache 2 | ✅ Apache 2 | ✅ Apache 2 | ✅ MIT | ✅ MIT | ✅ Apache 2 |
| **Standalone** | ✅ | ✅ | ✅ | ❌ App | ❌ Extension | ❌ Extension | ✅ | ✅ Docker | ✅ | ❌ Extension |
| **Stars** | — | 4.5K+ | 25K+ | N/A | 20K+ | 2K+ | 13K+ | 45K+ | 14K+ | 18K+ |

---

## Compression & Token Reduction

| Feature | TokMan | LLMLingua | Aider | Cursor | Continue | Cody | Letta | OpenHands | SWE-agent | Cline |
|---------|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|
| **Compression pipeline** | 37 layers | 3 stages | ❌ | Proprietary | ❌ | ❌ | 1 layer | ❌ | ❌ | ❌ |
| **Entropy filtering** | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Perplexity pruning** | ✅ Heuristic | ✅ Model-based | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **AST-aware compression** | ✅ | ❌ | ❌ | ✅ | ❌ | ✅ | ❌ | ❌ | ❌ | ❌ |
| **N-gram deduplication** | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Semantic chunking** | ✅ | ❌ | ❌ | ✅ | ❌ | ❌ | ✅ | ❌ | ❌ | ❌ |
| **H2O heavy-hitter** | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Attention sink** | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Meta-token lossless** | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Agent memory extraction** | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ | ❌ | ❌ | ❌ |
| **Query-aware filtering** | ✅ | ✅ | ❌ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Budget enforcement** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Reversible compression** | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Configurable tiers** | ✅ 7 tiers | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **TOML filter rules** | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Claimed reduction** | **60-90%** | 50-80% | 30-50% | Unknown | 20-40% | 20-40% | 40-60% | N/A | N/A | 20-40% |

---

## Code Intelligence & Context

| Feature | TokMan | LLMLingua | Aider | Cursor | Continue | Cody | Letta | OpenHands | SWE-agent | Cline |
|---------|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|
| **Repo map / code graph** | ❌ | ❌ | ✅ tree-sitter | ✅ Proprietary | ✅ RAG | ✅ SCIP | ❌ | ❌ | ❌ | ❌ |
| **Auto file discovery** | ❌ | ❌ | ✅ | ✅ | ✅ | ✅ | ❌ | Partial | Partial | Partial |
| **Semantic code search** | ❌ | ❌ | ❌ | ✅ | ✅ embeddings | ✅ embeddings | ❌ | ❌ | ❌ | ❌ |
| **Cross-file references** | ❌ | ❌ | ✅ | ✅ | ✅ | ✅ graph | ❌ | ❌ | ❌ | ❌ |
| **@-mention context** | ❌ | ❌ | ❌ | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ |
| **Smart file reading** | ✅ 7 modes | ❌ | Partial | ✅ | ❌ | ✅ | ❌ | ❌ | ❌ | Partial |
| **Conversation compaction** | ✅ | ❌ | ❌ | ✅ | ❌ | ❌ | ✅ | ❌ | ❌ | ✅ |
| **Session memory** | ✅ | ❌ | ❌ | ✅ | ❌ | ❌ | ✅ | ✅ | ❌ | ✅ |
| **Diff-based context** | Partial | ❌ | ✅ git-aware | ✅ | ❌ | ❌ | ❌ | ✅ | ✅ | ✅ |
| **Multi-step planning** | ❌ | ❌ | ❌ | ✅ | ❌ | ❌ | ❌ | ✅ | ✅ | ❌ |

---

## Agent Integration

| Feature | TokMan | LLMLingua | Aider | Cursor | Continue | Cody | Letta | OpenHands | SWE-agent | Cline |
|---------|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|
| **Claude Code** | ✅ Hook | ❌ | ❌ | N/A | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Cursor** | ✅ Hook | ❌ | ❌ | Native | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Copilot** | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Windsurf** | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Cline** | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | Native |
| **Gemini CLI** | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **OpenCode** | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Codex** | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Kiro** | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Multi-agent support** | **✅ 10+** | ❌ | Self only | Self only | Self only | Self only | Self only | Self only | Self only | Self only |
| **MCP server** | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ |
| **HTTP/API proxy** | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Transparent intercept** | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Zero-config setup** | ✅ `init -g` | pip install | ✅ | ✅ | ✅ | ✅ | ✅ | Docker | Manual | ✅ |

---

## Analytics & Observability

| Feature | TokMan | LLMLingua | Aider | Cursor | Continue | Cody | Letta | OpenHands | SWE-agent | Cline |
|---------|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|
| **Token tracking database** | ✅ SQLite | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Web dashboard** | ✅ | ❌ | ❌ | Basic | ❌ | ❌ | ❌ | ✅ | ❌ | ❌ |
| **Per-layer stats** | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Cost tracking** | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Cost projection** | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Daily/weekly reports** | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Model attribution** | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Team cost allocation** | ✅ | ❌ | ❌ | ❌ | ❌ | Enterprise | ❌ | ❌ | ❌ | ❌ |
| **Export CSV/JSON** | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Anomaly detection** | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Contribution graph** | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |

---

## Performance & Architecture

| Feature | TokMan | LLMLingua | Aider | Cursor | Continue | Cody | Letta | OpenHands | SWE-agent | Cline |
|---------|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|
| **Startup time** | **~15ms** | ~2s | ~1s | ~500ms | ~300ms | ~200ms | ~2s | ~5s | ~3s | ~300ms |
| **Binary / Install** | 12MB binary | pip (200MB+) | pip (50MB+) | 300MB app | Extension | Extension | pip (100MB+) | Docker 2GB+ | pip (100MB+) | Extension |
| **Offline capable** | ✅ Full | ❌ Needs model | Partial | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Max context** | 2M tokens | Model limit | Model limit | Model limit | Model limit | Model limit | Unlimited (virtual) | Model limit | Model limit | Model limit |
| **Streaming** | ✅ >500K | ❌ | ❌ | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ | ✅ |
| **Result caching** | ✅ SHA-based | ❌ | Partial | ✅ | ❌ | ✅ | ❌ | ❌ | ❌ | ❌ |
| **Concurrent safe** | ✅ | ✅ | ❌ | ✅ | ✅ | ✅ | ❌ | ✅ | ❌ | ✅ |
| **SIMD optimized** | Planned | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **gRPC API** | ✅ | ❌ | ❌ | ❌ | REST | ❌ | REST | REST | ❌ | ❌ |

---

## Extensibility & Ecosystem

| Feature | TokMan | LLMLingua | Aider | Cursor | Continue | Cody | Letta | OpenHands | SWE-agent | Cline |
|---------|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|
| **Plugin system** | Planned WASM | ❌ | ❌ | ✅ Extensions | ✅ TS plugins | ❌ | ✅ Tools | ❌ | ❌ | ❌ |
| **Custom filters/layers** | ✅ Go + TOML | ❌ | ❌ | ❌ | ✅ TS | ❌ | ✅ Python | ❌ | ❌ | ❌ |
| **Marketplace** | In progress | ❌ | ❌ | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **CLI completions** | ✅ | N/A | ✅ | N/A | N/A | N/A | ❌ | ❌ | ❌ | N/A |
| **SDK / client lib** | ✅ Go | ✅ Python | ❌ | ❌ | ✅ TS | ❌ | ✅ Python | ❌ | ❌ | ❌ |
| **Config file** | ✅ TOML | ❌ | ✅ YAML | JSON | JSON | JSON | ✅ | ✅ | ✅ YAML | JSON |

---

## Security & Safety

| Feature | TokMan | LLMLingua | Aider | Cursor | Continue | Cody | Letta | OpenHands | SWE-agent | Cline |
|---------|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|
| **Hook integrity (SHA-256)** | ✅ | N/A | N/A | N/A | N/A | N/A | N/A | N/A | N/A | N/A |
| **Command injection guard** | ✅ | N/A | Partial | ✅ | N/A | N/A | N/A | ✅ Docker | ❌ | Partial |
| **PII detection** | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Filter safety checks** | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Prompt injection guard** | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ | ❌ | ❌ |
| **Sandboxed execution** | ❌ | N/A | ❌ | ✅ | ❌ | ❌ | ❌ | ✅ Docker | ✅ Docker | ❌ |
| **Audit logging** | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ | ❌ | ❌ |
| **Data encryption at rest** | ✅ AES-GCM | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |

---

## Command Coverage

| Category | TokMan | Aider | Cursor | Cline |
|----------|:---:|:---:|:---:|:---:|
| **Git (status/diff/log)** | ✅ Compressed | ✅ Built-in | ✅ Built-in | ✅ Raw |
| **GitHub CLI (gh)** | ✅ Compressed | ❌ | ❌ | ❌ |
| **File ops (cat/ls/tree/find)** | ✅ Compressed | ✅ Raw | ✅ Built-in | ✅ Raw |
| **Grep/search** | ✅ Compressed | ✅ Raw | ✅ Built-in | ✅ Raw |
| **Test runners** | ✅ Compressed | ✅ Raw | ✅ Raw | ✅ Raw |
| **Build tools** | ✅ Compressed | ✅ Raw | ✅ Raw | ✅ Raw |
| **Linters** | ✅ Compressed | ✅ Raw | ✅ Raw | ✅ Raw |
| **Package managers** | ✅ Compressed | ✅ Raw | ✅ Raw | ✅ Raw |
| **Docker/K8s** | ✅ Compressed | ❌ | ❌ | ✅ Raw |
| **Cloud (AWS)** | ✅ Compressed | ❌ | ❌ | ✅ Raw |
| **curl/wget** | ✅ Compressed | ❌ | ❌ | ✅ Raw |

---

## Scorecard (1-10)

| Dimension | TokMan | LLMLingua | Aider | Cursor | Continue | Cody | Letta | OpenHands | SWE-agent | Cline |
|-----------|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|
| **Compression quality** | **10** | 8 | 2 | 5 | 2 | 2 | 5 | 1 | 1 | 2 |
| **Code intelligence** | **2** | 1 | 8 | 9 | 7 | 9 | 1 | 3 | 4 | 3 |
| **Agent integration** | **10** | 2 | 3 | 3 | 3 | 3 | 4 | 3 | 2 | 3 |
| **Analytics** | **10** | 1 | 1 | 2 | 1 | 2 | 1 | 3 | 1 | 1 |
| **Performance** | **10** | 4 | 5 | 7 | 6 | 7 | 4 | 3 | 3 | 6 |
| **Security** | **9** | 3 | 3 | 6 | 3 | 4 | 3 | 7 | 5 | 3 |
| **Extensibility** | **7** | 3 | 2 | 8 | 8 | 3 | 7 | 3 | 2 | 2 |
| **Ease of setup** | **9** | 6 | 8 | 9 | 8 | 7 | 5 | 4 | 3 | 8 |
| | | | | | | | | | | |
| **TOTAL** | **67** | 28 | 32 | 49 | 38 | 37 | 30 | 27 | 21 | 28 |

---

## Key Takeaway

```
TokMan's Unfilled Niche:

  Compression ████████████████████ 10/10  ← Best in class
  Analytics   ████████████████████ 10/10  ← Best in class
  Integration ████████████████████ 10/10  ← Only multi-agent middleware
  Performance ████████████████████ 10/10  ← Go >> Python
  Security    ██████████████████   9/10   ← Strong
  Extensibility ██████████████     7/10   ← WASM pending
  Code Intel  ████                 2/10   ← BIGGEST GAP
  
  Fix code intelligence → no competitor can match TokMan.
```
