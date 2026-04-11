package compaction

import (
	"regexp"
	"strings"
)

// Pre-compiled regexes for conversation detection
var (
	conversationMarkers = []*regexp.Regexp{
		regexp.MustCompile(`(?i)^(user|assistant|human|ai|bot):`),
		regexp.MustCompile(`(?i)\[(user|assistant)\]`),
		regexp.MustCompile(`(?i)^>\s`),
	}
	
	chatPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)(^|\n)(user|assistant|human|ai):\s*\n`),
		regexp.MustCompile(`(?i)\\n\\n(User|Assistant):\\s`),
	}
)

// ConversationDetector identifies conversation-style content
type ConversationDetector struct {
	contentTypes []string
}

// NewConversationDetector creates a new detector
func NewConversationDetector(contentTypes []string) *ConversationDetector {
	return &ConversationDetector{contentTypes: contentTypes}
}

// Detect checks if content is conversation-style
func (d *ConversationDetector) Detect(content string) (bool, string) {
	// Check conversation markers
	for _, marker := range conversationMarkers {
		if marker.MatchString(content) {
			return true, "conversation"
		}
	}
	
	// Check chat patterns
	for _, pattern := range chatPatterns {
		if pattern.MatchString(content) {
			return true, "chat"
		}
	}
	
	// Check line-based patterns
	lines := strings.Split(content, "\n")
	turnCount := 0
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "User:") || 
		   strings.HasPrefix(trimmed, "Assistant:") ||
		   strings.HasPrefix(trimmed, "Human:") ||
		   strings.HasPrefix(trimmed, "AI:") {
			turnCount++
		}
	}
	
	if turnCount >= 3 {
		return true, "chat"
	}
	
	return false, ""
}

// ShouldCompact checks if content should be compacted
func (d *ConversationDetector) ShouldCompact(content string, thresholdLines, thresholdTokens int) bool {
	isConversation, _ := d.Detect(content)
	if !isConversation {
		return false
	}
	
	lines := strings.Count(content, "\n")
	tokens := len(content) / 4 // Rough estimate
	
	return lines >= thresholdLines || tokens >= thresholdTokens
}
