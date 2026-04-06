# 🚀 Competitive Features - Quick Start Guide

This guide shows how to use TokMan's unique competitive features that set it apart from all other token compression tools.

## 📊 1. Automatic Quality Scoring

### Basic Usage

Analyze the quality of compression:

```bash
# From file
tokman quality input.txt

# From stdin
cat file.txt | tokman quality

# From command output
tokman git diff | tokman quality
```

**Example Output:**
```
Overall Quality: 87.3% (B+)

Breakdown:
  • Compression Ratio: 85.0%
  • Keywords Preserved: 92.0%
  • Structure Intact: 88.5%
  • Readability: 84.0%
  • Information Density: 79.0%
  • Semantic Preserved: 86.0%

Recommendations:
  ✅ Excellent compression quality!
```

### Compare All Modes

Compare minimal vs aggressive compression:

```bash
tokman quality --compare-all file.txt
```

**Output:**
```
🔍 Comparing All Compression Modes...

Processing Minimal mode... 10,000 tok → 3,000 tok (saved 7,000, 70.0%)
Processing Aggressive mode... 10,000 tok → 1,500 tok (saved 8,500, 85.0%)

📊 QUALITY COMPARISON:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
🏆 Minimal: 87.3% (B+)
     Compression: 70.0% | Keywords: 92.0% | Readability: 88.0%
   Aggressive: 82.1% (B-)
     Compression: 85.0% | Keywords: 78.0% | Readability: 72.0%

✅ RECOMMENDATION:
Use 'Minimal' mode for best quality (87.3% overall score)
```

**Use this to:**
- Choose the best compression mode
- Validate compression quality
- Get actionable recommendations
- Ensure important content is preserved

---

## 🎨 2. Visual Diff Tool

### Show Before/After Comparison

See exactly what changed with color-coded output:

```bash
# Show visual diff
tokman quality --diff input.txt

# Or from stdin
cat file.txt | tokman quality --diff
```

**Example Output:**
```
╔═══════════════════════════════════════════════════════╗
║           📊 VISUAL COMPRESSION COMPARISON 📊         ║
╚═══════════════════════════════════════════════════════╝

Original:    10,000 tokens
Compressed:  1,500 tokens
Saved:       8,500 tokens (85.0%)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
LINE-BY-LINE COMPARISON
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

✓   1 │ package main
✗   2 │ // This comment was removed
~   3 │ func ProcessData(data string) string {
       → func Process(d string) str {
✓   4 │     return result

[██████████████████████████████░░░░░░] 85.0% saved

KEY CHANGES:
  ✓ 45 lines kept verbatim
  ~ 23 lines compressed/modified
  ✗ 108 lines removed completely
```

### Export as HTML

Create a web-viewable diff:

```bash
tokman quality --html output.html input.txt
```

Opens in browser to show:
- Side-by-side comparison
- Color-coded changes
- Token statistics
- Interactive viewing

**Use this to:**
- Understand compression impact
- Document what was removed/kept
- Share results with team
- Debug compression issues

---

## 🔗 3. Multi-File Context Merging

### Merge Multiple Files

Combine files intelligently:

```bash
# Merge all .go files in current directory
tokman merge *.go

# Merge recursively
tokman merge -r src/

# With token budget
tokman merge --max-tokens 5000 src/**/*.go
```

**Example Output:**
```
📁 Merging 15 files...
📊 Total size: 245KB (65,000 tokens)
🔄 Compressing to meet 5000 token budget...
✅ Final: 4,850 tokens

═══════════════════════════════════════
File: src/main.go
═══════════════════════════════════════

package main
...

═══════════════════════════════════════
File: src/utils.go
═══════════════════════════════════════

package utils
...
```

### Intelligent Dependency-Aware Ordering

Let TokMan analyze and order files optimally:

```bash
tokman merge --intelligent src/*.go
```

**What it does:**
- Analyzes import/dependency relationships
- Orders files in logical sequence
- Deduplicates across files
- Preserves context for AI understanding

### Multiple Output Formats

```bash
# Markdown (default)
tokman merge *.go

# XML
tokman merge --format xml src/*.go

# JSON
tokman merge --format json src/*.go
```

**XML Output:**
```xml
<?xml version="1.0" encoding="UTF-8"?>
<merged-context>
  <metadata>
    <file-count>5</file-count>
  </metadata>
  <content><![CDATA[
    ... merged content ...
  ]]></content>
</merged-context>
```

**JSON Output:**
```json
{
  "files": [
    "src/main.go",
    "src/utils.go"
  ],
  "content": "... merged content ..."
}
```

**Use this to:**
- Merge entire codebases for AI
- Stay within token budgets
- Preserve cross-file context
- Automate context preparation

