package filter

import (
	"fmt"
	"os/exec"
	"strings"
	"sync"
)

// AutoValidationPipeline implements auto-validation after file changes.
// Inspired by lean-ctx's auto-validation pipeline.
type AutoValidationPipeline struct {
	mu         sync.RWMutex
	validators []Validator
	lastResult ValidationResult
}

// Validator represents a validation step.
type Validator struct {
	Name    string
	Command string
	Args    []string
}

// ValidationResult holds validation results.
type ValidationResult struct {
	Success  bool
	Errors   []string
	Warnings []string
	Duration int64
}

// NewAutoValidationPipeline creates a new auto-validation pipeline.
func NewAutoValidationPipeline() *AutoValidationPipeline {
	return &AutoValidationPipeline{}
}

// AddValidator adds a validation step.
func (avp *AutoValidationPipeline) AddValidator(name, command string, args ...string) {
	avp.mu.Lock()
	defer avp.mu.Unlock()
	avp.validators = append(avp.validators, Validator{Name: name, Command: command, Args: args})
}

// Validate runs all validators.
func (avp *AutoValidationPipeline) Validate() ValidationResult {
	avp.mu.Lock()
	defer avp.mu.Unlock()

	result := ValidationResult{Success: true}

	for _, v := range avp.validators {
		cmd := exec.Command(v.Command, v.Args...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			result.Success = false
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %s", v.Name, strings.TrimSpace(string(output))))
		} else if len(output) > 0 {
			result.Warnings = append(result.Warnings, fmt.Sprintf("%s: %s", v.Name, strings.TrimSpace(string(output))))
		}
	}

	avp.lastResult = result
	return result
}

// QualityScorer implements AST, identifier, and line preservation scoring.
// Inspired by lean-ctx's quality scorer.
type QualityScorer struct{}

// NewQualityScorer creates a new quality scorer.
func NewQualityScorer() *QualityScorer {
	return &QualityScorer{}
}

// Score computes quality score for compressed content.
func (qs *QualityScorer) Score(original, compressed string) float64 {
	if len(original) == 0 || len(compressed) == 0 {
		return 0
	}

	// AST preservation score (check for function/class signatures)
	astScore := computeASTPreservation(original, compressed)

	// Identifier preservation score
	idScore := computeIdentifierPreservation(original, compressed)

	// Line preservation score
	lineScore := computeLinePreservation(original, compressed)

	// Weighted average
	return astScore*0.4 + idScore*0.3 + lineScore*0.3
}

func computeASTPreservation(original, compressed string) float64 {
	origFuncs := countFunctions(original)
	compFuncs := countFunctions(compressed)
	if origFuncs == 0 {
		return 1.0
	}
	return float64(compFuncs) / float64(origFuncs)
}

func computeIdentifierPreservation(original, compressed string) float64 {
	origIdentifiers := extractIdentifiers(original)
	compIdentifiers := extractIdentifiers(compressed)
	if len(origIdentifiers) == 0 {
		return 1.0
	}
	preserved := 0
	for _, id := range origIdentifiers {
		if contains(compIdentifiers, id) {
			preserved++
		}
	}
	return float64(preserved) / float64(len(origIdentifiers))
}

func computeLinePreservation(original, compressed string) float64 {
	origLines := strings.Split(original, "\n")
	compLines := strings.Split(compressed, "\n")
	if len(origLines) == 0 {
		return 1.0
	}
	return float64(len(compLines)) / float64(len(origLines))
}

func countFunctions(content string) int {
	count := 0
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "func ") || strings.HasPrefix(trimmed, "def ") ||
			strings.HasPrefix(trimmed, "class ") || strings.HasPrefix(trimmed, "type ") {
			count++
		}
	}
	return count
}

func extractIdentifiers(content string) []string {
	var ids []string
	seen := make(map[string]bool)
	for _, word := range strings.Fields(content) {
		word = strings.Trim(word, "(),.;:{}[]\"'")
		if len(word) > 2 && !seen[word] {
			ids = append(ids, word)
			seen[word] = true
		}
	}
	return ids
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
