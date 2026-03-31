# TokMan 200-Task Master Plan

**Date:** 2026-03-31
**Status:** Planning Phase
**Sources:** rtk, claw-compactor, tokf, lean-ctx, kompact, tokscale, tamp, token-lens, token-enhancer, clawshield

---

## Phase 1: HTTP Proxy Mode (Tasks 1-25)
*Inspired by: kompact, tamp, token-lens, token-enhancer, clawshield*

- [ ] 1. Create `internal/proxy/proxy.go` - HTTP proxy server skeleton
- [ ] 2. Implement request interceptor for OpenAI-compatible API format
- [ ] 3. Implement request interceptor for Anthropic API format
- [ ] 4. Implement request interceptor for Gemini API format
- [ ] 5. Create message compressor for system prompts
- [ ] 6. Create message compressor for user messages
- [ ] 7. Create message compressor for assistant messages
- [ ] 8. Create message compressor for tool results
- [ ] 9. Implement streaming response passthrough
- [ ] 10. Add TLS certificate support for HTTPS proxy
- [ ] 11. Create proxy configuration (`~/.config/tokman/proxy.toml`)
- [ ] 12. Add `tokman proxy start` command
- [ ] 13. Add `tokman proxy stop` command
- [ ] 14. Add `tokman proxy status` command
- [ ] 15. Implement API key passthrough (preserve auth headers)
- [ ] 16. Add request/response logging for debugging
- [ ] 17. Implement per-request compression statistics
- [ ] 18. Add proxy health check endpoint (`/health`)
- [ ] 19. Add proxy metrics endpoint (`/metrics`)
- [ ] 20. Implement model aliasing (route `gpt-4` → `gpt-4o-mini`)
- [ ] 21. Add request deduplication (cache identical requests)
- [ ] 22. Implement fallback model chains
- [ ] 23. Add rate limiting for proxy requests
- [ ] 24. Create systemd service file for proxy
- [ ] 25. Write proxy integration tests

## Phase 2: KV-Cache Alignment (Tasks 26-35)
*Inspired by: claw-compactor (QuantumLock), kompact (cache_aligner), tamp*

- [ ] 26. Create `internal/filter/kv_cache.go` - KV-cache alignment layer
- [ ] 27. Implement stable prefix detection in system prompts
- [ ] 28. Implement dynamic content isolation (move to end of prompt)
- [ ] 29. Add byte-stable prefix preservation
- [ ] 30. Implement cache-aware compression mode
- [ ] 31. Add cacheability scoring (0-100) for content blocks
- [ ] 32. Implement static/dynamic content classification
- [ ] 33. Add cache hit rate estimation
- [ ] 34. Create `tokman cache-stats` command
- [ ] 35. Add KV-cache optimization to proxy mode

## Phase 3: Cross-Message Deduplication (Tasks 36-50)
*Inspired by: claw-compactor (SemanticDedup), tamp, lean-ctx*

- [ ] 36. Create `internal/filter/dedup.go` - Cross-message dedup layer
- [ ] 37. Implement SimHash fingerprint generation
- [ ] 38. Implement Hamming distance near-duplicate detection
- [ ] 39. Add conversation turn tracking
- [ ] 40. Implement cross-turn content fingerprinting
- [ ] 41. Add file content cache (avoid re-sending same file)
- [ ] 42. Implement command output deduplication across turns
- [ ] 43. Add diff generation for similar re-reads
- [ ] 44. Implement context block similarity detection
- [ ] 45. Add deduplication statistics tracking
- [ ] 46. Create `tokman dedup-stats` command
- [ ] 47. Implement cross-file deduplication (shared imports/boilerplate)
- [ ] 48. Add dedup-aware token counting
- [ ] 49. Implement dedup threshold configuration
- [ ] 50. Write deduplication benchmark tests

## Phase 4: Content-Type Auto-Detection (Tasks 51-65)
*Inspired by: claw-compactor (Cortex), lean-ctx, tamp*

