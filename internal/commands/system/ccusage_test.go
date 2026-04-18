package system

import "testing"

func TestBuildCcusageArgs(t *testing.T) {
	got := buildCcusageArgs("daily")
	want := []string{"daily", "--json", "--since", "20250101"}

	if len(got) != len(want) {
		t.Fatalf("len(args) = %d, want %d (%v)", len(got), len(want), got)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("args[%d] = %q, want %q (full: %v)", i, got[i], want[i], got)
		}
	}
}
