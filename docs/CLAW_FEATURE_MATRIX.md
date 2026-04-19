# Tok vs Claw Compactor: Feature Matrix

## Stage/Layer Mapping

| # | Claw Stage (Order) | Tok Layer | Status | Notes |
|---|-------------------|--------------|--------|-------|
| 1 | **QuantumLock (3)** | ❌ None | 🔥 IMPLEMENT | KV-cache alignment - HIGH IMPACT |
| 2 | **Cortex (5)** | ⚠️ Distributed | 🔵 Consider | Content detection - cleaner but not urgent |
| 3 | **Photon (8)** | ❌ None | 🔥 IMPLEMENT | Image compression - HIGH IMPACT |
| 4 | **RLE (10)** | ✅ Layer 6 (N-gram) | ✅ Better | Tok's n-gram is more sophisticated |
| 5 | **SemanticDedup (12)** | ✅ dedup.go | 🔥 UPGRADE | Add cross-message support |
| 6 | **Ionizer (15)** | ✅ json_sampler.go | ✅ Similar | JSON compression exists |
| 7 | **LogCrunch (16)** | ✅ log_crunch.go | 🟡 UPGRADE | Add stack traces, timestamps |
| 8 | **SearchCrunch (17)** | ✅ search_crunch.go | 🟡 UPGRADE | Add structured parsing |
| 9 | **DiffCrunch (18)** | ✅ diff_crunch.go | 🟡 UPGRADE | Add hunk parsing |
| 10 | **StructuralCollapse (20)** | ✅ structural_collapse.go | ✅ Exists | Similar implementation |
| 11 | **Neurosyntax (25)** | ✅ Layer 4 (AST) | ✅ Similar | Both use AST-aware compression |
| 12 | **Nexus (35)** | ✅ Layers 7,8,9 | ✅ Better | Tok has multiple ML layers |
| 13 | **TokenOpt (40)** | ✅ TOML filters | ✅ Better | Handled declaratively |
| 14 | **Abbrev (45)** | ✅ Layer 11 | ✅ Better | Semantic compression |

---

## Tok-Exclusive Layers (Not in Claw)

| Layer | Name | Research Paper | Status |
|-------|------|----------------|--------|
| 1 | Entropy Filtering | Selective Context (Mila 2023) | ✅ Production |
| 2 | Perplexity Pruning | LLMLingua (Microsoft 2023) | ✅ Production |
| 3 | Goal-Driven Selection | SWE-Pruner (Shanghai Jiao Tong 2025) | ✅ Production |
| 5 | Contrastive Ranking | LongLLMLingua (Microsoft 2024) | ✅ Production |
| 7 | Evaluator Heads | EHPC (Tsinghua/Huawei 2025) | ✅ Production |
| 8 | Gist Compression | Stanford/Berkeley (2023) | ✅ Production |
| 9 | Hierarchical Summary | AutoCompressor (Princeton/MIT 2023) | ✅ Production |
| 12 | Attribution Filter | ProCut (LinkedIn 2025) | ✅ Production |
| 13 | H2O Filter | Heavy-Hitter Oracle (NeurIPS 2023) | ✅ Production |
| 14 | Attention Sink | StreamingLLM (2023) | ✅ Production |
| 15 | Meta-Token | arXiv:2506.00307 (2025) | ✅ Production |
| 16 | Semantic Chunk | ChunkKV-style | ✅ Production |
| 17 | Sketch Store | KVReviver-style | ✅ Production |
| 18 | Lazy Pruner | LazyLLM (July 2024) | ✅ Production |
| 19 | Semantic Anchor | Attention Gradient | ✅ Production |
| 20 | Agent Memory | Focus-inspired | ✅ Production |

**Verdict:** Tok has 16 unique layers that Claw doesn't have.

---

## Claw-Exclusive Features (Not in Tok)

| Feature | Impact | Complexity | Recommendation |
|---------|--------|------------|----------------|
| **Cross-Message Dedup** | 🔥 CRITICAL | Medium | IMPLEMENT |
| **QuantumLock** | 🔥 HIGH | Low | IMPLEMENT |
| **Photon** | 🔥 HIGH | Medium | IMPLEMENT |
| **Cortex** | 🟡 MEDIUM | Medium | Consider |
| **Immutable Architecture** | 🟡 MEDIUM | High | Long-term |

**Verdict:** 3 critical features to implement, 2 to consider.

---

## Feature Comparison by Category

### Content Detection
| Feature | Tok | Claw | Winner |
|---------|--------|------|--------|
| Content-type detection | Distributed | Centralized (Cortex) | Claw (cleaner) |
| Language detection | Implicit | Explicit (16 langs) | Claw |
| Overhead | Zero | ~5ms | Tok |

