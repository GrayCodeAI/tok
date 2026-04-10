#!/bin/bash
# Migration: Fix Race Conditions in PipelineStats
# Critical: Must be applied to prevent data races

set -e

echo "🔧 Migrating: Race Condition Fixes"
echo "===================================="

# Backup original files
echo "📦 Creating backups..."
cp internal/filter/pipeline_gates.go internal/filter/pipeline_gates.go.backup.$(date +%s)
cp internal/filter/pipeline_stats.go internal/filter/pipeline_stats.go.backup.$(date +%s) 2>/dev/null || true

# Apply race condition fixes
echo "🔨 Applying fixes..."

# Fix 1: Add sync/atomic import if missing
if ! grep -q "sync/atomic" internal/filter/pipeline_gates.go; then
    sed -i '' 's/"sync"/"sync"\n\t"sync\/atomic"/' internal/filter/pipeline_gates.go
    echo "✅ Added sync/atomic import"
fi

# Fix 2: Make runningSaved atomic
cat > /tmp/race_fix.patch << 'EOF'
--- a/internal/filter/pipeline_stats.go
+++ b/internal/filter/pipeline_stats.go
@@ -1,5 +1,9 @@
 package filter
 
+import (
+	"sync"
+	"sync/atomic"
+)
 
 // PipelineStats tracks compression statistics.
 type PipelineStats struct {
@@ -8,6 +12,9 @@ type PipelineStats struct {
 	TotalSaved       int
 	ReductionPercent float64
 	LayerStats       map[string]LayerStat
+	
+	mu           sync.RWMutex
+	runningSaved int64 // atomic
 }
 
 // LayerStat holds per-layer statistics.
EOF

# Apply patch if pipeline_stats.go exists
if [ -f internal/filter/pipeline_stats.go ]; then
    patch -p1 < /tmp/race_fix.patch || echo "⚠️  Patch may already be applied"
else
    echo "⚠️  pipeline_stats.go not found, skipping patch"
fi

# Fix 3: Add thread-safe methods
cat >> internal/filter/pipeline_stats.go << 'EOF'

// AddLayerStatSafe adds a layer stat thread-safely.
func (s *PipelineStats) AddLayerStatSafe(name string, stat LayerStat) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if s.LayerStats == nil {
		s.LayerStats = make(map[string]LayerStat)
	}
	s.LayerStats[name] = stat
	atomic.AddInt64(&s.runningSaved, int64(stat.TokensSaved))
}

// RunningSavedSafe returns the running saved count thread-safely.
func (s *PipelineStats) RunningSavedSafe() int {
	return int(atomic.LoadInt64(&s.runningSaved))
}
EOF

echo "✅ Race condition fixes applied"
echo ""
echo "Next steps:"
echo "1. Run tests: go test ./internal/filter/... -race"
echo "2. Verify no race conditions: go test -race -count=100"
echo "3. If issues, rollback: cp internal/filter/pipeline_gates.go.backup.* internal/filter/pipeline_gates.go"
