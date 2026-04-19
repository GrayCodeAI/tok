package filter

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/lakshmanpatel/tok/internal/core"
)

// Message represents a chat message with role and content.
type Message struct {
	Role    string
	Content string
}

// ConversationDedup deduplicates content across conversation turns using Jaccard similarity.
type ConversationDedup struct {
	threshold float64 // Jaccard similarity threshold
}

// NewConversationDedup creates a new cross-message deduplicator.
func NewConversationDedup() *ConversationDedup {
	return &ConversationDedup{threshold: 0.8}
}

// messageFingerprint holds shingles for a message.
type messageFingerprint struct {
	index    int
	shingles map[string]bool
}

// DedupStats tracks deduplication statistics.
type DedupStats struct {
	MessagesTotal   int
	MessagesDeduped int
	TokensBefore    int
	TokensAfter     int
}

// DeduplicateMessages removes duplicate content across messages.
func (d *ConversationDedup) DeduplicateMessages(messages []Message) ([]Message, *DedupStats) {
	if len(messages) == 0 {
		return messages, &DedupStats{}
	}

	stats := &DedupStats{
		MessagesTotal: len(messages),
	}

	// Calculate tokens before
	for _, msg := range messages {
		stats.TokensBefore += core.EstimateTokens(msg.Content)
	}

	kept := []messageFingerprint{}
	result := make([]Message, 0, len(messages))

	for idx, msg := range messages {
		content := msg.Content
		if len(strings.TrimSpace(content)) < 50 {
			result = append(result, msg)
			continue
		}

		shingles := computeShingles(content, 3)
		if len(shingles) < 2 {
			result = append(result, msg)
			kept = append(kept, messageFingerprint{idx, shingles})
			continue
		}

		// Check similarity with previous messages
		duplicateOf := -1
		for _, prev := range kept {
			sim := jaccardSimilarity(shingles, prev.shingles)
			if sim >= d.threshold {
				duplicateOf = prev.index
				break
			}
		}

		if duplicateOf >= 0 {
			// Replace with reference
			newMsg := Message{
				Role:    msg.Role,
				Content: fmt.Sprintf("[content similar to message %d — omitted]", duplicateOf),
			}
			result = append(result, newMsg)
			stats.MessagesDeduped++
		} else {
			result = append(result, msg)
			kept = append(kept, messageFingerprint{idx, shingles})
		}
	}

	// Calculate tokens after
	for _, msg := range result {
		stats.TokensAfter += core.EstimateTokens(msg.Content)
	}

	return result, stats
}

// computeShingles creates n-word shingles from text.
func computeShingles(text string, n int) map[string]bool {
	tokens := tokenizeWords(text)
	if len(tokens) < n {
		return make(map[string]bool)
	}

	shingles := make(map[string]bool)
	for i := 0; i <= len(tokens)-n; i++ {
		shingle := strings.Join(tokens[i:i+n], " ")
		shingles[shingle] = true
	}
	return shingles
}

// tokenizeWords splits text into lowercase word tokens.
func tokenizeWords(text string) []string {
	re := regexp.MustCompile(`[a-z0-9]+`)
	return re.FindAllString(strings.ToLower(text), -1)
}

// jaccardSimilarity computes Jaccard similarity between two sets.
func jaccardSimilarity(a, b map[string]bool) float64 {
	if len(a) == 0 && len(b) == 0 {
		return 1.0
	}

	intersection := 0
	for k := range a {
		if b[k] {
			intersection++
		}
	}

	union := len(a) + len(b) - intersection
	if union == 0 {
		return 0.0
	}

	return float64(intersection) / float64(union)
}
