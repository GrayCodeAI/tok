package onboarding

import "strings"

type OnboardingStep struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Command     string `json:"command"`
	Completed   bool   `json:"completed"`
}

type OnboardingFlow struct {
	steps   []OnboardingStep
	current int
}

func NewOnboardingFlow() *OnboardingFlow {
	return &OnboardingFlow{
		steps: []OnboardingStep{
			{ID: 1, Title: "Install TokMan", Description: "Verify TokMan is installed and working", Command: "tokman --version"},
			{ID: 2, Title: "Initialize Config", Description: "Create your configuration file", Command: "tokman init"},
			{ID: 3, Title: "Run First Compression", Description: "Compress your first command output", Command: "tokman git status"},
			{ID: 4, Title: "Check Savings", Description: "View your token savings", Command: "tokman gain"},
			{ID: 5, Title: "Configure Hooks", Description: "Set up AI agent integration", Command: "tokman init -g"},
			{ID: 6, Title: "Explore Dashboard", Description: "Open the web dashboard", Command: "tokman tui"},
		},
	}
}

func (f *OnboardingFlow) CurrentStep() *OnboardingStep {
	if f.current < len(f.steps) {
		return &f.steps[f.current]
	}
	return nil
}

func (f *OnboardingFlow) CompleteStep() {
	if f.current < len(f.steps) {
		f.steps[f.current].Completed = true
		f.current++
	}
}

func (f *OnboardingFlow) Progress() float64 {
	if len(f.steps) == 0 {
		return 0
	}
	completed := 0
	for _, s := range f.steps {
		if s.Completed {
			completed++
		}
	}
	return float64(completed) / float64(len(f.steps)) * 100
}

func (f *OnboardingFlow) IsComplete() bool {
	return f.current >= len(f.steps)
}

func (f *OnboardingFlow) Steps() []OnboardingStep {
	return f.steps
}

type InteractiveTutorial struct {
	currentTopic int
	topics       []string
}

func NewInteractiveTutorial() *InteractiveTutorial {
	return &InteractiveTutorial{
		topics: []string{
			"What is TokMan?",
			"How Compression Works",
			"Filter System",
			"Pipeline Presets",
			"AI Agent Integration",
			"Dashboard & Analytics",
			"Advanced Features",
		},
	}
}

func (t *InteractiveTutorial) CurrentTopic() string {
	if t.currentTopic < len(t.topics) {
		return t.topics[t.currentTopic]
	}
	return ""
}

func (t *InteractiveTutorial) Next() {
	if t.currentTopic < len(t.topics) {
		t.currentTopic++
	}
}

func (t *InteractiveTutorial) Prev() {
	if t.currentTopic > 0 {
		t.currentTopic--
	}
}

func (t *InteractiveTutorial) Progress() float64 {
	if len(t.topics) == 0 {
		return 0
	}
	return float64(t.currentTopic) / float64(len(t.topics)) * 100
}

type HelpSystem struct {
	topics map[string]string
}

func NewHelpSystem() *HelpSystem {
	return &HelpSystem{
		topics: map[string]string{
			"compression": "TokMan uses a 31-layer compression pipeline to reduce token usage by 60-90%.",
			"filters":     "Filters are TOML-based rules that transform command output before it reaches the LLM.",
			"presets":     "Presets (fast, balanced, full) control how many compression layers are applied.",
			"hooks":       "Hooks integrate TokMan with AI coding assistants like Claude Code, Cursor, and Copilot.",
			"dashboard":   "The dashboard provides real-time analytics on token usage, savings, and costs.",
			"gateway":     "The gateway routes requests to multiple LLM providers with fallback chains.",
			"security":    "Security features include PII detection, prompt injection scanning, and secret redaction.",
		},
	}
}

func (h *HelpSystem) Get(topic string) string {
	return h.topics[topic]
}

func (h *HelpSystem) ListTopics() []string {
	var topics []string
	for k := range h.topics {
		topics = append(topics, k)
	}
	return topics
}

func (h *HelpSystem) Search(query string) map[string]string {
	results := make(map[string]string)
	query = strings.ToLower(query)
	for topic, content := range h.topics {
		if strings.Contains(strings.ToLower(topic), query) || strings.Contains(strings.ToLower(content), query) {
			results[topic] = content
		}
	}
	return results
}
