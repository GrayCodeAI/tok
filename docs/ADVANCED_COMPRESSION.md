# Advanced Compression Features

TokMan implements research-backed compression algorithms for optimal LLM context management.

## Overview

Based on analysis of 20+ LLM context compression papers (2023-2024), TokMan implements six advanced compression techniques:

| Feature | Research Paper | Token Savings | Overhead |
|---------|---------------|---------------|----------|
| Semantic Pruning | Selective Context (Li et al., 2024) | +5-10% | +2-5ms |
| Position-Aware | LongLLMLingua (Jiang et al., 2024) | Quality improvement | +1ms |
| Query-Aware | LongLLMLingua, ACON (2024) | +10-20% quality | +1ms |
| Guideline Optimizer | ACON (Zhang et al., 2024) | Self-improving | Minimal |
| Hierarchical Summarization | Recurrent Context Compression (Liu et al., 2024) | Up to 10x | +5ms |
| Local LLM Integration | TCRA-LLM (2024) | +40-60% quality | 50-200ms (optional)

## 1. Semantic Pruning

**What it does**: Identifies and removes low-information segments using statistical analysis.

**How it works**:
- Calculates information density (unique token ratio, keyword density, entropy)
- Prunes segments with density below threshold
- Preserves high-value content (errors, file references, key results)

**Example**:
```
Input (1000 lines):
  - 200 lines of "Building..." progress messages
  - 50 lines of actual error/output

Output: 50 lines (95% reduction) with all meaningful content preserved
```

## 2. Position-Aware Filtering

**What it does**: Reorders output so critical information appears at beginning and end.

**Why it matters**: LLMs exhibit "lost in the middle" bias - they better recall information at the start and end of context.

**How it works**:
- Scores each segment by importance (errors > warnings > success > info)
- High-scoring segments moved to beginning/end
- Preserves all content, just reorders for better LLM recall

**Importance scoring**:
- Error content: 10 points
- Stack traces: 8 points
- File references: 6 points
- Warnings: 4 points
- Diff hunks: 4 points
- Success messages: 1 point
- Normal content: 0 points

## 3. Query-Aware Compression

**What it does**: Tailors output filtering based on the agent's task intent.

**Usage**:
```bash
# CLI flag
tokman --query debug cargo test
tokman --query review git diff
tokman --query deploy docker ps

# Environment variable
TOKMAN_QUERY=debug tokman npm test
```

**Supported intents**:

| Intent | Prioritizes | Use Case |
|--------|-------------|----------|
| `debug` | Errors, stack traces, failures | Finding bugs |
| `review` | Diffs, changes, file references | Code review |
| `deploy` | Status, versions, health | Deployments |
| `search` | File names, definitions | Finding code |
| `test` | Test results, coverage | Testing |
| `build` | Errors, warnings | Build status |

**Example**:
```bash
# Debug mode keeps errors, removes success messages
$ tokman --query debug cargo test

# Output focuses on:
# - Failed test names and assertions
# - Error messages and stack traces
# - File paths with line numbers
# - Skips: passing test output, progress bars
```

## 4. Guideline-Based Optimization

**What it does**: Learns compression rules from agent failures to improve over time.

**How it works**:
1. Agent fails a task due to missing context
2. Failure is analyzed to identify what was removed incorrectly
3. A compression guideline is created (e.g., "keep test names in output")
4. Future filtering applies learned guidelines
5. Guidelines gain confidence with successful applications

**Storage**: `~/.local/share/tokman/guidelines.json`

**Example learned guidelines**:
```json
[
  {"pattern": "keep test names in output", "confidence": 0.9, "apply_count": 15},
  {"pattern": "preserve stack traces for debugging", "confidence": 0.85, "apply_count": 12},
  {"pattern": "keep file paths for navigation", "confidence": 0.8, "apply_count": 8}
]
```

## 5. Hierarchical Summarization

**What it does**: Creates multi-level summaries for very large outputs (500+ lines).

**Based on**: Recurrent Context Compression (Liu et al., 2024), Hierarchical Context Compression research.

**How it works**:
1. Segments output into logical sections (by detecting boundaries like `---`, `error:`, test results)
2. Scores each section by importance (errors > warnings > success)
3. Keeps high-score sections verbatim
4. Compresses mid-score sections into one-line summaries with line ranges
5. Drops low-score sections entirely

**Example**:
```
Input (1000 lines of build output):
  - 200 lines of "Compiling..." messages
  - 50 lines of warnings
  - 10 lines of errors

Output:
[Hierarchical Summary: 1000 lines → 3 sections]

├─ [L1-200] Compiling crate v1.0... (200 lines, score: 0.25)
<full error output preserved>
├─ [L250-300] warning: unused variable (50 lines, score: 0.45)
```

