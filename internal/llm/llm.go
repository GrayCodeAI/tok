// Package llm provides LLM-based summarization capabilities (stub implementation).
// NOTE: This is a stub package. The full implementation was removed as dead code.
// These stub types and functions maintain API compatibility.
package llm

// IsStub reports whether this package is a placeholder implementation.
func IsStub() bool {
	return true
}

// Summarizer provides LLM-based text summarization (stub).
type Summarizer struct {
	Provider string
	Model    string
	BaseURL  string
}

// SummaryRequest represents a summary request (stub).
type SummaryRequest struct {
	Text      string
	Content   string
	Intent    string
	MaxTokens int
}

// SummaryResponse represents a summary response (stub).
type SummaryResponse struct {
	Summary string
	Tokens  int
}

// NewSummarizer creates a new summarizer (stub).
func NewSummarizer(provider, model, baseURL string) *Summarizer {
	return &Summarizer{
		Provider: provider,
		Model:    model,
		BaseURL:  baseURL,
	}
}

// NewSummarizerFromEnv creates a summarizer from environment variables (stub).
func NewSummarizerFromEnv() *Summarizer {
	return &Summarizer{
		Provider: "stub",
		Model:    "stub",
		BaseURL:  "http://localhost:0",
	}
}

// Summarize summarizes text using LLM (stub - returns input unchanged).
func (s *Summarizer) Summarize(text string, maxTokens int) (string, error) {
	return text, nil
}

// SummarizeWithRequest summarizes text with a request object (stub).
func (s *Summarizer) SummarizeWithRequest(req SummaryRequest) (SummaryResponse, error) {
	return SummaryResponse{
		Summary: req.Text,
		Tokens:  len(req.Text) / 4,
	}, nil
}

// IsAvailable checks if LLM is available (stub - always returns false).
func (s *Summarizer) IsAvailable() bool {
	return false
}