- [ ] 51. Create `internal/detect/content.go` - Content type detector
- [ ] 52. Implement code detection (Python, JS, TS, Go, Rust, Ruby, etc.)
- [ ] 53. Implement JSON detection and schema extraction
- [ ] 54. Implement log detection and format identification
- [ ] 55. Implement diff detection (unified, context, git)
- [ ] 56. Implement search result detection (grep, rg, find)
- [ ] 57. Implement natural language text detection
- [ ] 58. Implement XML/HTML detection
- [ ] 59. Implement YAML/TOML detection
- [ ] 60. Implement CSV/TSV detection
- [ ] 61. Add language detection for code (16+ languages)
- [ ] 62. Implement content-type routing to appropriate compressors
- [ ] 63. Add content-type detection to pipeline coordinator
- [ ] 64. Create `tokman detect <file>` command for testing
- [ ] 65. Write content detection test suite

## Phase 5: Reversible Compression (Tasks 66-80)
*Inspired by: claw-compactor (RewindStore)*

- [ ] 66. Create `internal/filter/reversible.go` - Reversible compression engine
- [ ] 67. Implement hash-addressed storage for compressed sections
- [ ] 68. Implement marker generation (`[rewind:abc123...]`)
- [ ] 69. Create rewind store with LRU eviction
- [ ] 70. Implement `tokman rewind <marker>` command
- [ ] 71. Add rewind tool for AI agents (LLM can retrieve originals)
- [ ] 72. Implement compression level selection (lossless vs lossy)
- [ ] 73. Add rewind statistics tracking
- [ ] 74. Implement reversible JSON compression
- [ ] 75. Implement reversible code compression
- [ ] 76. Implement reversible log compression
- [ ] 77. Add rewind store persistence (survives restarts)
- [ ] 78. Create `tokman rewind-stats` command
- [ ] 79. Add rewind store size limits and cleanup
- [ ] 80. Write reversible compression integration tests

## Phase 6: TOON Columnar Encoding (Tasks 81-90)
*Inspired by: kompact, tamp*

- [ ] 81. Create `internal/filter/toon.go` - Columnar encoder
- [ ] 82. Implement JSON array type detection
- [ ] 83. Implement column extraction from homogeneous arrays
- [ ] 84. Implement columnar format generation
- [ ] 85. Add columnar format decoder
- [ ] 86. Implement JSON metadata pruning (npm URLs, integrity hashes)
- [ ] 87. Implement line number stripping from tool output
- [ ] 88. Add columnar compression statistics
- [ ] 89. Create `tokman toon <file>` command for testing
- [ ] 90. Write TOON encoding benchmark tests

## Phase 7: LLM-Based Compression (Tasks 91-100)
*Inspired by: claw-compactor (Nexus), tamp (textpress), clawshield*

- [ ] 91. Create `internal/filter/llm_compress.go` - LLM compression layer
- [ ] 92. Implement Ollama integration for local compression
- [ ] 93. Implement OpenRouter free model integration
- [ ] 94. Implement semantic text compression via LLM
- [ ] 95. Add compression quality validation
- [ ] 96. Implement fallback to rule-based compression
- [ ] 97. Add LLM compression cost tracking
- [ ] 98. Create `tokman llm-compress <file>` command
- [ ] 99. Add LLM compression configuration
- [ ] 100. Write LLM compression quality benchmarks

## Phase 8: Adaptive Context Scaling (Tasks 101-110)
*Inspired by: kompact, lean-ctx*

- [ ] 101. Create `internal/filter/adaptive.go` - Adaptive scaling engine
- [ ] 102. Implement context size detection (short/medium/long/very-long)
- [ ] 103. Implement auto-adjusted compression aggressiveness
- [ ] 104. Add per-model compression profiles
- [ ] 105. Implement budget-aware adaptive scaling
- [ ] 106. Add quality-preserving compression limits
- [ ] 107. Implement feedback-based threshold adjustment
- [ ] 108. Create `tokman adaptive-config` command
- [ ] 109. Add adaptive scaling statistics
- [ ] 110. Write adaptive scaling benchmark tests

