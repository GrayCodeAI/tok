#!/bin/bash

echo "Deep Feature Analysis of All Competitors"
echo "========================================"
echo ""

# Function to analyze a repo
analyze_repo() {
    local repo=$1
    local name=$2
    
    if [ ! -d "$repo" ]; then
        echo "⚠️  $name: Not found"
        return
    fi
    
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "📦 Analyzing: $name"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    
    cd "$repo"
    
    # Check project structure
    echo ""
    echo "Project Structure:"
    ls -1 | head -20
    
    # Check for key files
    echo ""
    echo "Key Files:"
    echo "  Makefile: $([ -f Makefile ] && echo '✅' || echo '❌')"
    echo "  Dockerfile: $([ -f Dockerfile ] && echo '✅' || echo '❌')"
    echo "  .goreleaser.yml: $([ -f .goreleaser.yml ] && echo '✅' || echo '❌')"
    echo "  Cargo.toml: $([ -f Cargo.toml ] && echo '✅' || echo '❌')"
    echo "  package.json: $([ -f package.json ] && echo '✅' || echo '❌')"
    echo "  CI/CD: $([ -d .github/workflows ] && echo '✅' || echo '❌')"
    
    # Check documentation
    echo ""
    echo "Documentation:"
    find docs -name "*.md" 2>/dev/null | head -10 || echo "  No docs/ folder"
    
    # Check tests
    echo ""
    echo "Tests:"
    find . -name "*_test.*" -o -name "test_*" 2>/dev/null | head -5 | wc -l | xargs echo "  Test files:"
    
    # Check source structure
    echo ""
    echo "Source Structure:"
    if [ -d "src" ]; then
        echo "  src/ layout:"
        ls -1 src/ | head -10
    fi
    if [ -d "internal" ]; then
        echo "  internal/ layout:"
        ls -1 internal/ | head -10
    fi
    if [ -d "pkg" ]; then
        echo "  pkg/ layout:"
        ls -1 pkg/ | head -10
    fi
    
    cd ..
    echo ""
}

# Analyze top competitors
analyze_repo "rtk" "RTK (Rust)"
analyze_repo "snip" "Snip (Go)"
analyze_repo "context-compressor" "Context-Compressor (Python)"
analyze_repo "token-optimizer-mcp" "Token-Optimizer-MCP (JS)"

