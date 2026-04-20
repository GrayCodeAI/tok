package tui

import (
	"fmt"
	"math"
	"strings"
)

// BrailleLineChart renders a sequence of values as a multi-row Braille
// line chart. Each Braille glyph encodes a 2×4 dot matrix, so one
// character column carries two x-samples and the chart height in cells
// × 4 is the effective y-resolution.
//
// The chart does not draw axes; callers overlay labels themselves. We
// take this tradeoff because sections need full control over surrounding
// layout and because Braille axes are visually noisy at small sizes.
//
// width is the target character width; height is the number of character
// rows. Width must be ≥ 4; height ≥ 2. Smaller sizes degrade gracefully
// to a single-row sparkline. Non-UTF-8 terminals should use
// BrailleLineChartASCII instead — Braille can't be faked in ASCII.
func BrailleLineChart(values []float64, width, height int) string {
	if len(values) == 0 || width < 4 {
		return ""
	}
	if height < 2 {
		return BrailleSparkline(values, width)
	}

	// Downsample or stretch the values to 2*width x-samples.
	xSamples := width * 2
	samples := resample(values, xSamples)

	// Find y-range for scaling.
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

	rows := height
	yCells := rows * 4

	// Build a 2D grid of 8-bit Braille dots indexed [row][col].
	grid := make([][]rune, rows)
	for i := range grid {
		grid[i] = make([]rune, width)
		for j := range grid[i] {
			grid[i][j] = 0x2800 // blank Braille cell
		}
	}

	for i, v := range samples {
		col := i / 2
		subx := i % 2
		norm := (v - minV) / span
		if norm < 0 {
			norm = 0
		}
		if norm > 1 {
			norm = 1
		}
		yAbs := int(math.Round(norm * float64(yCells-1)))
		if yAbs >= yCells {
			yAbs = yCells - 1
		}
		// Braille origin is top-left of the cell; our y axis runs bottom-up.
		rowFromBottom := yAbs / 4
		subyFromBottom := yAbs % 4
		row := rows - 1 - rowFromBottom
		if row < 0 || row >= rows || col < 0 || col >= width {
			continue
		}
		grid[row][col] |= brailleDot(subx, subyFromBottom)
	}

	lines := make([]string, rows)
	for i, row := range grid {
		lines[i] = string(row)
	}
	return strings.Join(lines, "\n")
}

// BrailleSparkline is the 1-row variant — same bins as a classic
// block-glyph sparkline but higher vertical resolution thanks to the
// 4-dot Braille cell.
func BrailleSparkline(values []float64, width int) string {
	if len(values) == 0 || width <= 0 {
		return ""
	}
	samples := resample(values, width*2)
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
	var b strings.Builder
	for col := 0; col < width; col++ {
		cell := rune(0x2800)
		for sub := 0; sub < 2; sub++ {
			idx := col*2 + sub
			if idx >= len(samples) {
				break
			}
			norm := (samples[idx] - minV) / span
			y := int(math.Round(norm * 3))
			if y < 0 {
				y = 0
			}
			if y > 3 {
				y = 3
			}
			cell |= brailleDot(sub, y)
		}
		b.WriteRune(cell)
	}
	return b.String()
}

// brailleDot maps (sub-column 0/1, sub-row 0..3 from bottom) to the
// bit within the Braille cell that lights that dot.
//
// Braille dot numbering (ISO 11548):
//
//	   col 0  col 1
//	row 3:  1     4
//	row 2:  2     5
//	row 1:  3     6
//	row 0:  7     8
//
// Converted to bit offsets from 0x2800:
//
//	dot 1 → 0x01     dot 4 → 0x08
//	dot 2 → 0x02     dot 5 → 0x10
//	dot 3 → 0x04     dot 6 → 0x20
//	dot 7 → 0x40     dot 8 → 0x80
//
// Our y axis runs bottom (y=0) to top (y=3) so we map row 0→dot 7/8,
// row 1→dot 3/6, row 2→dot 2/5, row 3→dot 1/4.
func brailleDot(subx, suby int) rune {
	switch {
	case subx == 0 && suby == 0:
		return 0x40 // dot 7
	case subx == 0 && suby == 1:
		return 0x04 // dot 3
	case subx == 0 && suby == 2:
		return 0x02 // dot 2
	case subx == 0 && suby == 3:
		return 0x01 // dot 1
	case subx == 1 && suby == 0:
		return 0x80 // dot 8
	case subx == 1 && suby == 1:
		return 0x20 // dot 6
	case subx == 1 && suby == 2:
		return 0x10 // dot 5
	case subx == 1 && suby == 3:
		return 0x08 // dot 4
	}
	return 0
}

// resample stretches or compresses a value slice to exactly n elements
// using nearest-neighbor sampling. Good enough for visualization where
// we're just preserving shape, not exact magnitudes.
func resample(values []float64, n int) []float64 {
	if n <= 0 {
		return nil
	}
	if len(values) == 0 {
		return make([]float64, n)
	}
	if len(values) == n {
		out := make([]float64, n)
		copy(out, values)
		return out
	}
	out := make([]float64, n)
	for i := range out {
		idx := int(math.Round(float64(i) * float64(len(values)-1) / float64(n-1)))
		if idx < 0 {
			idx = 0
		}
		if idx >= len(values) {
			idx = len(values) - 1
		}
		out[i] = values[idx]
	}
	return out
}

// LineChart returns BrailleLineChart on UTF-8 terminals and an
// ASCII-only substitute otherwise. Sections that care about unicode
// availability should call this helper rather than BrailleLineChart
// directly.
func LineChart(values []float64, width, height int, utf8 bool) string {
	if utf8 {
		return BrailleLineChart(values, width, height)
	}
	return asciiLineChart(values, width, height)
}

// asciiLineChart draws a multi-row line plot with only ASCII chars.
// It quantizes each sample to one of `height*2` vertical buckets and
// places '*' / '+' at the correct (row, col). Less dense than Braille
// but readable everywhere — and correct in a non-UTF-8 locale where
// Braille would render as ? boxes.
func asciiLineChart(values []float64, width, height int) string {
	if len(values) == 0 || width < 4 || height < 1 {
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
	rows := height
	yMax := rows - 1

	grid := make([][]rune, rows)
	for i := range grid {
		grid[i] = make([]rune, width)
		for j := range grid[i] {
			grid[i][j] = ' '
		}
	}
	for col, v := range samples {
		norm := (v - minV) / span
		if norm < 0 {
			norm = 0
		}
		if norm > 1 {
			norm = 1
		}
		y := int(norm * float64(yMax))
		if y < 0 {
			y = 0
		}
		if y > yMax {
			y = yMax
		}
		row := rows - 1 - y
		if row >= 0 && row < rows && col < width {
			grid[row][col] = '*'
		}
	}
	lines := make([]string, rows)
	for i, row := range grid {
		lines[i] = string(row)
	}
	return strings.Join(lines, "\n")
}

// FormatChartRange renders a "min ··· max" label sized for the given
// width, useful as a caption below a BrailleLineChart.
func FormatChartRange(values []float64, width int) string {
	if len(values) == 0 || width < 6 {
		return ""
	}
	minV, maxV := values[0], values[0]
	for _, v := range values {
		if v < minV {
			minV = v
		}
		if v > maxV {
			maxV = v
		}
	}
	return fmt.Sprintf("%.0f %s %.0f", minV, strings.Repeat("·", max(1, width-20)), maxV)
}