## Phase 9: LITM-Aware Positioning (Tasks 111-120)
*Inspired by: lean-ctx*

- [ ] 111. Create `internal/filter/positioning.go` - Attention-optimal ordering
- [ ] 112. Implement primacy/recency model for Claude
- [ ] 113. Implement primacy/recency model for GPT
- [ ] 114. Implement primacy/recency model for Gemini
- [ ] 115. Add attention prediction model (U-curve weighting)
- [ ] 116. Implement structural importance scoring
- [ ] 117. Add content reordering for attention optimization
- [ ] 118. Create per-model positioning profiles
- [ ] 119. Add positioning statistics tracking
- [ ] 120. Write LITM-aware positioning benchmark tests

## Phase 10: Token Dense Dialect (Tasks 121-130)
*Inspired by: lean-ctx (TDD)*

- [ ] 121. Create `internal/filter/tdd.go` - Token Dense Dialect
- [ ] 122. Implement symbol shorthand (λ, ∂, ∫, τ, ε, etc.)
- [ ] 123. Implement ROI-based identifier mapping
- [ ] 124. Add symbol dictionary for common programming terms
- [ ] 125. Implement TDD encoder
- [ ] 126. Implement TDD decoder (for reversibility)
- [ ] 127. Add TDD compression statistics
- [ ] 128. Create `tokman tdd <file>` command for testing
- [ ] 129. Add TDD configuration (enable/disable per language)
- [ ] 130. Write TDD benchmark tests

## Phase 11: Cross-Session Memory (Tasks 131-145)
*Inspired by: lean-ctx (CCP), claw-compactor (Engram)*

- [ ] 131. Create `internal/memory/memory.go` - Cross-session memory
- [ ] 132. Implement task persistence across sessions
- [ ] 133. Implement findings persistence
- [ ] 134. Implement decision history tracking
- [ ] 135. Create knowledge store with query support
- [ ] 136. Implement multi-agent context sharing
- [ ] 137. Add agent scratchpad for inter-agent communication
- [ ] 138. Implement Engram Observer (LLM-driven memory compression)
- [ ] 139. Implement Engram Reflector (memory consolidation)
- [ ] 140. Add tiered summaries (L0/L1/L2)
- [ ] 141. Implement observation/reflection daemon mode
- [ ] 142. Add memory retrieval by query/category
- [ ] 143. Create `tokman memory` command suite
- [ ] 144. Add memory statistics and management
- [ ] 145. Write cross-session memory integration tests

## Phase 12: Project Intelligence Graph (Tasks 146-155)
*Inspired by: lean-ctx (ctx_graph)*

- [ ] 146. Create `internal/graph/graph.go` - Project intelligence graph
- [ ] 147. Implement dependency analysis
- [ ] 148. Implement related file discovery
- [ ] 149. Implement impact analysis for file changes
- [ ] 150. Add semantic intent detection for queries
- [ ] 151. Implement auto-loading of relevant files
- [ ] 152. Add adaptive mode selection for file reads
- [ ] 153. Implement incremental deltas (Myers diff for changes)
- [ ] 154. Create `tokman graph` command suite
- [ ] 155. Write project graph integration tests

## Phase 13: TOML Filter DSL Enhancement (Tasks 156-170)
*Inspired by: tokf*

- [ ] 156. Enhance TOML filter DSL with template pipes
- [ ] 157. Implement `join` pipe operation
- [ ] 158. Implement `each` pipe operation
- [ ] 159. Implement `truncate` pipe operation
- [ ] 160. Implement `lines` pipe operation
- [ ] 161. Implement `where` filter operation
- [ ] 162. Add JSONPath extraction (RFC 9535)
- [ ] 163. Implement filter variants (file-based + output-pattern)
- [ ] 164. Add Lua script escape hatch (sandboxed)
- [ ] 165. Implement passthrough args (skip filter on conflicting flags)
- [ ] 166. Add color passthrough (strip for match, restore in output)
- [ ] 167. Implement prefer-less mode (compare filtered vs piped)
- [ ] 168. Add task runner wrapping (filter each make/just recipe line)
- [ ] 169. Create community filter registry (publish/share filters)
- [ ] 170. Write TOML DSL test suite with safety checks

