package rtksmart

import (
	"strings"
)

type SmartReader struct {
	maxLines int
}

func NewSmartReader(maxLines int) *SmartReader {
	if maxLines == 0 {
		maxLines = 50
	}
	return &SmartReader{maxLines: maxLines}
}

func (r *SmartReader) ReadSignaturesOnly(content string) string {
	lines := strings.Split(content, "\n")
	var signatures []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if r.isSignature(trimmed) {
			signatures = append(signatures, line)
		}
	}
	if len(signatures) > r.maxLines {
		return strings.Join(signatures[:r.maxLines], "\n") + "\n..."
	}
	return strings.Join(signatures, "\n")
}

func (r *SmartReader) isSignature(line string) bool {
	patterns := []string{
		"func ", "function ", "def ", "class ", "interface ", "type ",
		"export ", "import ", "from ", "require(", "pub fn ",
		"struct ", "enum ", "trait ", "impl ", "mod ",
		"public ", "private ", "protected ", "static ",
		"#include", "package ", "module ", "namespace ",
		"@", "describe(", "it(", "test(", "func Test",
	}
	for _, p := range patterns {
		if strings.Contains(line, p) {
			return true
		}
	}
	return false
}

func (r *SmartReader) SmartSummary(content string) string {
	lines := strings.Split(content, "\n")
	if len(lines) <= 2 {
		return content
	}
	return lines[0] + "\n" + lines[1] + "\n[...]"
}

type CompactFind struct {
	Path  string `json:"path"`
	Match string `json:"match"`
	Line  int    `json:"line"`
}

func CompactFindResults(results []string) []CompactFind {
	var compact []CompactFind
	for _, r := range results {
		parts := strings.SplitN(r, ":", 3)
		if len(parts) >= 3 {
			compact = append(compact, CompactFind{
				Path:  parts[0],
				Match: parts[2],
			})
		}
	}
	return compact
}

type CompactGrep struct {
	Pattern string      `json:"pattern"`
	Groups  []GrepGroup `json:"groups"`
}

type GrepGroup struct {
	FilePath string `json:"file_path"`
	Lines    []int  `json:"lines"`
	Count    int    `json:"count"`
}

func GroupGrepResults(pattern string, results []string) *CompactGrep {
	groups := make(map[string]*GrepGroup)
	for _, r := range results {
		parts := strings.SplitN(r, ":", 3)
		if len(parts) >= 2 {
			file := parts[0]
			if _, ok := groups[file]; !ok {
				groups[file] = &GrepGroup{FilePath: file}
			}
			groups[file].Count++
		}
	}
	var groupList []GrepGroup
	for _, g := range groups {
		groupList = append(groupList, *g)
	}
	return &CompactGrep{Pattern: pattern, Groups: groupList}
}

type PipeStripper struct{}

func NewPipeStripper() *PipeStripper {
	return &PipeStripper{}
}

func (p *PipeStripper) Strip(command string) string {
	stripped := command
	pipeStrippers := []string{
		" | head ", " | tail ", " | grep ", " | sort ",
		" | uniq ", " | wc ", " | cut ", " | awk ",
		" | sed ", " | tr ", " | xargs ",
	}
	for _, ps := range pipeStrippers {
		if idx := strings.Index(stripped, ps); idx >= 0 {
			stripped = strings.TrimSpace(stripped[:idx])
		}
	}
	return stripped
}

type LogDedup struct{}

func NewLogDedup() *LogDedup {
	return &LogDedup{}
}

func (d *LogDedup) Deduplicate(input string) string {
	lines := strings.Split(input, "\n")
	seen := make(map[string]int)
	var result []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		count := seen[trimmed]
		if count == 0 {
			result = append(result, line)
		} else if count == 1 {
			result = append(result, "[repeated: "+trimmed[:min(50, len(trimmed))]+"]")
		}
		seen[trimmed] = count + 1
	}
	return strings.Join(result, "\n")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

type SessionTracker struct {
	sessions map[string]int
}

func NewSessionTracker() *SessionTracker {
	return &SessionTracker{
		sessions: make(map[string]int),
	}
}

func (t *SessionTracker) Record(sessionID string) {
	t.sessions[sessionID]++
}

func (t *SessionTracker) GetCount(sessionID string) int {
	return t.sessions[sessionID]
}

func (t *SessionTracker) GetAll() map[string]int {
	return t.sessions
}
