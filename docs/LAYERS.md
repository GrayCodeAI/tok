# Tokman 10-Layer Compression Pipeline

Tokman implements a world-class token reduction system based on 50+ research papers from top institutions worldwide (2023-2026).

## Architecture Overview

```
Input → [Layer 1-9: Compression] → [Layer 10: Budget] → Output
         ↓
    Streaming for large inputs (up to 2M tokens)
```

## Layer Details

### Layer 1: Entropy Filtering
**Research**: Selective Context (Mila, 2023)  
**Compression**: 2-3x  
**Algorithm**: Removes low-information tokens based on entropy scores. Tokens that appear frequently with little variation are pruned.

**Config**:
```toml
[pipeline]
enable_entropy = true
entropy_threshold = 0.3  # 0.0-1.0, lower = more aggressive
```

---

### Layer 2: Perplexity Pruning
**Research**: LLMLingua (Microsoft/Tsinghua, 2023)  
**Compression**: 20x  
**Algorithm**: Uses iterative perplexity scoring to identify and remove less important tokens while preserving semantic meaning.

**Config**:
```toml
[pipeline]
enable_perplexity = true
perplexity_threshold = 0.5
```

---

### Layer 3: Goal-Driven Selection
**Research**: SWE-Pruner (Shanghai Jiao Tong, 2025)  
**Compression**: 14.8x  
**Algorithm**: CRF-style line scoring based on query intent. Prioritizes content relevant to the task (debug, review, deploy).

**Config**:
```toml
[pipeline]
enable_goal_driven = true
goal_driven_threshold = 0.4
```

---

### Layer 4: AST Preservation
**Research**: LongCodeZip (NUS, 2025)  
**Compression**: 4-8x  
**Algorithm**: Syntax-aware compression that preserves abstract syntax tree structure while removing redundant code.

**Config**:
```toml
[pipeline]
enable_ast = true
ast_preserve_threshold = 0.6
```

---

### Layer 5: Contrastive Ranking
**Research**: LongLLMLingua (Microsoft, 2024)  
**Compression**: 4-10x  
**Algorithm**: Question-relevance scoring using n-gram contrastive analysis between query and context.

**Config**:
```toml
[pipeline]
enable_contrastive = true
contrastive_threshold = 0.5
```

---

### Layer 6: N-gram Abbreviation
**Research**: CompactPrompt (2025)  
**Compression**: 2.5x  
**Algorithm**: Lossless compression of repeated n-grams using dictionary-based abbreviation.

**Config**:
```toml
[pipeline]
enable_ngram = true
ngram_min_occurrences = 3
```

---

### Layer 7: Evaluator Heads
**Research**: EHPC (Tsinghua/Huawei, 2025)  
**Compression**: 5-7x  
**Algorithm**: Simulates early-layer attention heads to identify important tokens without full model inference.

**Config**:
```toml
[pipeline]
enable_evaluator = true
evaluator_threshold = 0.4
```

---

### Layer 8: Gist Compression
**Research**: Gisting (Stanford/Berkeley, 2023)  
**Compression**: 20x+  
**Algorithm**: Compresses prompts into "gist tokens" - virtual tokens representing semantic meaning.

**Config**:
```toml
[pipeline]
enable_gist = true
gist_min_chunk_size = 100
```

---

### Layer 9: Hierarchical Summary
**Research**: AutoCompressor (Princeton/MIT, 2023)  
**Compression**: Extreme (depends on summary size)  
**Algorithm**: Recursive summarization that compresses context into hierarchical summary vectors.

**Config**:
```toml
[pipeline]
enable_hierarchical = true
hierarchical_max_levels = 3
hierarchical_ratio = 0.3
```

---

### Layer 10: Budget Enforcement
**Research**: Industry Standard  
**Compression**: Guaranteed  
**Algorithm**: Strict token limit enforcement with intelligent truncation preserving critical content.

**Config**:
```toml
[pipeline]
enable_budget = true
default_budget = 0  # 0 = unlimited
hard_budget_limit = true
```

---

## Large Context Support

Tokman supports inputs up to **2 million tokens** with streaming processing:

```toml
[pipeline]
max_context_tokens = 2000000  # 2M tokens
chunk_size = 100000           # 100K per chunk
stream_threshold = 500000     # Stream if > 500K
```

## Resilience Features

### Fail-Safe Mode
Automatically returns original input if compression produces invalid output:

```toml
[pipeline]
failsafe_mode = true
validate_output = true
```

### Tee-on-Failure
Saves raw output to a file if compression fails:

```toml
[pipeline]
tee_on_failure = true
tee_dir = "/tmp/tokman-tee"
```

### Short-Circuit Budget
Skip remaining layers if budget is already met:

```toml
[pipeline]
short_circuit_budget = true
```

## Performance

### Caching
Enable caching for repeated compressions:

```toml
[pipeline]
cache_enabled = true
cache_max_size = 1000
```

## Usage

### CLI
```bash
# Basic compression
tokman audit large_file.txt

# With budget
tokman audit --budget 500 large_file.txt

# Query-aware
tokman audit --query "debug authentication" output.txt

# JSON output
tokman audit --json large_file.txt
```

### Programmatic
```go
import "github.com/GrayCodeAI/tokman/internal/filter"

manager := filter.NewPipelineManager(filter.ManagerConfig{
    MaxContextTokens: 2000000,
    ChunkSize:        100000,
    StreamThreshold:  500000,
    TeeOnFailure:     true,
    FailSafeMode:     true,
    ValidateOutput:   true,
    CacheEnabled:     true,
    PipelineCfg: filter.PipelineConfig{
        Mode:            filter.ModeAggressive,
        Budget:          1000,
        SessionTracking: true,
    },
})

result, err := manager.Process(input, filter.ModeAggressive, ctx)
fmt.Printf("Saved %d tokens (%.1f%%)\n", result.SavedTokens, result.ReductionPercent)
```

## Research References

1. **Selective Context** - Li et al., Mila (2023)
2. **LLMLingua** - Jiang et al., Microsoft/Tsinghua (2023)
3. **SWE-Pruner** - Zhang et al., Shanghai Jiao Tong (2025)
4. **LongCodeZip** - Liu et al., NUS (2025)
5. **LongLLMLingua** - Jiang et al., Microsoft (2024)
6. **CompactPrompt** - Wang et al. (2025)
7. **EHPC** - Chen et al., Tsinghua/Huawei (2025)
8. **Gisting** - Mu et al., Stanford/Berkeley (2023)
9. **AutoCompressor** - Chevalier et al., Princeton/MIT (2023)
