package core

import (
	"strings"
	"testing"
)

func TestDetectProvider(t *testing.T) {
	tests := []struct {
		header []string
		want   string
	}{
		{
			header: []string{"date", "model", "input_tokens", "output_tokens", "cost_usd", "api_key_name", "cache_creation"},
			want:   "anthropic",
		},
		{
			header: []string{"date", "prompt_tokens", "completion_tokens", "cost"},
			want:   "openai",
		},
		{
			header: []string{"date", "input_tokens", "model"},
			want:   "openai", // no anthropic-specific markers
		},
		{
			header: []string{"foo", "bar"},
			want:   "",
		},
	}
	for _, tt := range tests {
		got := detectProvider(tt.header)
		if got != tt.want {
			t.Errorf("detectProvider(%v) = %q, want %q", tt.header, got, tt.want)
		}
	}
}

func TestParseAnthropicRow(t *testing.T) {
	header := []string{"date", "model", "input_tokens", "output_tokens", "cost_usd"}
	rec := []string{"2026-04-15", "claude-opus-4-7", "125000", "8200", "4.52"}
	row, err := parseAnthropicRow(header, rec)
	if err != nil {
		t.Fatalf("parseAnthropicRow: %v", err)
	}
	if row.inputTokens != 125000 {
		t.Errorf("inputTokens = %d, want 125000", row.inputTokens)
	}
	if row.outputTokens != 8200 {
		t.Errorf("outputTokens = %d, want 8200", row.outputTokens)
	}
	if row.amountUSD != 4.52 {
		t.Errorf("amountUSD = %f, want 4.52", row.amountUSD)
	}
	if row.date.Format("2006-01-02") != "2026-04-15" {
		t.Errorf("date = %s, want 2026-04-15", row.date)
	}
}

func TestParseOpenAIRow(t *testing.T) {
	header := []string{"date", "prompt_tokens", "completion_tokens", "cost"}
	rec := []string{"2026-04-15", "50000", "2000", "1.25"}
	row, err := parseOpenAIRow(header, rec)
	if err != nil {
		t.Fatalf("parseOpenAIRow: %v", err)
	}
	if row.inputTokens != 50000 || row.outputTokens != 2000 || row.amountUSD != 1.25 {
		t.Errorf("unexpected row: %+v", row)
	}
}

func TestParseFloatHandlesCurrencySymbols(t *testing.T) {
	cases := map[string]float64{
		"4.52":    4.52,
		"$4.52":   4.52,
		"1,250.5": 1250.5,
	}
	for in, want := range cases {
		got := parseFloat(in)
		if got != want {
			t.Errorf("parseFloat(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestParseDateFlexibleAcceptsCommonFormats(t *testing.T) {
	for _, s := range []string{
		"2026-04-15",
		"2026-04-15T10:30:00Z",
		"2026-04-15T10:30:00",
		"04/15/2026",
	} {
		if _, err := parseDateFlexible(s); err != nil {
			t.Errorf("parseDateFlexible(%q) failed: %v", s, err)
		}
	}
	if _, err := parseDateFlexible("not-a-date"); err == nil {
		t.Error("parseDateFlexible accepted garbage input")
	}
}

func TestHeaderMapCaseInsensitive(t *testing.T) {
	header := []string{"Date", "INPUT_TOKENS", "cost_usd"}
	rec := []string{"2026-04-15", "100", "0.50"}
	m := headerMap(header, rec)
	if m["date"] != "2026-04-15" {
		t.Errorf("date not normalized to lowercase key")
	}
	if m["input_tokens"] != "100" {
		t.Errorf("INPUT_TOKENS not normalized")
	}
}

func TestParseAnthropicRejectsEmptyDate(t *testing.T) {
	header := []string{"date", "input_tokens", "output_tokens", "cost_usd"}
	rec := []string{"", "100", "50", "0.10"}
	_, err := parseAnthropicRow(header, rec)
	if err == nil {
		t.Error("expected error for empty date")
	}
	if !strings.Contains(err.Error(), "empty date") {
		t.Errorf("wrong error: %v", err)
	}
}
