# Changelog

## [2.0.0] - 2026-03-31

### Major Features
- HTTP Proxy Mode for OpenAI, Anthropic, Gemini APIs
- KV-Cache Alignment for provider-level caching optimization
- Cross-Message Deduplication with SimHash near-duplicate detection
- Content-Type Auto-Detection (8 types, 16 languages)
- Reversible Compression with LLM retrieval support
- TOON Columnar Encoding for JSON arrays (40-80% compression)
- Token Dense Dialect with Unicode symbol shorthand
- LLM-Based Compression via external processes
- Adaptive Context Scaling with per-model profiles
- Position-Aware/LITM content ordering
- Engram Memory System with tiered summaries (L0/L1/L2)
- Feedback Loop Learning for threshold optimization
- Information Bottleneck Filter
- 6 Read Modes (full, map, signatures, diff, aggressive, entropy)
- Incremental Deltas for file changes
- Project Intelligence Graph with dependency analysis
- Cross-Session Memory with persistent storage
- Multi-Agent Context Sharing
- Analytics: anomaly detection, forecasting, right-sizing, heatmap, waste detection
- Multi-Platform Tracking (16+ AI clients)
- Analytics TUI dashboard
- AI Gateway with kill switches, quotas, model aliasing, fallback chains
- Content Guardrails (PII, injection detection)
- Security Scanner (SQLi, XSS, SSRF, path traversal)
- eBPF Monitoring and Network Firewall
- SIEM Integration with OCSF format
- Decision Explainability with forensic audit records
- Template Pipes (join, truncate, lines, keep, where, each)
- JSONPath Extraction (RFC 9535)
- Filter Variants and Safety Checks
- Community Filter Registry
- Filter Test Suites
- Auto-Validation Pipeline
- Developer Playground
- Live Monitor (htop-style)
- PATH Shim Injection
- Color Passthrough
- Prefer-Less Mode
- Task Runner Wrapping
- Photon (base64 image compression)
- LogCrunch, DiffCrunch, StructuralCollapse
- Dictionary Encoding with auto-learned codebook
- Leaderboard, Wrapped Year-in-Review, GitHub Profile Widget

### Performance
- BPE tokenization (tiktoken cl100k_base)
- O(1) LRU cache with doubly-linked list
- SIMD support (Go 1.26+)
- Parallel pipeline execution
- Streaming for large inputs

### Breaking Changes
- Go version requirement: 1.26+
- Pipeline configuration format updated
