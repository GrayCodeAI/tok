package tui

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestBrailleLineChartDimensions(t *testing.T) {
	values := []float64{1, 3, 2, 7, 5, 4, 8, 2, 6, 9}
	chart := BrailleLineChart(values, 20, 5)
	lines := strings.Split(chart, "\n")
	if len(lines) != 5 {
		t.Fatalf("expected 5 rows, got %d", len(lines))
	}
	for i, line := range lines {
		if utf8.RuneCountInString(line) != 20 {
			t.Fatalf("row %d width = %d runes, want 20", i, utf8.RuneCountInString(line))
		}
		for _, r := range line {
			if r < 0x2800 || r > 0x28FF {
				t.Fatalf("row %d contains non-Braille rune %U", i, r)
			}
		}
	}
}

func TestBrailleSparklineFallsBackForShortHeight(t *testing.T) {
	// height=1 should route through sparkline path and return one row.
	chart := BrailleLineChart([]float64{1, 2, 3, 4}, 10, 1)
	if strings.Contains(chart, "\n") {
		t.Fatalf("height=1 chart should be single row, got:\n%s", chart)
	}
	if utf8.RuneCountInString(chart) != 10 {
		t.Fatalf("width mismatch: got %d runes, want 10", utf8.RuneCountInString(chart))
	}
}

func TestBrailleLineChartHandlesFlatSeries(t *testing.T) {
	// All values equal — no division-by-zero panic and chart still
	// renders with Braille cells.
	chart := BrailleLineChart([]float64{5, 5, 5, 5, 5}, 12, 3)
	if chart == "" {
		t.Fatal("flat series should still render")
	}
}

func TestResampleInvariants(t *testing.T) {
	cases := []struct {
		name string
		in   []float64
		n    int
	}{
		{"upsample", []float64{1, 2, 3}, 9},
		{"downsample", []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, 3},
		{"identity", []float64{1, 2, 3}, 3},
		{"empty", nil, 5},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := resample(tc.in, tc.n)
			if len(got) != tc.n {
				t.Fatalf("%s: len(got) = %d, want %d", tc.name, len(got), tc.n)
			}
		})
	}
}
