package policydiff

import (
	"encoding/json"
	"time"
)

type PolicyChange struct {
	Type      string    `json:"type"`
	RuleID    string    `json:"rule_id"`
	OldValue  string    `json:"old_value,omitempty"`
	NewValue  string    `json:"new_value,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

type PolicyDiffLogger struct {
	changes []PolicyChange
}

func NewPolicyDiffLogger() *PolicyDiffLogger {
	return &PolicyDiffLogger{}
}

func (l *PolicyDiffLogger) LogAdded(ruleID, value string) {
	l.changes = append(l.changes, PolicyChange{
		Type:      "added",
		RuleID:    ruleID,
		NewValue:  value,
		Timestamp: time.Now(),
	})
}

func (l *PolicyDiffLogger) LogRemoved(ruleID, value string) {
	l.changes = append(l.changes, PolicyChange{
		Type:      "removed",
		RuleID:    ruleID,
		OldValue:  value,
		Timestamp: time.Now(),
	})
}

func (l *PolicyDiffLogger) LogChanged(ruleID, oldVal, newVal string) {
	l.changes = append(l.changes, PolicyChange{
		Type:      "changed",
		RuleID:    ruleID,
		OldValue:  oldVal,
		NewValue:  newVal,
		Timestamp: time.Now(),
	})
}

func (l *PolicyDiffLogger) GetChanges() []PolicyChange {
	return l.changes
}

func (l *PolicyDiffLogger) ExportJSON() ([]byte, error) {
	return json.MarshalIndent(l.changes, "", "  ")
}

func (l *PolicyDiffLogger) Count() int {
	return len(l.changes)
}

type ContextCarrier struct {
	context map[string]string
}

func NewContextCarrier() *ContextCarrier {
	return &ContextCarrier{
		context: make(map[string]string),
	}
}

func (c *ContextCarrier) CarryRequest(key, value string) {
	c.context[key] = value
}

func (c *ContextCarrier) GetResponse(key string) string {
	return c.context[key]
}

func (c *ContextCarrier) GetAll() map[string]string {
	return c.context
}

func (c *ContextCarrier) Clear() {
	c.context = make(map[string]string)
}
