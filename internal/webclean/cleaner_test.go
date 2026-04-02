package webclean

import "testing"

func TestHTMLCleaner(t *testing.T) {
	c := NewHTMLCleaner()

	html := `<html><head><style>body{}</style></head><body><nav>menu</nav><main>content here</main><script>alert()</script></body></html>`
	cleaned := c.Clean(html)

	if cleaned == "" {
		t.Error("Expected non-empty output")
	}
}

func TestNormalizeWhitespace(t *testing.T) {
	input := "hello    world\n\n\nfoo  bar"
	output := normalizeWhitespace(input)
	if len(output) >= len(input) {
		t.Error("Expected shorter output after normalization")
	}
}

func TestStripHTMLTags(t *testing.T) {
	html := "<p>Hello <b>World</b></p>"
	text := stripHTMLTags(html)
	if text != "Hello World" {
		t.Errorf("Expected 'Hello World', got %s", text)
	}
}

func TestContentExtractor(t *testing.T) {
	e := NewContentExtractor()

	text := "This is a short paragraph.\n\nThis is a very long paragraph with much more content that should score higher because it has more words and more meaningful content for extraction."
	score, _ := e.ExtractScored(text)
	if score == "" {
		t.Error("Expected non-empty extraction")
	}
}

func TestCalculateReduction(t *testing.T) {
	result := CalculateReduction("long original text", "short")
	if result.OriginalTokens == 0 {
		t.Error("Expected non-zero original tokens")
	}
	if result.ReductionPct == 0 {
		t.Error("Expected non-zero reduction")
	}
}
