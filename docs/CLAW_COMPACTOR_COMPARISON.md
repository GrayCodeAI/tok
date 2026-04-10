# TokMan vs Claw Compactor: Comprehensive Technical Comparison

**Date:** April 2026  
**TokMan Version:** Current (20-layer pipeline)  
**Claw Compactor Version:** v7.1.0 (14-stage Fusion Pipeline)

---

## Executive Summary

Both TokMan and Claw Compactor are research-backed LLM token compression engines designed to reduce context window usage for AI coding assistants. While they share similar goals and some overlapping techniques, they differ significantly in architecture, implementation language, and design philosophy.

**Key Verdict:**
- **TokMan**: Go-based CLI proxy with 20 practical layers, CLI-first design, production-ready for terminal workflows
- **Claw Compactor**: Python-based library with 14 fusion stages, API-first design, optimized for agent integration

---

## Architecture Comparison

### Core Design Philosophy

| Aspect | TokMan | Claw Compactor |
|--------|--------|----------------|
| **Language** | Go 1.26+ | Python 3.9+ |
| **Architecture** | Mutable pipeline with stage gates | Immutable data flow (frozen dataclasses) |
| **Primary Use Case** | CLI command interception | Library/API integration |
| **Data Model** | Mutable string processing | Immutable `FusionContext` → `FusionResult` |
| **Stage Communication** | Direct mutation + stats tracking | Context passing (no shared state) |
| **Execution Model** | Sequential with early-exit | Sequential with immutable threading |

### Pipeline Structure

#### TokMan (20 Layers)
```
Input → TOML Pre-filter → Core (1-9) → Semantic (11-20) → Research (21-25) → Budget → Output
         ↓                  ↓              ↓                  ↓
    Declarative      Entropy, AST,    Compaction,      Experimental
    Filters          Perplexity       H2O, Attention   Layers
```

**Layer Organization:**
- **Pre-filters:** TOML declarative filters, adaptive routing
- **Core (1-9):** Entropy, Perplexity, Goal-Driven, AST, Contrastive, N-gram, Evaluator Heads, Gist, Hierarchical
- **Semantic (11-20):** Compaction, Attribution, H2O, Attention Sink, Meta-Token, Semantic Chunk, Sketch Store, Lazy Pruner, Semantic Anchor, Agent Memory
- **Research (21-25):** Experimental layers (not in production by default)

#### Claw Compactor (14 Stages)
```
Input → Cross-msg Dedup → Fusion Pipeline (order 3-45) → Output + RewindStore
                           ↓
        QuantumLock(3) → Cortex(5) → Photon(8) → RLE(10) → SemanticDedup(12)
        → Ionizer(15) → LogCrunch(16) → SearchCrunch(17) → DiffCrunch(18)
        → StructuralCollapse(20) → Neurosyntax(25) → Nexus(35) → TokenOpt(40) → Abbrev(45)
```

**Stage Organization:**
- **Order 3-10:** Pre-processing (KV-cache, detection, image, path compression)
- **Order 12-20:** Content-specific compression (JSON, logs, search, diffs, code structure)
- **Order 25-45:** Deep compression (AST, ML tokens, format optimization, abbreviation)

---

## Layer/Stage Mapping

### Overlapping Techniques

| TokMan Layer | Claw Stage | Technique | Notes |
|--------------|------------|-----------|-------|
| **Layer 4: AST Preservation** | **Neurosyntax (25)** | AST-aware code compression | TokMan: syntax preservation; Claw: tree-sitter with identifier shortening |
| **Layer 13: H2O Filter** | *(Not present)* | Heavy-Hitter Oracle (30x compression) | TokMan-specific |
| **Layer 14: Attention Sink** | *(Not present)* | StreamingLLM stability | TokMan-specific |
| **Layer 11: Compaction** | **Nexus (35)** | Semantic compression | TokMan: MemGPT-style; Claw: ML token-level |
| **Layer 6: N-gram** | **RLE (10)** | Pattern compression | TokMan: lossless n-gram; Claw: path/IP/enum |
| **Layer 17: Sketch Store** | **Ionizer (15)** | Reversible compression | TokMan: KVReviver-style; Claw: JSON sampling + RewindStore |
| *(Not present)* | **QuantumLock (3)** | KV-cache alignment | Claw-specific |
| *(Not present)* | **Cortex (5)** | Content-type detection | Claw-specific (TokMan uses inline detection) |
| *(Not present)* | **Photon (8)** | Image/base64 compression | Claw-specific |
| *(Not present)* | **LogCrunch (16)** | Log folding | Claw-specific (TokMan has general log handling) |
| *(Not present)* | **SearchCrunch (17)** | Search result dedup | Claw-specific |
| *(Not present)* | **DiffCrunch (18)** | Diff context folding | Claw-specific |

