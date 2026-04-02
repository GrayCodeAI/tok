package webclean

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var ErrBadURL = errors.New("bad URL")

type URLFetcher struct {
	client  *http.Client
	timeout time.Duration
}

func NewURLFetcher(timeout time.Duration) *URLFetcher {
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	return &URLFetcher{
		client:  &http.Client{Timeout: timeout},
		timeout: timeout,
	}
}

func (f *URLFetcher) Fetch(rawURL string) (string, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", ErrBadURL
	}

	resp, err := f.client.Get(rawURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", ErrBadURL
	}

	buf := make([]byte, 1024*1024)
	n, _ := resp.Body.Read(buf)
	return string(buf[:n]), nil
}

func ValidateURL(rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return err
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return ErrBadURL
	}
	return nil
}

type HTMLCleaner struct {
	removeTags  []string
	removeAttrs []string
}

func NewHTMLCleaner() *HTMLCleaner {
	return &HTMLCleaner{
		removeTags: []string{
			"script", "style", "nav", "footer", "aside", "iframe",
			"noscript", "svg", "img", "video", "audio", "canvas",
			"header", "menu", "advertisement", "banner",
		},
		removeAttrs: []string{
			"onclick", "onload", "onerror", "style", "class",
		},
	}
}

func (c *HTMLCleaner) Clean(html string) string {
	result := html

	for _, tag := range c.removeTags {
		result = removeTag(result, tag)
	}

	result = stripHTMLTags(result)
	result = normalizeWhitespace(result)

	return result
}

func removeTag(html, tag string) string {
	lower := strings.ToLower(html)
	for {
		start := strings.Index(lower, "<"+tag)
		if start == -1 {
			break
		}
		endTag := "</" + tag + ">"
		end := strings.Index(lower[start:], endTag)
		if end == -1 {
			end = strings.Index(lower[start:], ">")
			if end == -1 {
				break
			}
			html = html[:start] + html[start+end+1:]
		} else {
			html = html[:start] + html[start+end+len(endTag):]
		}
		lower = strings.ToLower(html)
	}
	return html
}

func stripHTMLTags(html string) string {
	var result strings.Builder
	inTag := false
	for _, ch := range html {
		if ch == '<' {
			inTag = true
			continue
		}
		if ch == '>' {
			inTag = false
			continue
		}
		if !inTag {
			result.WriteRune(ch)
		}
	}
	return result.String()
}

func normalizeWhitespace(text string) string {
	var result strings.Builder
	prevSpace := false
	for _, ch := range text {
		isSpace := ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
		if isSpace {
			if !prevSpace {
				result.WriteRune(' ')
			}
			prevSpace = true
		} else {
			result.WriteRune(ch)
			prevSpace = false
		}
	}
	return strings.TrimSpace(result.String())
}

type ContentExtractor struct{}

func NewContentExtractor() *ContentExtractor {
	return &ContentExtractor{}
}

func (e *ContentExtractor) ExtractScored(text string) (string, float64) {
	paragraphs := strings.Split(text, "\n\n")
	var best string
	var bestScore float64

	for _, p := range paragraphs {
		p = strings.TrimSpace(p)
		if len(p) < 50 {
			continue
		}
		score := contentScore(p)
		if score > bestScore {
			best = p
			bestScore = score
		}
	}

	return best, bestScore
}

func contentScore(text string) float64 {
	score := float64(len(text))
	words := strings.Fields(text)
	if len(words) < 10 {
		return 0
	}
	score += float64(len(words)) * 2

	linkRatio := float64(strings.Count(text, "<a")) / float64(len(words))
	if linkRatio > 0.5 {
		score *= 0.5
	}

	return score
}

type JSONCleaner struct{}

func NewJSONCleaner() *JSONCleaner {
	return &JSONCleaner{}
}

func (c *JSONCleaner) Clean(jsonStr string) string {
	result := jsonStr
	result = removeJSONNoise(result)
	return result
}

func removeJSONNoise(jsonStr string) string {
	lines := strings.Split(jsonStr, "\n")
	var clean []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, "copyright") ||
			strings.Contains(trimmed, "license") ||
			strings.Contains(trimmed, "_comment") {
			continue
		}
		clean = append(clean, line)
	}
	return strings.Join(clean, "\n")
}

type WebCleanResult struct {
	OriginalTokens int     `json:"original_tokens"`
	CleanTokens    int     `json:"clean_tokens"`
	ReductionPct   float64 `json:"reduction_pct"`
	ContentType    string  `json:"content_type"`
}

func CalculateReduction(original, cleaned string) *WebCleanResult {
	origTokens := len(original) / 4
	cleanTokens := len(cleaned) / 4
	reduction := 0.0
	if origTokens > 0 {
		reduction = float64(origTokens-cleanTokens) / float64(origTokens) * 100
	}
	return &WebCleanResult{
		OriginalTokens: origTokens,
		CleanTokens:    cleanTokens,
		ReductionPct:   reduction,
	}
}
