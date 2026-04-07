# TokMan Architecture

## Overview

TokMan is a token-aware CLI proxy that intercepts CLI commands and applies a 31-layer compression pipeline to reduce token usage for AI coding assistants.

## System Architecture

```
┌─────────────────────────────────────────────────────────┐
│                     AI Coding Assistant                   │
│              (Claude Code, Cursor, etc.)                  │
└────────────────────────┬────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────┐
│                      Hook Layer                          │
│  ┌─────────────┐  ┌──────────────┐  ┌────────────────┐ │
│  │ PreToolUse  │  │ PostToolUse  │  │ SessionStart   │ │
│  │   Hook      │  │   Hook       │  │   Hook         │ │
│  └─────────────┘  └──────────────┘  └────────────────┘ │
└────────────────────────┬────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────┐
│                   TokMan CLI Proxy                       │
│  ┌──────────────────────────────────────────────────┐  │
│  │              Command Router                       │  │
│  │  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌───────┐ │  │
│  │  │ Filter  │ │ Commands│ │ Analysis│ │System │ │  │
│  │  │ Pipeline│ │ Runner  │ │ Engine  │ │ Utils │ │  │
│  │  └─────────┘ └─────────┘ └─────────┘ └───────┘ │  │
│  └──────────────────────────────────────────────────┘  │
└────────────────────────┬────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────┐
│                  Filter Pipeline (31 Layers)              │
│  ┌────────────────────────────────────────────────────┐ │
│  │  Stage 1: Information Analysis                     │ │
│  │  ├─ L1: Entropy Filtering                          │ │
│  │  ├─ L2: Perplexity Pruning                         │ │
│  │  ├─ L3: Goal-Driven Selection                      │ │
│  │  └─ L4: AST Preservation                           │ │
│  ├────────────────────────────────────────────────────┤ │
│  │  Stage 2: Semantic Compression                     │ │
│  │  ├─ L5: Contrastive Ranking                        │ │
│  │  ├─ L6: N-gram Abbreviation                        │ │
│  │  ├─ L7: Evaluator Heads                            │ │
│  │  ├─ L8: Gist Compression                           │ │
│  │  └─ L9: Hierarchical Summary                       │ │
│  ├────────────────────────────────────────────────────┤ │
│  │  Stage 3: Advanced Filtering                       │ │
│  │  ├─ L10: Budget Enforcement                        │ │
│  │  ├─ L11: Compaction                                │ │
│  │  ├─ L12: Attribution Filter                        │ │
│  │  ├─ L13: H2O Filter                                │ │
│  │  └─ L14: Attention Sink                            │ │
│  ├────────────────────────────────────────────────────┤ │
│  │  Stage 4: Intelligence Enhancement                 │ │
│  │  ├─ L15: Meta-Token Compression                    │ │
│  │  ├─ L16: Semantic Chunking                         │ │
│  │  ├─ L17: Sketch Store                              │ │
│  │  ├─ L18: Lazy Pruner                               │ │
│  │  ├─ L19: Semantic Anchor                           │ │
│  │  └─ L20: Agent Memory                              │ │
│  └────────────────────────────────────────────────────┘ │
└────────────────────────┬────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────┐
│                  Output & Analytics                      │
│  ┌─────────────┐  ┌──────────────┐  ┌────────────────┐ │
│  │ Filtered    │  │ Quality      │  │ Tracking &     │ │
│  │ Output      │  │ Metrics      │  │ Analytics      │ │
│  └─────────────┘  └──────────────┘  └────────────────┘ │
└─────────────────────────────────────────────────────────┘
```

## Directory Structure

```
tokman/
├── cmd/tokman/main.go              # Entry point
├── internal/
│   ├── commands/                   # CLI command definitions
│   │   ├── root.go                 # Root command
│   │   ├── registry/               # Command registry pattern
│   │   ├── shared/                 # Shared state
│   │   └── [categories]/           # Command categories
│   ├── filter/                     # 31-layer compression pipeline
│   │   ├── pipeline.go             # Pipeline coordinator
│   │   ├── filter.go               # Filter types and modes
│   │   ├── presets.go              # Pipeline presets
│   │   └── [layer_*.go]           # Individual layers
│   ├── core/                       # Core functionality
│   │   ├── runner.go               # Command runner
│   │   ├── estimator.go            # Token estimation
│   │   └── interfaces.go           # Core interfaces
│   ├── config/                     # Configuration loading
│   ├── tracking/                   # Command tracking & SQLite
│   ├── toml/                       # TOML filter configuration
│   ├── tee/                        # Output tee/logging
│   ├── dashboard/                  # Dashboard web interface
│   ├── economics/                  # Cost analysis
│   ├── telemetry/                  # Telemetry collection
│   ├── integrity/                  # Hook integrity verification
│   └── utils/                      # Logging utilities
├── config/                         # Default config files
├── templates/                      # Init templates
├── tests/                          # Integration tests
└── benchmarks/                     # Performance benchmarks
```

