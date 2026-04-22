# tok Code Review TODO

## 🔴 CRITICAL (Fix First)

### 1. Duplicate Output Printing
**File:** `internal/commands/shared/executor.go`
**Lines:** 115-126
**Issue:** `out.Global().Print(filtered)` called twice for every successful command.
**Fix:** Remove the duplicate block.

### 2. Out-of-Bounds Slice Access
**File:** `internal/filter/six_layer_pipeline.go`
**Issue:** Hardcoded `p.layers[0]` through `p.layers[5]` with no bounds checking. If `buildLayers()` hasn't run or returns fewer layers → panic.
**Fix:** Bounds-check before indexing or use named accessors.

### 3. Nil PipelineCoordinator Process
**File:** `internal/filter/pipeline_process.go`
**Issue:** `Process()` returns a stats struct but `runLayer1Preprocess` etc. may access `p.layers` on a nil receiver if caller doesn't check.
**Fix:** Already has nil check at top — verify all call sites handle it.

## 🟠 HIGH (Fix Second)

### 4. SQLite Connection Thrashing
**File:** `internal/commands/shared/executor.go:31`
**Issue:** `tracking.NewTracker()` opens a 25-connection pool for every command recording.
**Fix:** Reuse `tracking.GetGlobalTracker()` singleton.

### 5. Unbounded Global Cache
**File:** `internal/cache/cache.go`
**Issue:** `FingerprintCache` has no size limit or eviction — permanent memory leak.
**Fix:** Add TTL + LRU eviction or size cap.

### 6. Regex Recompilation on Every Call
**File:** `internal/security/scanner.go:267`
**Issue:** `IsSuspiciousContent()` compiles 10 regexes from scratch every invocation.
**Fix:** Use `sync.Once` or package-level precompiled vars.

### 7. Tee File Rotation Bug
**File:** `internal/commands/shared/fallback.go:393`
**Issue:** `os.ReadDir` order is undefined; deletes random files instead of oldest.
**Fix:** Sort by `ModTime` before deleting.

### 8. Config Loaded Repeatedly
**File:** `internal/tracking/tracker.go:101`, `internal/tee/tee.go`
**Issue:** `config.Load("")` reads disk + env on every call.
**Fix:** Cache config after first load.

## 🟡 MEDIUM (Fix Third)

### 9. syncFromGlobals Called on Every Access
**File:** `internal/commands/shared/globals.go`
**Issue:** Every flag accessor (IsVerbose, IsUltraCompact, etc.) triggers full state sync under mutex.
**Fix:** Cache synced state; eliminate dual-copy pattern.

### 10. Progress Callback Race
**File:** `internal/filter/progress.go:6`
**Issue:** Package-level `ProgressCallback` var with no mutex; set/restored from goroutines.
**Fix:** Protect with `atomic.Value` or mutex.

### 11. Chunk Joiner Inflates Tokens
**File:** `internal/filter/manager.go:187`
**Issue:** `\n\n--- Chunk Boundary ---\n\n` adds 11 tokens per chunk.
**Fix:** Use minimal delimiter `\n---\n`.

### 12. runGuardrailFallback Expensive Copy
**File:** `internal/filter/pipeline_process.go:48`
**Issue:** Copies entire 100+ field `PipelineConfig` struct by value.
**Fix:** Pass pointer and mutate selectively.

### 13. RecordCommand Stores Full Output
**File:** `internal/commands/shared/executor.go`
**Issue:** `OriginalOutput` and `FilteredOutput` fields store megabytes of text per row.
**Fix:** Truncate or hash large outputs; store metadata only.

### 14. RunAndCapture Pipe Handling
**File:** `internal/commands/shared/executor.go`
**Issue:** `Wait()` before reading errCh; pipes never explicitly closed.
**Fix:** Use `io.ReadAll` or drain pipes properly.

### 15. RedactPII Does Not Sort Findings
**File:** `internal/security/scanner.go`
**Issue:** Overlapping findings in non-deterministic order cause skipped redactions.
**Fix:** Sort findings by Position before iterating.

## 🔵 ARCHITECTURAL (Refactor)

### 16. Massive PipelineConfig Struct
**File:** `internal/filter/pipeline_types.go`
**Issue:** 100+ fields passed by value; mixes concerns.
**Fix:** Split into nested structs or builder pattern.

### 17. Dual Global State System
**File:** `internal/commands/shared/state.go` + `globals.go`
**Issue:** Two parallel copies of every flag with sync dance.
**Fix:** Eliminate package globals; read from AppState directly.

### 18. PipelineCoordinator Rebuilt Per Request
**File:** `internal/commands/shared/fallback.go:285`
**Issue:** New coordinator + 20+ filter allocations per command.
**Fix:** Pool coordinators or make reusable.

### 19. Layer Index Coupling
**File:** `internal/filter/pipeline_init.go`, `six_layer_pipeline.go`
**Issue:** Hardcoded indices shift when layers are added.
**Fix:** Named constants or map-based lookup.

## 🛡️ SECURITY (Audit)

### 20. saveTee Path Traversal
**File:** `internal/filter/manager.go:287`
**Issue:** `strings.HasPrefix` for path checks has edge cases with symlinks.
**Fix:** Use `filepath.EvalSymlinks` before prefix check.

### 21. validateCommandName Only Checks Binary
**File:** `internal/core/runner.go`
**Issue:** Arguments are sanitized (control chars stripped) but shell meta-chars in args are allowed.
**Fix:** Validate all args, not just command name.
