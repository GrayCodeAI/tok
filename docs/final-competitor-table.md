# TokMan vs Top 10 Real Competitors — Final Comparison

## The Field

| # | Project | Stars | Lang | Type | Since |
|---|---------|:-----:|------|------|-------|
| 1 | **[tokf](https://github.com/mpecan/tokf)** | 140 | Rust | CLI proxy | Feb 2026 |
| 2 | **[CLOV](https://github.com/alexandephilia/clov-ai)** | 33 | Go/Rust/Python | MCP proxy + CLI | Mar 2026 |
| 3 | **[claude-context-optimizer](https://github.com/AzozzALFiras/claude-context-optimizer)** | 23 | TypeScript | MCP server | Mar 2026 |
| 4 | **[claude-modular](https://github.com/oxygen-fragment/claude-modular)** | 276 | N/A (config) | Framework | Jul 2025 |
| 5 | **[th0th](https://github.com/S1LV4/th0th)** | 130 | TypeScript | Semantic search API | Feb 2026 |
| 6 | **[opencode-magic-context](https://github.com/cortexkit/opencode-magic-context)** | 46 | TypeScript | Plugin | Mar 2026 |
| 7 | **[chop](https://github.com/AgusRdz/chop)** | 7 | Go | CLI proxy | Mar 2026 |
| 8 | **[claude-shorthand](https://github.com/gladehq/claude-shorthand)** | 4 | Python | Hook | Mar 2026 |
| 9 | **[hush](https://github.com/omergulen/hush)** | 3 | Shell | Hook | Mar 2026 |
| 10 | **[mycelium](https://github.com/basidiocarp/mycelium)** | 0 | Rust | CLI proxy | Mar 2026 |

## Direct Feature Comparison

| Feature | TokMan | tokf | CLOV | context-opt | chop | hush | mycelium |
|---------|:---:|:---:|:---:|:---:|:---:|:---:|:---:|
| **CLI proxy** | ✅ | ✅ | ✅ | ❌ MCP | ✅ | ✅ Hook | ✅ |
| **MCP proxy** | ❌ | ❌ | ✅ | ✅ | ❌ | ❌ | ✅ |
| **Hook install** | ✅ 10+ agents | ✅ 3 agents | ✅ | ❌ | ✅ | ✅ Claude/Cursor | ✅ |
| **Compression layers** | ✅ **37** | ❌ TOML rules | ❌ structured | ❌ per-tool | ❌ basic | ❌ basic | ❌ 5 strat |
| **Smart read / semantic read** | ❌ | ❌ | ❌ | ✅ 99% | ❌ | ❌ | ✅ |
| **Project map** | ❌ | ❌ | ❌ | ✅ 99% | ❌ | ❌ | ✅ |
| **Read modes** | ✅ **7 modes** | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ 3 levels |
| **TOML custom filters** | ✅ | ✅ | ✅ TOML | ❌ | ❌ | ✅ conf file | ❌ |
| **Streaming** | ✅ >500K | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Reversible compression** | ✅ | ❌ | ✅ Tee mode | ❌ | ❌ | ✅ breadcrumbs | ❌ |
| **Color passthrough** | ✅ | ✅ | ✅ strip ANSI | ❌ | ❌ | ❌ | ❌ |
| **Prefer-less mode** | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Task runner wrapping** | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Budget enforcement** | ✅ | ❌ | ✅ --max-tokens | ❌ | ❌ | ❌ | ❌ |
| **Tracking database** | ✅ SQLite | ❌ | ✅ SQLite | ✅ Session | ❌ | ❌ | ✅ SQLite |
| **Web analytics dashboard** | ✅ **35 endpoints** | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Per-layer statistics** | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Cost tracking** | ✅ | ❌ | ✅ | ❌ | ❌ | ❌ | ❌ |
| **Cost projection** | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Team allocation** | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **PII detection** | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Encryption at rest** | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Filter safety checks** | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Filter test/verify** | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Filter eject/customize** | ❌ | ✅ | ✅ config | ❌ | ❌ | ✅ conf file | ❌ |
| **HTTP proxy** | ✅ | ❌ | ✅ MCP only | ❌ | ❌ | ❌ | ❌ |
| **gRPC API** | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Plugin system** | ❌ planned | ❌ | ❌ | ✅ (skills) | ❌ | ❌ | ✅ |
| **MCP server** | ✅ | ❌ | ❌ | ✅ | ❌ | ❌ | ✅ |
| **Published benchmarks** | ❌ | ❌ | ✅ table | ✅ **reproducible** | ❌ | ❌ | ❌ |
| **Homebrew/cargo install** | ❌ Go only | ✅ brew+cargo | ✅ brew+cargo | ✅ npm | ✅ | ❌ git only | ✅ cargo |
| **Website** | ✅ tokman.dev | ✅ tokf.net | ❌ | ❌ | ❌ | ❌ | ❌ |
| **7 compression tiers** | ✅ | ❌ | ✅ presets | ❌ | ❌ | ❌ | ❌ |
| **LLM compression layer** | ❌ heuristic | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Anomaly detection** | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Session management** | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Agent attribution** | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Context quality scores** | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Contribution graphs** | ✅ | ✅ pulse | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Hook integrity (SHA-256)** | ✅ | ❌ | ✅ | ❌ | ❌ | ❌ | ❌ |
| **Command coverage** | ✅ 142 | ✅ ~40 | ✅ ~40 | ✅ 6 tools | ❌ limited | ✅ ~90 | ✅ 50+ |

## Compression Approach Comparison

| | TokMan | tokf | CLOV | context-opt | chop | mycelium |
|---|---|---|---|---|---|---|
| **What it compresses** | CLI output + API | CLI output | CLI + MCP responses | File reads + logs | CLI output | CLI output |
| **Method** | 37 research layers | TOML pattern rules | Structured JSON + CLI rules | 6 specialized MCP tools | Basic dedup + summary | 5 filtering strategies |
| **Token scoring** | Heuristic (frequency, position, entropy) | Regex patterns | Structure-aware + tokenizer profiles | Semantic | Basic text dedup | Positional + frequency |
| **ML/model used** | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Claimed reduction** | 60-90% | 60-90% | 76-95% | 80-99% | 50-90% | 60-90% |
| **Proof** | ❌ No public benchmarks | ❌ No public benchmarks | ❌ Table claims | ✅ Reproducible test | ❌ Claims | ❌ Same table as TokMan |
| **Token type** | Tiktoken/BPE heuristic | Regex/line counting | Approx + tokenizer profiles | Real tokenizer (tiktoken) | Length-based | Length-based |

## Management Comparison

| | TokMan | tokf | CLOV | context-opt | th0th |
|---|---|---|---|---|---|
| **Tracking** | ✅ SQLite, full records | ❌ | ✅ SQLite, basic | ✅ Session state | ✅ Vector DB |
| **Dashboard** | ✅ Web, 35 endpoints | ❌ | ❌ | ❌ | ❌ API only |
| **Cost analytics** | ✅ Per model, project, team | ❌ | ❌ | ❌ | ❌ |
| **Forecasting** | ✅ Cost projection | ❌ | ❌ | ❌ | ❌ |
| **Pulse/Stats** | ✅ `tokman gain` | ❌ | ✅ `clov pulse` | ❌ | ❌ |
| **Ecosystem** | ❌ Standalone | ❌ | ❌ | ❌ | ✅ (search API) |
| **RAG/Memory** | ❌ | ❌ | ❌ | ✅ Session | ✅ Full RAG |
| **Semantic search** | ❌ | ❌ | ❌ | ✅ bulk_search | ✅ 98% reduction |

## Distribution & Reach

| | TokMan | tokf | CLOV | context-opt | th0th |
|---|---|---|---|---|---|
| **Package manager** | Go install only | ✅ brew + crates.io | ✅ brew + cargo | ✅ npm | bun |
| **Website** | ✅ tokman.dev | ✅ tokf.net | ❌ | ❌ | ❌ |
| **Stars** | — | **140** | **33** | **23** | **130** |
| **Pre-built binaries** | ❌ | ✅ releases | ✅ releases | ❌ (npx) | ❌ |
| **Zero-config** | ✅ `tokman init -g` | ✅ `tokf setup` | ✅ `clov hook --global` | ❌ manual MCP | ✅ setup script |
| **CI/CD** | ✅ GitHub Actions | ✅ GitHub Actions | ? | ? | ? |
| **Documentation** | AGENTS.md | README + website | README | README | README |

## What Makes Each Unique

| Project | Unique Position |
|---------|----------------|
| **TokMan** | Only complete token lifecycle: 37-layer optimization + full management suite (dashboard, costs, teams, alerts, attribution) |
| **tokf** | Best distribution: brew + crates + homebrew tap. Filter test suites (`tokf verify`). Most polished UX. 140 stars. |
| **CLOV** | Only MCP response proxy. Intercepts web search, database connectors, not just CLI. Built on RTK. Go+Rust+Python. |
| **context-opt** | Most proven: published benchmarks (98% reproducible). Smart specialized tools (project_map, smart_read). 23 stars. |
| **th0th** | Vector semantic search with 98% reduction. Ollama-based, fully offline. 130 stars. Different approach entirely. |
| **claude-modular** | Claude Code framework with 30+ commands and progressive disclosure. 276 stars but it's a config template, not a tool. |
| **chop** | Cleanest explanation of cascade effect. Supports 4 agents (Claude, Gemini, Codex, Antigravity). Go-based. |
| **hush** | Zero dependencies — pure bash + jq. Claude Code plugin marketplace. ~90 commands. Simplest option. |
| **claude-shorthand** | Only tool using real ML compression (LLMLingua-2). ~55% reduction. Trade speed for quality. |
| **mycelium** | Ecosystem play: Rhizome (code intel) + Hyphae (memory) + Cap (dashboard) + Lamella (plugins). Nearly identical to TokMan CLI. |

## The Honest Summary

**TokMan wins on:**
- Deepest compression: 37 layers vs everything else (5 at most)
- Only full management suite: dashboard, cost tracking, projections, team allocation, anomaly detection
- Most security: PII, encryption, SHA-256 integrity, safety checks
- Most commands: 142 vs 30-90 range
- Most agent support: 10+ vs 1-3
- Only with HTTP proxy and gRPC API

**TokMan loses on:**
- No published benchmarks (vs context-opt's reproducible proof)
- No Homebrew/crates.io distribution (vs tokf's brew install)
- No `project_map` or `smart_read` (vs context-opt, mycelium, th0th)
- No MCP response filtering (vs CLOV's unique MCP proxy)
- Plugin system not shipped (vs mycelium, context-opt, hush)
- Lower star count than tokf (140), th0th (130), claude-modular (276)