### Unique to TokMan

| Layer | Technique | Research Paper | Purpose |
|-------|-----------|----------------|---------|
| **Layer 1: Entropy** | Selective Context | Mila 2023 | Remove low-information tokens |
| **Layer 2: Perplexity** | LLMLingua | Microsoft 2023 | Iterative token removal |
| **Layer 3: Goal-Driven** | SWE-Pruner | Shanghai Jiao Tong 2025 | CRF-style line scoring |
| **Layer 5: Contrastive** | LongLLMLingua | Microsoft 2024 | Question-relevance scoring |
| **Layer 7: Evaluator Heads** | EHPC | Tsinghua/Huawei 2025 | Early-layer attention simulation |
| **Layer 9: Hierarchical** | AutoCompressor | Princeton/MIT 2023 | Recursive summarization |
| **Layer 12: Attribution** | ProCut | LinkedIn 2025 | 78% pruning |
| **Layer 15: Meta-Token** | arXiv:2506.00307 | 2025 | 27% lossless compression |
| **Layer 16: Semantic Chunk** | ChunkKV-style | Context-aware | Context-aware boundaries |
| **Layer 18: Lazy Pruner** | LazyLLM | July 2024 | 2.34x speedup |
| **Layer 19: Semantic Anchor** | Attention Gradient | Custom | Context preservation |
| **Layer 20: Agent Memory** | Focus-inspired | Custom | Knowledge graph extraction |

### Unique to Claw Compactor

| Stage | Technique | Purpose |
|-------|-----------|---------|
| **QuantumLock (3)** | KV-cache alignment | Maximize cache hit rate by isolating dynamic content |
| **Cortex (5)** | Content-type detection | Auto-detect 16 languages + content types |
| **Photon (8)** | Image compression | Re-encode images, compress base64 blobs |
| **RLE (10)** | Path/IP/enum compression | Structural pattern compression |
| **SemanticDedup (12)** | SimHash deduplication | Within-message near-duplicate removal |
| **Ionizer (15)** | JSON array sampling | Statistical sampling with schema discovery |
| **LogCrunch (16)** | Log folding | Fold repetitive log lines with counts |
| **SearchCrunch (17)** | Search result dedup | Merge near-duplicate search results |
| **DiffCrunch (18)** | Diff folding | Fold unchanged context in git diffs |
| **StructuralCollapse (20)** | Import/assertion collapse | Collapse repetitive code patterns |
| **TokenOpt (40)** | Format optimization | Remove markdown decorators, normalize whitespace |
| **Abbrev (45)** | Natural language abbreviation | Text-only abbreviation (never touches code) |

---

## Compression Performance

### Benchmark Comparison

#### TokMan (30-minute Claude Code session)
| Command | Uses | Before | After | Savings |
|---------|------|--------|-------|---------|
| `ls`/`tree` | 10× | 2,000 | 400 | **80%** |
| `cat`/`read` | 20× | 40,000 | 12,000 | **70%** |
| `grep`/`rg` | 8× | 16,000 | 3,200 | **80%** |
| `git status` | 10× | 3,000 | 600 | **80%** |
| `git diff` | 5× | 10,000 | 2,500 | **75%** |
| `npm test` | 5× | 25,000 | 2,500 | **90%** |
| **Total** | | **~118,000** | **~23,500** | **80%** |

#### Claw Compactor (Real-World Content)
| Content Type | Legacy Regex | FusionEngine | Improvement |
|--------------|--------------|--------------|-------------|
| Python source | 7.3% | **25.0%** | 3.4x |
| JSON (100 items) | 12.6% | **81.9%** | 6.5x |
| Build logs | 5.5% | **24.1%** | 4.4x |
| Agent conversation | 5.7% | **31.0%** | 5.4x |
| Git diff | 6.2% | **15.0%** | 2.4x |
| Search results | 5.3% | **40.7%** | 7.7x |
| **Weighted average** | **9.2%** | **36.3%** | **3.9x** |

