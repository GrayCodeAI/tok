package filter

import (
	"strings"
	"testing"
)

func TestTieredSummaryFilter_New(t *testing.T) {
	tsf := NewTieredSummaryFilter()
	if tsf == nil {
		t.Fatal("expected non-nil TieredSummaryFilter")
	}
	if !tsf.enabled {
		t.Error("expected filter to be enabled")
	}
}

func TestTieredSummaryFilter_Name(t *testing.T) {
	tsf := NewTieredSummaryFilter()
	if tsf.Name() != "tiered_summary" {
		t.Errorf("expected name 'tiered_summary', got '%s'", tsf.Name())
	}
}

func TestTieredSummaryFilter_Apply(t *testing.T) {
	tsf := NewTieredSummaryFilter()

	// Test with short content - should pass through
	shortInput := "short content"
	output, saved := tsf.Apply(shortInput, ModeMinimal)
	if output != shortInput {
		t.Error("expected unchanged output for short content")
	}
	if saved != 0 {
		t.Error("expected 0 savings for short content")
	}

	// Test with mode None
	longInput := strings.Repeat("This is a long paragraph with lots of text. ", 20)
	output, saved = tsf.Apply(longInput, ModeNone)
	if output != longInput {
		t.Error("expected unchanged output with ModeNone")
	}
	if saved != 0 {
		t.Error("expected 0 savings with ModeNone")
	}
}

func TestTieredSummaryFilter_GenerateTiers(t *testing.T) {
	tsf := NewTieredSummaryFilter()

	content := `# Introduction to Go Programming

Go is a statically typed, compiled programming language designed at Google.

## Key Features
Go has several key features that make it popular:
- Concurrency support with goroutines
- Fast compilation
- Simple syntax
- Built-in testing

## Installation
To install Go, download it from golang.org.

## Conclusion
Go is a great language for building scalable systems.`

	result := tsf.GenerateTiers(content)

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	// Check L0
	if result.L0 == nil {
		t.Error("expected L0 summary")
	} else {
		if len(result.L0.Keywords) == 0 {
			t.Error("expected L0 to have keywords")
		}
	}

	// Check L1
	if result.L1 == nil {
		t.Error("expected L1 summary")
	} else {
		if len(result.L1.Sections) == 0 {
			t.Error("expected L1 to have sections")
		}
	}

	// Check L2
	if result.L2 == nil {
		t.Error("expected L2 summary")
	} else {
		if result.L2.Summary == "" {
			t.Error("expected L2 to have summary")
		}
	}
}

func TestTieredSummaryFilter_L0Generation(t *testing.T) {
	tsf := NewTieredSummaryFilter()

	content := `The quick brown fox jumps over the lazy dog.
The Go programming language is designed for simplicity and efficiency.
Go supports concurrency through goroutines and channels.
Many developers use Go for building scalable backend services.`

	l0 := tsf.generateL0(content)

	if l0 == nil {
		t.Fatal("expected non-nil L0 summary")
	}

	// Should have some keywords
	if len(l0.Keywords) == 0 {
		t.Error("expected L0 to extract keywords")
	}

	// Should have tokens counted
	if l0.TokenCount == 0 {
		t.Error("expected L0 to have token count")
	}
}

func TestTieredSummaryFilter_L1Generation(t *testing.T) {
	tsf := NewTieredSummaryFilter()

	content := `# Main Title

## Section 1
This is the content of section 1.
It has multiple lines.

## Section 2
This is section 2 content.
More content here.`

	l1 := tsf.generateL1(content)

	if l1 == nil {
		t.Fatal("expected non-nil L1 summary")
	}

	// Should have title
	if l1.Title == "" {
		t.Error("expected L1 to extract title")
	}

	// Should have sections
	if len(l1.Sections) == 0 {
		t.Error("expected L1 to extract sections")
	}
}

func TestTieredSummaryFilter_L2Generation(t *testing.T) {
	tsf := NewTieredSummaryFilter()

	content := `First paragraph with important information about the topic.

Second paragraph with more details and explanations.

Third paragraph concluding the discussion.`

	l2 := tsf.generateL2(content)

	if l2 == nil {
		t.Fatal("expected non-nil L2 summary")
	}

	// Should have summary
	if l2.Summary == "" {
		t.Error("expected L2 to have summary")
	}

	// Should have key points
	if len(l2.KeyPoints) == 0 {
		t.Error("expected L2 to have key points")
	}
}

