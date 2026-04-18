# TokMan Quick Reference Card

## 🚨 Top 3 Critical Issues

### 1. Global State (100+ variables)
**File:** `internal/commands/root.go`
**Problem:** Cannot test in parallel, race conditions
**Fix:** Use config struct + dependency injection
**Effort:** 2 weeks

### 2. Memory Waste (50MB per pipeline)
**File:** `internal/filter/pipeline_init.go`
**Problem:** All filters initialized even if disabled
**Fix:** Lazy initialization
**Effort:** 3 days

### 3. No Panic Recovery
**File:** `cmd/tokman/main.go`
**Problem:** Crashes on panic
**Fix:** Add `defer recover()`
**Effort:** 1 hour

---

## ⚡ Quick Wins (< 1 day each)

| Issue | File | Fix | Speedup |
|-------|------|-----|---------|
| Regex compilation | `internal/toml/filter.go` | Pre-compile with `sync.Once` | 10-100x |
| No config validation | `internal/config/config.go` | Add `Validate()` method | Better UX |
| Inaccurate tokens | `internal/core/estimator.go` | Use tiktoken sampling | 50% → 95% |
| No panic recovery | `cmd/tokman/main.go` | Add `defer recover()` | Stability |

---

## 📊 Performance Targets

| Metric | Current | Target | How |
|--------|---------|--------|-----|
| Small input | 883μs | 420μs | Lazy init + parallel |
| Medium input | 8.2ms | 2.8ms | Lazy init + parallel |
| Large input | 82ms | 28ms | Lazy init + parallel + streaming |
| Memory | 50MB | 5-10MB | Lazy initialization |
| Tracker batch | 500ms | 50ms | Batch inserts |

---

## 🏗️ Architecture Patterns

### ❌ Don't Do This
```go
// Global variables
var verbose int
var dryRun bool

// Shared mutable state
package shared
var rootCmd *cobra.Command
func SetFlags(cfg FlagConfig) { ... }
```

### ✅ Do This Instead
```go
// Config struct
type GlobalConfig struct {
    Verbose int
    DryRun  bool
}

// Dependency injection
func NewCommand(cfg *GlobalConfig) *cobra.Command { ... }

// Context-based state
ctx = WithConfig(ctx, cfg)
cfg := GetConfig(ctx)
```

---

## 🔧 Common Patterns

### Lazy Initialization
```go
type PipelineCoordinator struct {
    entropyFilter *EntropyFilter
    initOnce      sync.Once
}

func (p *PipelineCoordinator) getEntropyFilter() *EntropyFilter {
    if p.entropyFilter == nil && p.config.EnableEntropy {
        p.entropyFilter = NewEntropyFilter()
    }
    return p.entropyFilter
}
```

### Parallel Processing
```go
var wg sync.WaitGroup
results := make(chan result, len(layers))

for _, layer := range layers {
    wg.Add(1)
    go func(l filterLayer) {
        defer wg.Done()
        output, saved := l.filter.Apply(input, mode)
        results <- result{output, saved}
    }(layer)
}

wg.Wait()
close(results)
```

### Batch Operations
```go
batch := make([]*Record, 0, 100)
ticker := time.NewTicker(5 * time.Second)

for {
    select {
    case record := <-ch:
        batch = append(batch, record)
        if len(batch) >= 100 {
            flushBatch(batch)
            batch = batch[:0]
        }
    case <-ticker.C:
        if len(batch) > 0 {
            flushBatch(batch)
            batch = batch[:0]
        }
    }
}
```

---

## 📁 File Organization

```
tokman/
├── cmd/tokman/main.go              # Entry point (add panic recovery)
├── internal/
│   ├── commands/
│   │   ├── root.go                 # 🔴 100+ global vars (refactor)
│   │   └── shared/shared.go        # 🔴 Global state (remove)
│   ├── filter/
│   │   ├── pipeline_init.go        # 🔴 Eager init (make lazy)
│   │   ├── pipeline_process.go     # 🟠 Sequential (parallelize)
│   │   └── *.go                    # 20+ layer implementations
│   ├── config/
│   │   └── config.go               # 🟡 No validation (add)
│   ├── core/
│   │   ├── runner.go               # 🟡 No limits (add)
│   │   └── estimator.go            # 🟡 Inaccurate (fix)
│   ├── tracking/
│   │   └── tracker.go              # 🟠 Single inserts (batch)
│   └── toml/
│       └── filter.go               # 🟡 Regex recompile (cache)
```

---

## 🧪 Testing Checklist

### Before Refactoring
- [ ] Run `make test` (baseline)
- [ ] Run `make benchmark` (baseline)
- [ ] Check memory usage
- [ ] Verify no race conditions

### After Refactoring
- [ ] All tests pass
- [ ] Benchmarks improved
- [ ] Memory reduced
- [ ] No new race conditions
- [ ] Coverage increased

---

## 📈 Metrics to Track

```bash
# Performance
make benchmark

# Memory
go test -memprofile=mem.prof
go tool pprof mem.prof

# Race conditions
go test -race ./...

# Coverage
make test-cover
```

---

## 🎯 Implementation Order

### Week 1: Quick Wins
1. ✅ Panic recovery (1 hour)
2. ✅ Pre-compile regex (4 hours)
3. ✅ Config validation (1 day)
4. ✅ Lazy initialization (3 days)

### Week 2-3: Architecture
5. ✅ Remove global state (2 weeks)
   - Create GlobalConfig struct
   - Refactor all commands
   - Update tests

### Week 4-5: Performance
6. ✅ Parallel processing (1 week)
7. ✅ Streaming API (1 week)
8. ✅ Batch inserts (2 days)

### Week 6-7: Quality
9. ✅ Accurate token estimation (2 days)
10. ✅ Integration tests (1 week)
11. ✅ Structured logging (3 days)

### Week 8: Polish
12. ✅ Metrics export (2 days)
13. ✅ Documentation (3 days)

---

## 🔍 Code Review Checklist

### Architecture
- [ ] No global variables
- [ ] Dependency injection used
- [ ] Context for state passing
- [ ] Interfaces for testability

### Performance
- [ ] Lazy initialization
- [ ] Parallel where possible
- [ ] Streaming for large inputs
- [ ] Batch operations
- [ ] Caching enabled

### Quality
- [ ] Config validated
- [ ] Errors handled
- [ ] Panic recovery
- [ ] Tests added
- [ ] Benchmarks run

### Security
- [ ] Input sanitized
- [ ] Output size limited
- [ ] Timeouts set
- [ ] Signatures verified

---

## 💻 Useful Commands

```bash
# Build
make build

# Test
make test
make test-cover
go test -race ./...

# Benchmark
make benchmark
go test -bench=. -benchmem ./internal/filter

# Profile
go test -cpuprofile=cpu.prof -bench=.
go tool pprof cpu.prof

# Lint
make lint

# Format
make fmt

# Check everything
make check
```

---

## 📚 Further Reading

- [Part 1: Entry Point](./TOKMAN_ANALYSIS_PART1_ENTRY_POINT.md)
- [Part 2: Command System](./TOKMAN_ANALYSIS_PART2_COMMAND_SYSTEM.md)
- [Part 3: Pipeline](./TOKMAN_ANALYSIS_PART3_PIPELINE.md)
- [Part 7: Summary](./TOKMAN_ANALYSIS_PART7_SUMMARY.md)
- [Master Index](./TOKMAN_COMPLETE_ANALYSIS_INDEX.md)

---

**Print this card and keep it handy while refactoring!**
