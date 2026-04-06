# TokMan vs RTK — The Honest Comparison

## The Relationship

| Check | RTK | TokMan |
|-------|-----|--------|
| Same Discord server (1470188214710046894) | ✅ | ✅ |
| Identical savings table (ls/tree/grep/git/docker) | ✅ | ✅ |
| Same CLI structure (`init -g`, `gain`, `discover`) | ✅ | ✅ |
| Same hook pattern (PreToolUse rewrite) | ✅ | ✅ |
| Same command coverage (git, gh, cargo, test, docker, aws) | ✅ | ✅ |
| Same TOML config pattern | ✅ | ✅ |
| Same flag names (`-v`, `-u`, `--skip-env`) | ✅ | ✅ |

**TokMan and RTK share the same lineage.** RTK is the public-facing project with 18,838 stars. TokMan is the internal/extended version.

---

## What Makes Them Different

| Feature | RTK (Public, 18.8K★) | TokMan (Internal) |
|---------|:---:|:---:|
| **Language** | Rust | Go |
| **Compression approach** | 4 strategies (filter, group, truncate, dedup) | **37-layer research pipeline** |
| **Commands** | ~50 top-level | ~142 top-level |
| **Analytics** | `rtk gain` (CLI only) | **Web dashboard (35 endpoints)** |
| **Cost tracking** | ❌ | **✅ Per model, project, team** |
| **Cost projection** | ❌ | **✅ Forecast future spending** |
| **Team allocation** | ❌ | **✅ Per-developer cost breakdown** |
| **Budget enforcement** | ❌ | **✅ Hard token limits** |
| **Budget alerts** | ❌ | **✅ Daily/weekly limits** |
| **Anomaly detection** | ❌ | **✅ Unusual usage patterns** |
| **Contribution graphs** | ✅ ASCII only | **✅ Full web graphs** |
| **Per-layer stats** | ❌ | **✅ Which layers are effective** |
| **Audit system** | ❌ | **✅ SHA-256 hook integrity** |
| **PII detection** | ❌ | **✅ Redact sensitive data** |
| **Encryption at rest** | ❌ | **✅ AES-GCM** |
| **Filter safety checks** | ❌ | **✅ Injection detection** |
| **Learn system** | ❌ | **✅ Auto-tune filters** |
| **Context quality scores** | ❌ | **✅ Per-read effectiveness** |
| **Parse failure tracking** | ❌ | **✅ Failed commands tracked** |
| **HTTP proxy** | ❌ | **✅ LLM API compression** |
| **gRPC API** | ❌ | **✅ Compression-as-a-service** |
| **Read modes** | ❌ | **✅ 7 modes (full/map/sig/diff/aggressive/entropy/lines)** |
| **Reversible compression** | ❌ (tee only) | **✅ Full undo** |
| **Streaming** | ❌ | **✅ >500K token chunked processing** |
| **Budget tiers** | ❌ | **✅ 7 tiers (surface/trim/extract/core/code/log/thread)** |
| **Telemetry** | ✅ (enabled by default, cloud) | ✅ (opt-in, local) |
| **Distribution** | Homebrew, cargo, releases, install.sh | Go install, source build |
| **Stars** | **18,838** | — |
| **Website** | rtk-ai.app | tokman.dev |
| **Languages** | English, FR, ZH, JA, KO, ES | English |

---

## What RTK Does Better

| Feature | Why It's Better |
|---------|----------------|
| **Homebrew formula** | `brew install rtk` — TokMan has no package manager support |
| **Cargo registry** | `cargo install` — clean installation |
| **Pre-built binaries** | macOS, Linux, Windows binaries on releases |
| **6 languages** | README translated to FR, ZH, JA, KO, ES |
| **412 open issues** | Active community filing bugs and features |
| **1,026 forks** | Ecosystem of plugins and integrations |
| **OpenCode plugin** | Built-in TypeScript plugin |
| **OpenClaw plugin** | Plugin ecosystem support |
| **Apache-2.0 license** | More permissive than MIT for enterprise |
| **GitHub Actions CI** | Formal CI/CD pipeline |
| **Security policy** | SECURITY.md with responsible disclosure |

---

## What TokMan Does Better

| Feature | Why It Matters |
|---------|---------------|
| **37-layer pipeline vs 4 strategies** | 7x more compression techniques. RTK uses simple filter+group+truncate+dedup. TokMan uses entropy, H2O, attention sink, semantic chunking, lazy pruning, agent memory, etc. |
| **Full web dashboard** | RTK only has `rtk gain` in CLI. TokMan has 35 API endpoints, charts, projections. |
| **Team cost allocation** | RTK has no team features. TokMan tracks per-developer costs. |
| **Security** | PII detection, encryption, safety checks — RTK has none of this. |
| **Reversible compression** | TokMan can undo compression. RTK can only save raw output on failure via tee. |
| **7 read modes** | TokMan has full/map/signatures/diff/aggressive/entropy/lines. RTK has basic read only. |
| **7 compression tiers** | TokMan has surface/trim/extract/core/code/log/thread. RTK has no tiers. |
| **HTTP/gRPC proxy** | TokMan can compress LLM API calls, not just CLI output. |
| **Local-first telemetry** | RTK sends telemetry to cloud by default. TokMan is opt-in and local. |
| **Agent attribution** | TokMan tracks which AI agent ran each command (Claude, Cursor, etc.). RTK doesn't. |

