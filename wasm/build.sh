#!/bin/bash
# Build TokMan WASM module

set -e

echo "Building TokMan WASM module..."

# Copy wasm_exec.js from Go
cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" ./wasm_exec.js

# Build WASM
GOOS=js GOARCH=wasm go build -o tokman.wasm ./main.go

# Optimize with wasm-opt if available
if command -v wasm-opt &> /dev/null; then
    echo "Optimizing with wasm-opt..."
    wasm-opt -Oz tokman.wasm -o tokman.wasm
fi

echo "Build complete!"
ls -lh tokman.wasm wasm_exec.js tokman.js
