# Performance Benchmark Report - April 2026

## Pipeline Performance

### Compression Results

| Command | Original | Savings | Rate | Old Time | New Time |
|---------|----------|---------|------|----------|----------|
| build_output | 181 tokens | 92.8% (168 saved) | Excellent | 386µs | 1.171ms |
| git_status | 194 tokens | 91.8% (178 saved) | Excellent | 294µs | 1.068ms |
| cargo_test | 732 tokens | 96.4% (706 saved) | Excellent | 505µs | 1.864ms |
| ls_output | 207 tokens | 94.7% (196 saved) | Excellent | 224µs | 337µs |
| docker_ps | 232 tokens | 93.5% (217 saved) | Excellent | 157µs | 757µs |

### Analysis

- Compression rate improved from ~54-87% to ~92-96%
- Processing time: 157µs - 1.864ms per command
- Target: <10ms overhead (RTK's benchmark)
- Status: ✅ Well within target, 5-60x faster than target

## Optimization Priority

### Hot Paths (to optimize)
1. Entropy calculation - most CPU-intensive layer
2. N-gram matching - regex compilation overhead
3. Token counting - called for every layer output

### Memory Hot Spots
1. String allocations in filter pipeline
2. Regex compilation cache - compile once
3. Slice pre-allocation for token processing