### Analysis

**TokMan Strengths:**
- Higher compression on CLI commands (80% average)
- Optimized for terminal output patterns
- Better performance on test output (90%)

**Claw Compactor Strengths:**
- Exceptional JSON compression (81.9%)
- Better on search results (40.7%)
- More consistent across content types

**Verdict:** TokMan excels at CLI workflows; Claw Compactor excels at structured data and agent conversations.

---

## Reversible Compression

### TokMan: Sketch Store (Layer 17)
```go
// KVReviver-style semantic caching
type SketchStoreFilter struct {
    store map[string]string  // hash → original content
}

// Stores compressed content with hash marker
func (s *SketchStoreFilter) Apply(input string, mode Mode) (string, int) {
    hash := computeHash(input)
    s.store[hash] = input
    return fmt.Sprintf("[SKETCH:%s]", hash), len(input)
}
```

**Features:**
- Semantic similarity-based retrieval
- Budget-aware (only fires when budget set)
- Integrated with LazyPruner (Layer 18)

### Claw Compactor: RewindStore
```python
# Hash-addressed immutable store
class RewindStore:
    def __init__(self):
        self._store: dict[str, str] = {}  # SHA-256 → content
    
    def put(self, content: str) -> str:
        hash_id = hashlib.sha256(content.encode()).hexdigest()[:16]
        self._store[hash_id] = content
        return hash_id
    
    def retrieve(self, hash_id: str) -> str | None:
        return self._store.get(hash_id)
```

**Features:**
- Used by Ionizer (JSON), LogCrunch, SearchCrunch, Photon
- LLM tool call: `rewind_retrieve(hash)`
- Append-only, immutable

**Comparison:**

| Feature | TokMan Sketch Store | Claw RewindStore |
|---------|---------------------|------------------|
| **Storage** | In-memory map | In-memory map (optional persistent) |
| **Key** | Custom hash | SHA-256 (16 chars) |
| **Retrieval** | Semantic similarity | Exact hash match |
| **Integration** | Single layer (17) | Multiple stages (6, 7, 8, 17) |
| **LLM Access** | Manual | Tool call (`rewind_retrieve`) |

**Verdict:** Claw's RewindStore is more mature and better integrated across stages. TokMan's Sketch Store is more experimental.

---

## Content-Type Detection

### TokMan: Inline Detection
- No dedicated layer
- Detection happens within individual layers (e.g., AST layer detects code)
- TOML filters provide declarative content matching

### Claw Compactor: Cortex Stage (Order 5)
```python
class Cortex(FusionStage):
    order = 5
    name = "cortex"
    
    def apply(self, ctx: FusionContext) -> FusionResult:
        content_type = self._detect_type(ctx.content)  # code/json/log/diff/search/text
        language = self._detect_language(ctx.content)  # 16 languages
        
        return FusionResult(
            content=ctx.content,
            metadata={"content_type": content_type, "language": language}
        )
```

**Detected Types:** `code`, `json`, `log`, `diff`, `search`, `text`  
**Detected Languages:** Python, JavaScript, TypeScript, Java, C, C++, C#, Go, Rust, Ruby, PHP, Swift, Kotlin, Scala, Shell, SQL

**Comparison:**

| Aspect | TokMan | Claw Compactor |
|--------|--------|----------------|
| **Approach** | Distributed (per-layer) | Centralized (Cortex stage) |
| **Languages** | Implicit (AST layer) | Explicit (16 languages) |
| **Content Types** | TOML patterns | 6 types (code/json/log/diff/search/text) |
| **Overhead** | Zero (no dedicated layer) | ~5ms (one-time cost) |

**Verdict:** Claw's centralized detection is cleaner and more maintainable. TokMan's distributed approach is more flexible but harder to debug.

---

## Stage Gating (Early Exit)