### Deduplication
| Feature | Tok | Claw | Winner |
|---------|--------|------|--------|
| Within-message | SimHash | 3-word shingles | Similar |
| Cross-message | ❌ None | ✅ Yes | **Claw** |
| Threshold | Hamming distance | Jaccard > 0.8 | Similar |

### Image Handling
| Feature | Tok | Claw | Winner |
|---------|--------|------|--------|
| Base64 detection | ❌ None | ✅ Regex | **Claw** |
| Image resize | ❌ None | ✅ Pillow | **Claw** |
| Format conversion | ❌ None | ✅ PNG→JPEG | **Claw** |
| Vision token optimization | ❌ None | ✅ detail:low | **Claw** |

### Cache Optimization
| Feature | Tok | Claw | Winner |
|---------|--------|------|--------|
| KV-cache alignment | ❌ None | ✅ QuantumLock | **Claw** |
| Dynamic content detection | ❌ None | ✅ 6 patterns | **Claw** |
| Prefix stabilization | ❌ None | ✅ Yes | **Claw** |

### Log Compression
| Feature | Tok | Claw | Winner |
|---------|--------|------|--------|
| Error preservation | ✅ Yes | ✅ Yes | Tie |
| Stack trace detection | ❌ None | ✅ Yes | **Claw** |
| Timestamp normalization | ❌ None | ✅ Yes | **Claw** |
| Occurrence counts | ❌ None | ✅ Yes | **Claw** |

### Diff Compression
| Feature | Tok | Claw | Winner |
|---------|--------|------|--------|
| Context folding | ✅ Simple | ✅ Sophisticated | **Claw** |
| Hunk parsing | ❌ None | ✅ Yes | **Claw** |
| Change preservation | ✅ Yes | ✅ Yes | Tie |

### Search Compression
| Feature | Tok | Claw | Winner |
|---------|--------|------|--------|
| Line deduplication | ✅ Yes | ✅ Yes | Tie |
| Structured parsing | ❌ None | ✅ Yes | **Claw** |
| Snippet dedup | ❌ None | ✅ SimHash | **Claw** |

### Architecture
| Feature | Tok | Claw | Winner |
|---------|--------|------|--------|
| Mutability | Mutable | Immutable | **Claw** (testability) |
| Performance | Fast (Go) | Slower (Python) | **Tok** |
| Memory | Low | Higher | **Tok** |
| Testability | Good | Excellent | **Claw** |

### Integration
| Feature | Tok | Claw | Winner |
|---------|--------|------|--------|
| CLI interception | ✅ Yes | ❌ None | **Tok** |
| Library API | ⚠️ Limited | ✅ Full | **Claw** |
| Agent support | ✅ 7+ agents | ✅ OpenClaw | **Tok** |
| Dashboard | ✅ Built-in | ❌ None | **Tok** |

---

## Overall Score

### Tok Strengths (10/14)
- ✅ More layers (20 vs 14)
- ✅ Broader research coverage (120+ papers)
- ✅ CLI-first design
- ✅ Faster (Go vs Python)
- ✅ Built-in dashboard
- ✅ TOML filters (97+)
- ✅ Agent integrations (7+)
- ✅ Quality metrics (6-metric grading)
- ✅ SIMD optimization
- ✅ Hook integrity verification

### Claw Strengths (4/14)
- ✅ Cross-message deduplication
- ✅ KV-cache alignment
- ✅ Image compression
- ✅ Immutable architecture

### Verdict
**Tok wins overall (10-4), but Claw has 3 critical features Tok needs.**

---

## Recommended Actions

### Immediate (Week 1)
1. ✅ Implement Cross-Message Dedup
2. ✅ Implement QuantumLock
3. ✅ Implement Photon

### Short-term (Week 2)
4. ✅ Upgrade LogCrunch
5. ✅ Upgrade DiffCrunch
6. ✅ Upgrade SearchCrunch

### Long-term (v2.0+)
7. 🔵 Consider Cortex refactor
8. 🔵 Consider immutable architecture

---

## Impact Summary

### Current State
- Tok: 60-90% compression on CLI
- Multi-turn: ~70%
- Vision: Not supported
- Cache: Not optimized

### After Implementation
- Tok: 60-90% (maintained)
- Multi-turn: **80-85%** (+10-15%)
- Vision: **50-70%** (NEW)
- Cache: **70-90%** hit rate (NEW)

### ROI
- **Effort:** 2 weeks
- **Impact:** 20-40% better compression
- **Risk:** Low (additive features)
- **Value:** HIGH

---

**Conclusion:** Implement the 3 critical Claw features, upgrade 3 existing features, maintain Tok's unique advantages.
