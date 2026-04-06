#!/bin/bash

echo "Analyzing All Competitors..."
echo "============================"
echo ""

for repo in rtk context-compressor cntxtpy cntxtjs tokenpacker snip token-optimizer-mcp tore tokenreduction lightcompress toonify omni zon-format pact awesome-collection; do
    if [ -d "$repo" ]; then
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo "📦 Repository: $repo"
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        
        cd "$repo"
        
        # Check language
        echo "Language:"
        if [ -f "package.json" ]; then
            echo "  → JavaScript/TypeScript (Node.js)"
        elif [ -f "setup.py" ] || [ -f "pyproject.toml" ] || [ -f "requirements.txt" ]; then
            echo "  → Python"
        elif [ -f "go.mod" ]; then
            echo "  → Go"
        elif [ -f "Cargo.toml" ]; then
            echo "  → Rust"
        elif [ -f "pom.xml" ] || [ -f "build.gradle" ]; then
            echo "  → Java"
        else
            echo "  → Unknown (check files)"
        fi
        
        # Check README
        echo ""
        echo "README Found:"
        ls -1 README* 2>/dev/null | head -5 || echo "  → No README found"
        
        # Count files
        echo ""
        echo "File Count:"
        echo "  → $(find . -type f | grep -v ".git" | wc -l) files"
        
        # Last commit
        echo ""
        echo "Last Commit:"
        git log -1 --format="  → %ar (%ad)" 2>/dev/null || echo "  → Unable to check"
        
        # Stars (from git remote)
        echo ""
        
        cd ..
        echo ""
    fi
done