### TokMan: Stage Gates
```go
func (p *PipelineCoordinator) shouldSkipEntropy(output string) bool {
    return len(output) < 50  // Skip if too short
}

func (p *PipelineCoordinator) shouldSkipQueryDependent() bool {
    return p.config.QueryIntent == ""  // Skip if no query
}

func (p *PipelineCoordinator) shouldEarlyExit(stats *PipelineStats) bool {
    if p.config.Budget > 0 {
        currentTokens := core.EstimateTokens(output)
        return currentTokens <= p.config.Budget
    }
    return false
}
```

**Gate Types:**
- Content-length gates (entropy, H2O, attention sink)
- Query-dependent gates (goal-driven, contrastive)
- Budget-dependent gates (sketch store, lazy pruner)
- Early exit when budget met

### Claw Compactor: should_apply()
```python
class MyStage(FusionStage):
    def should_apply(self, ctx: FusionContext) -> bool:
        # Return False immediately if irrelevant
        return ctx.content_type == "code" and len(ctx.content) > 100
    
    def apply(self, ctx: FusionContext) -> FusionResult:
        # Only called if should_apply() returned True
        ...
```

**Gate Types:**
- Content-type gates (all stages check `ctx.content_type`)
- Length gates (minimum thresholds)
- Language gates (Neurosyntax checks supported languages)
- No early exit (all stages run if applicable)

**Comparison:**

| Feature | TokMan | Claw Compactor |
|---------|--------|----------------|
| **Gate Location** | Coordinator method | Stage method |
| **Early Exit** | Yes (budget-based) | No (all stages run) |
| **Gate Cost** | O(1) checks | O(1) checks |
| **Flexibility** | Centralized logic | Distributed logic |

**Verdict:** TokMan's early-exit is more efficient for budget-constrained scenarios. Claw's per-stage gates are cleaner architecturally.

---

## Immutability

### TokMan: Mutable Pipeline
```go
func (p *PipelineCoordinator) Process(input string) (string, *PipelineStats) {
    output := input  // Mutable string
    
    output = p.processLayer(layer1, output, stats)  // Mutates output
    output = p.processLayer(layer2, output, stats)  // Mutates output
    
    return output, stats
}
```

**Characteristics:**
- Mutable string passed through pipeline
- Stats accumulated in mutable `PipelineStats` struct
- Efficient (no copying)
- Harder to debug (state changes in-place)

### Claw Compactor: Immutable Data Flow
```python
@dataclass(frozen=True)
class FusionContext:
    content: str
    content_type: str
    language: str | None
    metadata: dict  # Frozen copy-on-write

@dataclass(frozen=True)
class FusionResult:
    content: str
    original_tokens: int
    compressed_tokens: int
    markers: list[str]
```

**Characteristics:**
- Frozen dataclasses (immutable)
- Each stage returns new `FusionResult`
- Next stage receives new `FusionContext` derived from previous result
- Easier to debug (no side effects)
- Slightly higher memory overhead

**Comparison:**

| Aspect | TokMan | Claw Compactor |
|--------|--------|----------------|
| **Mutability** | Mutable | Immutable |
| **Memory** | Efficient (in-place) | Higher (copying) |
| **Debugging** | Harder (state changes) | Easier (no side effects) |
| **Testability** | Requires mocking | Easy (pure functions) |
| **Concurrency** | Not thread-safe | Thread-safe |

**Verdict:** Claw's immutability is better for correctness and testing. TokMan's mutability is more performant.

---

## Language & Performance

### TokMan (Go)
```go
// Compiled binary, native performance
// SIMD auto-vectorization (Go 1.26+)
// Concurrent goroutines for parallel processing

func (p *PipelineCoordinator) Process(input string) (string, *PipelineStats) {
    // Sub-millisecond processing for most commands
    // <20ms overhead typical
}
```

**Performance:**
- **Latency:** <20ms for typical CLI commands
- **Throughput:** Handles 500K+ token inputs with streaming
- **Binary Size:** ~5MB (with `make build-tiny`)
- **Memory:** Low overhead (Go GC)
- **SIMD:** AVX2, AVX-512, ARM NEON support (Go 1.26+)

### Claw Compactor (Python)
```python
# Interpreted, slower than Go
# No SIMD (unless using NumPy/tree-sitter)
# GIL limits concurrency

def compress(self, text: str, ...) -> dict[str, Any]:
    # 10-80ms for typical inputs
    # Scales with input length
```

