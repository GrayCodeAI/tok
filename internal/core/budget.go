// Package core provides budget enforcement for token optimization.
package core

import (
	"fmt"
	"strings"
)

// BudgetMode represents the budget enforcement mode.
type BudgetMode int

const (
	BudgetModeSoft     BudgetMode = iota // Warn but don't enforce
	BudgetModeStrict                     // Strict enforcement - truncate content
	BudgetModeAdaptive                   // Adapt compression based on budget
)

// BudgetConfig configures budget enforcement.
type BudgetConfig struct {
	MaxTokens     int
	Mode          BudgetMode
	PreserveRatio float64 // Ratio of budget for high-priority content
}

// DefaultBudgetConfig returns default configuration.
func DefaultBudgetConfig() BudgetConfig {
	return BudgetConfig{
		MaxTokens:     4000,
		Mode:          BudgetModeAdaptive,
		PreserveRatio: 0.7,
	}
}

// BudgetEnforcer enforces token budget constraints.
type BudgetEnforcer struct {
	config BudgetConfig
}

// NewBudgetEnforcer creates a new budget enforcer.
func NewBudgetEnforcer(config BudgetConfig) *BudgetEnforcer {
	return &BudgetEnforcer{config: config}
}

// BudgetStatus represents the current budget status.
type BudgetStatus struct {
	CurrentTokens   int
	MaxTokens       int
	RemainingTokens int
	PercentUsed     float64
	Exceeded        bool
}

// Check checks current content against budget.
func (be *BudgetEnforcer) Check(content string) BudgetStatus {
	tokens := EstimateTokens(content)
	return BudgetStatus{
		CurrentTokens:   tokens,
		MaxTokens:       be.config.MaxTokens,
		RemainingTokens: be.config.MaxTokens - tokens,
		PercentUsed:     float64(tokens) / float64(be.config.MaxTokens) * 100,
		Exceeded:        tokens > be.config.MaxTokens,
	}
}

// Enforce enforces the budget on content.
func (be *BudgetEnforcer) Enforce(content string) (string, EnforceResult) {
	status := be.Check(content)

	if !status.Exceeded {
		return content, EnforceResult{
			OriginalTokens:  status.CurrentTokens,
			FinalTokens:     status.CurrentTokens,
			WasEnforced:     false,
			ReductionMethod: "none",
		}
	}

	// Budget exceeded, need to enforce
	switch be.config.Mode {
	case BudgetModeSoft:
		return content, EnforceResult{
			OriginalTokens:  status.CurrentTokens,
			FinalTokens:     status.CurrentTokens,
			WasEnforced:     false,
			ReductionMethod: "warned",
			Warning:         fmt.Sprintf("Budget warning: %d/%d tokens (%.1f%%)", status.CurrentTokens, status.MaxTokens, status.PercentUsed),
		}

	case BudgetModeStrict:
		return be.enforceStrict(content, status)

	case BudgetModeAdaptive:
		return be.enforceAdaptive(content, status)

	default:
		return content, EnforceResult{
			OriginalTokens: status.CurrentTokens,
			FinalTokens:    status.CurrentTokens,
			WasEnforced:    false,
			Warning:        "Unknown budget mode",
		}
	}
}

// EnforceResult contains enforcement results.
type EnforceResult struct {
	OriginalTokens  int
	FinalTokens     int
	WasEnforced     bool
	ReductionMethod string
	StagesApplied   []string
	Warning         string
}

// TokensSaved returns tokens saved.
func (er EnforceResult) TokensSaved() int {
	return er.OriginalTokens - er.FinalTokens
}

// ReductionPercent returns reduction percentage.
func (er EnforceResult) ReductionPercent() float64 {
	if er.OriginalTokens == 0 {
		return 0
	}
	return float64(er.TokensSaved()) / float64(er.OriginalTokens) * 100
}

// enforceStrict truncates content strictly.
func (be *BudgetEnforcer) enforceStrict(content string, status BudgetStatus) (string, EnforceResult) {
	result := EnforceResult{
		OriginalTokens:  status.CurrentTokens,
		WasEnforced:     true,
		ReductionMethod: "truncate",
		StagesApplied:   []string{"truncate"},
	}

	// Keep the most important parts
	lines := strings.Split(content, "\n")
	var kept []string
	currentTokens := 0

	// First pass: keep essential lines
	for _, line := range lines {
		if be.isEssentialLine(line) {
			lineTokens := EstimateTokens(line)
			if currentTokens+lineTokens <= be.config.MaxTokens {
				kept = append(kept, line)
				currentTokens += lineTokens
			}
		}
	}

	// Second pass: fill remaining budget with non-essential lines
	for _, line := range lines {
		if be.isEssentialLine(line) {
			continue // Already handled
		}
		lineTokens := EstimateTokens(line)
		if currentTokens+lineTokens <= be.config.MaxTokens {
			kept = append(kept, line)
			currentTokens += lineTokens
		}
	}

	result.FinalTokens = currentTokens
	return strings.Join(kept, "\n"), result
}

