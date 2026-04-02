package fetchclean

import (
	"net/http"
	"strings"
	"time"
)

type FetchCleaner struct {
	client *http.Client
}

func NewFetcherCleaner() *FetchCleaner {
	return &FetchCleaner{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (f *FetchCleaner) FetchClean(url string) (string, error) {
	resp, err := f.client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	buf := make([]byte, 1024*1024)
	n, _ := resp.Body.Read(buf)
	content := string(buf[:n])

	return f.cleanHTML(content), nil
}

func (f *FetchCleaner) FetchCleanBatch(urls []string) (map[string]string, error) {
	results := make(map[string]string)
	for _, url := range urls {
		clean, err := f.FetchClean(url)
		if err == nil {
			results[url] = clean
		}
	}
	return results, nil
}

func (f *FetchCleaner) RefinePrompt(input string) string {
	result := input
	fillers := []string{
		"Here is the ", "Here's the ", "The following ",
		"Note that ", "Keep in mind ", "As mentioned ",
		"Of course ", "Certainly ", "Let me ",
		"I will ", "I'll ", "I can ",
	}
	for _, filler := range fillers {
		result = strings.ReplaceAll(result, filler, "")
	}
	return strings.TrimSpace(result)
}

func (f *FetchCleaner) cleanHTML(html string) string {
	result := html
	tags := []string{"script", "style", "nav", "footer", "aside", "header", "noscript", "iframe"}
	for _, tag := range tags {
		for {
			start := strings.Index(strings.ToLower(result), "<"+tag)
			if start == -1 {
				break
			}
			endTag := "</" + tag + ">"
			end := strings.Index(strings.ToLower(result[start:]), endTag)
			if end == -1 {
				end = strings.Index(result[start:], ">")
				if end == -1 {
					break
				}
				result = result[:start] + result[start+end+1:]
			} else {
				result = result[:start] + result[start+end+len(endTag):]
			}
		}
	}
	var sb strings.Builder
	inTag := false
	for _, c := range result {
		if c == '<' {
			inTag = true
			continue
		}
		if c == '>' {
			inTag = false
			sb.WriteRune(' ')
			continue
		}
		if !inTag {
			sb.WriteRune(c)
		}
	}
	return strings.TrimSpace(sb.String())
}

type LLMCompressionEngine struct {
	endpoint string
}

func NewLLMCompressionEngine(endpoint string) *LLMCompressionEngine {
	if endpoint == "" {
		endpoint = "http://localhost:11434/api/generate"
	}
	return &LLMCompressionEngine{endpoint: endpoint}
}

func (e *LLMCompressionEngine) Compress(input string) string {
	if len(input) < 50 {
		return input
	}
	lines := strings.Split(input, "\n")
	var compressed []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if len(trimmed) > 10 {
			compressed = append(compressed, trimmed)
		}
	}
	if len(compressed) > 5 {
		compressed = compressed[:5]
	}
	return strings.Join(compressed, "\n")
}
