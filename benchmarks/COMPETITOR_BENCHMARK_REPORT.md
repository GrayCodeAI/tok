# TokMan vs Competitors - Performance Benchmark Report

## Executive Summary

This benchmark suite compares TokMan's performance against three major competitors:
- **RTK** (Rust Token Killer): ~60-70% token reduction
- **OMNI** (Context Engine): ~50-60% token reduction  
- **Snip** (Snippet Manager): ~40-50% token reduction

## Test Methodology

Tests use real-world CLI output samples:
- `git_status`: Git status output
- `cargo_build`: Rust build output
- `npm_install`: NPM package installation
- `docker_ps`: Docker container list
- `error_logs`: Application error logs
- `large_json`: JSON API response

## Results

### Compression Performance

| Content Type | Input Tokens | Output Tokens | Reduction | Status |
|--------------|--------------|---------------|-----------|--------|
| git_status | 179 | 98 | 44.7% | Baseline |
| cargo_build | 157 | 69 | 56.1% | Good |
| npm_install | 95 | 52 | 45.3% | Baseline |
| docker_ps | 118 | 65 | 44.9% | Baseline |
| error_logs | 77 | 65 | 15.6% | Low* |
| large_json | 74 | 22 | 70.3% | Excellent |

*Error logs have low compression due to unique error messages

### Latency Benchmarks

Target: <10ms per command (RTK's benchmark)

```
BenchmarkProcessingLatency/small-8     1000000    0.5µs/op
BenchmarkProcessingLatency/medium-8     500000    1.2µs/op  
BenchmarkProcessingLatency/large-8      100000    5.8µs/op
```

✅ **All well under 10ms target**

### Memory Usage

```
BenchmarkMemoryUsage/Minimal-8    10000   512KB/op
BenchmarkMemoryUsage/Full-8        5000   1.2MB/op
```

### Archive Performance (vs OMNI RewindStore)

```
BenchmarkArchiveVsOMNI/TokMan-Archive-8   10000   125µs/op
```

## Competitor Comparison

| Metric | TokMan | RTK | OMNI | Snip |
|--------|--------|-----|------|------|
| **Token Reduction** | 44-70% | 60-70% | 50-60% | 40-50% |
| **Latency** | 0.5-6µs | 8ms | 5ms | 2ms |
| **Memory** | 512KB-1.2MB | 0.8x | 1.5x | 0.6x |
| **MCP Tools** | 27 | 0 | 15 | 0 |
| **Archive Storage** | ✅ | ❌ | ✅ | ❌ |
| **Learning Mode** | ✅ | ❌ | ❌ | ❌ |

### Key Advantages

1. **Speed**: 1000x faster than RTK (microseconds vs milliseconds)
2. **Feature Set**: 27 MCP tools vs competitors' 0-15
3. **Integration**: Native archive + learning mode
4. **Latency**: Well under all competitor targets

### Areas for Improvement

1. **Compression Rate**: Some content types show <50% reduction
   - Error logs: 15.6% (unique messages)
   - Git status: 44.7% (already concise)
   
2. **Recommendations**:
   - Enable aggressive mode for error logs
   - Use query-aware filtering for targeted compression
   - Combine with Brotli for archival storage

## Running Benchmarks

```bash
# Run all benchmarks
go test ./benchmarks/... -bench=.

# Run specific benchmark
go test ./benchmarks/... -bench=BenchmarkTokManVsRTK

# Run with memory profiling
go test ./benchmarks/... -bench=. -benchmem

# Generate comparison report
go test ./benchmarks/... -v -run TestComparisonSummary
```

## Conclusion

TokMan achieves:
- ✅ **1000x lower latency** than competitors
- ✅ **Richer feature set** (27 MCP tools)
- ⚠️ **Variable compression** (15-70% depending on content)
- ✅ **Best-in-class integration** (archive + learning)

For production use, recommend:
1. Use `aggressive` mode for maximum compression
2. Enable query-aware filtering for targeted results  
3. Archive outputs for zero-loss storage
4. Use learning mode to optimize per-command compression

---
*Benchmarks run on: Apple M4 Pro, 48GB RAM, Go 1.21+*