package core

import (
	"strings"
	"testing"
)

func TestGroupByFile_SplitsProperly(t *testing.T) {
	findings := []string{
		"src/a.go:10 🔴 bug hard-coded credential. load from env.",
		"src/a.go:22 🟡 risk TODO. resolve.",
		"src/b.go:5 🔵 nit console.log. remove.",
	}
	groups := groupByFile(findings)
	if len(groups) != 2 {
		t.Fatalf("want 2 files, got %d", len(groups))
	}
	if len(groups["src/a.go"]) != 2 {
		t.Errorf("src/a.go should have 2 findings, got %d", len(groups["src/a.go"]))
	}
	if len(groups["src/b.go"]) != 1 {
		t.Errorf("src/b.go should have 1 finding, got %d", len(groups["src/b.go"]))
	}
	// Ensure file prefix is stripped — the rest should start with line number.
	for _, v := range groups["src/a.go"] {
		if strings.HasPrefix(v, "src/a.go") {
			t.Errorf("grouped finding still carries file prefix: %q", v)
		}
	}
}

func TestGroupByFile_SkipsMalformedFindings(t *testing.T) {
	findings := []string{
		"no-colon-anywhere",
		"src/a.go:1 ok",
	}
	groups := groupByFile(findings)
	if len(groups) != 1 {
		t.Errorf("want 1 file (malformed skipped), got %d", len(groups))
	}
}

func TestSortedKeys_Ordering(t *testing.T) {
	m := map[string][]string{
		"zeta":  {"x"},
		"alpha": {"y"},
		"mid":   {"z"},
	}
	got := sortedKeys(m)
	want := []string{"alpha", "mid", "zeta"}
	if len(got) != len(want) {
		t.Fatalf("len mismatch: %v vs %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("pos %d: got %q, want %q", i, got[i], want[i])
		}
	}
}

func TestSortedKeys_EmptyMap(t *testing.T) {
	got := sortedKeys(map[string][]string{})
	if len(got) != 0 {
		t.Errorf("empty map should yield empty slice, got %v", got)
	}
}
