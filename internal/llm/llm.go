package llm

type LLM struct{}

func New() *LLM {
	return &LLM{}
}

type Summarizer struct{}

func NewSummarizerFromEnv() *Summarizer {
	return &Summarizer{}
}

func (s *Summarizer) Summarize(req SummaryRequest) (SummaryResponse, error) {
	return SummaryResponse{Summary: req.Content}, nil
}

func (s *Summarizer) IsAvailable() bool {
	return false
}

func (s *Summarizer) GetProvider() string {
	return ""
}

func (s *Summarizer) GetModel() string {
	return ""
}

type SummaryRequest struct {
	Content   string
	MaxTokens int
	Intent    string
}

type SummaryResponse struct {
	Summary string
}
