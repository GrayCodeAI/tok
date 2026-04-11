package filter

import "testing"

func TestTokenDenseDialect_Encode(t *testing.T) {
	tdd := NewTokenDenseDialect(DefaultTDDConfig())
	input := "function hello() { return true; }"
	encoded, count := tdd.Encode(input)
	if count < 0 {
		t.Errorf("expected non-negative count, got %d", count)
	}
	if len(encoded) == 0 {
		t.Error("expected non-empty encoded output")
	}
}

func TestTokenDenseDialect_Decode(t *testing.T) {
	tdd := NewTokenDenseDialect(DefaultTDDConfig())
	original := "function hello() { return true; }"
	encoded, _ := tdd.Encode(original)
	decoded := tdd.Decode(encoded)
	if len(decoded) == 0 {
		t.Error("expected non-empty decoded output")
	}
}

func TestTokenDenseDialect_EncodeWithStats(t *testing.T) {
	tdd := NewTokenDenseDialect(DefaultTDDConfig())
	input := "function hello() { return true; }"
	_, stats := tdd.EncodeWithStats(input)
	if stats.OriginalLen != len(input) {
		t.Errorf("expected original len %d, got %d", len(input), stats.OriginalLen)
	}
	if stats.Replacements < 0 {
		t.Errorf("expected non-negative replacements, got %d", stats.Replacements)
	}
}

func TestFormatTDDStats(t *testing.T) {
	stats := TDDStats{Replacements: 5, OriginalLen: 100, CompressedLen: 80, SavingsPct: 20.0}
	result := FormatTDDStats(stats)
	if result == "" {
		t.Error("expected non-empty stats string")
	}
}