**Performance:**
- **Latency:** 10-80ms for 8K-128K token inputs
- **Throughput:** Handles large inputs but slower than Go
- **Dependencies:** Zero required (tiktoken/tree-sitter optional)
- **Memory:** Higher overhead (Python runtime)
- **SIMD:** Limited (tree-sitter uses C bindings)

**Comparison:**

| Metric | TokMan (Go) | Claw Compactor (Python) |
|--------|-------------|-------------------------|
| **Latency** | <20ms | 10-80ms |
| **Binary Size** | ~5MB | N/A (interpreted) |
| **Memory** | Low | Medium-High |
| **SIMD** | Yes (Go 1.26+) | Limited |
| **Concurrency** | Goroutines | GIL-limited |
| **Startup Time** | Instant | ~100ms (Python import) |

**Verdict:** TokMan is significantly faster for CLI use cases. Claw Compactor is acceptable for API/library use.

---

## Integration & Deployment

### TokMan: CLI Proxy
```bash
# Install as CLI tool
brew install tokman

# Intercept commands
tokman init -g  # Claude Code
tokman init --cursor  # Cursor

# Commands automatically compressed
git status  # Intercepted by TokMan
docker ps   # Intercepted by TokMan
```

**Deployment:**
- Standalone binary (Linux, macOS, Windows)
- Hook-based interception (shell integration)
- Agent-specific installers (Claude, Cursor, Copilot, etc.)
- Dashboard: `tokman dashboard` (web UI)

### Claw Compactor: Python Library
```python
# Install as library
pip install claw-compactor

# Use in code
from claw_compactor.fusion.engine import FusionEngine

engine = FusionEngine()
result = engine.compress(text="...", content_type="code")
print(result["compressed"])
```

**Deployment:**
- PyPI package
- API integration (OpenAI-compatible proxy)
- Used by OpenClaw agents
- No CLI interception (library-first)

**Comparison:**

| Aspect | TokMan | Claw Compactor |
|--------|--------|----------------|
| **Primary Use** | CLI interception | Library/API |
| **Installation** | Binary (Homebrew) | pip install |
| **Integration** | Shell hooks | Python import |
| **Agent Support** | 7+ agents | OpenClaw agents |
| **Dashboard** | Built-in web UI | None |

**Verdict:** TokMan is better for CLI workflows. Claw Compactor is better for programmatic integration.

---

## Testing & Quality

### TokMan
- **Tests:** 144 packages with tests
- **Coverage:** Improving (not specified)
- **Benchmarks:** `make benchmark` suite
- **Fuzz Testing:** `filter/fuzz_test.go`
- **Quality Metrics:** 6-metric grading (A+ to F)

### Claw Compactor
- **Tests:** 1,600+ tests
- **Coverage:** Not specified
- **Benchmarks:** Real-world SWE-bench tasks
- **ROUGE-L:** 0.653 @ 0.3 compression, 0.723 @ 0.5
- **Quality Metrics:** ROUGE-L fidelity

**Comparison:**

| Metric | TokMan | Claw Compactor |
|--------|--------|----------------|
| **Test Count** | 144 packages | 1,600+ tests |
| **Quality Metrics** | 6-metric grading | ROUGE-L |
| **Benchmarks** | CLI commands | SWE-bench tasks |
| **Fuzz Testing** | Yes | Not mentioned |

**Verdict:** Claw Compactor has more comprehensive testing. TokMan has better quality metrics.

---

## Research Foundation

### TokMan: 120+ Papers
- **Entropy:** Selective Context (Mila 2023)
- **Perplexity:** LLMLingua (Microsoft 2023)
- **Goal-Driven:** SWE-Pruner (Shanghai Jiao Tong 2025)
- **AST:** LongCodeZip (NUS 2025)
- **Contrastive:** LongLLMLingua (Microsoft 2024)
- **N-gram:** CompactPrompt (2025)
- **Evaluator Heads:** EHPC (Tsinghua/Huawei 2025)
- **Gist:** Stanford/Berkeley (2023)
- **Hierarchical:** AutoCompressor (Princeton/MIT 2023)
- **Compaction:** MemGPT (UC Berkeley 2023)
- **Attribution:** ProCut (LinkedIn 2025)
- **H2O:** Heavy-Hitter Oracle (NeurIPS 2023)
- **Attention Sink:** StreamingLLM (2023)
- **Meta-Token:** arXiv:2506.00307 (2025)
- **Lazy Pruner:** LazyLLM (July 2024)

