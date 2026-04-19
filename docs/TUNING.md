# Tok Configuration Tuning Guide

This guide explains how to tune Tok's compression pipeline for optimal performance.

## Configuration File

Tok uses TOML configuration. Default location: `~/.config/tok/config.toml`

```bash
# Create default config
tok config --init
```

## Complete Configuration Example

```toml
# Tok Configuration

[tracking]
enabled = true
database_path = ""  # Default: ~/.local/share/tok/tracking.db
telemetry = false

[filter]
mode = "minimal"  # "minimal" or "aggressive"
max_width = 0     # 0 = auto-detect

[filter.noise_dirs]
dirs = [".git", "node_modules", "target", "__pycache__", ".venv", "vendor"]

[filter.ignore_files]
patterns = ["*.lock", "*.min.js", "*.min.css", "*.map"]

# 10-Layer Pipeline Configuration
[pipeline]
# Context limits (supports up to 2M tokens)
max_context_tokens = 2000000
chunk_size = 100000
stream_threshold = 500000

# Layer enable/disable
enable_entropy = true
enable_perplexity = true
enable_goal_driven = true
enable_ast = true
enable_contrastive = true
enable_ngram = true
enable_evaluator = true
enable_gist = true
enable_hierarchical = true
enable_budget = true

# Layer thresholds
entropy_threshold = 0.3
perplexity_threshold = 0.5
goal_driven_threshold = 0.4
ast_preserve_threshold = 0.6
contrastive_threshold = 0.5
ngram_min_occurrences = 3
evaluator_threshold = 0.4
gist_min_chunk_size = 100
hierarchical_max_levels = 3
hierarchical_ratio = 0.3

# Budget
default_budget = 0
hard_budget_limit = true
budget_overflow_file = ""

# Resilience
tee_on_failure = true
failsafe_mode = true
validate_output = true
short_circuit_budget = true

# Performance
parallel_layers = false
cache_enabled = true
cache_max_size = 1000

[hooks]
excluded_commands = []
audit_dir = ""
tee_dir = ""

[dashboard]
port = 8080
bind = "localhost"
update_interval = 30000
theme = "dark"
enable_export = true

[alerts]
enabled = true
daily_token_limit = 1000000
weekly_token_limit = 5000000
usage_spike_threshold = 2.0

[export]
default_format = "json"
include_timestamps = true
max_records = 0
```

## Tuning Strategies

### Maximum Compression (High Token Savings)

```toml
[filter]
mode = "aggressive"

[pipeline]
enable_entropy = true
entropy_threshold = 0.2  # More aggressive
enable_perplexity = true
perplexity_threshold = 0.3
enable_hierarchical = true
hierarchical_ratio = 0.2  # Smaller summaries
```

### Balanced Mode (Default)

```toml
[filter]
mode = "minimal"

[pipeline]
entropy_threshold = 0.3
perplexity_threshold = 0.5
hierarchical_ratio = 0.3
```

### Preserve Code Structure

```toml
[pipeline]
enable_ast = true
ast_preserve_threshold = 0.8  # More preservation
enable_entropy = false        # Disable aggressive text pruning
enable_gist = false           # Keep full code blocks
```

### Debug Mode (Preserve Errors)

```toml
[pipeline]
enable_entropy = false
enable_perplexity = false
enable_goal_driven = true
goal_driven_threshold = 0.2  # Keep more error context
```

### Large File Processing

```toml
[pipeline]
max_context_tokens = 2000000  # 2M tokens
chunk_size = 200000           # Larger chunks
stream_threshold = 100000     # Stream earlier
cache_enabled = true
cache_max_size = 5000
```

### Fast Processing (Lower Quality)

```toml
[pipeline]
enable_perplexity = false  # Skip slow perplexity scoring
enable_contrastive = false
enable_hierarchical = false
parallel_layers = true
cache_enabled = true
```

## Environment Variables

Override configuration with environment variables:

```bash
# Budget
export TOK_PIPELINE_DEFAULT_BUDGET=1000

# Enable layers
export TOK_PIPELINE_ENABLE_ENTROPY=true
export TOK_PIPELINE_ENABLE_PERPLEXITY=false

# Thresholds
export TOK_PIPELINE_ENTROPY_THRESHOLD=0.3

# Resilience
export TOK_PIPELINE_FAILSAFE_MODE=true
export TOK_PIPELINE_VALIDATE_OUTPUT=true

# Cache
export TOK_PIPELINE_CACHE_ENABLED=true
```

## CLI Flags

Override config at runtime:

```bash
# Set budget
tok --budget 500 audit file.txt

# Set mode
tok audit --mode aggressive file.txt

# Set query intent
tok audit --query "debug error" file.txt

# Use LLM compression
tok --llm audit file.txt
```

## Layer-Specific Tuning

### Entropy Layer

Lower threshold = more compression (may lose content):

| Threshold | Effect |
|-----------|--------|
| 0.1 | Very aggressive, may remove useful content |
| 0.3 | Balanced (default) |
| 0.5 | Conservative, preserves more content |
| 0.8 | Minimal pruning |

### Perplexity Layer

Lower threshold = more aggressive pruning:

| Threshold | Effect |
|-----------|--------|
| 0.2 | Maximum compression |
| 0.5 | Balanced (default) |
| 0.7 | Conservative |

### Hierarchical Layer

| Ratio | Effect |
|-------|--------|
| 0.2 | Very compressed summaries |
| 0.3 | Balanced (default) |
| 0.5 | Detailed summaries |

| Max Levels | Effect |
|------------|--------|
| 2 | Shallow hierarchy |
| 3 | Balanced (default) |
| 5 | Deep hierarchy |

## Monitoring

### View Current Config

```bash
tok config
```

### Audit Compression Performance

```bash
tok audit large_file.txt
```

### View Layer Statistics

```bash
tok layers --verbose
```

## Performance Tips

1. **Enable caching** for repeated compressions
2. **Use short-circuit budget** to skip layers when budget is met
3. **Disable slow layers** (perplexity, hierarchical) for fast processing
4. **Increase chunk size** for very large inputs
5. **Use streaming** for inputs > 500K tokens (automatic)
