package permissions

import (
	"sync"
)

type PermissionType string

const (
	PermissionRead   PermissionType = "read"
	PermissionWrite  PermissionType = "write"
	PermissionDelete PermissionType = "delete"
	PermissionAdmin  PermissionType = "admin"
)

type PermissionRule struct {
	Resource string         `json:"resource"`
	Action   PermissionType `json:"action"`
	Allowed  bool           `json:"allowed"`
}

type PermissionEngine struct {
	rules map[string][]PermissionRule
	mu    sync.RWMutex
}

func NewPermissionEngine() *PermissionEngine {
	return &PermissionEngine{
		rules: make(map[string][]PermissionRule),
	}
}

func (e *PermissionEngine) Grant(role string, resource string, action PermissionType) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.rules[role] = append(e.rules[role], PermissionRule{
		Resource: resource,
		Action:   action,
		Allowed:  true,
	})
}

func (e *PermissionEngine) Check(role, resource string, action PermissionType) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	for _, rule := range e.rules[role] {
		if rule.Resource == resource && rule.Action == action {
			return rule.Allowed
		}
		if rule.Resource == "*" && rule.Action == PermissionAdmin {
			return true
		}
	}
	return false
}

func (e *PermissionEngine) ListRules(role string) []PermissionRule {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.rules[role]
}

type EnvVarPrefix struct {
	Prefixes []string
}

func NewEnvVarPrefix(prefixes []string) *EnvVarPrefix {
	if len(prefixes) == 0 {
		prefixes = []string{"TOKMAN_", "TK_", "TOKF_"}
	}
	return &EnvVarPrefix{Prefixes: prefixes}
}

func (p *EnvVarPrefix) GetPrefixes() []string {
	return p.Prefixes
}

func (p *EnvVarPrefix) HasPrefix(key string) bool {
	for _, prefix := range p.Prefixes {
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			return true
		}
	}
	return false
}

type InfoCommand struct {
	ConfigPath  string `json:"config_path"`
	DBPath      string `json:"db_path"`
	FilterCount int    `json:"filter_count"`
	Version     string `json:"version"`
}

func NewInfoCommand(configPath, dbPath string, filterCount int, version string) *InfoCommand {
	return &InfoCommand{
		ConfigPath:  configPath,
		DBPath:      dbPath,
		FilterCount: filterCount,
		Version:     version,
	}
}

func (i *InfoCommand) Format() string {
	return "Config: " + i.ConfigPath + "\n" +
		"DB: " + i.DBPath + "\n" +
		"Filters: " + string(rune(i.FilterCount+'0')) + "\n" +
		"Version: " + i.Version
}