func TestTieredSummaryFilter_Formatters(t *testing.T) {
	tsf := NewTieredSummaryFilter()

	// Test L0 formatting
	l0 := &L0Summary{
		Keywords:   []string{"go", "programming", "language"},
		Entities:   []string{"Google"},
		Topics:     []string{"programming"},
		TokenCount: 10,
	}
	formatted := tsf.formatL0(l0)
	if formatted == "" {
		t.Error("expected non-empty L0 format")
	}
	if !strings.Contains(formatted, "go") {
		t.Error("expected L0 format to contain keywords")
	}

	// Test L1 formatting
	l1 := &L1Summary{
		Title:    "Test Document",
		Sections: []Section{{Heading: "Section 1", Level: 1}},
		Outline:  "- Section 1",
	}
	formatted = tsf.formatL1(l1)
	if formatted == "" {
		t.Error("expected non-empty L1 format")
	}
	if !strings.Contains(formatted, "Test Document") {
		t.Error("expected L1 format to contain title")
	}

	// Test L2 formatting
	l2 := &L2Summary{
		Summary:   "This is a summary",
		KeyPoints: []string{"Point 1", "Point 2"},
	}
	formatted = tsf.formatL2(l2)
	if formatted == "" {
		t.Error("expected non-empty L2 format")
	}
	if !strings.Contains(formatted, "summary") {
		t.Error("expected L2 format to contain summary")
	}
}

func TestTieredSummaryFilter_TierSelection(t *testing.T) {
	tsf := NewTieredSummaryFilter()

	// Very long content should get L2
	veryLong := strings.Repeat("word ", 5000)
	result := tsf.GenerateTiers(veryLong)
	_ = result

	// Medium content should get L1
	medium := strings.Repeat("word ", 1000)
	result = tsf.GenerateTiers(medium)
	_ = result

	// Short content should get L0
	short := "short content"
	result = tsf.GenerateTiers(short)
	_ = result
}

func TestTieredSummaryFilter_TierName(t *testing.T) {
	tests := []struct {
		tier     SummaryTier
		expected string
	}{
		{TierL0, "L0"},
		{TierL1, "L1"},
		{TierL2, "L2"},
		{SummaryTier(99), "unknown"},
	}

	for _, tt := range tests {
		result := tierName(tt.tier)
		if result != tt.expected {
			t.Errorf("tierName(%v) = %s, expected %s", tt.tier, result, tt.expected)
		}
	}
}

func TestTieredSummaryFilter_ExtractKeywords(t *testing.T) {
	content := `The Go programming language is designed at Google.
Go is simple and efficient for building software.`

	keywords := extractKeywords(content, 5)

	if len(keywords) == 0 {
		t.Error("expected keywords to be extracted")
	}

	// Should not have stopwords
	stopwords := []string{"the", "is", "at", "and", "for"}
	for _, kw := range keywords {
		for _, sw := range stopwords {
			if kw == sw {
				t.Errorf("expected no stopwords, found '%s'", kw)
			}
		}
	}
}

func TestTieredSummaryFilter_ExtractEntities(t *testing.T) {
	content := `Google developed the Go language.
Many companies like Apple and Microsoft use it.`

	entities := extractEntities(content)

	if len(entities) == 0 {
		t.Error("expected entities to be extracted")
	}
}

func TestTieredSummaryFilter_InferTopics(t *testing.T) {
	keywords := []string{"function", "class", "variable", "import"}
	topics := inferTopics(keywords)

	if len(topics) == 0 {
		t.Error("expected topics to be inferred")
	}

	// Should detect code topic
	foundCode := false
	for _, t := range topics {
		if t == "code" {
			foundCode = true
			break
		}
	}
	if !foundCode {
		t.Error("expected 'code' topic to be inferred from code keywords")
	}
}

func TestTieredSummaryFilter_BuildOutline(t *testing.T) {
	sections := []Section{
		{Heading: "Section 1", Level: 1},
		{Heading: "Subsection", Level: 2},
		{Heading: "Section 2", Level: 1},
	}

	outline := buildOutline(sections)

	if outline == "" {
		t.Error("expected non-empty outline")
	}

	if !strings.Contains(outline, "Section 1") {
		t.Error("expected outline to contain section headings")
	}
}

func TestTieredSummaryFilter_SplitSentences(t *testing.T) {
	text := "First sentence. Second sentence! Third question?"
	sentences := splitSentences(text)

	if len(sentences) != 3 {
		t.Errorf("expected 3 sentences, got %d", len(sentences))
	}
}

func TestTieredSummaryFilter_DeduplicateStrings(t *testing.T) {
	input := []string{"a", "b", "a", "c", "b"}
	result := deduplicateStrings(input)

	if len(result) != 3 {
		t.Errorf("expected 3 unique strings, got %d", len(result))
	}
}
