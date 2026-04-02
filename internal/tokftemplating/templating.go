package tokftemplating

import (
	"strings"
)

type TemplateEngine struct{}

func NewTemplateEngine() *TemplateEngine {
	return &TemplateEngine{}
}

func (e *TemplateEngine) Render(template string, data map[string]interface{}) string {
	result := template
	for key, value := range data {
		result = strings.ReplaceAll(result, "{{"+key+"}}", toString(value))
	}
	return result
}

func (e *TemplateEngine) CarryForward(data map[string]interface{}, keys []string) map[string]interface{} {
	result := make(map[string]interface{})
	for _, key := range keys {
		if val, ok := data[key]; ok {
			result[key] = val
		}
	}
	return result
}

func (e *TemplateEngine) ChildrenAs(data map[string]interface{}, key, delimiter string) string {
	if val, ok := data[key]; ok {
		if m, ok := val.(map[string]interface{}); ok {
			var items []string
			for k, v := range m {
				items = append(items, k+"="+toString(v))
			}
			return strings.Join(items, delimiter)
		}
	}
	return ""
}

func toString(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case int:
		return string(rune(val + '0'))
	case float64:
		return string(rune(int(val) + '0'))
	default:
		return ""
	}
}

type PermissionEngine struct {
	allowed map[string]bool
}

func NewPermissionEngine() *PermissionEngine {
	return &PermissionEngine{
		allowed: make(map[string]bool),
	}
}

func (e *PermissionEngine) Allow(command string) {
	e.allowed[command] = true
}

func (e *PermissionEngine) Deny(command string) {
	e.allowed[command] = false
}

func (e *PermissionEngine) IsAllowed(command string) bool {
	return e.allowed[command]
}

func (e *PermissionEngine) ListAllowed() []string {
	var result []string
	for cmd, allowed := range e.allowed {
		if allowed {
			result = append(result, cmd)
		}
	}
	return result
}

type InfoCommand struct{}

func NewInfoCommand() *InfoCommand {
	return &InfoCommand{}
}

func (c *InfoCommand) Show(configPath, dbPath string, filterCount int, version string) string {
	return "TokMan Info\n" +
		"Config: " + configPath + "\n" +
		"DB: " + dbPath + "\n" +
		"Filters: " + string(rune(filterCount+'0')) + "\n" +
		"Version: " + version + "\n"
}

type EnvVarHandler struct {
	prefixes []string
}

func NewEnvVarHandler(prefixes []string) *EnvVarHandler {
	if len(prefixes) == 0 {
		prefixes = []string{"TOKMAN_", "TK_", "TOKF_"}
	}
	return &EnvVarHandler{prefixes: prefixes}
}

func (h *EnvVarHandler) HasPrefix(key string) bool {
	for _, prefix := range h.prefixes {
		if strings.HasPrefix(key, prefix) {
			return true
		}
	}
	return false
}

func (h *EnvVarHandler) GetPrefixes() []string {
	return h.prefixes
}
