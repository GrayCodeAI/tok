package tui

import (
	"runtime"
	"strings"
	"testing"
)

func TestDetectUTF8FromLang(t *testing.T) {
	// On darwin/linux the no-env-vars case defaults to true (modern terminals
	// are UTF-8 by default). On other platforms it defaults to false.
	emptyWant := runtime.GOOS == "darwin" || runtime.GOOS == "linux" ||
		runtime.GOOS == "freebsd" || runtime.GOOS == "openbsd" || runtime.GOOS == "netbsd"

	cases := []struct {
		name string
		env  map[string]string
		want bool
	}{
		{"explicit utf-8", map[string]string{"LANG": "en_US.UTF-8"}, true},
		{"explicit utf8", map[string]string{"LC_ALL": "en_US.utf8"}, true},
		{"POSIX", map[string]string{"LANG": "POSIX"}, false},
		{"latin1", map[string]string{"LANG": "en_US.ISO-8859-1"}, false},
		{"empty", map[string]string{}, emptyWant},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Clear the vars we inspect, then set only what the case wants.
			for _, k := range []string{"LC_ALL", "LC_CTYPE", "LANG"} {
				t.Setenv(k, "")
			}
			for k, v := range tc.env {
				t.Setenv(k, v)
			}
			if got := detectUTF8(); got != tc.want {
				t.Errorf("detectUTF8() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestSparklineASCIIFallback(t *testing.T) {
	values := []int64{1, 2, 3, 4, 5}
	got := sparklineGlyphs(values, false)
	// ASCII fallback glyph set is ".-=#"; output must use only those.
	for _, r := range got {
		if !strings.ContainsRune(".-=#", r) {
			t.Fatalf("expected ASCII fallback glyphs only, got %q (rune %U)", got, r)
		}
	}
}

func TestLineChartASCIIFallback(t *testing.T) {
	values := []float64{1, 3, 2, 5, 4, 7}
	chart := LineChart(values, 20, 4, false)
	lines := strings.Split(chart, "\n")
	if len(lines) != 4 {
		t.Fatalf("ascii line chart rows = %d, want 4", len(lines))
	}
	// Should contain at least one '*' and no Braille codepoint.
	if !strings.ContainsRune(chart, '*') {
		t.Fatalf("ascii chart missing '*':\n%s", chart)
	}
	for _, r := range chart {
		if r >= 0x2800 && r <= 0x28FF {
			t.Fatalf("ascii chart contained Braille rune %U", r)
		}
	}
}

func TestRenderBarASCIIGlyphs(t *testing.T) {
	th := newTheme()
	bar := renderBarGlyphs(th, 5, 10, 10, 0, false)
	// Strip ANSI for safer containment check — but since we use '#' and
	// '-' the raw check suffices: neither U+2588 nor U+2591 should show.
	for _, r := range bar {
		if r == '█' || r == '░' {
			t.Fatalf("ascii bar leaked Unicode glyph in output: %q", bar)
		}
	}
	if !strings.ContainsRune(bar, '#') {
		t.Fatalf("ascii bar missing '#':\n%s", bar)
	}
}
