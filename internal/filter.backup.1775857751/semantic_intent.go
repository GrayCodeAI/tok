package filter

import (
	"strings"
)

// SemanticIntentDetector detects semantic intent in queries.
// Inspired by lean-ctx's semantic intent detection.
type SemanticIntentDetector struct {
	intents map[string][]string
}

// NewSemanticIntentDetector creates a new semantic intent detector.
func NewSemanticIntentDetector() *SemanticIntentDetector {
	return &SemanticIntentDetector{
		intents: map[string][]string{
			"debug":    {"error", "bug", "fail", "crash", "panic", "stack trace", "exception"},
			"explain":  {"explain", "what does", "how does", "why", "describe", "understand"},
			"fix":      {"fix", "repair", "correct", "resolve", "patch"},
			"optimize": {"optimize", "improve", "faster", "performance", "slow"},
			"test":     {"test", "unit test", "integration", "coverage", "assert"},
			"refactor": {"refactor", "clean up", "restructure", "reorganize"},
			"review":   {"review", "check", "audit", "inspect", "analyze"},
			"generate": {"generate", "create", "write", "implement", "add"},
		},
	}
}

// DetectIntent detects the semantic intent of a query.
func (sid *SemanticIntentDetector) DetectIntent(query string) (string, float64) {
	queryLower := strings.ToLower(query)
	bestIntent := "unknown"
	bestScore := 0.0

	for intent, keywords := range sid.intents {
		score := 0.0
		for _, kw := range keywords {
			if strings.Contains(queryLower, kw) {
				score += 1.0
			}
		}
		if score > bestScore {
			bestScore = score
			bestIntent = intent
		}
	}

	if bestScore == 0 {
		return "general", 0.0
	}

	return bestIntent, bestScore / float64(len(sid.intents[bestIntent]))
}

// MultiAgentContextSharing implements multi-agent context sharing.
// Inspired by lean-ctx's multi-agent context sharing.
type MultiAgentContextSharing struct {
	agents     map[string]*AgentContext
	scratchpad []ScratchMessage
}

// AgentContext holds context for a single agent.
type AgentContext struct {
	Name       string
	Status     string
	LastActive string
	Tasks      []string
}

// ScratchMessage represents a scratchpad message.
type ScratchMessage struct {
	From      string
	To        string
	Content   string
	Timestamp string
}

// NewMultiAgentContextSharing creates a new multi-agent context sharing system.
func NewMultiAgentContextSharing() *MultiAgentContextSharing {
	return &MultiAgentContextSharing{
		agents: make(map[string]*AgentContext),
	}
}

// RegisterAgent registers an agent.
func (macs *MultiAgentContextSharing) RegisterAgent(name string) {
	macs.agents[name] = &AgentContext{Name: name, Status: "active"}
}

// PostMessage posts a message to the scratchpad.
func (macs *MultiAgentContextSharing) PostMessage(from, to, content string) {
	macs.scratchpad = append(macs.scratchpad, ScratchMessage{
		From:    from,
		To:      to,
		Content: content,
	})
}

// GetMessages gets messages for an agent.
func (macs *MultiAgentContextSharing) GetMessages(agent string) []ScratchMessage {
	var msgs []ScratchMessage
	for _, m := range macs.scratchpad {
		if m.To == agent || m.To == "*" {
			msgs = append(msgs, m)
		}
	}
	return msgs
}

// PersistentKnowledgeStore implements persistent knowledge storage.
// Inspired by lean-ctx's persistent knowledge store.
type PersistentKnowledgeStore struct {
	facts map[string]string
}

// NewPersistentKnowledgeStore creates a new knowledge store.
func NewPersistentKnowledgeStore() *PersistentKnowledgeStore {
	return &PersistentKnowledgeStore{facts: make(map[string]string)}
}

// Remember stores a fact.
func (pks *PersistentKnowledgeStore) Remember(key, value string) {
	pks.facts[key] = value
}

// Recall retrieves a fact.
func (pks *PersistentKnowledgeStore) Recall(key string) string {
	return pks.facts[key]
}

// QueryByCategory queries facts by category prefix.
func (pks *PersistentKnowledgeStore) QueryByCategory(category string) map[string]string {
	results := make(map[string]string)
	for k, v := range pks.facts {
		if strings.HasPrefix(k, category+":") {
			results[k] = v
		}
	}
	return results
}
