#!/bin/bash
# Performance profiling script for Tok
# Usage: ./scripts/profile.sh [cpu|mem|trace] [duration]

set -e

PROFILE_TYPE=${1:-cpu}
DURATION=${2:-30s}
OUTPUT_DIR=${3:-./profiles}

echo "🔍 Tok Performance Profiling"
echo "═══════════════════════════════════════════════════════"
echo "Profile Type: $PROFILE_TYPE"
echo "Duration: $DURATION"
echo "Output Directory: $OUTPUT_DIR"
echo ""

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Timestamp for filenames
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
OUTPUT_FILE="$OUTPUT_DIR/${PROFILE_TYPE}_${TIMESTAMP}.prof"

# Build with profiling support
echo "📦 Building with profiling support..."
go build -o /tmp/tok-profiled ./cmd/tok

case $PROFILE_TYPE in
    cpu)
        echo "🚀 Running CPU profiler..."
        /tmp/tok-profiled profile --cpu --duration="$DURATION" --output="$OUTPUT_FILE"
        echo ""
        echo "📊 Analyzing CPU profile..."
        go tool pprof -top -cum "$OUTPUT_FILE" | head -30
        echo ""
        echo "💡 View with: go tool pprof -http=:8080 $OUTPUT_FILE"
        ;;
    
    mem)
        echo "🧠 Running memory profiler..."
        /tmp/tok-profiled profile --mem --duration="$DURATION" --output="$OUTPUT_FILE"
        echo ""
        echo "📊 Analyzing memory profile..."
        go tool pprof -top -cum "$OUTPUT_FILE" | head -30
        echo ""
        echo "💡 View with: go tool pprof -http=:8080 $OUTPUT_FILE"
        ;;
    
    trace)
        echo "🎯 Running execution tracer..."
        TRACE_FILE="$OUTPUT_DIR/trace_${TIMESTAMP}.out"
        /tmp/tok-profiled profile --trace --duration="$DURATION" --output="$TRACE_FILE"
        echo ""
        echo "💡 View with: go tool trace $TRACE_FILE"
        ;;
    
    rewrite)
        echo "🔄 Profiling rewrite system..."
        go test -bench=BenchmarkRewriteCommand -benchmem -cpuprofile="$OUTPUT_DIR/rewrite_cpu.prof" -memprofile="$OUTPUT_DIR/rewrite_mem.prof" ./internal/discover/
        echo ""
        echo "📊 CPU Hotspots:"
        go tool pprof -top "$OUTPUT_DIR/rewrite_cpu.prof" | head -20
        echo ""
        echo "📊 Memory Allocations:"
        go tool pprof -top "$OUTPUT_DIR/rewrite_mem.prof" | head -20
        ;;
    
    quota)
        echo "💰 Profiling quota calculations..."
        go test -bench=BenchmarkQuota -benchmem -cpuprofile="$OUTPUT_DIR/quota_cpu.prof" -memprofile="$OUTPUT_DIR/quota_mem.prof" ./internal/commands/core/
        echo ""
        echo "📊 Results:"
        go tool pprof -top "$OUTPUT_DIR/quota_cpu.prof" | head -20
        ;;
    
    all)
        echo "🔥 Running all profilers..."
        $0 cpu "$DURATION" "$OUTPUT_DIR"
        $0 mem "$DURATION" "$OUTPUT_DIR"
        $0 rewrite "" "$OUTPUT_DIR"
        $0 quota "" "$OUTPUT_DIR"
        ;;
    
    *)
        echo "❌ Unknown profile type: $PROFILE_TYPE"
        echo "Usage: $0 [cpu|mem|trace|rewrite|quota|all] [duration] [output_dir]"
        exit 1
        ;;
esac

echo ""
echo "✅ Profiling complete!"
echo "Output saved to: $OUTPUT_DIR"
echo ""
echo "📁 Files:"
ls -lh "$OUTPUT_DIR"/*.prof "$OUTPUT_DIR"/*.out 2>/dev/null || true
