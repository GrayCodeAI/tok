package core

import (
	"regexp"
	"strings"
)

type MissedSaving struct {
	Command     string
	Reason      string
	Suggestion  string
	EstTokens   int
	EstSavings  int
}

type DiscoverAnalyzer struct {
	patterns []missedPattern
}

type missedPattern struct {
	name     string
	regex    *regexp.Regexp
	suggest  string
	estSave  int
}

func NewDiscoverAnalyzer() *DiscoverAnalyzer {
	d := &DiscoverAnalyzer{}
	d.patterns = []missedPattern{
		{name: "long_cat", regex: regexp.MustCompile(`\bcat\s+\S+`), suggest: "Use 'tokman read' instead of 'cat'", estSave: 80},
		{name: "verbose_git", regex: regexp.MustCompile(`\bgit\s+(status|diff|log)\b`), suggest: "Already optimized by tokman", estSave: 0},
		{name: "raw_ls", regex: regexp.MustCompile(`\bls\s+(-la|-lR|--long)`), suggest: "Use 'tokman ls' for token-optimized output", estSave: 70},
		{name: "full_test", regex: regexp.MustCompile(`\b(go\s+test|npm\s+test|cargo\s+test|pytest)\b`), suggest: "Use 'tokman test' to show only failures", estSave: 90},
		{name: "verbose_docker", regex: regexp.MustCompile(`\bdocker\s+(ps|images|logs)\b`), suggest: "Use 'tokman docker' for compact output", estSave: 80},
		{name: "raw_grep", regex: regexp.MustCompile(`\b(grep|rg)\s+[^|]+\s+\.`), suggest: "Use 'tokman grep' for grouped results", estSave: 75},
		{name: "full_build", regex: regexp.MustCompile(`\b(cargo\s+build|npm\s+run\s+build|go\s+build)\b`), suggest: "Use 'tokman build' to filter noise", estSave: 60},
		{name: "raw_env", regex: regexp.MustCompile(`\benv\b`), suggest: "Use 'tokman env -f' to filter sensitive values", estSave: 50},
		{name: "verbose_kubectl", regex: regexp.MustCompile(`\bkubectl\s+get\b`), suggest: "Use 'tokman kubectl' for compact output", estSave: 75},
		{name: "raw_curl", regex: regexp.MustCompile(`\bcurl\s+`), suggest: "Use 'tokman curl' for auto-JSON schema output", estSave: 60},
	}
	return d
}

func (d *DiscoverAnalyzer) Analyze(command string) []MissedSaving {
	var results []MissedSaving
	for _, p := range d.patterns {
		if p.regex.MatchString(command) {
			tokens := EstimateTokens(command)
			results = append(results, MissedSaving{
				Command:    command,
				Reason:     p.name,
				Suggestion: p.suggest,
				EstTokens:  tokens,
				EstSavings: tokens * p.estSave / 100,
			})
		}
	}
	return results
}

func (d *DiscoverAnalyzer) AnalyzeBatch(commands []string) []MissedSaving {
	var all []MissedSaving
	seen := make(map[string]bool)
	for _, cmd := range commands {
		results := d.Analyze(cmd)
		for _, r := range results {
			key := r.Command + r.Reason
			if !seen[key] {
				seen[key] = true
				all = append(all, r)
			}
		}
	}
	return all
}

func normalizeCommandForDiscover(cmd string) string {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return ""
	}
	return strings.ToLower(parts[0])
}
