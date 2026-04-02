package policy

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"
)

type PolicyMode string

const (
	ModeEnforce PolicyMode = "enforce"
	ModeShadow  PolicyMode = "shadow"
	ModeCanary  PolicyMode = "canary"
)

type PolicyRule struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Action  string `json:"action"`
	Pattern string `json:"pattern"`
	Enabled bool   `json:"enabled"`
}

type SecurityPolicy struct {
	Version   string       `json:"version"`
	SHA256    string       `json:"sha256"`
	Mode      PolicyMode   `json:"mode"`
	Rules     []PolicyRule `json:"rules"`
	UpdatedAt time.Time    `json:"updated_at"`
}

type PolicyEngine struct {
	policy   *SecurityPolicy
	mu       sync.RWMutex
	onChange []func(*SecurityPolicy)
}

func NewPolicyEngine() *PolicyEngine {
	return &PolicyEngine{
		policy: &SecurityPolicy{
			Version: "1.0.0",
			Mode:    ModeEnforce,
			Rules:   []PolicyRule{},
		},
	}
}

func (e *PolicyEngine) LoadPolicy(rules []PolicyRule) *SecurityPolicy {
	e.mu.Lock()
	defer e.mu.Unlock()

	oldPolicy := e.policy
	e.policy = &SecurityPolicy{
		Version:   incrementVersion(oldPolicy.Version),
		Mode:      oldPolicy.Mode,
		Rules:     rules,
		UpdatedAt: time.Now(),
	}
	e.policy.SHA256 = e.calculateHash()
	e.policy.Version = e.policy.SHA256[:8]

	for _, fn := range e.onChange {
		go fn(e.policy)
	}

	return e.policy
}

func (e *PolicyEngine) SetMode(mode PolicyMode) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.policy.Mode = mode
}

func (e *PolicyEngine) GetMode() PolicyMode {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.policy.Mode
}

func (e *PolicyEngine) GetPolicy() *SecurityPolicy {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.policy
}

func (e *PolicyEngine) OnChange(fn func(*SecurityPolicy)) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.onChange = append(e.onChange, fn)
}

func (e *PolicyEngine) Diff(newRules []PolicyRule) (added, removed, changed []PolicyRule) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	oldMap := make(map[string]PolicyRule)
	for _, r := range e.policy.Rules {
		oldMap[r.ID] = r
	}

	newMap := make(map[string]PolicyRule)
	for _, r := range newRules {
		newMap[r.ID] = r
	}

	for id, rule := range newMap {
		if _, exists := oldMap[id]; !exists {
			added = append(added, rule)
		} else if oldMap[id].Pattern != rule.Pattern || oldMap[id].Action != rule.Action {
			changed = append(changed, rule)
		}
	}

	for id, rule := range oldMap {
		if _, exists := newMap[id]; !exists {
			removed = append(removed, rule)
		}
	}

	return
}

func (e *PolicyEngine) Validate() []string {
	var errors []string
	if e.policy.Mode == "" {
		errors = append(errors, "policy mode is required")
	}
	for _, rule := range e.policy.Rules {
		if rule.ID == "" {
			errors = append(errors, "rule ID is required")
		}
		if rule.Pattern == "" {
			errors = append(errors, "rule pattern is required for rule "+rule.ID)
		}
	}
	return errors
}

func (e *PolicyEngine) calculateHash() string {
	data := ""
	for _, rule := range e.policy.Rules {
		data += rule.ID + rule.Pattern + rule.Action
	}
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

func incrementVersion(version string) string {
	return version + ".1"
}
