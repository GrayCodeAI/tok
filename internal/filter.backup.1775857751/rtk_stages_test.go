package filter

import "testing"

func TestColorPassthrough_StripAndStore(t *testing.T) {
	cp := &ColorPassthrough{}
	content := "\x1b[31mred\x1b[0m text"
	result := cp.StripAndStore(content)
	if result == "" {
		t.Error("expected non-empty stripped content")
	}
}

func TestColorPassthrough_RestoreCodes(t *testing.T) {
	cp := &ColorPassthrough{}
	content := "\x1b[31mred\x1b[0m text"
	cp.StripAndStore(content)
	restored := cp.RestoreCodes(cp.stripped)
	if len(restored) == 0 {
		t.Error("expected non-empty restored content")
	}
}

func TestPreferLessMode(t *testing.T) {
	original := "this is a longer output string"
	filtered := "shorter"
	result := PreferLessMode(original, filtered)
	if result != filtered {
		t.Error("expected filtered output when shorter")
	}
}

func TestPreferLessMode_Original(t *testing.T) {
	original := "short"
	filtered := "this is a much longer filtered output"
	result := PreferLessMode(original, filtered)
	if result != original {
		t.Error("expected original output when shorter")
	}
}

func TestTaskRunnerWrapping_Wrap(t *testing.T) {
	trw := NewTaskRunnerWrapping("make", "tokman")
	content := "build:\n\tgo build ./..."
	result := trw.Wrap(content)
	if len(result) == 0 {
		t.Error("expected non-empty result")
	}
}
