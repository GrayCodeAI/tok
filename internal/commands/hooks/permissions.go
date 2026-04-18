package hooks

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// PermissionVerdict models Claude-style allow/ask/deny shell rules.
type PermissionVerdict string

const (
	PermissionAllow   PermissionVerdict = "allow"
	PermissionAsk     PermissionVerdict = "ask"
	PermissionDeny    PermissionVerdict = "deny"
	PermissionDefault PermissionVerdict = "default"
)

var checkCommandPermissions = func(cmd string) PermissionVerdict {
	denyRules, askRules, allowRules := loadPermissionRules()
	return checkCommandWithRules(cmd, denyRules, askRules, allowRules)
}

func checkCommandWithRules(cmd string, denyRules, askRules, allowRules []string) PermissionVerdict {
	segments := splitPermissionCommand(cmd)
	anyAsk := false
	allAllowed := true
	sawSegment := false

	for _, segment := range segments {
		segment = strings.TrimSpace(segment)
		if segment == "" {
			continue
		}
		sawSegment = true

		for _, pattern := range denyRules {
			if commandMatchesPattern(segment, pattern) {
				return PermissionDeny
			}
		}

		if !anyAsk {
			for _, pattern := range askRules {
				if commandMatchesPattern(segment, pattern) {
					anyAsk = true
					break
				}
			}
		}

		if allAllowed {
			matched := false
			for _, pattern := range allowRules {
				if commandMatchesPattern(segment, pattern) {
					matched = true
					break
				}
			}
			if !matched {
				allAllowed = false
			}
		}
	}

	if anyAsk {
		return PermissionAsk
	}
	if sawSegment && allAllowed && len(allowRules) > 0 {
		return PermissionAllow
	}
	return PermissionDefault
}

func loadPermissionRules() (denyRules, askRules, allowRules []string) {
	for _, path := range permissionSettingsPaths() {
		content, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		var root map[string]any
		if err := json.Unmarshal(content, &root); err != nil {
			continue
		}

		permissions, _ := root["permissions"].(map[string]any)
		if permissions == nil {
			continue
		}

		denyRules = append(denyRules, extractBashRules(permissions["deny"])...)
		askRules = append(askRules, extractBashRules(permissions["ask"])...)
		allowRules = append(allowRules, extractBashRules(permissions["allow"])...)
	}
	return denyRules, askRules, allowRules
}

func permissionSettingsPaths() []string {
	var paths []string
	if cwd, err := os.Getwd(); err == nil {
		for dir := cwd; dir != filepath.Dir(dir); dir = filepath.Dir(dir) {
			paths = append(paths,
				filepath.Join(dir, ".claude", "settings.json"),
				filepath.Join(dir, ".claude", "settings.local.json"),
			)
		}
	}
	if home, err := os.UserHomeDir(); err == nil {
		paths = append(paths,
			filepath.Join(home, ".claude", "settings.json"),
			filepath.Join(home, ".claude", "settings.local.json"),
		)
	}
	return paths
}

func extractBashRules(value any) []string {
	items, _ := value.([]any)
	if len(items) == 0 {
		return nil
	}
	rules := make([]string, 0, len(items))
	for _, item := range items {
		text, _ := item.(string)
		if strings.HasPrefix(text, "Bash(") && strings.HasSuffix(text, ")") {
			rules = append(rules, strings.TrimSuffix(strings.TrimPrefix(text, "Bash("), ")"))
		}
	}
	return rules
}

func splitPermissionCommand(cmd string) []string {
	replacer := strings.NewReplacer("&&", "\n", "||", "\n", ";", "\n")
	parts := strings.Split(replacer.Replace(cmd), "\n")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func commandMatchesPattern(cmd, pattern string) bool {
	if pattern == "*" {
		return true
	}

	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSpace(strings.TrimSuffix(pattern, "*"))
		prefix = strings.TrimSuffix(prefix, ":")
		if prefix == "" {
			return true
		}
		if !strings.Contains(prefix, "*") {
			return cmd == prefix || strings.HasPrefix(cmd, prefix+" ")
		}
	}

	if strings.Contains(pattern, "*") {
		return globMatches(cmd, pattern)
	}

	return cmd == pattern || strings.HasPrefix(cmd, pattern+" ")
}

func globMatches(cmd, pattern string) bool {
	normalized := strings.ReplaceAll(strings.ReplaceAll(pattern, ":*", " *"), "*:", "* ")
	parts := strings.Split(normalized, "*")
	if len(parts) == 0 {
		return false
	}
	if strings.TrimSpace(strings.Join(parts, "")) == "" {
		return true
	}

	searchFrom := 0
	for i, part := range parts {
		if part == "" {
			continue
		}
		if i == 0 && !strings.HasPrefix(normalized, "*") {
			if !strings.HasPrefix(cmd, part) {
				return false
			}
			searchFrom = len(part)
			continue
		}
		index := strings.Index(cmd[searchFrom:], part)
		if index < 0 {
			return false
		}
		searchFrom += index + len(part)
	}

	last := parts[len(parts)-1]
	if last != "" && !strings.HasSuffix(normalized, "*") && !strings.HasSuffix(cmd, last) {
		return false
	}
	return true
}