## Phase 14: Analytics & Tracking (Tasks 171-185)
*Inspired by: tokscale, token-lens*

- [ ] 171. Add multi-platform token tracking (16 AI clients)
- [ ] 172. Implement real-time pricing (LiteLLM + OpenRouter)
- [ ] 173. Add cost anomaly detection (rolling mean + 2sigma)
- [ ] 174. Implement spend forecasting
- [ ] 175. Add model right-sizing recommendations
- [ ] 176. Implement history bloat tracking
- [ ] 177. Add token heatmap (system/tools/context/history/query)
- [ ] 178. Implement wrapped year-in-review reports
- [ ] 179. Add leaderboard/social features
- [ ] 180. Create TUI dashboard (Ratatui-style in Go)
- [ ] 181. Add 3D contribution graph for token usage
- [ ] 182. Implement GitHub profile embed widget
- [ ] 183. Add webhook notifications for events
- [ ] 184. Implement CSV export for analytics
- [ ] 185. Add OpenTelemetry export

## Phase 15: Security (Tasks 186-200)
*Inspired by: clawshield, token-lens, tokf*

- [ ] 186. Create `internal/security/security.go` - Security engine
- [ ] 187. Implement prompt injection detection
- [ ] 188. Implement PII detection and redaction
- [ ] 189. Implement secrets detection (API keys, tokens)
- [ ] 190. Add vulnerability scanning (SQLi, SSRF, XSS)
- [ ] 191. Implement filter safety checks (tokf-style)
- [ ] 192. Add streaming response scanning
- [ ] 193. Implement policy hot-reload
- [ ] 194. Add kill switches and quotas
- [ ] 195. Implement decision explainability (audit records)
- [ ] 196. Add SIEM integration (OCSF format)
- [ ] 197. Implement URL/web fetch compression
- [ ] 198. Add entity protection engine for prompt refinement
- [ ] 199. Create `tokman security` command suite
- [ ] 200. Write security integration tests

---

## Priority Tiers

### Tier 1: Immediate Impact (Tasks 1-50)
HTTP Proxy Mode, KV-Cache Alignment, Cross-Message Dedup, Content-Type Detection

### Tier 2: Strong Competitive Advantage (Tasks 51-110)
Reversible Compression, TOON Encoding, LLM Compression, Adaptive Scaling

### Tier 3: Differentiation (Tasks 111-160)
LITM Positioning, Token Dense Dialect, Cross-Session Memory, TOML DSL Enhancement

### Tier 4: Expansion (Tasks 161-200)
Analytics, Security, Project Graph, Community Features

---

## Estimated Timeline

| Phase | Tasks | Estimated Time |
|-------|-------|----------------|
| Phase 1: HTTP Proxy | 1-25 | 2 weeks |
| Phase 2: KV-Cache | 26-35 | 1 week |
| Phase 3: Cross-Message Dedup | 36-50 | 1.5 weeks |
| Phase 4: Content Detection | 51-65 | 1.5 weeks |
| Phase 5: Reversible | 66-80 | 1.5 weeks |
| Phase 6: TOON Encoding | 81-90 | 1 week |
| Phase 7: LLM Compression | 91-100 | 1 week |
| Phase 8: Adaptive Scaling | 101-110 | 1 week |
| Phase 9: LITM Positioning | 111-120 | 1 week |
| Phase 10: Token Dense | 121-130 | 1 week |
| Phase 11: Cross-Session Memory | 131-145 | 2 weeks |
| Phase 12: Project Graph | 146-155 | 1.5 weeks |
| Phase 13: TOML DSL | 156-170 | 2 weeks |
| Phase 14: Analytics | 171-185 | 2 weeks |
| Phase 15: Security | 186-200 | 2 weeks |
| **Total** | **200 tasks** | **~21 weeks** |
