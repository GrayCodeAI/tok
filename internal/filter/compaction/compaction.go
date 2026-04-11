package compaction

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
	
	"github.com/GrayCodeAI/tokman/internal/core"
	"github.com/GrayCodeAI/tokman/internal/filter"
)

// Compact performs semantic compaction on conversation content
func (cl *CompactionLayer) Compact(input string, mode filter.Mode) (string, int, error) {
	// Check if compaction should be applied
	detector := NewConversationDetector(cl.config.ContentTypes)
	shouldCompact := detector.ShouldCompact(input, cl.config.ThresholdLines, cl.config.ThresholdTokens)
	
	if !shouldCompact {
		return input, 0, nil
	}
	
	// Check cache
	cacheKey := ""
	if cl.config.CacheEnabled {
		cacheKey = cl.hashInput(input)
		if cached := cl.getCached(cacheKey); cached != nil {
			return cached.Snapshot.String(), cached.SavedTokens, nil
		}
	}
	
	// Extract content
	extractor := NewContentExtractor(cl.config.MaxContextEntries)
	turns := extractor.ParseTurns(input)
	critical := extractor.ExtractCritical(input)
	nextAction := extractor.ExtractNextAction(input)
	
	// Build snapshot
	snapshot := &StateSnapshot{
		SessionHistory: cl.buildSessionHistory(turns),
		CurrentState: CurrentState{
			Focus:      "conversation",
			NextAction: nextAction,
		},
		Context: SnapshotContext{
			Critical: critical,
			Working:  cl.extractWorkingContext(turns),
		},
	}
	
	// Format output
	output := snapshot.String()
	saved := core.EstimateTokens(input) - core.EstimateTokens(output)
	
	// Cache result
	if cl.config.CacheEnabled {
		cl.cacheResult(cacheKey, &CompactionResult{
			Snapshot:         snapshot,
			OriginalTokens:   core.EstimateTokens(input),
			FinalTokens:      core.EstimateTokens(output),
			SavedTokens:      saved,
			CompressionRatio: float64(saved) / float64(core.EstimateTokens(input)),
			Cached:           false,
			Timestamp:        time.Now(),
		})
	}
	
	return output, saved, nil
}

// buildSessionHistory builds session history from turns
func (cl *CompactionLayer) buildSessionHistory(turns []Turn) SessionHistory {
	history := SessionHistory{
		UserQueries: []string{},
		ActivityLog: []string{},
	}
	
	for _, turn := range turns {
		if turn.Role == "user" {
			history.UserQueries = append(history.UserQueries, 
				truncate(turn.Content, 100))
		} else {
			history.ActivityLog = append(history.ActivityLog, 
				truncate(turn.Content, 100))
		}
	}
	
	return history
}

// extractWorkingContext extracts working context
func (cl *CompactionLayer) extractWorkingContext(turns []Turn) []string {
	var working []string
	recentTurns := turns
	if len(turns) > cl.config.PreserveRecentTurns {
		recentTurns = turns[len(turns)-cl.config.PreserveRecentTurns:]
	}
	
	for _, turn := range recentTurns {
		working = append(working, truncate(turn.Content, 200))
	}
	
	return working
}

// hashInput creates cache key
func (cl *CompactionLayer) hashInput(input string) string {
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:8]) // First 8 bytes is enough
}

// getCached retrieves cached result
func (cl *CompactionLayer) getCached(key string) *CompactionResult {
	cl.cacheMu.RLock()
	defer cl.cacheMu.RUnlock()
	
	if result, ok := cl.cache[key]; ok {
		// Check if not expired (5 minutes)
		if time.Since(result.Timestamp) < 5*time.Minute {
			return result
		}
	}
	return nil
}

// cacheResult stores result in cache
func (cl *CompactionLayer) cacheResult(key string, result *CompactionResult) {
	cl.cacheMu.Lock()
	defer cl.cacheMu.Unlock()
	
	// Simple LRU: if cache full, clear it
	if len(cl.cache) >= 100 {
		cl.cache = make(map[string]*CompactionResult)
	}
	
	cl.cache[key] = result
}

// String formats snapshot as string
func (s *StateSnapshot) String() string {
	return fmt.Sprintf(
		"Session History: %d queries, %d activities\n"+
		"Current Focus: %s\n"+
		"Next Action: %s\n"+
		"Critical Items: %d\n"+
		"Working Context: %d items",
		len(s.SessionHistory.UserQueries),
		len(s.SessionHistory.ActivityLog),
		s.CurrentState.Focus,
		s.CurrentState.NextAction,
		len(s.Context.Critical),
		len(s.Context.Working),
	)
}

// truncate truncates string to max length
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