---

## External Competitors (Not Related to RTK/TokMan)

| Project | Stars | Type | Key Difference from RTK/TokMan |
|---------|:-----:|------|-------------------------------|
| **[tokf](https://github.com/mpecan/tokf)** | 140 | Rust CLI | Filter test suites (`tokf verify`), filter eject, best distribution |
| **[claude-context-optimizer](https://github.com/AzozzALFiras/claude-context-optimizer)** | 23 | TS MCP | `project_map` tool (95K→815 tokens), published reproducible benchmarks |
| **[CLOV](https://github.com/alexandephilia/clov-ai)** | 33 | Go/Rust MCP | MCP response interception (web search, databases) — intercepts more than CLI |
| **[th0th](https://github.com/S1LV4/th0th)** | 130 | TS RAG | Semantic vector search, 98% reduction, Ollama-based, fully offline |
| **[claude-modular](https://github.com/oxygen-fragment/claude-modular)** | 276 | Config | Progressive disclosure framework, 20+ Claude Code commands |
| **[claude-shorthand](https://github.com/gladehq/claude-shorthand)** | 4 | Python Hook | Only ML compression (LLMLingua-2) — real model-based scoring |
| **[hush](https://github.com/omergulen/hush)** | 3 | Bash | Zero dependencies, pure bash+jq, ~90 commands via config |
| **[chop](https://github.com/AgusRdz/chop)** | 7 | Go CLI | Supports 4 agents (Claude, Gemini, Codex, Antigravity) |
| **[opencode-magic-context](https://github.com/cortexkit/opencode-magic-context)** | 46 | TS Plugin | Cache-aware infinite context, cross-session memory |
| **[mycelium](https://github.com/basidiocarp/mycelium)** | 0 | Rust CLI | Ecosystem: Rhizome (code intel) + Hyphae (memory) + Cap (dashboard) |

---

## The Numbers

| Metric | RTK | TokMan |
|--------|-----|--------|
| **Stars** | 18,838 | — |
| **Forks** | 1,026 | — |
| **Open issues** | 412 | — |
| **Languages supported** | 6 | 1 |
| **Package managers** | brew, cargo, releases | go install only |
| **CLI analytics** | `rtk gain` (basic) | `tokman gain` + full dashboard |
| **Web analytics** | ❌ | ✅ 35 endpoints |
| **Filter layers** | 4 strategies | 37 layers |
| **Read modes** | 1 | 7 |
| **Compression tiers** | ❌ | 7 |
| **Agent integrations** | 10 | 10+ |
| **Command coverage** | ~50 | ~142 |
| **Team features** | ❌ | ✅ |
| **Security features** | ❌ | ✅ PII + encryption + safety |
| **HTTP proxy** | ❌ | ✅ |
| **gRPC API** | ❌ | ✅ |
| **Created** | Jan 22, 2026 | Similar timeframe |
| **License** | Apache-2.0 | MIT |
| **Website** | rtk-ai.app | tokman.dev |
| **Telemetry** | Cloud, enabled by default | Local, opt-in |

---

## The Positioning

**RTK** is the mass-market product: easy to install, broad language support, 18.8K stars growing daily. It does the one thing well — compress CLI output — and does it simply.

**TokMan** is the power-user/internal product: 37 research layers, full analytics dashboard, team management, security features. It's the complete token lifecycle platform.

**The gap:** TokMan has deeper compression (37 vs 4 layers) but worse distribution (no homebrew, no pre-built binaries, no translations). RTK has better visibility and reach.

## What TokMan Should Learn from RTK

| Do This | Why |
|---------|-----|
| Get on Homebrew | `brew install tokman` is 10x easier than `go install` |
| Build pre-binned binaries | macOS, Linux, Windows releases |
| Support more agents natively | RTK integrates OpenClaw, OpenCode, Mistral Vibe (planned) |
| Publish SECURITY.md | Professional security policy |
| Add GitHub Actions CI | Formal testing pipeline |
| Translate README | 6 languages = global reach |
| Track issues openly | 412 open issues = active community |

## Where RTK Can Learn from TokMan

| Do This | Why |
|---------|-----|
| Add web dashboard | CLI-only analytics is limiting |
| Add team cost tracking | Enterprise needs |
| Add PII detection | Security-conscious users |
| Add filter safety checks | Enterprise compliance |
| Add encryption at rest | Data protection |
| Add multiple read modes | Power users need flexibility |
| Add compression tiers | Different use cases need different depth |
| Local-default telemetry | Privacy-first positioning |
