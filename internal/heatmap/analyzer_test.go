package heatmap

import (
	"strings"
	"testing"
)

func TestClassifyLine(t *testing.T) {
	analyzer := NewSectionAnalyzer()

	tests := []struct {
		line     string
		expected SectionType
	}{
		{"<system>You are a helpful assistant", SectionSystem},
		{"your role is to help", SectionSystem},
		{"<tool>bash", SectionTools},
		{"tool_call: read_file", SectionTools},
		{"<context>file: main.go", SectionContext},
		{"```go\nfunc main()", SectionContext},
		{"<history>previous conversation", SectionHistory},
		{"user: what is 2+2", SectionHistory},
		{"assistant: 4", SectionHistory},
		{"fix the bug in this code", SectionQuery},
		{"", SectionUnknown},
	}

	for _, tt := range tests {
		result := analyzer.ClassifyLine(tt.line)
		if result != tt.expected {
			t.Errorf("ClassifyLine(%q) = %v, want %v", tt.line, result, tt.expected)
		}
	}
}

func TestAnalyze(t *testing.T) {
	analyzer := NewSectionAnalyzer()
	input := "<system>You are a helpful assistant\nYour role is to help with code\n<tool>bash\ntool_call: ls\n<context>file: main.go\nfunc main() {}\nuser: fix this\nassistant: ok\nfix the bug please"

	sections := analyzer.Analyze(input)
	if len(sections) == 0 {
		t.Fatal("Expected sections, got none")
	}

	totalTokens := 0
	for _, s := range sections {
		totalTokens += s.TokenCount
		if s.Percentage < 0 || s.Percentage > 100 {
			t.Errorf("Invalid percentage: %.2f", s.Percentage)
		}
	}

	if totalTokens == 0 {
		t.Error("Expected non-zero tokens")
	}
}

func TestHeatmapGenerator(t *testing.T) {
	gen := NewHeatmapGenerator()
	input := "system prompt here\nuser query here\nassistant response"

	data := gen.Generate(input)
	if data.TotalTokens == 0 {
		t.Error("Expected non-zero tokens")
	}

	jsonData, err := gen.ToJSON(data)
	if err != nil {
		t.Fatalf("ToJSON error: %v", err)
	}
	if len(jsonData) == 0 {
		t.Error("Expected JSON output")
	}

	csvData := gen.ToCSV(data)
	if !strings.Contains(csvData, "type,token_count") {
		t.Error("Expected CSV header")
	}

	summary := gen.Summary(data)
	if !strings.Contains(summary, "Total Tokens") {
		t.Error("Expected summary to contain Total Tokens")
	}
}

func TestEstimateTokens(t *testing.T) {
	tests := []struct {
		text      string
		minTokens int
	}{
		{"", 0},
		{"hello", 1},
		{"hello world foo bar", 4},
		{strings.Repeat("word ", 100), 100},
	}

	for _, tt := range tests {
		got := estimateTokens(tt.text)
		if got < tt.minTokens {
			t.Errorf("estimateTokens(%q) = %d, want >= %d", tt.text, got, tt.minTokens)
		}
	}
}
