# TokMan Performance Guide

This guide covers performance optimization features in TokMan, including caching, batching, and profiling.

## Overview

TokMan includes several performance optimizations:

1. **Command Rewrite Caching**: 28x speedup for repeated commands
2. **Telemetry Batching**: Reduces HTTP requests by 10x
3. **SIMD Optimizations**: AVX2/AVX-512/NEON acceleration
4. **Streaming Mode**: Handles large inputs efficiently

## Command Rewrite Caching

The rewrite system caches command rewrites to avoid reprocessing.

### How It Works

```
First call:  cargo test → tokman test-runner cargo test  (230 ns)
Cached call: cargo test → [cache lookup] → cached result  (8 ns)
```

### Performance Gains

- **Without cache**: 230 ns/op
- **With cache**: 8.2 ns/op
- **Speedup**: 28x faster
- **Memory**: Zero allocations

### Cache Statistics

View cache performance:

```bash
tokman doctor --cache-stats
```

Output:
```
Cache Statistics:
  Hits:   1,234
  Misses: 56
  Hit Rate: 95.6%
```

### Manual Cache Management

```bash
# Clear cache
tokman doctor --clear-cache

# View cache size
tokman doctor --cache-info
```

### Disabling Cache

For debugging or testing:

```go
opts := &discover.RewriteOptions{
    DisableCache: true,
}
```

## Telemetry Batching

Telemetry events are batched to reduce network overhead.

### How It Works

```
Without batching:
  Event 1 → HTTP POST
  Event 2 → HTTP POST
  Event 3 → HTTP POST

With batching:
  Event 1 → [batch]
  Event 2 → [batch]
  Event 3 → [batch]
  [timer expires or batch full] → HTTP POST (all events)
```

### Configuration

Default batch settings:
- **Batch size**: 10 events
- **Flush timeout**: 30 seconds

### Performance Impact

| Metric | Without Batching | With Batching | Improvement |
|--------|------------------|---------------|-------------|
| HTTP Requests | 100/min | 10/min | 10x fewer |
| Latency | ~50ms | ~5ms | 10x faster |
| CPU Usage | 2% | 0.2% | 10x less |

### Manual Flush

Force immediate flush:

```bash
# Flush all pending telemetry
tokman telemetry --flush

# Disable batching (debug only)
export TOKMAN_TELEMETRY_BATCH_SIZE=1
```

## Profiling

TokMan includes built-in profiling capabilities.

### CPU Profiling

Profile CPU usage:

```bash
# Built-in profiler
tokman profile --cpu --duration=30s --output=cpu.prof

# View results
go tool pprof -http=:8080 cpu.prof

# Top functions
go tool pprof -top cpu.prof | head -20
```

### Memory Profiling

Profile memory allocations:

```bash
# Built-in profiler
tokman profile --mem --duration=30s --output=mem.prof

# View results
go tool pprof -http=:8080 mem.prof

# Top allocations
go tool pprof -top mem.prof | head -20
```

### Execution Tracing

Trace execution flow:

```bash
# Built-in tracer
tokman profile --trace --duration=5s --output=trace.out

# View results
go tool trace trace.out
```

### Script-Based Profiling

Use the profiling script for comprehensive analysis:

```bash
# Profile everything
./scripts/profile.sh all

# Profile specific components
./scripts/profile.sh cpu 30s ./profiles
./scripts/profile.sh mem 10s ./profiles
./scripts/profile.sh rewrite
./scripts/profile.sh quota
```

## Benchmarks

Run benchmarks to measure performance:

### Discover Rewrite Benchmarks

```bash
# Run all rewrite benchmarks
go test -bench=. ./internal/discover/

# Run with memory profiling
go test -bench=BenchmarkRewriteCommand -benchmem ./internal/discover/

# Run specific benchmark
go test -bench=BenchmarkRewriteCommandWithCaching ./internal/discover/
```

### Expected Results

```
BenchmarkRewriteCommand-8                    5232486    230.2 ns/op
BenchmarkRewriteCommandWithCaching-8        145809583    8.221 ns/op
BenchmarkRewriteCommandMemoryAllocation-8   42185068    28.63 ns/op     0 B/op    0 allocs/op
```

### Quota Calculation Benchmarks

