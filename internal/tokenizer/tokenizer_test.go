package tokenizer

import (
	"strings"
	"testing"
)

func TestModelToEncoding_MapCoverage(t *testing.T) {
	if len(ModelToEncoding) < 20 {
		t.Errorf("expected at least 20 model mappings, got %d", len(ModelToEncoding))
	}

	tests := []struct {
		model    string
		expected Encoding
	}{
		{"gpt-4o", O200kBase},
		{"gpt-4o-mini", O200kBase},
		{"gpt-4", Cl100kBase},
		{"gpt-3.5-turbo", Cl100kBase},
		{"text-embedding-ada-002", Cl100kBase},
		{"davinci", P50kBase},
		{"claude-3-opus", Cl100kBase},
		{"claude-3.5-sonnet", Cl100kBase},
	}

	for _, tt := range tests {
		t.Run(tt.model, func(t *testing.T) {
			enc, ok := ModelToEncoding[tt.model]
			if !ok {
				t.Fatalf("model %s not found in map", tt.model)
			}
			if enc != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, enc)
			}
		})
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		enc     Encoding
		wantErr bool
	}{
		{Cl100kBase, false},
		{O200kBase, false},
		{P50kBase, false},
		{R50kBase, false},
		{Encoding("invalid"), true},
	}

	for _, tt := range tests {
		t.Run(string(tt.enc), func(t *testing.T) {
			tok, err := New(tt.enc)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tok == nil {
				t.Fatal("expected tokenizer, got nil")
			}
			if tok.encName != tt.enc {
				t.Errorf("expected encoding %s, got %s", tt.enc, tok.encName)
			}
		})
	}
}

func TestNewForModel(t *testing.T) {
	tests := []struct {
		model    string
		expected Encoding
	}{
		{"gpt-4o", O200kBase},
		{"gpt-4", Cl100kBase},
		{"davinci", P50kBase},
		{"unknown-model", Cl100kBase}, // defaults to cl100k_base
		{"", Cl100kBase},              // defaults to cl100k_base
	}

	for _, tt := range tests {
		t.Run(tt.model, func(t *testing.T) {
			tok, err := NewForModel(tt.model)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tok.encName != tt.expected {
				t.Errorf("expected encoding %s for model %s, got %s", tt.expected, tt.model, tok.encName)
			}
		})
	}
}

func TestTokenizer_Count(t *testing.T) {
	tok, err := New(Cl100kBase)
	if err != nil {
		t.Fatalf("failed to create tokenizer: %v", err)
	}

	tests := []struct {
		name     string
		input    string
		minCount int
	}{
		{"empty", "", 0},
		{"single word", "hello", 1},
		{"sentence", "The quick brown fox jumps over the lazy dog", 9},
		{"code snippet", "func main() { fmt.Println(\"hello\") }", 10},
		{"multiline", "line1\nline2\nline3", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count := tok.Count(tt.input)
			if count < tt.minCount {
				t.Errorf("expected at least %d tokens for %q, got %d", tt.minCount, tt.input, count)
			}
		})
	}
}

func TestTokenizer_Count_Empty(t *testing.T) {
	tok, err := New(Cl100kBase)
	if err != nil {
		t.Fatalf("failed to create tokenizer: %v", err)
	}
	if count := tok.Count(""); count != 0 {
		t.Errorf("expected 0 for empty string, got %d", count)
	}
}

func TestEstimateTokens(t *testing.T) {
	tests := []struct {
		input    string
		minCount int
	}{
		{"", 0},
		{"hello", 1},
		{"a b c d e f g h i j", 2},
	}

	for _, tt := range tests {
		count := EstimateTokens(tt.input)
		if count < tt.minCount {
			t.Errorf("expected at least %d for %q, got %d", tt.minCount, tt.input, count)
		}
	}
}

func TestCompareCounts(t *testing.T) {
	text := "The quick brown fox jumps over the lazy dog"
	heuristic, actual, diff := CompareCounts(text)

	if heuristic <= 0 {
		t.Error("expected positive heuristic count")
	}
	if actual <= 0 {
		t.Error("expected positive actual count")
	}
	if diff == 0 && heuristic != actual {
		t.Error("expected non-zero difference when counts differ")
	}
}

func TestCompareCounts_Empty(t *testing.T) {
	heuristic, actual, diff := CompareCounts("")
	if heuristic != 0 {
		t.Errorf("expected 0 heuristic for empty, got %d", heuristic)
	}
	if actual != 0 {
		t.Errorf("expected 0 actual for empty, got %d", actual)
	}
	if diff != 0 {
		t.Errorf("expected 0 diff for empty, got %f", diff)
	}
}

func TestCountStats_Summary(t *testing.T) {
	stats := &CountStats{
		TotalTokens: 1000,
		TotalChars:  4000,
		TotalLines:  50,
		FilesCount:  3,
		Encoding:    Cl100kBase,
	}

	summary := stats.Summary()
	if !strings.Contains(summary, "1000") {
		t.Error("summary should contain token count")
	}
	if !strings.Contains(summary, "4000") {
		t.Error("summary should contain char count")
	}
	if !strings.Contains(summary, "cl100k_base") {
		t.Error("summary should contain encoding name")
	}
	if !strings.Contains(summary, "Token/Char ratio") {
		t.Error("summary should contain ratio")
	}
}

func TestCountStats_Summary_NoFiles(t *testing.T) {
	stats := &CountStats{
		TotalTokens: 500,
		TotalChars:  2000,
		TotalLines:  25,
		FilesCount:  0,
		Encoding:    O200kBase,
	}

	summary := stats.Summary()
	if strings.Contains(summary, "Files:") {
		t.Error("summary should not contain Files when count is 0")
	}
}