---

## 🎯 Real-World Use Cases

### Use Case 1: Code Review Preparation

```bash
# Get all changed files from git, merge and analyze quality
git diff --name-only | xargs tokman merge --max-tokens 8000 | tokman quality
```

### Use Case 2: AI Assistant Context

```bash
# Merge entire feature into one context for Claude/GPT
tokman merge -r src/features/new-feature/ --max-tokens 10000 > context.txt
```

### Use Case 3: Documentation Quality Check

```bash
# Check if compression preserves key documentation
tokman quality --diff --html report.html docs/*.md
```

### Use Case 4: Multi-File Debugging

```bash
# Merge relevant files for debugging session
tokman merge --intelligent \
  src/controllers/*.go \
  src/models/*.go \
  src/services/*.go \
  --max-tokens 15000
```

### Use Case 5: CI/CD Quality Gates

```bash
# Fail build if compression quality is too low
quality_score=$(tokman quality input.txt | grep "Overall" | awk '{print $3}' | tr -d '%')
if [ "$quality_score" -lt 80 ]; then
  echo "Compression quality too low: $quality_score%"
  exit 1
fi
```

---

## 📊 Command Reference

### Quality Command

```bash
# Basic quality analysis
tokman quality [file]

# Show visual diff
tokman quality --diff [file]

# Compare all modes
tokman quality --compare-all [file]

# Export HTML
tokman quality --html output.html [file]

# From stdin
command | tokman quality
```

### Merge Command

```bash
# Basic merge
tokman merge [files...]

# Recursive
tokman merge -r [directory]

# With budget
tokman merge --max-tokens N [files...]

# Intelligent ordering
tokman merge --intelligent [files...]

# Custom format
tokman merge --format [markdown|xml|json] [files...]

# Disable headers
tokman merge --headers=false [files...]
```

---

## 🏆 Why These Features Matter

### vs LLMLingua (Microsoft)

| Feature | TokMan | LLMLingua |
|---------|--------|-----------|
| Quality scoring | ✅ Full metrics | ❌ Perplexity only |
| Visual diff | ✅ Color-coded | ❌ None |
| Multi-file | ✅ Intelligent | ❌ Single file |

### vs AutoCompressor (Princeton/MIT)

| Feature | TokMan | AutoCompressor |
|---------|--------|----------------|
| Quality scoring | ✅ 6 metrics | ❌ None |
| Visual diff | ✅ Yes | ❌ No |
| Multi-file | ✅ Yes | ❌ No |
| Production ready | ✅ Yes | ❌ Academic only |

### vs Generic Tools

| Feature | TokMan | Generic |
|---------|--------|---------|
| Quality analysis | ✅ Comprehensive | ❌ Token count |
| Visual comparison | ✅ Yes | ❌ No |
| Multi-file merging | ✅ Yes | ❌ No |
| Recommendations | ✅ Actionable | ❌ None |

---

## 💡 Pro Tips

### 1. Always Check Quality

Before using compressed output:
```bash
tokman quality --compare-all input.txt
```

### 2. Use Visual Diff for Understanding

See what was removed:
```bash
tokman quality --diff input.txt | less -R
```

### 3. Merge with Budget for AI

Stay within context limits:
```bash
tokman merge -r src/ --max-tokens 20000 > ai-context.txt
```

### 4. Export Reports for Team

Share compression results:
```bash
tokman quality --html report.html src/**/*.go
```

### 5. Intelligent Ordering Matters

Use `--intelligent` for complex codebases:
```bash
tokman merge --intelligent src/**/*.go
```

---

## 🎓 Learning More

- Read [COMPETITIVE_FEATURES.md](COMPETITIVE_FEATURES.md) for detailed comparison
- See [README.md](../README.md) for general TokMan usage
- Check [CONTRIBUTING.md](../CONTRIBUTING.md) for development guide

---

## 🐛 Troubleshooting

### Quality Score is Low

**Problem:** Getting quality scores below 70%

**Solutions:**
- Try `--compare-all` to find best mode
- Use `--diff` to see what's being removed
- Reduce compression with `--mode minimal`
- Increase budget with `--budget N`

### Merge Exceeds Token Budget

**Problem:** Merged output too large

**Solutions:**
- Reduce `--max-tokens` value
- TokMan will automatically compress to fit
- Use more aggressive compression
- Filter files more selectively

### Files Not Ordering Correctly

**Problem:** Dependency order is wrong

**Solutions:**
- Use `--intelligent` flag
- Check import statements are correct
- Manually specify order if needed

---

<div align="center">

**TokMan: The only tool with quality scoring, visual diff, and intelligent merging.**

[GitHub →](https://github.com/GrayCodeAI/tokman)

</div>