```bash
# Run quota benchmarks
go test -bench=BenchmarkQuota ./internal/commands/core/

# With profiling
go test -bench=BenchmarkQuota -cpuprofile=cpu.prof -memprofile=mem.prof ./internal/commands/core/
```

## SIMD Optimizations

TokMan uses SIMD instructions for performance-critical operations.

### Supported Instructions

| Platform | Instructions | Speedup |
|----------|--------------|---------|
| x86_64 | AVX2, AVX-512 | 2-3x |
| ARM64 | NEON | 2-3x |

### Enabling SIMD

SIMD is automatically enabled when available:

```bash
# Check SIMD support
tokman doctor --simd-info
```

Output:
```
SIMD Support:
  AVX2:     ✓ Available
  AVX-512:  ✗ Not available
  NEON:     ✗ (not ARM)
```

### SIMD Build Tags

Force SIMD build:

```bash
# Build with AVX2
go build -tags=simd_avx2 ./cmd/tokman

# Build with NEON
go build -tags=simd_neon ./cmd/tokman
```

## Streaming Mode

For large inputs (>500K tokens), TokMan uses streaming mode.

### How It Works

```
Input (>500K tokens)
    ↓
[Streaming Mode]
    ↓
Chunk 1 → Process → Output
Chunk 2 → Process → Output
Chunk N → Process → Output
```

### Performance

| Input Size | Regular Mode | Streaming Mode | Memory |
|------------|--------------|----------------|--------|
| 100K | 50ms | 55ms | 10MB |
| 500K | 250ms | 280ms | 50MB |
| 1M | OOM | 500ms | 100MB |
| 5M | OOM | 2.5s | 100MB |

### Configuration

```toml
[pipeline]
stream_threshold = 500000  # tokens
stream_chunk_size = 100000  # tokens per chunk
```

## Performance Tuning

### Optimize for Speed

```toml
[filter]
preset = "fast"  # Fewer layers, faster processing

[pipeline]
enable_caching = true
cache_size = 10000
```

### Optimize for Memory

```toml
[pipeline]
stream_threshold = 100000  # Lower threshold for streaming
max_memory_mb = 512

[cache]
max_entries = 1000  # Smaller cache
```

### Optimize for Compression

```toml
[filter]
preset = "full"  # All 20 layers
mode = "aggressive"

[pipeline]
budget = 10000  # Strict token budget
```

## Monitoring Performance

### Built-in Metrics

```bash
# View performance metrics
tokman gain --metrics

# View cache hit rate
tokman doctor --cache-stats

# View telemetry stats
tokman telemetry --stats
```

### External Monitoring

```bash
# Export metrics in Prometheus format
tokman metrics --format=prometheus

# Export in JSON
tokman metrics --format=json
```

## Troubleshooting Performance

### Slow Command Rewriting

If rewriting is slow:

```bash
# Check cache hit rate
tokman doctor --cache-stats

# Clear and rebuild cache
tokman doctor --clear-cache

# Profile rewrite system
./scripts/profile.sh rewrite
```

### High Memory Usage

If memory usage is high:

```bash
# Check for memory leaks
./scripts/profile.sh mem

# Enable streaming mode
tokman config set pipeline.stream_threshold 100000

# Reduce cache size
tokman config set cache.max_entries 1000
```

### Slow Telemetry

If telemetry is slowing down commands:

```bash
# Check telemetry queue
tokman telemetry --queue-size

# Flush pending events
tokman telemetry --flush

# Disable telemetry (last resort)
tokman telemetry --disable
```

## Performance Checklist

Before production deployment:

- [ ] Run benchmarks: `go test -bench=. ./...`
- [ ] Profile hot paths: `./scripts/profile.sh all`
- [ ] Check cache hit rate: > 90% expected
- [ ] Verify SIMD support: `tokman doctor --simd-info`
- [ ] Test with large inputs: > 500K tokens
- [ ] Monitor memory usage: < 512MB expected
- [ ] Verify telemetry batching: < 10% overhead

## See Also

- [Benchmarks](./BENCHMARKS.md)
- [Tuning Guide](./TUNING.md)
- [Deployment Guide](./DEPLOYMENT.md)
- [API Reference](./API.md)
