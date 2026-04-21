package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// BrailleHistogram renders a distribution as a histogram using Braille patterns.
// buckets: counts per bucket, labels: optional labels for each bucket.
func BrailleHistogram(buckets []int, labels []string, width, height int, utf8 bool) string {
	if len(buckets) == 0 || width < 4 || height < 2 {
		return ""
	}

	// Find max for scaling
	maxVal := 0
	for _, v := range buckets {
		if v > maxVal {
			maxVal = v
		}
	}
	if maxVal == 0 {
		return strings.Repeat("░", width)
	}

	if utf8 {
		return brailleHistogramUTF8(buckets, labels, width, height, maxVal)
	}
	return asciiHistogram(buckets, labels, width, height, maxVal)
}

func brailleHistogramUTF8(buckets []int, labels []string, width, height int, maxVal int) string {
	// Braille patterns for different heights (quadrants)
	blocks := []rune{' ', '⠂', '⠆', '⠖', '⠶', '⡶', '⡾', '⡿', '⣿'}

	cols := min(len(buckets), width/2)
	if cols < 1 {
		cols = 1
	}

	// Group buckets into columns
	groupSize := len(buckets) / cols
	if groupSize < 1 {
		groupSize = 1
	}

	result := make([]string, height)
	for row := 0; row < height; row++ {
		var line strings.Builder
		for col := 0; col < cols; col++ {
			// Get max value in this column's bucket group
			startIdx := col * groupSize
			endIdx := min(startIdx+groupSize, len(buckets))
			colMax := 0
			for i := startIdx; i < endIdx; i++ {
				if buckets[i] > colMax {
					colMax = buckets[i]
				}
			}

			// Calculate which block to show based on height position
			normalized := float64(colMax) / float64(maxVal)
			blockIdx := int(normalized * float64(len(blocks)-1))
			if blockIdx >= len(blocks) {
				blockIdx = len(blocks) - 1
			}

			// Invert for top-down rendering
			rowThreshold := float64(height-row-1) / float64(height)
			if normalized >= rowThreshold {
				line.WriteRune(blocks[blockIdx])
				line.WriteRune(' ') // spacing
			} else {
				line.WriteString("  ")
			}
		}
		result[row] = line.String()
	}

	return strings.Join(result, "\n")
}

func asciiHistogram(buckets []int, labels []string, width, height int, maxVal int) string {
	chars := []byte{' ', '.', ':', '+', '*', '#'}

	cols := min(len(buckets), width/2)
	if cols < 1 {
		cols = 1
	}

	groupSize := len(buckets) / cols
	if groupSize < 1 {
		groupSize = 1
	}

	result := make([]string, height)
	for row := 0; row < height; row++ {
		var line strings.Builder
		for col := 0; col < cols; col++ {
			startIdx := col * groupSize
			endIdx := min(startIdx+groupSize, len(buckets))
			colMax := 0
			for i := startIdx; i < endIdx; i++ {
				if buckets[i] > colMax {
					colMax = buckets[i]
				}
			}

			normalized := float64(colMax) / float64(maxVal)
			charIdx := int(normalized * float64(len(chars)-1))
			if charIdx >= len(chars) {
				charIdx = len(chars) - 1
			}

			rowThreshold := float64(height-row-1) / float64(height)
			if normalized >= rowThreshold {
				line.WriteByte(chars[charIdx])
				line.WriteByte(' ')
			} else {
				line.WriteString("  ")
			}
		}
		result[row] = line.String()
	}

	return strings.Join(result, "\n")
}

// StackedBar renders a horizontal stacked bar chart.
// segments: values for each segment, labels: names for each segment.
func StackedBar(segments []int, labels []string, width int, th theme, utf8 bool) string {
	if len(segments) == 0 || width < 10 {
		return ""
	}

	total := 0
	for _, v := range segments {
		total += v
	}
	if total == 0 {
		return strings.Repeat("░", width)
	}

	// Calculate segment widths
	segmentWidths := make([]int, len(segments))
	remaining := width
	for i, v := range segments {
		if i == len(segments)-1 {
			segmentWidths[i] = remaining
		} else {
			w := int(float64(v) * float64(width) / float64(total))
			if w < 1 && v > 0 {
				w = 1
			}
			segmentWidths[i] = w
			remaining -= w
		}
	}

	// Build the bar
	var bar strings.Builder
	fillChar := '█'
	if !utf8 {
		fillChar = '#'
	}

	accentColors := th.AccentColors
	for i, w := range segmentWidths {
		if w <= 0 {
			continue
		}
		color := accentColors[i%len(accentColors)]
		segment := strings.Repeat(string(fillChar), w)
		bar.WriteString(lipgloss.NewStyle().Foreground(color).Render(segment))
	}

	// Build legend
	var legend strings.Builder
	for i, label := range labels {
		if i >= len(segments) {
			break
		}
		if i > 0 {
			legend.WriteString("  ")
		}
		color := accentColors[i%len(accentColors)]
		legend.WriteString(lipgloss.NewStyle().Foreground(color).Render("●"))
		legend.WriteString(fmt.Sprintf(" %s (%d%%)", label, segments[i]*100/total))
	}

	return bar.String() + "\n" + legend.String()
}

// SparklineBars renders a mini bar chart for trend data.
func SparklineBars(values []float64, width int, utf8 bool) string {
	if len(values) == 0 || width < 3 {
		return ""
	}

	samples := resample(values, width)

	minV, maxV := samples[0], samples[0]
	for _, v := range samples {
		if v < minV {
			minV = v
		}
		if v > maxV {
			maxV = v
		}
	}
	span := maxV - minV
	if span <= 0 {
		span = 1
	}

	var result strings.Builder
	if utf8 {
		bars := []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}
		for _, v := range samples {
			normalized := (v - minV) / span
			idx := int(normalized * float64(len(bars)-1))
			if idx < 0 {
				idx = 0
			}
			if idx >= len(bars) {
				idx = len(bars) - 1
			}
			result.WriteRune(bars[idx])
		}
	} else {
		chars := []byte{'.', ':', '+', '*', '#'}
		for _, v := range samples {
			normalized := (v - minV) / span
			idx := int(normalized * float64(len(chars)-1))
			if idx < 0 {
				idx = 0
			}
			if idx >= len(chars) {
				idx = len(chars) - 1
			}
			result.WriteByte(chars[idx])
		}
	}

	return result.String()
}
