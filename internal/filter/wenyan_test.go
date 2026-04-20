package filter

import (
	"strings"
	"testing"
)

func TestWenyanCompress_StripsFiller(t *testing.T) {
	in := "This is really just a simple test."
	out := WenyanCompress(in, WenyanLite)
	for _, filler := range []string{"really", "just"} {
		if strings.Contains(strings.ToLower(out), filler) {
			t.Errorf("WenyanLite kept filler %q: %q", filler, out)
		}
	}
}

func TestWenyanCompress_ArrowCausality(t *testing.T) {
	cases := []struct{ in, want string }{
		{"A because B", "A → B"},
		{"Fails therefore retry", "Fails → retry"},
		{"Bug causes crash", "Bug → crash"},
	}
	for _, c := range cases {
		got := WenyanCompress(c.in, WenyanFull)
		if !strings.Contains(got, "→") {
			t.Errorf("WenyanFull missing arrow for %q: got %q", c.in, got)
		}
	}
}

func TestWenyanCompress_Abbreviations(t *testing.T) {
	in := "configuration documentation environment function database"
	out := WenyanCompress(in, WenyanFull)
	for long, short := range map[string]string{
		"configuration":  "config",
		"documentation":  "docs",
		"environment":    "env",
		"function":       "fn",
		"database":       "db",
	} {
		if strings.Contains(strings.ToLower(out), long) {
			t.Errorf("WenyanFull kept %q (should become %q): got %q", long, short, out)
		}
		if !strings.Contains(strings.ToLower(out), short) {
			t.Errorf("WenyanFull missing %q: got %q", short, out)
		}
	}
}

func TestWenyanCompress_UltraStripsConnectives(t *testing.T) {
	in := "foo and bar but baz however qux"
	out := WenyanCompress(in, WenyanUltra)
	for _, c := range []string{" and ", " but ", " however "} {
		if strings.Contains(out, c) {
			t.Errorf("WenyanUltra kept connective %q: %q", c, out)
		}
	}
}

func TestWenyanCompress_UltraAbbrevs(t *testing.T) {
	in := "connect without password through proxy"
	out := WenyanCompress(in, WenyanUltra)
	for _, want := range []string{"w/o", "thru"} {
		if !strings.Contains(out, want) {
			t.Errorf("WenyanUltra missing %q in %q", want, out)
		}
	}
}

func TestWenyanCompress_LiteKeepsGrammar(t *testing.T) {
	// Lite should not introduce arrows or strip connectives.
	in := "foo and bar because baz"
	out := WenyanCompress(in, WenyanLite)
	if strings.Contains(out, "→") {
		t.Errorf("WenyanLite should not use arrow causality: %q", out)
	}
	if !strings.Contains(out, "and") || !strings.Contains(out, "because") {
		t.Errorf("WenyanLite dropped connective or causal word: %q", out)
	}
}

func TestWenyanCompress_CJKParticleStripping(t *testing.T) {
	in := "我們的系統是好的"
	out := WenyanCompress(in, WenyanFull)
	for _, p := range []string{"的"} {
		if strings.Contains(out, p) {
			t.Errorf("Wenyan kept modern particle %q in CJK input: %q", p, out)
		}
	}
}

func TestWenyanCompress_EmptyInput(t *testing.T) {
	if got := WenyanCompress("", WenyanFull); got != "" {
		t.Errorf("empty input should stay empty, got %q", got)
	}
}

func TestWenyanFilter_Apply_ModeNone(t *testing.T) {
	f := NewWenyanFilter(WenyanFull)
	in := "the really long configuration documentation"
	out, saved := f.Apply(in, ModeNone)
	if out != in || saved != 0 {
		t.Errorf("ModeNone must pass through unchanged; got %q saved=%d", out, saved)
	}
}

func TestWenyanFilter_Apply_SavesBytes(t *testing.T) {
	f := NewWenyanFilter(WenyanUltra)
	in := "The really long configuration documentation environment without extra."
	out, saved := f.Apply(in, ModeMinimal)
	if len(out) >= len(in) {
		t.Errorf("expected compression, got len %d >= %d: %q", len(out), len(in), out)
	}
	if saved <= 0 {
		t.Errorf("expected saved > 0, got %d", saved)
	}
}

func TestWenyanFilter_Name(t *testing.T) {
	cases := []struct {
		mode WenyanMode
		want string
	}{
		{WenyanLite, "wenyan_lite"},
		{WenyanFull, "wenyan_full"},
		{WenyanUltra, "wenyan_ultra"},
	}
	for _, c := range cases {
		if got := NewWenyanFilter(c.mode).Name(); got != c.want {
			t.Errorf("Name(%d) = %q, want %q", c.mode, got, c.want)
		}
	}
}

func TestContainsCJK(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"hello world", false},
		{"你好", true},
		{"mix 中文 text", true},
		{"", false},
	}
	for _, c := range cases {
		if got := containsCJK(c.in); got != c.want {
			t.Errorf("containsCJK(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}