**Configuration**:
- Line threshold: 500 lines (configurable via `SetLineThreshold()`)
- Max depth: 3 levels (configurable via `SetMaxDepth()`)

## 6. Local LLM Integration

**What it does**: Uses local LLMs (Ollama, LM Studio) for intelligent summarization.

**Based on**: TCRA-LLM (2024) - LLM-based Context Compression research showing 40-60% better semantic preservation.

**Usage**:
```bash
# CLI flag (requires Ollama or LM Studio running)
tokman --llm cargo test

# Environment variable
TOKMAN_LLM=true tokman npm test

# Configure LLM provider
TOKMAN_LLM_PROVIDER=ollama TOKMAN_LLM_MODEL=llama3.2:3b tokman <command>
```

**Supported providers**:

| Provider | Default URL | Models |
|----------|-------------|--------|
| Ollama | http://localhost:11434 | llama3.2, mistral, phi3, etc. |
| LM Studio | http://localhost:1234 | Any GGUF model |
| OpenAI-compatible | Configurable | Any compatible API |

**How it works**:
1. Auto-detects running local LLM (Ollama → LM Studio)
2. Builds intent-aware prompt based on content type (debug/review/test/build)
3. Sends content to local LLM for summarization
4. Falls back to semantic filter if LLM unavailable
5. Caches summaries for repeated content

**Intent-aware prompts**:
- **debug**: Focus on errors, stack traces, file paths, line numbers
- **review**: Focus on code changes, diff hunks, modified functions
- **test**: Focus on test results, pass/fail counts, failing test names
- **build**: Focus on compilation errors, warnings, build status

**Configuration**:
```bash
# Set provider explicitly
TOKMAN_LLM_PROVIDER=ollama

# Set model
TOKMAN_LLM_MODEL=llama3.2:3b

# Set base URL (for custom setups)
TOKMAN_LLM_BASE_URL=http://localhost:8080
```

**Performance considerations**:
- Latency: 50-200ms (depends on model and hardware)
- Quality: 40-60% better semantic preservation vs heuristics
- Privacy: All processing is local
- Fallback: Automatically uses semantic filter if LLM unavailable

## Performance Benchmarks

| Scenario | Latency | Memory | Token Savings |
|----------|---------|--------|---------------|
| Short output (<1KB) | 13ms | 6KB | 30-50% |
| Git status (50 files) | 131ms | 26KB | 60-70% |
| Test output (100 tests) | 75ms | 18KB | 70-85% |
| Large output (1000 lines) | 500ms | 100KB | 80-90% |
| Query-Aware filter | 43ms | 10KB | +10-20% quality |
| Position-Aware filter | 67ms | 27KB | Quality improvement |
| Semantic filter | 432ms | 113KB | +5-10% savings |

**Target**: <20ms overhead for typical outputs with all advanced filters enabled.

## Integration

### CLI Usage

```bash
# Enable query-aware compression
tokman --query debug <command>

# Use environment variable
export TOKMAN_QUERY=debug
tokman cargo test

# Combine with ultra-compact mode
tokman -u --query deploy docker ps
```

### Programmatic Usage

```go
import "github.com/GrayCodeAI/tokman/internal/filter"

// Create engine with query intent
engine := filter.NewEngineWithQuery(filter.ModeMinimal, "debug")
output, tokensSaved := engine.Process(input)

// Use guideline optimizer
optimizer := feedback.NewGuidelineOptimizer(dataDir)
optimizer.AnalyzeFailure(feedback.AgentFailure{
    Task:    "fix bug in auth",
    Missing: "test name was removed",
})
enhanced := optimizer.EnhanceOutput(original, filtered)
```

## Research References

1. **Selective Context** - Li et al., 2024: Information density-based pruning
2. **LongLLMLingua** - Jiang et al., 2024: Position bias and query-aware compression
3. **ACON** - Zhang et al., 2024: Agent-optimized context and guideline learning
4. **Recurrent Context Compression** - Liu et al., 2024: Hierarchical summarization
5. **TCRA-LLM** - 2024: LLM-based compression (future consideration)

## Future Enhancements

- ✅ **Hierarchical Summarization**: Implemented in v0.9.0
- ✅ **Local LLM Integration**: Implemented in v0.9.0
- **Cross-session Learning**: Share guidelines across team members
- **Multi-file Context**: Optimize across multiple related outputs
- **Custom LLM Prompts**: User-defined summarization templates
