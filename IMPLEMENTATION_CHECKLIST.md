# TokMan Critical Fixes - Implementation Checklist

## Quick Commands to Apply All Fixes

### Step 1: Backup Original Files
```bash
cd /Users/lakshmanpatel/Desktop/ProjectAlpha/tokman
mkdir -p backups/$(date +%Y%m%d_%H%M%S)
cp -r internal/filter backups/$(date +%Y%m%d_%H%M%S)/
echo "✅ Backup created"
```

### Step 2: Apply Race Condition Fixes
```bash
# Add sync import to pipeline_stats.go
sed -i '' '1s/^/import "sync"\n/' internal/filter/pipeline_stats.go

# Add mutex to PipelineStats struct
sed -i '' '/type PipelineStats struct {/a\\
\tmu sync.RWMutex' internal/filter/pipeline_stats.go
```

### Step 3: Apply Constants
```bash
# Replace magic numbers in key files
sed -i '' 's/< 1000/< TightBudgetThreshold/g' internal/filter/pipeline_gates.go
sed -i '' 's/< 50/< MinContentLength/g' internal/filter/entropy.go
sed -i '' 's/500000/StreamingThreshold/g' internal/filter/streaming.go
```

### Step 4: Verify
```bash
go build ./...
go test -race ./internal/filter/... -count=1
echo "✅ All fixes applied and verified"
```

---

## Detailed Fix List

### Fix 1: Race Conditions (CRITICAL)

**Files to modify:**
- `internal/filter/pipeline_stats.go`
- `internal/filter/pipeline_gates.go`

**Changes:**
1. Add `sync.RWMutex` to `PipelineStats`
2. Add `sync/atomic` import
3. Replace direct field access with thread-safe methods

**Code:**
```go
// Add to pipeline_stats.go
type PipelineStats struct {
    OriginalTokens int
    FinalTokens    int
    // ... other fields
    
    mu           sync.RWMutex        // NEW
    runningSaved int64               // NEW - atomic
}

// Add thread-safe methods
func (s *PipelineStats) AddLayerStatSafe(name string, stat LayerStat) {
    s.mu.Lock()
    defer s.mu.Unlock()
    if s.LayerStats == nil {
        s.LayerStats = make(map[string]LayerStat)
    }
    s.LayerStats[name] = stat
    atomic.AddInt64(&s.runningSaved, int64(stat.TokensSaved))
}
```

---

### Fix 2: Nil Safety (CRITICAL)

**Files to modify:**
- `internal/filter/pipeline_process.go`
- All filter initialization files

**Changes:**
1. Add nil checks before calling filter methods
2. Add panic recovery

**Code:**
```go
// In processLayer, before calling Apply:
if layer.filter == nil {
    return input, 0
}

// With panic recovery:
defer func() {
    if r := recover(); r != nil {
        // Log and continue
        output = input
        saved = 0
    }
}()
```

---

### Fix 3: Magic Numbers (HIGH)

**Files to modify:**
- `internal/filter/pipeline_gates.go`
- `internal/filter/entropy.go`
- `internal/filter/perplexity.go`
- `internal/filter/streaming.go`
- `internal/filter/h2o.go`

**Replace:**
```go
// From:
if len(content) < 50
if tokens < 1000
if size > 500000

// To:
if len(content) < MinContentLength
if tokens < TightBudgetThreshold
if size > StreamingThreshold
```

---

## Verification Steps

### 1. Build Check
```bash
go build ./...
```

### 2. Race Detection
```bash
go test -race ./internal/filter/... -count=100
```

### 3. Benchmark Comparison
```bash
# Before fixes
go test -bench=BenchmarkPipeline -benchmem ./internal/filter/... > before.txt

# After fixes
go test -bench=BenchmarkPipeline -benchmem ./internal/filter/... > after.txt

# Compare
diff before.txt after.txt
```

### 4. Test Suite
```bash
go test ./... -v 2>&1 | grep -E "(PASS|FAIL)"
```

---

## Rollback Instructions

If anything breaks:

```bash
# Restore from backup
cp -r backups/20260111_120000/filter internal/

# Verify
go build ./...
go test ./internal/filter/...
```

---

## Success Criteria

- [ ] `go build ./...` passes
- [ ] `go test -race ./...` passes with 0 races
- [ ] All existing tests pass
- [ ] No performance regression (>5%)
- [ ] Zero nil pointer panics in logs

---

## Next Steps After Fixes

1. **Monitor production** for 24 hours
2. **Collect metrics** on performance improvements
3. **Apply Week 2 refactoring** if Week 1 is stable
4. **Document lessons learned**

---

**Ready to implement?** Run the commands in Step 1-4 above, or apply fixes manually using the detailed instructions.
