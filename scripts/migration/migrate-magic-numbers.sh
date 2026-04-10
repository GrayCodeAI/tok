#!/bin/bash
# Migration: Replace Magic Numbers with Constants
# Improves code readability and maintainability

set -e

echo "🔢 Migrating: Magic Numbers → Constants"
echo "=========================================="

# Create constants file if not exists
if [ ! -f internal/filter/constants.go ]; then
    echo "⚠️  constants.go not found. Please ensure it exists."
    exit 1
fi

echo "📦 Creating backups..."
find internal/filter -name "*.go" -type f ! -name "*_test.go" -exec cp {} {}.backup.$(date +%s) \;

echo "🔨 Replacing magic numbers..."

# Replace magic numbers in pipeline_gates.go
if [ -f internal/filter/pipeline_gates.go ]; then
    # Budget threshold
    sed -i '' 's/< 1000/< TightBudgetThreshold/g' internal/filter/pipeline_gates.go
    
    # Early exit interval
    sed -i '' 's/% 3 != 0/% EarlyExitCheckInterval != 0/g' internal/filter/pipeline_gates.go
    
    echo "✅ Updated pipeline_gates.go"
fi

# Replace in entropy.go
if [ -f internal/filter/entropy.go ]; then
    sed -i '' 's/< 50/< MinContentLength/g' internal/filter/entropy.go
    sed -i '' 's/> 500/> MediumContentThreshold/g' internal/filter/entropy.go
    echo "✅ Updated entropy.go"
fi

# Replace in perplexity.go
if [ -f internal/filter/perplexity.go ]; then
    sed -i '' 's/< 5/< PerplexityMinLines/g' internal/filter/perplexity.go
    echo "✅ Updated perplexity.go"
fi

# Replace in streaming.go
if [ -f internal/filter/streaming.go ]; then
    sed -i '' 's/500000/StreamingThreshold/g' internal/filter/streaming.go
    sed -i '' 's/100000/ChunkSize/g' internal/filter/streaming.go
    echo "✅ Updated streaming.go"
fi

# Replace in h2o.go
if [ -f internal/filter/h2o.go ]; then
    sed -i '' 's/< 50/< H2OMinTokens/g' internal/filter/h2o.go
    echo "✅ Updated h2o.go"
fi

# Replace in attention_sink.go
if [ -f internal/filter/attention_sink.go ]; then
    sed -i '' 's/< 3/< AttentionSinkMinLines/g' internal/filter/attention_sink.go
    echo "✅ Updated attention_sink.go"
fi

# Replace in meta_token.go
if [ -f internal/filter/meta_token.go ]; then
    sed -i '' 's/< 500/< MetaTokenMinLength/g' internal/filter/meta_token.go
    echo "✅ Updated meta_token.go"
fi

# Replace in semantic_chunk.go
if [ -f internal/filter/semantic_chunk.go ]; then
    sed -i '' 's/< 300/< SemanticChunkMinLength/g' internal/filter/semantic_chunk.go
    echo "✅ Updated semantic_chunk.go"
fi

echo ""
echo "✅ Magic number migration complete"
echo ""
echo "Next steps:"
echo "1. Build: go build ./..."
echo "2. Test: go test ./internal/filter/..."
echo "3. Review changes: git diff"
