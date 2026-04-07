// Package visual provides visual comparison tools for compression results.
// Competitive feature vs other tools that only show token counts.
package visual

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
)

// DiffResult represents a visual comparison of before/after compression.
type DiffResult struct {
	OriginalLines   int
	CompressedLines int
	LinesRemoved    int
	LinesChanged    int
	LinesKept       int

	OriginalTokens   int
	CompressedTokens int
	TokensSaved      int

	Highlights []DiffHighlight
}

// DiffHighlight represents a notable change in the compression.
type DiffHighlight struct {
	Type           string // "removed", "kept", "changed", "compressed"
	OriginalLine   string
	CompressedLine string
	Explanation    string
}

// VisualDiff creates a side-by-side comparison with highlights.
func VisualDiff(original, compressed string, originalTokens, compressedTokens int) string {
	var output strings.Builder

	// Header
	output.WriteString(color.New(color.FgCyan, color.Bold).Sprint("╔═══════════════════════════════════════════════════════════════╗\n"))
	output.WriteString(color.New(color.FgCyan, color.Bold).Sprint("║           📊 VISUAL COMPRESSION COMPARISON 📊                 ║\n"))
	output.WriteString(color.New(color.FgCyan, color.Bold).Sprint("╚═══════════════════════════════════════════════════════════════╝\n\n"))

	// Stats summary
	tokensSaved := originalTokens - compressedTokens
	savingsPercent := 0.0
	if originalTokens > 0 {
		savingsPercent = (float64(tokensSaved) / float64(originalTokens)) * 100
	}

	output.WriteString(fmt.Sprintf("Original:    %s tokens\n",
		color.RedString("%d", originalTokens)))
	output.WriteString(fmt.Sprintf("Compressed:  %s tokens\n",
		color.GreenString("%d", compressedTokens)))
	output.WriteString(fmt.Sprintf("Saved:       %s tokens (%s%%)\n\n",
		color.YellowString("%d", tokensSaved),
		color.YellowString("%.1f", savingsPercent)))

	// Line-by-line comparison
	origLines := strings.Split(original, "\n")
	compLines := strings.Split(compressed, "\n")

	output.WriteString(color.New(color.FgBlue, color.Bold).Sprint("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"))
	output.WriteString(color.New(color.FgBlue, color.Bold).Sprint("LINE-BY-LINE COMPARISON\n"))
	output.WriteString(color.New(color.FgBlue, color.Bold).Sprint("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n"))

	// Show first 10 lines comparison
	maxLines := min(10, max(len(origLines), len(compLines)))

	for i := 0; i < maxLines; i++ {
		origLine := ""
		compLine := ""

		if i < len(origLines) {
			origLine = truncate(origLines[i], 50)
		}
		if i < len(compLines) {
			compLine = truncate(compLines[i], 50)
		}

		// Determine change type
		if origLine == compLine {
			// Kept unchanged
			output.WriteString(fmt.Sprintf("%s %3d │ %s\n",
				color.GreenString("✓"), i+1, origLine))
		} else if compLine == "" {
			// Line removed
			output.WriteString(fmt.Sprintf("%s %3d │ %s\n",
				color.RedString("✗"), i+1,
				color.RedString(origLine)))
		} else if origLine == "" {
			// Line added (shouldn't happen in compression)
			output.WriteString(fmt.Sprintf("%s %3d │ %s\n",
				color.GreenString("+"), i+1,
				color.GreenString(compLine)))
		} else {
			// Line changed/compressed
			output.WriteString(fmt.Sprintf("%s %3d │ %s\n",
				color.YellowString("~"), i+1, origLine))
			output.WriteString(fmt.Sprintf("       → %s\n",
				color.CyanString(compLine)))
		}
	}

	if len(origLines) > maxLines || len(compLines) > maxLines {
		output.WriteString(fmt.Sprintf("\n... (%d more lines)\n",
			max(len(origLines), len(compLines))-maxLines))
	}

	// Key changes summary
	output.WriteString(fmt.Sprintf("\n%s\n",
		color.New(color.FgMagenta, color.Bold).Sprint("KEY CHANGES:")))

	removedCount := countRemovedLines(origLines, compLines)
	changedCount := countChangedLines(origLines, compLines)
	keptCount := countKeptLines(origLines, compLines)

	output.WriteString(fmt.Sprintf("  %s %d lines kept verbatim\n",
		color.GreenString("✓"), keptCount))
	output.WriteString(fmt.Sprintf("  %s %d lines compressed/modified\n",
		color.YellowString("~"), changedCount))
	output.WriteString(fmt.Sprintf("  %s %d lines removed completely\n",
		color.RedString("✗"), removedCount))

	return output.String()
}

