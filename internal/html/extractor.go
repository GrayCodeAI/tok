// Package html provides HTML content extraction for token savings.
package html

import (
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

// Extractor extracts content from HTML.
type Extractor struct {
	selectors map[string]string
}

// NewExtractor creates a new HTML extractor.
func NewExtractor() *Extractor {
	return &Extractor{
		selectors: map[string]string{
			"title":    "title",
			"content":  "article, .content, .main, #content, #main",
			"author":   ".author, [rel=author], .byline",
			"date":     ".date, time, [datetime]",
			"summary":  ".summary, .description, meta[name=description]",
			"comments": ".comments, #comments",
		},
	}
}

// ExtractResult contains extracted content.
type ExtractResult struct {
	Title    string
	Content  string
	Author   string
	Date     string
	Summary  string
	Links    []string
	Images   []string
	SiteName string
}

// Extract extracts content from HTML.
func (e *Extractor) Extract(htmlContent string) (*ExtractResult, error) {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return nil, err
	}

	result := &ExtractResult{
		Links:  []string{},
		Images: []string{},
	}

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode {
			switch n.Data {
			case "title":
				if result.Title == "" && n.FirstChild != nil {
					result.Title = n.FirstChild.Data
				}
			case "meta":
				var name, content string
				for _, attr := range n.Attr {
					if attr.Key == "name" || attr.Key == "property" {
						name = attr.Val
					}
					if attr.Key == "content" {
						content = attr.Val
					}
				}
				if name == "description" && result.Summary == "" {
					result.Summary = content
				}
				if name == "og:site_name" && result.SiteName == "" {
					result.SiteName = content
				}
			case "a":
				for _, attr := range n.Attr {
					if attr.Key == "href" {
						result.Links = append(result.Links, attr.Val)
					}
				}
			case "img":
				for _, attr := range n.Attr {
					if attr.Key == "src" {
						result.Images = append(result.Images, attr.Val)
					}
				}
			case "time":
				if result.Date == "" {
					for _, attr := range n.Attr {
						if attr.Key == "datetime" {
							result.Date = attr.Val
						}
					}
				}
			}

			// Extract text from main content areas
			if e.isContentElement(n.Data, n.Attr) {
				text := e.extractText(n)
				if len(text) > len(result.Content) {
					result.Content = text
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	return result, nil
}

func (e *Extractor) isContentElement(tag string, attrs []html.Attribute) bool {
	contentClasses := []string{
		"content", "article", "main", "post", "entry",
		"#content", "#main", "#article",
	}

	for _, attr := range attrs {
		if attr.Key == "class" || attr.Key == "id" {
			for _, cls := range contentClasses {
				if strings.Contains(attr.Val, cls) {
					return true
				}
			}
		}
	}

	return tag == "article" || tag == "main"
}

func (e *Extractor) extractText(n *html.Node) string {
	var text strings.Builder
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.TextNode {
			text.WriteString(n.Data)
			text.WriteString(" ")
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(n)
	return strings.TrimSpace(text.String())
}

// SiteSpecificExtractors provide extractors for popular sites.
type SiteSpecificExtractors struct {
	extractors map[string]SiteExtractor
}

// SiteExtractor extracts content from a specific site.
type SiteExtractor interface {
	Extract(htmlContent string) (*ExtractResult, error)
}

// NewSiteSpecificExtractors creates site-specific extractors.
func NewSiteSpecificExtractors() *SiteSpecificExtractors {
	return &SiteSpecificExtractors{
		extractors: map[string]SiteExtractor{
			"news.ycombinator.com": &HackerNewsExtractor{},
			"github.com":           &GitHubExtractor{},
			"stackoverflow.com":    &StackOverflowExtractor{},
			"wikipedia.org":        &WikipediaExtractor{},
		},
	}
}

// Extract extracts content using site-specific extractor if available.
func (s *SiteSpecificExtractors) Extract(url, htmlContent string) (*ExtractResult, error) {
	for domain, extractor := range s.extractors {
		if strings.Contains(url, domain) {
			return extractor.Extract(htmlContent)
		}
	}

	// Fall back to generic extractor
	return NewExtractor().Extract(htmlContent)
}

// HackerNewsExtractor extracts content from Hacker News.
type HackerNewsExtractor struct{}

func (h *HackerNewsExtractor) Extract(htmlContent string) (*ExtractResult, error) {
	result := &ExtractResult{
		SiteName: "Hacker News",
		Links:    []string{},
	}

	// Extract title
	titleRe := regexp.MustCompile(`<span class="titleline"><a[^>]*>([^<]+)</a>`)
	if matches := titleRe.FindStringSubmatch(htmlContent); len(matches) > 1 {
		result.Title = matches[1]
	}

	// Extract comments
	// Simplified extraction for demo
	result.Content = "Hacker News discussion"

	return result, nil
}

// GitHubExtractor extracts content from GitHub.
type GitHubExtractor struct{}

func (g *GitHubExtractor) Extract(htmlContent string) (*ExtractResult, error) {
	result := &ExtractResult{
		SiteName: "GitHub",
		Links:    []string{},
	}

	// Extract repo name
	titleRe := regexp.MustCompile(`<title>([^<]+)</title>`)
	if matches := titleRe.FindStringSubmatch(htmlContent); len(matches) > 1 {
		result.Title = matches[1]
	}

	result.Content = "GitHub repository page"

	return result, nil
}

// StackOverflowExtractor extracts from Stack Overflow.
type StackOverflowExtractor struct{}

func (s *StackOverflowExtractor) Extract(htmlContent string) (*ExtractResult, error) {
	result := &ExtractResult{
		SiteName: "Stack Overflow",
		Links:    []string{},
	}

	// Extract question title
	titleRe := regexp.MustCompile(`<h1[^>]*class="[^"]*fs-headline1[^"]*"[^>]*>([^<]+)</h1>`)
	if matches := titleRe.FindStringSubmatch(htmlContent); len(matches) > 1 {
		result.Title = matches[1]
	}

	result.Content = "Stack Overflow question"

	return result, nil
}

// WikipediaExtractor extracts from Wikipedia.
type WikipediaExtractor struct{}

func (w *WikipediaExtractor) Extract(htmlContent string) (*ExtractResult, error) {
	result := &ExtractResult{
		SiteName: "Wikipedia",
		Links:    []string{},
	}

	// Extract title
	titleRe := regexp.MustCompile(`<h1[^>]*id="firstHeading"[^>]*>([^<]+)</h1>`)
	if matches := titleRe.FindStringSubmatch(htmlContent); len(matches) > 1 {
		result.Title = matches[1]
	}

	// Extract main content
	contentRe := regexp.MustCompile(`<div[^>]*id="mw-content-text"[^>]*>(.*?)</div>`)
	if matches := contentRe.FindStringSubmatch(htmlContent); len(matches) > 1 {
		// Strip HTML tags
		re := regexp.MustCompile(`<[^>]+>`)
		result.Content = re.ReplaceAllString(matches[1], " ")
	}

	return result, nil
}

// FormatResult formats extraction result for display.
func FormatResult(result *ExtractResult) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("# %s\n\n", result.Title))

	if result.SiteName != "" {
		b.WriteString(fmt.Sprintf("Source: %s\n", result.SiteName))
	}

	if result.Author != "" {
		b.WriteString(fmt.Sprintf("Author: %s\n", result.Author))
	}

	if result.Date != "" {
		b.WriteString(fmt.Sprintf("Date: %s\n", result.Date))
	}

	if result.Summary != "" {
		b.WriteString(fmt.Sprintf("\n> %s\n", result.Summary))
	}

	if result.Content != "" {
		b.WriteString(fmt.Sprintf("\n%s\n", result.Content))
	}

	if len(result.Links) > 0 {
		b.WriteString("\n## Links\n")
		for _, link := range result.Links[:min(10, len(result.Links))] {
			b.WriteString(fmt.Sprintf("- %s\n", link))
		}
	}

	return b.String()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
