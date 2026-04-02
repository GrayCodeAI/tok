package ragkb

import (
	"fmt"
	"strings"
	"sync"
)

type KnowledgeEntry struct {
	ID      string `json:"id"`
	Agent   string `json:"agent"`
	Content string `json:"content"`
	Tag     string `json:"tag"`
}

type RAGKnowledgeBase struct {
	entries map[string][]*KnowledgeEntry
	mu      sync.RWMutex
}

func NewRAGKnowledgeBase() *RAGKnowledgeBase {
	return &RAGKnowledgeBase{
		entries: make(map[string][]*KnowledgeEntry),
	}
}

func (kb *RAGKnowledgeBase) Add(agent string, entry *KnowledgeEntry) {
	kb.mu.Lock()
	defer kb.mu.Unlock()
	entry.Agent = agent
	kb.entries[agent] = append(kb.entries[agent], entry)
}

func (kb *RAGKnowledgeBase) Query(agent, query string) []*KnowledgeEntry {
	kb.mu.RLock()
	defer kb.mu.RUnlock()
	query = strings.ToLower(query)
	var results []*KnowledgeEntry
	for _, entry := range kb.entries[agent] {
		if strings.Contains(strings.ToLower(entry.Content), query) ||
			strings.Contains(strings.ToLower(entry.Tag), query) {
			results = append(results, entry)
		}
	}
	return results
}

func (kb *RAGKnowledgeBase) GetAll(agent string) []*KnowledgeEntry {
	kb.mu.RLock()
	defer kb.mu.RUnlock()
	return kb.entries[agent]
}

func (kb *RAGKnowledgeBase) Count(agent string) int {
	kb.mu.RLock()
	defer kb.mu.RUnlock()
	return len(kb.entries[agent])
}

type OTelMetric struct {
	Name   string            `json:"name"`
	Value  float64           `json:"value"`
	Labels map[string]string `json:"labels"`
}

type OTelCollector struct {
	metrics []OTelMetric
	mu      sync.RWMutex
}

func NewOTelCollector() *OTelCollector {
	return &OTelCollector{}
}

func (c *OTelCollector) Record(name string, value float64, labels map[string]string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.metrics = append(c.metrics, OTelMetric{Name: name, Value: value, Labels: labels})
}

func (c *OTelCollector) GetMetrics() []OTelMetric {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return append([]OTelMetric{}, c.metrics...)
}

func (c *OTelCollector) Export() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var result string
	for _, m := range c.metrics {
		labels := ""
		for k, v := range m.Labels {
			labels += k + "=" + v + ","
		}
		result += m.Name + "{" + labels + "} " + fmt.Sprintf("%v", m.Value) + "\n"
	}
	return result
}

func (c *OTelCollector) Trace(name string) string {
	c.Record("trace."+name, 1, map[string]string{"span": name})
	return name
}
