package filter

import "testing"

func TestNewTemplatePipe(t *testing.T) {
	pipe := NewTemplatePipe("join | truncate 50 | lines 10")
	if len(pipe.operations) != 3 {
		t.Errorf("expected 3 operations, got %d", len(pipe.operations))
	}
}

func TestTemplatePipe_Process(t *testing.T) {
	pipe := NewTemplatePipe("truncate 10")
	input := "this is a long string that should be truncated"
	result := pipe.Process(input)
	if len(result) > 13 {
		t.Errorf("expected truncated result, got %d chars", len(result))
	}
}

func TestTemplatePipe_Keep(t *testing.T) {
	pipe := NewTemplatePipe("keep error")
	input := "info: starting\nerror: failed\ninfo: done\nerror: timeout"
	result := pipe.Process(input)
	if len(result) == 0 {
		t.Error("expected non-empty result")
	}
}

func TestTemplatePipe_Where(t *testing.T) {
	pipe := NewTemplatePipe("where debug")
	input := "info: starting\nerror: failed\ninfo: done"
	result := pipe.Process(input)
	if len(result) == 0 {
		t.Error("expected non-empty result")
	}
}

func TestJSONPathExtract(t *testing.T) {
	jsonStr := `{"name": "test", "version": "1.0"}`
	result := JSONPathExtract(jsonStr, "name")
	if result == "" {
		t.Error("expected non-empty result")
	}
}

func TestJSONPathExtract_Nested(t *testing.T) {
	jsonStr := `{"outer": {"inner": "value"}}`
	result := JSONPathExtract(jsonStr, "outer.inner")
	if result == "" {
		t.Error("expected non-empty result")
	}
}
