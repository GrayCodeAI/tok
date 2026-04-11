package filter

import (
	"strings"
	"testing"
)

func TestFilterProgressBars(t *testing.T) {
	input := `Downloading packages...
[========================================] 100% (2.5 MB/s)
Installation complete
`
	result := FilterProgressBars(input)
	if !strings.Contains(result, "Downloading") {
		t.Error("should keep non-progress lines")
	}
	if !strings.Contains(result, "Installation complete") {
		t.Error("should keep completion line")
	}
}

func TestFilterProgressBars_NoProgress(t *testing.T) {
	input := "line1\nline2\nline3\n"
	result := FilterProgressBars(input)
	if result == "" {
		t.Error("should return non-empty output")
	}
}

func TestFilterNoisyOutput(t *testing.T) {
	input := "\rDownloading...\rProgress: 50%\rProgress: 100%\nDone\n\n\n\n"
	result := FilterNoisyOutput(input)
	// Should remove \r carriage returns
	if strings.Contains(result, "\r") {
		t.Error("should strip carriage returns")
	}
	// Should collapse blank lines
	if strings.Count(result, "\n\n\n") > 0 {
		t.Error("should collapse 3+ blank lines")
	}
	if !strings.Contains(result, "Done") {
		t.Error("should keep meaningful content")
	}
}

func TestFilterNoisyOutput_Empty(t *testing.T) {
	result := FilterNoisyOutput("")
	if result != "" {
		t.Errorf("empty input should return empty, got: %q", result)
	}
}

func TestFilterNoisyOutput_ANSI(t *testing.T) {
	input := "\x1b[32mDownloading\x1b[0m \x1b[33m50%\x1b[0m\nDone\n"
	result := FilterNoisyOutput(input)
	// Should strip ANSI codes
	if strings.Contains(result, "\x1b[") {
		t.Error("should strip ANSI codes")
	}
}