// CompactDiff creates a compact one-line comparison.
func CompactDiff(originalTokens, compressedTokens int) string {
	saved := originalTokens - compressedTokens
	percent := 0.0
	if originalTokens > 0 {
		percent = (float64(saved) / float64(originalTokens)) * 100
	}

	return fmt.Sprintf("%s → %s (saved %s, %.1f%%)",
		color.RedString("%d tok", originalTokens),
		color.GreenString("%d tok", compressedTokens),
		color.YellowString("%d", saved),
		percent)
}

// DiffStats creates a visual bar chart of compression.
func DiffStats(originalTokens, compressedTokens int) string {
	if originalTokens == 0 {
		return ""
	}

	saved := originalTokens - compressedTokens
	percent := (float64(saved) / float64(originalTokens)) * 100

	// Create progress bar
	barWidth := 50
	savedBars := int((percent / 100.0) * float64(barWidth))
	keptBars := barWidth - savedBars

	var bar strings.Builder
	bar.WriteString(color.RedString(strings.Repeat("█", keptBars)))
	bar.WriteString(color.GreenString(strings.Repeat("█", savedBars)))

	return fmt.Sprintf("[%s] %.1f%% saved", bar.String(), percent)
}

// ShowHighlights displays key compression highlights.
func ShowHighlights(original, compressed string) []string {
	highlights := []string{}

	// Check for specific compression patterns
	if strings.Contains(original, "error:") && !strings.Contains(compressed, "error:") {
		highlights = append(highlights, "⚠️  Warning: error messages were removed")
	}

	if countWords(original) > countWords(compressed)*2 {
		highlights = append(highlights, "✓ Aggressive word reduction applied")
	}

	if strings.Count(original, "\n") > strings.Count(compressed, "\n")*2 {
		highlights = append(highlights, "✓ Significant line deduplication")
	}

	// Check for preserved structure
	if strings.Contains(original, "{") && strings.Contains(compressed, "{") {
		highlights = append(highlights, "✓ Code structure preserved")
	}

	return highlights
}

// Helper functions

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func countWords(s string) int {
	return len(strings.Fields(s))
}

func countRemovedLines(original, compressed []string) int {
	count := 0
	compSet := make(map[string]bool)

	for _, line := range compressed {
		compSet[strings.TrimSpace(line)] = true
	}

	for _, line := range original {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !compSet[trimmed] {
			count++
		}
	}

	return count
}

func countChangedLines(original, compressed []string) int {
	count := 0
	maxLen := min(len(original), len(compressed))

	for i := 0; i < maxLen; i++ {
		orig := strings.TrimSpace(original[i])
		comp := strings.TrimSpace(compressed[i])

		if orig != "" && comp != "" && orig != comp {
			count++
		}
	}

	return count
}

func countKeptLines(original, compressed []string) int {
	count := 0
	maxLen := min(len(original), len(compressed))

	for i := 0; i < maxLen; i++ {
		orig := strings.TrimSpace(original[i])
		comp := strings.TrimSpace(compressed[i])

		if orig != "" && comp != "" && orig == comp {
			count++
		}
	}

	return count
}

// AnimatedDiff shows an animated compression visualization (for terminals that support it).
func AnimatedDiff(originalTokens, compressedTokens int, durationMs int) string {
	// For now, just return a static version
	// TODO: Implement actual animation with term control codes
	return DiffStats(originalTokens, compressedTokens)
}

// ExportDiffHTML exports diff as HTML for web viewing.
func ExportDiffHTML(original, compressed string, originalTokens, compressedTokens int) string {
	var html strings.Builder

	html.WriteString(`<!DOCTYPE html>
<html>
<head>
	<title>TokMan Compression Diff</title>
	<style>
		body { font-family: monospace; margin: 20px; }
		.original { background: #ffe0e0; }
		.compressed { background: #e0ffe0; }
		.removed { text-decoration: line-through; color: #a00; }
		.kept { color: #0a0; }
		.changed { background: #fffacd; }
		.stats { background: #f0f0f0; padding: 10px; margin: 10px 0; }
	</style>
</head>
<body>
	<h1>🔍 Compression Comparison</h1>
`)

	saved := originalTokens - compressedTokens
	percent := 0.0
	if originalTokens > 0 {
		percent = (float64(saved) / float64(originalTokens)) * 100
	}

	html.WriteString(fmt.Sprintf(`
	<div class="stats">
		<p><strong>Original:</strong> %d tokens</p>
		<p><strong>Compressed:</strong> %d tokens</p>
		<p><strong>Saved:</strong> %d tokens (%.1f%%)</p>
	</div>
`, originalTokens, compressedTokens, saved, percent))

	html.WriteString(`<h2>Original</h2><pre class="original">`)
	html.WriteString(htmlEscape(original))
	html.WriteString(`</pre>`)

	html.WriteString(`<h2>Compressed</h2><pre class="compressed">`)
	html.WriteString(htmlEscape(compressed))
	html.WriteString(`</pre>`)

	html.WriteString(`</body></html>`)

	return html.String()
}

func htmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	return s
}