### Claw Compactor: 30+ Papers
- **Neurosyntax:** AST-aware compression (tree-sitter)
- **Ionizer:** JSON statistical sampling
- **SemanticDedup:** SimHash deduplication
- **Nexus:** ML token-level compression
- **QuantumLock:** KV-cache alignment
- **Photon:** Image compression

**Comparison:**

| Aspect | TokMan | Claw Compactor |
|--------|--------|----------------|
| **Papers** | 120+ | 30+ |
| **Depth** | More layers, more papers | Fewer stages, focused |
| **Novelty** | Combines many techniques | Practical fusion |

**Verdict:** TokMan has broader research coverage. Claw Compactor is more focused and practical.

---

## Unique Features

### TokMan Only
1. **TOML Filters:** 97+ declarative filters for popular tools
2. **Agent Integration:** 7+ AI agents (Claude, Cursor, Copilot, etc.)
3. **Dashboard:** Built-in web analytics dashboard
4. **Cost Analysis:** Per-command cost tracking
5. **SIMD Optimization:** AVX2, AVX-512, ARM NEON (Go 1.26+)
6. **Quality Metrics:** 6-metric grading system
7. **Hook Integrity:** Verification system for shell hooks
8. **Session Tracking:** SQLite-based command history

### Claw Compactor Only
1. **Cross-Message Dedup:** Deduplication across conversation turns
2. **RewindStore:** Mature reversible compression with tool calls
3. **QuantumLock:** KV-cache alignment for system prompts
4. **Cortex:** Centralized content-type detection (16 languages)
5. **Photon:** Image/base64 compression
6. **Content-Specific Stages:** LogCrunch, SearchCrunch, DiffCrunch
7. **Immutable Architecture:** Frozen dataclasses, no side effects
8. **ROUGE-L Validation:** Academic-grade quality metrics

---

## Recommendations

### Use TokMan If:
- ✅ You need CLI command interception
- ✅ You're using Claude Code, Cursor, or GitHub Copilot
- ✅ You want a standalone binary (no dependencies)
- ✅ You need sub-20ms latency
- ✅ You want built-in analytics dashboard
- ✅ You need TOML declarative filters
- ✅ You're working in Go ecosystem

### Use Claw Compactor If:
- ✅ You need Python library integration
- ✅ You're building AI agents (OpenClaw)
- ✅ You need cross-message deduplication
- ✅ You want immutable, testable architecture
- ✅ You need exceptional JSON compression (81.9%)
- ✅ You want content-specific stages (logs, diffs, search)
- ✅ You're working in Python ecosystem

---

## Conclusion

**TokMan** and **Claw Compactor** are both excellent tools with different strengths:

- **TokMan** is a production-ready CLI proxy optimized for terminal workflows, with broader research coverage (120+ papers) and faster performance (Go).
- **Claw Compactor** is a mature Python library optimized for agent integration, with cleaner architecture (immutable) and better structured data compression.

**For CLI users:** TokMan is the clear winner.  
**For Python developers:** Claw Compactor is the better choice.  
**For maximum compression:** Use both! TokMan for CLI, Claw for API.

---

## Future Opportunities

### TokMan Could Learn From Claw:
1. **Immutable Architecture:** Adopt frozen dataclasses for better testability
2. **Cross-Message Dedup:** Add conversation-level deduplication
3. **Content-Specific Stages:** Add LogCrunch, SearchCrunch, DiffCrunch equivalents
4. **Centralized Detection:** Add a Cortex-like content detection layer
5. **RewindStore Maturity:** Improve Sketch Store with tool call integration

### Claw Could Learn From TokMan:
1. **CLI Interception:** Add shell hook system for command interception
2. **TOML Filters:** Add declarative filter system
3. **Dashboard:** Add web analytics UI
4. **Quality Metrics:** Add 6-metric grading system
5. **SIMD Optimization:** Add native SIMD support (via Cython/NumPy)
6. **Go Port:** Consider Go rewrite for performance

---

**Document Version:** 1.0  
**Last Updated:** April 10, 2026  
**Authors:** TokMan Team  
**License:** MIT