## Core Components

### 1. Command Runner (`internal/core/runner.go`)

The `OSCommandRunner` executes shell commands via `os/exec`:

```go
type OSCommandRunner struct {
    logger  Logger
    timeout time.Duration
}

func (r *OSCommandRunner) Run(ctx context.Context, cmd string, args ...string) (string, error) {
    // 1. Create command
    command := exec.CommandContext(ctx, cmd, args...)
    
    // 2. Capture stdout/stderr
    var stdout, stderr bytes.Buffer
    command.Stdout = &stdout
    command.Stderr = &stderr
    
    // 3. Execute with timeout
    err := command.Run()
    if err != nil {
        return "", fmt.Errorf("command failed: %w, stderr: %s", err, stderr.String())
    }
    
    return stdout.String(), nil
}
```

### 2. Filter Pipeline (`internal/filter/pipeline.go`)

The `PipelineCoordinator` orchestrates all 31 layers with early-exit support and stage gates:

```go
type PipelineCoordinator struct {
    config  PipelineConfig
    layers  []Layer
    stats   PipelineStats
}

func (p *PipelineCoordinator) Process(input string) (string, PipelineStats) {
    output := input
    
    for _, layer := range p.layers {
        // Stage gate: skip if layer wouldn't add value
        if p.shouldSkipLayer(layer) {
            continue
        }
        
        // Apply layer
        filtered, saved := layer.Apply(output, p.config.Mode)
        output = filtered
        
        // Update stats
        p.stats.TokensSaved += saved
        
        // Early exit if budget met
        if p.config.Budget > 0 && p.stats.TokensSaved >= p.config.Budget {
            break
        }
    }
    
    return output, p.stats
}
```

### 3. Token Estimator (`internal/core/estimator.go`)

Unified token estimation using `len(text) / 4` heuristic:

```go
func EstimateTokens(text string) int {
    // Approximation: 1 token ≈ 4 characters
    return len(text) / 4
}
```

### 4. Configuration (`internal/config/config.go`)

Viper-based configuration loading:

```go
type Config struct {
    Tracking  TrackingConfig
    Filter    FilterConfig
    Pipeline  PipelineConfig
    Hooks     HooksConfig
    Dashboard DashboardConfig
}
```

### 5. Tracking (`internal/tracking/tracker.go`)

SQLite-based command tracking:

```go
type Tracker struct {
    db *sql.DB
}

func (t *Tracker) Record(record CommandRecord) error {
    // Insert into SQLite
    _, err := t.db.Exec(
        "INSERT INTO commands (name, args, output_tokens, saved_tokens) VALUES (?, ?, ?, ?)",
        record.Name, record.Args, record.OutputTokens, record.SavedTokens,
    )
    return err
}
```

## Filter Layer Architecture

Each layer implements the `Layer` interface:

```go
type Layer interface {
    Apply(input string, mode Mode) (string, int)
    Name() string
    ShouldSkip(input string, config PipelineConfig) bool
}
```

### Layer Execution Flow

```
Input Text
    │
    ▼
┌─────────────────┐
│  Stage Gate     │ ← Skip check (cost ≈ 0)
│  shouldSkip?    │
└────────┬────────┘
         │ No
         ▼
┌─────────────────┐
│  Apply Filter   │ ← Core filtering logic
│  layer.Apply()  │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Update Stats   │ ← Tokens saved, time taken
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Early Exit?    │ ← Budget met?
└────────┬────────┘
         │ No
         ▼
    Next Layer
```

## Hook System Architecture

### PreToolUse Hook

Intercepts commands before execution:

```bash
#!/bin/bash
# tokman-pre-hook.sh

# Check if command should be intercepted
if should_intercept "$command"; then
    # Rewrite to tokman
    exec tokman "$command" "$@"
fi
```

