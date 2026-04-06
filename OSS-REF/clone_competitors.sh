#!/bin/bash

echo "Cloning all token reduction competitors..."
echo "=========================================="
echo ""

# High Priority Competitors
echo "🔴 HIGH PRIORITY COMPETITORS:"
echo ""

echo "1. Cloning RTK..."
git clone https://github.com/rtk-ai/rtk.git rtk 2>&1 | tail -3

echo "2. Cloning Context-Compressor..."
git clone https://github.com/Huzaifa785/context-compressor.git context-compressor 2>&1 | tail -3

echo "3. Cloning CntxtPY..."
git clone https://github.com/brandondocusen/CntxtPY.git cntxtpy 2>&1 | tail -3

echo "4. Cloning CntxtJS..."
git clone https://github.com/brandondocusen/CntxtJS.git cntxtjs 2>&1 | tail -3

echo "5. Cloning TokenPacker..."
git clone https://github.com/CircleRadon/TokenPacker.git tokenpacker 2>&1 | tail -3

echo "6. Cloning Snip..."
git clone https://github.com/edouard-claude/snip.git snip 2>&1 | tail -3

echo "7. Cloning Token-Optimizer-MCP..."
git clone https://github.com/ooples/token-optimizer-mcp.git token-optimizer-mcp 2>&1 | tail -3

echo ""
echo "🟡 MEDIUM PRIORITY COMPETITORS:"
echo ""

echo "8. Cloning TORE..."
git clone https://github.com/Frank-ZY-Dou/TORE.git tore 2>&1 | tail -3

echo "9. Cloning TokenReduction..."
git clone https://github.com/JoakimHaurum/TokenReduction.git tokenreduction 2>&1 | tail -3

echo "10. Cloning LightCompress..."
git clone https://github.com/ModelTC/LightCompress.git lightcompress 2>&1 | tail -3

echo "11. Cloning Toonify..."
git clone https://github.com/ScrapeGraphAI/toonify.git toonify 2>&1 | tail -3

echo "12. Cloning Omni..."
git clone https://github.com/fajarhide/omni.git omni 2>&1 | tail -3

echo "13. Cloning ZON-Format..."
git clone https://github.com/ZON-Format/zon-TS.git zon-format 2>&1 | tail -3

echo "14. Cloning PACT..."
git clone https://github.com/orailix/PACT.git pact 2>&1 | tail -3

echo ""
echo "📚 COLLECTION:"
echo ""

echo "15. Cloning Awesome-Collection-Token-Reduction..."
git clone https://github.com/ZLKong/Awesome-Collection-Token-Reduction.git awesome-collection 2>&1 | tail -3

echo ""
echo "=========================================="
echo "✅ Cloning complete!"
echo ""
echo "Cloned repositories:"
ls -1

