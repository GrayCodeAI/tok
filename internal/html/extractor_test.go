// Package html provides tests for HTML content extraction.
package html

import (
	"strings"
	"testing"
)

func TestExtractor_Extract(t *testing.T) {
	htmlContent := `<!DOCTYPE html>
<html>
<head>
	<title>Test Page</title>
	<meta name="description" content="Test description">
</head>
<body>
	<article>
		<h1>Article Title</h1>
		<p>This is the main content.</p>
	</article>
</body>
</html>`

	extractor := NewExtractor()
	result, err := extractor.Extract(htmlContent)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	if result.Title != "Test Page" {
		t.Errorf("expected title 'Test Page', got %q", result.Title)
	}

	if result.Summary != "Test description" {
		t.Errorf("expected summary 'Test description', got %q", result.Summary)
	}

	if !strings.Contains(result.Content, "Article Title") {
		t.Error("expected content to contain article text")
	}
}

func TestSiteSpecificExtractors_Extract(t *testing.T) {
	extractors := NewSiteSpecificExtractors()

	tests := []struct {
		name     string
		url      string
		html     string
		expected string
	}{
		{
			name:     "wikipedia",
			url:      "https://en.wikipedia.org/wiki/Go_(programming_language)",
			html:     `<h1 id="firstHeading">Go (programming language)</h1><div id="mw-content-text">Go is a programming language</div>`,
			expected: "Go (programming language)",
		},
		{
			name:     "github",
			url:      "https://github.com/example/repo",
			html:     `<title>example/repo: Description</title>`,
			expected: "example/repo",
		},
		{
			name:     "hackernews",
			url:      "https://news.ycombinator.com/item?id=123",
			html:     `<span class="titleline"><a href="#">Interesting Article</a></span>`,
			expected: "Interesting Article",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := extractors.Extract(tc.url, tc.html)
			if err != nil {
				t.Fatalf("Extract failed: %v", err)
			}

			if !strings.Contains(result.Title, tc.expected) {
				t.Errorf("expected title to contain %q, got %q", tc.expected, result.Title)
			}

			if result.SiteName == "" {
				t.Error("expected SiteName to be set")
			}
		})
	}
}

func TestFormatResult(t *testing.T) {
	result := &ExtractResult{
		Title:    "Test Title",
		SiteName: "Example Site",
		Author:   "John Doe",
		Date:     "2024-01-01",
		Summary:  "Test summary",
		Content:  "Full content here",
		Links:    []string{"https://example.com/1", "https://example.com/2"},
	}

	formatted := FormatResult(result)

	if !strings.Contains(formatted, "Test Title") {
		t.Error("expected formatted output to contain title")
	}

	if !strings.Contains(formatted, "Example Site") {
		t.Error("expected formatted output to contain site name")
	}

	if !strings.Contains(formatted, "John Doe") {
		t.Error("expected formatted output to contain author")
	}
}
