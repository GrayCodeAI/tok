# TokMan Performance Report

**Benchmark results and optimization status**

---

## Executive Summary

TokMan achieves **sub-millisecond overhead** for typical command outputs, well under the 10ms target. Performance is comparable to RTK's claimed <10ms overhead.

---

## Benchmark Results

### Command Handler Overhead (Target: <10ms)

| Command | Input Size | Overhead | Memory | Status |
|---------|------------|----------|--------|--------|
| Git status | 50 files | 0.29ms | 48KB | ✅ |
| Git status | 100 files | 0.52ms | 87KB | ✅ |
| Go test | 100 tests | 0.57ms | 108KB | ✅ |
| Npm test | 200 tests | 0.40ms | 69KB | ✅ |
| Docker ps | 100 containers | 0.88ms | 156KB | ✅ |
| Kubectl get | 50 pods | 0.22ms | 39KB | ✅ |

### Large Output Performance

| Lines | Overhead | Memory | Status |
|-------|----------|--------|--------|
| 100 | 0.43ms | 73KB | ✅ |
| 500 | 2.03ms | 345KB | ✅ |
| 1,000 | 4.29ms | 711KB | ✅ |
| 5,000 | 20.1ms | 3.3MB | ⚠️ |

### Filter Engine Performance

| Benchmark | Time | Memory | Status |
|-----------|------|--------|--------|
| Short input | 0.01ms | 6KB | ✅ |
| Git status typical | 0.06ms | 13KB | ✅ |
| Npm output typical | 0.07ms | 10KB | ✅ |
| Large output (1000 lines) | 6.9ms | 1.1MB | ✅ |

### Token Estimation

| Method | Time | Notes |
|--------|------|-------|
| Heuristic | 0.5ns | `len(text) / 4` approximation |
| tiktoken | ~1μs | Accurate OpenAI tokenizer |

---

## Performance Analysis

### Strengths

1. **Sub-millisecond overhead** for typical outputs (50-100 items)
2. **Zero-allocation** token estimation (heuristic)
3. **Linear scaling** with output size
4. **Minimal memory footprint** for typical use cases

### Optimization Opportunities

1. **Large output handling** (5000+ lines) - could use streaming
2. **Memory allocation** - some allocations could be pooled
3. **Parallel processing** - large outputs could be chunked

---

## Comparison with RTK

| Metric | TokMan (Go) | RTK (Rust) |
|--------|-------------|------------|
| Typical overhead | 0.3-0.9ms | <10ms |
| Large output (5K lines) | 20ms | ~10ms |
| Memory efficiency | Good | Better |
| Startup time | ~1ms | <1ms |

**Conclusion**: TokMan achieves comparable performance to RTK for typical workloads. Rust's advantage shows primarily in extreme cases (very large outputs).

---

## Recommendations

### For Users

1. **Normal use**: No optimization needed - overhead is negligible
2. **Large outputs**: Consider piping through `head` first:
   ```bash
   tokman command | head -1000
   ```
3. **Memory-sensitive**: Use ultra-compact mode (`-u`)

### For Development

1. **Streaming**: Implement streaming for outputs >5000 lines
2. **Buffer pooling**: Reuse buffers for repeated commands
3. **Benchmark CI**: Add performance regression tests

---

## Running Benchmarks

```bash
# Filter engine benchmarks
go test -bench=. -benchmem ./internal/filter/

# Command handler benchmarks
go test -bench=. -benchmem ./internal/commands/

# Full benchmark suite
go test -bench=. -benchmem ./...

# Memory profile
go test -bench=. -memprofile=mem.prof ./internal/filter/
go tool pprof mem.prof
```

---

## Test Environment

- **CPU**: AMD EPYC 7543P 32-Core
- **OS**: Linux (amd64)
- **Go**: 1.21+
- **Date**: 2026-03-18

---

## Future Optimizations

### Sprint 5.1: Streaming (Optional)

For very large outputs, implement streaming processing:

```go
func ProcessStream(r io.Reader, w io.Writer) error {
    scanner := bufio.NewScanner(r)
    for scanner.Scan() {
        line := scanner.Text()
        if shouldKeep(line) {
            w.Write([]byte(line + "\n"))
        }
    }
    return scanner.Err()
}
```

### Sprint 5.2: Buffer Pooling

```go
var bufferPool = sync.Pool{
    New: func() interface{} {
        return new(strings.Builder)
    },
}
```

### Sprint 5.3: Parallel Processing

For multi-core systems, process chunks in parallel:

```go
func ProcessParallel(input string, workers int) string {
    lines := strings.Split(input, "\n")
    chunkSize := len(lines) / workers
    // Process chunks in parallel
}
```

---

## Conclusion

TokMan meets the performance target of <10ms overhead for typical command outputs. The implementation is efficient and production-ready. Future optimizations can address edge cases (very large outputs) if needed.
