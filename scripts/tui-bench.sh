#!/usr/bin/env bash
# Run the TUI performance benchmarks and save the report to artifacts/.
#
# The benchmarks in internal/tui/bench_test.go establish targets:
#   BrailleLineChart_Wide     < 250 µs/op
#   TableRender_1000Rows      < 6 ms/op
#   ModelView_FullFrame       < 16 ms/op  (single-frame @ 60 fps)
#
# Run from repo root; writes artifacts/tui-bench.txt.

set -euo pipefail

cd "$(dirname "$0")/.."

mkdir -p artifacts
go test -bench=. -benchmem -run='^$' -benchtime=500ms ./internal/tui/ \
    | tee artifacts/tui-bench.txt

echo ""
echo "Targets (indicative):"
echo "  BrailleLineChart_Wide  < 250µs"
echo "  TableRender_1000Rows   <   6ms"
echo "  ModelView_FullFrame    <  16ms"
