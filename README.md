# TokMan - Token-Aware CLI Proxy

**31-layer compression pipeline** for AI coding assistants. Built on research from 120+ papers.

## Features

### Core Compression
- **31-layer pipeline** - Most comprehensive compression of any tool
- **BPE tokenization** (tiktoken) - Accurate token counting
- **SIMD support** - Go 1.26+ vectorized operations
- **Parallel pipeline** - Concurrent layer execution
- **Streaming** - Handles large inputs (>500K tokens)
- **O(1) LRU cache** - Fingerprint-based result caching

### HTTP Proxy Mode (NEW)
- Transparent proxy for OpenAI, Anthropic, Gemini APIs
- Model aliasing and fallback chains
- Request deduplication and caching
- Health check and metrics endpoints

### Advanced Compression Layers
- **KV-Cache Alignment** - Maximize provider-level caching
- **Cross-Message Dedup** - SimHash-based near-duplicate detection
- **Content-Type Detection** - 8 content types, 16 languages
- **Reversible Compression** - LLM can retrieve originals on demand
- **TOON Columnar Encoding** - 40-80% compression on JSON arrays
- **Token Dense Dialect** - Unicode symbol shorthand (8-25% extra savings)
- **LLM-Based Compression** - External LLM semantic compression
- **Adaptive Context Scaling** - Auto-adjusts based on context size
- **Position-Aware/LITM** - Attention-optimal content ordering
- **Photon** - Base64 image compression
- **LogCrunch/DiffCrunch** - Log and diff folding
- **StructuralCollapse** - Import block merging
- **Dictionary Encoding** - Auto-learned codebook substitution

### Intelligence Features
- **Engram Memory** - LLM-driven observational memory
- **Tiered Summaries (L0/L1/L2)** - Multi-resolution memory
- **Feedback Loop Learning** - Learns optimal thresholds
- **Information Bottleneck** - Entropy + task-relevance filtering
- **6 Read Modes** - Full, map, signatures, diff, aggressive, entropy
- **Incremental Deltas** - Myers diff for file changes
- **Project Graph** - Dependency analysis and impact tracking
- **Cross-Session Memory** - Persistent tasks, findings, decisions, facts
- **Multi-Agent Context Sharing** - Inter-agent scratchpad

### Analytics & Monitoring
- **Anomaly Detection** - Rolling mean + 2sigma
- **Spend Forecasting** - Monthly cost projection
- **Model Right-Sizing** - Complexity-based recommendations
- **Token Heatmap** - System/tools/context/history/query breakdown
- **Cacheability Scoring** - 0-100 cache efficiency score
- **Waste Detection** - Duplicate request identification
- **History Bloat Tracking** - Conversation history monitoring
- **Multi-Platform Tracking** - 16+ AI clients
- **Analytics TUI** - Interactive terminal dashboard
- **Leaderboard** - Token savings competition
- **Wrapped Year-in-Review** - Shareable annual summary

### AI Gateway
- **Kill Switches** - Per-model emergency stops
- **Quotas** - Token budget enforcement
- **Model Aliasing** - Route requests to different models
- **Fallback Chains** - Automatic failover
- **Content Guardrails** - PII and injection detection
- **Developer Playground** - Test prompts with cost preview
- **Live Monitor** - Real-time API traffic monitoring

### Security
- **Security Scanner** - SQLi, XSS, SSRF, path traversal detection
- **Secrets Detection** - API keys, tokens, passwords
- **Prompt Injection Detection** - Jailbreak attempt identification
- **PII Redaction** - Email, phone, SSN, credit card detection
- **eBPF Monitoring** - Kernel-level syscall monitoring
- **Network Firewall** - iptables-based egress filtering
- **SIEM Integration** - OCSF v1.1 format
- **Decision Explainability** - Forensic audit records

### Filter System
- **TOML DSL** - Rich filter configuration language
- **Template Pipes** - join, truncate, lines, keep, where, each
- **JSONPath Extraction** - RFC 9535 support
- **Filter Variants** - File-based + output-pattern detection
- **Filter Safety Checks** - Injection and Unicode detection
- **Community Filter Registry** - Publish/share filters
- **Filter Test Suites** - Declarative test system
- **Remote Gain Sync** - Cross-machine usage aggregation
- **Auto-Validation Pipeline** - Post-change validation

### Agent Integrations
- Claude Code, Cursor, Copilot, Gemini CLI, Windsurf, Cline
- Codex, OpenCode, Aider, and 10+ more

## Quick Start

```bash
# Install
go install github.com/GrayCodeAI/tokman/cmd/tokman@latest

# Compress output
tokman summary --preset full < input.txt

# Start HTTP proxy
tokman http-proxy start --listen :8080

# Analytics
tokman analytics --action anomaly
tokman tui

# Security scan
tokman security --action scan < input.txt
```

## Performance

| Metric | Value |
|--------|-------|
| Compression | 60-90% on common dev operations |
| Layers | 31 (research-based) |
| Tokenizer | BPE (tiktoken cl100k_base) |
| Cache | O(1) LRU with fingerprinting |
| SIMD | Go 1.26+ vectorized |

## License

MIT