### PostToolUse Hook

Processes output after execution:

```bash
#!/bin/bash
# tokman-post-hook.sh

# Read output from file
output=$(cat "$output_file")

# Process through tokman
processed=$(tokman filter --input "$output")

# Replace with filtered version
echo "$processed" > "$output_file"
```

## Quality Metrics System

6-metric analysis for each filtered output:

```go
type QualityMetrics struct {
    SemanticScore     float64  // 0.0-1.0, how much meaning preserved
    SignalToNoise     float64  // 0.0-1.0, relevant vs noise ratio
    ContextRetention  float64  // 0.0-1.0, context preserved
    Readability       float64  // 0.0-1.0, human readability
    Completeness      float64  // 0.0-1.0, information completeness
    CompressionRatio  float64  // 0.0-1.0, size reduction
}

func CalculateGrade(metrics QualityMetrics) string {
    // A+: 0.95-1.0, A: 0.90-0.95, etc.
    // F: < 0.50
}
```

## Data Flow

### Command Execution Flow

```
1. User runs: git status
2. Hook intercepts
3. tokman command router receives
4. Command runner executes: git status
5. Output captured
6. Pipeline processes through 31 layers
7. Quality metrics calculated
8. Filtered output returned
9. Tracking recorded to SQLite
10. Telemetry sent (if enabled)
```

### Configuration Flow

```
1. tokman starts
2. Load ~/.config/tokman/config.toml
3. Override with env vars (TOKMAN_*)
4. Override with CLI flags
5. Initialize components
6. Ready to process
```

## Preset System

Three presets via `--preset`:

```go
type Preset struct {
    Name        string
    Mode        Mode
    Layers      map[string]bool
    Description string
}

var Presets = map[string]Preset{
    "fast": {
        Name:        "fast",
        Mode:        ModeMinimal,
        Layers:      fastLayers,
        Description: "Fewer layers, faster processing",
    },
    "balanced": {
        Name:        "balanced",
        Mode:        ModeMinimal,
        Layers:      balancedLayers,
        Description: "Default mix of speed and compression",
    },
    "full": {
        Name:        "full",
        Mode:        ModeAggressive,
        Layers:      allLayers,
        Description: "All layers enabled, maximum compression",
    },
}
```

## Extension Points

### TOML Filters

Custom filter definitions:

```toml
[my_command]
match = "^my-tool (build|test)"
output_patterns = ["^Building...", "^Testing..."]
strip_lines_matching = ["^INFO:"]
```

### Commands

Registry pattern for easy addition:

```go
func init() {
    registry.Add(func() { registry.Register(myCmd) })
}
```

### Filter Layers

Add to pipeline:

1. Create `internal/filter/my_layer.go`
2. Implement `Layer` interface
3. Add to `PipelineConfig`
4. Add to `PipelineCoordinator`
5. Initialize and execute in pipeline

## Performance Considerations

### SIMD Optimization

- Auto-vectorized by Go compiler
- Native SIMD planned for Go 1.26+
- `internal/simd/` package for optimizations

### Streaming

- Large inputs (>500K tokens) use streaming
- Process in chunks to avoid memory issues
- `internal/filter/stream.go`

### Caching

- Fingerprint-based result caching
- `internal/cache/` package
- Skip reprocessing identical outputs

### Stage Gates

- Skip layers when not applicable (zero cost)
- Early exit when budget met
- Minimal overhead per layer

## Security Considerations

### Hook Integrity

- SHA-256 checksums for hooks
- Tamper detection via `tokman doctor`
- `tokman hook-audit` for detailed reports

### Input Validation

- Command allowlist
- Path traversal protection
- SQL injection prevention

### Data Protection

- No secrets logged
- No credentials in telemetry
- Local-only SQLite database

## Testing Strategy

### Unit Tests

- Table-driven tests for all layers
- Edge cases, error paths, nil inputs
- `*_test.go` alongside source

### Integration Tests

- End-to-end command execution
- Real AI tool integration
- `tests/` directory

### Benchmarks

- Performance baselines
- Layer-by-layer benchmarks
- `benchmarks/` directory

### Fuzz Testing

- Parser fuzz tests
- Filter input fuzz tests
- `filter/fuzz_test.go`

---

**This document is a living reference for TokMan's architecture.** Update it as the system evolves.