// enforceAdaptive applies progressive compression.
func (be *BudgetEnforcer) enforceAdaptive(content string, status BudgetStatus) (string, EnforceResult) {
	result := EnforceResult{
		OriginalTokens: status.CurrentTokens,
		WasEnforced:    true,
		StagesApplied:  []string{},
	}

	currentContent := content
	_ = status.CurrentTokens - be.config.MaxTokens // target reduction if needed

	// Stage 1: Remove comments
	if status.CurrentTokens > be.config.MaxTokens {
		currentContent = removeComments(currentContent)
		result.StagesApplied = append(result.StagesApplied, "comments_removed")
		status = be.Check(currentContent)
	}

	// Stage 2: Normalize whitespace
	if status.CurrentTokens > be.config.MaxTokens {
		currentContent = normalizeWhitespace(currentContent)
		result.StagesApplied = append(result.StagesApplied, "whitespace_normalized")
		status = be.Check(currentContent)
	}

	// Stage 3: Collapse imports
	if status.CurrentTokens > be.config.MaxTokens {
		currentContent = collapseImports(currentContent)
		result.StagesApplied = append(result.StagesApplied, "imports_collapsed")
		status = be.Check(currentContent)
	}

	// Stage 4: Truncate non-essential sections
	if status.CurrentTokens > be.config.MaxTokens {
		currentContent = truncateNonEssential(currentContent, be.config.MaxTokens)
		result.StagesApplied = append(result.StagesApplied, "truncate")
		status = be.Check(currentContent)
	}

	result.FinalTokens = status.CurrentTokens
	result.ReductionMethod = "adaptive"

	return currentContent, result
}

// isEssentialLine determines if a line is essential.
func (be *BudgetEnforcer) isEssentialLine(line string) bool {
	trimmed := strings.TrimSpace(line)

	// Essential patterns
	essentialPatterns := []string{
		"func ",
		"class ",
		"struct ",
		"interface ",
		"package ",
		"import ",
		"error",
		"Error",
		"// ",
	}

	for _, pattern := range essentialPatterns {
		if strings.Contains(trimmed, pattern) {
			return true
		}
	}

	return false
}

// Helper functions

func removeComments(content string) string {
	lines := strings.Split(content, "\n")
	var result []string

	inBlockComment := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Block comment handling
		if strings.HasPrefix(trimmed, "/*") {
			inBlockComment = true
		}
		if inBlockComment {
			if strings.Contains(line, "*/") {
				inBlockComment = false
			}
			continue
		}

		// Line comments (preserve TODO/FIXME)
		if strings.HasPrefix(trimmed, "//") {
			if strings.Contains(trimmed, "TODO") ||
				strings.Contains(trimmed, "FIXME") ||
				strings.Contains(trimmed, "NOTE") {
				result = append(result, line)
			}
			continue
		}

		result = append(result, line)
	}

	return strings.Join(result, "\n")
}

func normalizeWhitespace(content string) string {
	// Replace multiple spaces with single space
	content = strings.ReplaceAll(content, "    ", "\t")
	// Remove trailing whitespace
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " \t")
	}
	return strings.Join(lines, "\n")
}

func collapseImports(content string) string {
	// Simple import collapsing
	lines := strings.Split(content, "\n")
	var result []string
	inImportBlock := false
	importCount := 0

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "import (") {
			inImportBlock = true
			continue
		}

		if inImportBlock && trimmed == ")" {
			inImportBlock = false
			if importCount > 3 {
				result = append(result, "import ( ... // "+fmt.Sprintf("%d imports", importCount))
			} else {
				// Restore original imports
				result = append(result, "import (")
				for j := i - importCount; j < i; j++ {
					if j >= 0 && j < len(lines) {
						result = append(result, lines[j])
					}
				}
				result = append(result, ")")
			}
			importCount = 0
			continue
		}

		if inImportBlock {
			importCount++
			continue
		}

		result = append(result, line)
	}

	return strings.Join(result, "\n")
}

func truncateNonEssential(content string, maxTokens int) string {
	lines := strings.Split(content, "\n")
	var essential []string
	var nonEssential []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		isEss := false

		// Check if essential
		for _, pattern := range []string{"func ", "class ", "struct ", "error", "// "} {
			if strings.Contains(trimmed, pattern) {
				isEss = true
				break
			}
		}

		if isEss {
			essential = append(essential, line)
		} else {
			nonEssential = append(nonEssential, line)
		}
	}

	// Add essential lines first
	var result []string
	currentTokens := 0

	for _, line := range essential {
		tokens := EstimateTokens(line)
		if currentTokens+tokens <= maxTokens {
			result = append(result, line)
			currentTokens += tokens
		}
	}

	// Fill with non-essential
	for _, line := range nonEssential {
		tokens := EstimateTokens(line)
		if currentTokens+tokens <= maxTokens {
			result = append(result, line)
			currentTokens += tokens
		}
	}

	return strings.Join(result, "\n")
}
