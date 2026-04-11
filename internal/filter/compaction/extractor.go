package compaction

import (
	"regexp"
	"strings"
)

// Pre-compiled regexes for extraction
var (
	reCriticalError = regexp.MustCompile(`(?i)(error|failed|exception)[:：]\s*(.+)`)
	reCriticalFile  = regexp.MustCompile(`(?i)(file|path)[:：]\s*(.+)`)
	reCriticalTodo  = regexp.MustCompile(`(?i)(todo|fixme|important|note)[:：]\s*(.+)`)
	reKVGeneral     = regexp.MustCompile(`(?i)(\w+)\s*[:：=]\s*([^\n]+)`)
	reNextAction    = regexp.MustCompile(`next[:：]\s*(.+)`)
)

// ContentExtractor extracts critical information from content
type ContentExtractor struct {
	maxEntries int
}

// NewContentExtractor creates a new extractor
func NewContentExtractor(maxEntries int) *ContentExtractor {
	return &ContentExtractor{maxEntries: maxEntries}
}

// ExtractCritical extracts critical information
func (e *ContentExtractor) ExtractCritical(content string) []string {
	var critical []string
	
	// Extract errors
	matches := reCriticalError.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) > 1 {
			critical = append(critical, "Error: "+match[2])
		}
	}
	
	// Extract file references
	matches = reCriticalFile.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) > 1 {
			critical = append(critical, "File: "+match[2])
		}
	}
	
	// Extract TODOs
	matches = reCriticalTodo.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) > 1 {
			critical = append(critical, "TODO: "+match[2])
		}
	}
	
	// Limit entries
	if len(critical) > e.maxEntries {
		critical = critical[:e.maxEntries]
	}
	
	return critical
}

// ExtractKeyValuePairs extracts key-value pairs
func (e *ContentExtractor) ExtractKeyValuePairs(content string) map[string]string {
	pairs := make(map[string]string)
	
	matches := reKVGeneral.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) > 2 {
			key := strings.TrimSpace(match[1])
			value := strings.TrimSpace(match[2])
			if key != "" && value != "" {
				pairs[key] = value
			}
		}
	}
	
	return pairs
}

// ExtractNextAction extracts next action hints
func (e *ContentExtractor) ExtractNextAction(content string) string {
	matches := reNextAction.FindStringSubmatch(content)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}

// ParseTurns parses conversation into turns
func (e *ContentExtractor) ParseTurns(content string) []Turn {
	var turns []Turn
	lines := strings.Split(content, "\n")
	
	var currentTurn Turn
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		
		if strings.HasPrefix(trimmed, "User:") || 
		   strings.HasPrefix(trimmed, "Human:") {
			if currentTurn.Role != "" {
				turns = append(turns, currentTurn)
			}
			currentTurn = Turn{
				Role:    "user",
				Content: strings.TrimPrefix(trimmed, "User:"),
			}
		} else if strings.HasPrefix(trimmed, "Assistant:") || 
		          strings.HasPrefix(trimmed, "AI:") {
			if currentTurn.Role != "" {
				turns = append(turns, currentTurn)
			}
			currentTurn = Turn{
				Role:    "assistant",
				Content: strings.TrimPrefix(trimmed, "Assistant:"),
			}
		} else if currentTurn.Role != "" {
			currentTurn.Content += "\n" + line
		}
	}
	
	if currentTurn.Role != "" {
		turns = append(turns, currentTurn)
	}
	
	return turns
}

// Turn represents a single conversation turn
type Turn struct {
	Role    string
	Content string
}
