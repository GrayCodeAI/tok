package prune

import (
	"regexp"
	"strings"
)

type PruneEngine struct {
	lockfilePatterns []*regexp.Regexp
}

func NewPruneEngine() *PruneEngine {
	return &PruneEngine{
		lockfilePatterns: []*regexp.Regexp{
			regexp.MustCompile(`(?m)^name:\s+.+$`),
			regexp.MustCompile(`(?m)^version:\s+.+$`),
			regexp.MustCompile(`(?m)^resolved:\s+.+$`),
			regexp.MustCompile(`(?m)^integrity:\s+.+$`),
			regexp.MustCompile(`(?m)^\s+version:\s+.+$`),
		},
	}
}

func (e *PruneEngine) Prune(input string) string {
	lines := strings.Split(input, "\n")
	var filtered []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || trimmed == "---" || trimmed == "..." {
			continue
		}
		if strings.Contains(trimmed, "resolved") || strings.Contains(trimmed, "integrity") || strings.Contains(trimmed, "tarball") {
			continue
		}
		filtered = append(filtered, line)
	}
	return strings.Join(filtered, "\n")
}

func (e *PruneEngine) IsLockfile(input string) bool {
	markers := []string{"resolved", "integrity", "shasum", "tarball", "lockfileVersion"}
	lower := strings.ToLower(input)
	for _, m := range markers {
		if strings.Contains(lower, m) {
			return true
		}
	}
	return false
}

type TextPressEngine struct {
	endpoint string
}

func NewTextPressEngine(endpoint string) *TextPressEngine {
	if endpoint == "" {
		endpoint = "http://localhost:11434/api/generate"
	}
	return &TextPressEngine{endpoint: endpoint}
}

func (e *TextPressEngine) Compress(input string) string {
	if len(input) < 100 {
		return input
	}
	lines := strings.Split(input, "\n")
	var important []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if len(trimmed) > 20 && !strings.HasPrefix(trimmed, "#") && !strings.HasPrefix(trimmed, "//") {
			important = append(important, trimmed)
		}
	}
	if len(important) > 10 {
		important = important[:10]
	}
	return strings.Join(important, "\n")
}

type CacheSafe struct {
	enabled bool
}

func NewCacheSafe() *CacheSafe {
	return &CacheSafe{enabled: true}
}

func (c *CacheSafe) SetEnabled(enabled bool) {
	c.enabled = enabled
}

func (c *CacheSafe) IsEnabled() bool {
	return c.enabled
}

func (c *CacheSafe) Process(input string) string {
	if !c.enabled {
		return input
	}
	return input
}
