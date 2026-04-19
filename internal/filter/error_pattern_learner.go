package filter

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"
)

// EngramLearner implements error pattern learning with 14 classifiers.
// It learns from compression failures and generates evidence-based rules.
type EngramLearner struct {
	mu          sync.RWMutex
	rules       []EngramRule
	patterns    map[string]*ErrorPattern
	storagePath string
	enabled     bool
}

// EngramRule represents a learned compression rule.
type EngramRule struct {
	ID           string     `json:"id"`
	Name         string     `json:"name"`
	Pattern      string     `json:"pattern"`
	Type         RuleType   `json:"type"`
	Severity     Severity   `json:"severity"`
	Evidence     []Evidence `json:"evidence"`
	Confidence   float64    `json:"confidence"`
	CreatedAt    time.Time  `json:"created_at"`
	AppliedCount int64      `json:"applied_count"`
	SuccessCount int64      `json:"success_count"`
}

// RuleType defines the type of engram rule.
type RuleType string

const (
	RuleTypePreserve  RuleType = "preserve"   // Always preserve matching content
	RuleTypeCompress  RuleType = "compress"   // Aggressively compress
	RuleTypeSkipLayer RuleType = "skip_layer" // Skip specific layer for this content
	RuleTypeBoost     RuleType = "boost"      // Boost importance score
	RuleTypeReduce    RuleType = "reduce"     // Reduce importance score
)

// Severity defines rule severity.
type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityHigh     Severity = "high"
	SeverityMedium   Severity = "medium"
	SeverityLow      Severity = "low"
)

// Evidence represents a single observation that supports the rule.
type Evidence struct {
	Timestamp   time.Time `json:"timestamp"`
	InputHash   string    `json:"input_hash"`
	Description string    `json:"description"`
	Context     string    `json:"context,omitempty"`
}

// ErrorPattern tracks error occurrences for pattern detection.
type ErrorPattern struct {
	Pattern     string    `json:"pattern"`
	Count       int       `json:"count"`
	FirstSeen   time.Time `json:"first_seen"`
	LastSeen    time.Time `json:"last_seen"`
	SampleInput string    `json:"sample_input,omitempty"`
}

// ErrorClassifier defines a specific error pattern classifier.
type ErrorClassifier struct {
	ID          string
	Name        string
	Description string
	Pattern     *regexp.Regexp
	RuleType    RuleType
	Severity    Severity
}

// The 14 error pattern classifiers from Claw Compactor.
var defaultClassifiers = []ErrorClassifier{
	{
		ID:          "stack_trace_loss",
		Name:        "Stack Trace Compression Loss",
		Description: "Detects when stack traces are over-compressed losing critical line info",
		Pattern:     regexp.MustCompile(`(?i)(at\s+\w+\.\w+\(|line\s+\d+|goroutine\s+\d+|traceback|stack)`),
		RuleType:    RuleTypePreserve,
		Severity:    SeverityCritical,
	},
	{
		ID:          "error_code_drop",
		Name:        "Error Code Dropped",
		Description: "Detects when error codes or error types are removed",
		Pattern:     regexp.MustCompile(`(?i)(error\s*[:\-]?\s*\w+|errno|exit\s*code|status\s*:\s*\d+)`),
		RuleType:    RuleTypePreserve,
		Severity:    SeverityCritical,
	},
	{
		ID:          "json_schema_broken",
		Name:        "JSON Schema Breakage",
		Description: "Detects when JSON keys are removed breaking schema",
		Pattern:     regexp.MustCompile(`"[a-zA-Z_][a-zA-Z0-9_]*"\s*:`),
		RuleType:    RuleTypePreserve,
		Severity:    SeverityHigh,
	},
	{
		ID:          "import_stripped",
		Name:        "Import Statement Stripped",
		Description: "Detects when imports/includes are over-compressed",
		Pattern:     regexp.MustCompile(`(?i)^(import\s|#include|require\s*\(|from\s+\w+\s+import)`),
		RuleType:    RuleTypePreserve,
		Severity:    SeverityHigh,
	},
	{
		ID:          "function_sig_loss",
		Name:        "Function Signature Loss",
		Description: "Detects when function signatures are mangled",
		Pattern:     regexp.MustCompile(`(?i)(func\s+\w+|def\s+\w+\s*\(|function\s+\w+)`),
		RuleType:    RuleTypePreserve,
		Severity:    SeverityHigh,
	},
	{
		ID:          "type_annotation_drop",
		Name:        "Type Annotation Removal",
		Description: "Detects when type annotations are stripped",
		Pattern:     regexp.MustCompile(`(?i)(:\s*(string|int|bool|float|void|any)|\-\>\s*\w+)`),
		RuleType:    RuleTypeBoost,
		Severity:    SeverityMedium,
	},
	{
		ID:          "url_truncation",
		Name:        "URL Truncation",
		Description: "Detects when URLs are truncated",
		Pattern:     regexp.MustCompile(`https?://[^\s]+`),
		RuleType:    RuleTypePreserve,
		Severity:    SeverityMedium,
	},
	{
		ID:          "uuid_id_loss",
		Name:        "UUID/ID Loss",
		Description: "Detects when UUIDs or IDs are compressed",
		Pattern:     regexp.MustCompile(`\b[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}\b|\b[A-Z0-9]{16,}\b`),
		RuleType:    RuleTypePreserve,
		Severity:    SeverityHigh,
	},
	{
		ID:          "timestamp_drop",
		Name:        "Timestamp Removal",
		Description: "Detects when timestamps are stripped",
		Pattern:     regexp.MustCompile(`\d{4}-\d{2}-\d{2}[T ]\d{2}:\d{2}:\d{2}`),
		RuleType:    RuleTypeBoost,
		Severity:    SeverityLow,
	},
	{
		ID:          "path_shortening_issue",
		Name:        "Path Over-Shortening",
		Description: "Detects when file paths are shortened too aggressively",
		Pattern:     regexp.MustCompile(`(?i)(/home/\w+|/users/\w+|[a-z]:\\\w+|/\w+\.\w{2,4})`),
		RuleType:    RuleTypeBoost,
		Severity:    SeverityMedium,
	},
	{
		ID:          "log_level_drop",
		Name:        "Log Level Stripping",
		Description: "Detects when log levels are removed",
		Pattern:     regexp.MustCompile(`(?i)\[(debug|info|warn|error|fatal|trace)\]|\b(DEBUG|INFO|WARN|ERROR|FATAL)\b`),
		RuleType:    RuleTypePreserve,
		Severity:    SeverityMedium,
	},
	{
		ID:          "config_key_loss",
		Name:        "Configuration Key Loss",
		Description: "Detects when config keys are compressed",
		Pattern:     regexp.MustCompile(`(?i)^[a-z_][a-z0-9_]*\s*[=:]\s*\w+`),
		RuleType:    RuleTypePreserve,
		Severity:    SeverityHigh,
	},
	{
		ID:          "variable_name_mangle",
		Name:        "Variable Name Mangling",
		Description: "Detects when variable names are shortened",
		Pattern:     regexp.MustCompile(`\b[a-z_][a-z0-9_]{2,}\s*=`),
		RuleType:    RuleTypeBoost,
		Severity:    SeverityLow,
	},
	{
		ID:          "comment_over_removal",
		Name:        "Comment Over-Removal",
		Description: "Detects when docstrings/comments with value are removed",
		Pattern:     regexp.MustCompile(`(?i)(TODO|FIXME|BUG|HACK|XXX|NOTE):?\s*\w+`),
		RuleType:    RuleTypeBoost,
		Severity:    SeverityLow,
	},
}

// NewEngramLearner creates a new engram learner.
func NewEngramLearner() *EngramLearner {
	storagePath := getEngramStoragePath()
	el := &EngramLearner{
		rules:       make([]EngramRule, 0),
		patterns:    make(map[string]*ErrorPattern),
		storagePath: storagePath,
		enabled:     true,
	}
	el.LoadRules()
	return el
}

// getEngramStoragePath returns the storage path for engram rules.
func getEngramStoragePath() string {
	if path := os.Getenv("TOK_ENGRAM_PATH"); path != "" {
		return path
	}
	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, ".config", "tok", "engram_rules.json")
	}
	return "engram_rules.json"
}

// Name returns the filter name.
func (el *EngramLearner) Name() string { return "engram_learner" }

// Apply runs the engram learner to generate rules from input.
func (el *EngramLearner) Apply(input string, mode Mode) (string, int) {
	if !el.enabled || mode == ModeNone {
		return input, 0
	}

	// Analyze input for error patterns
	el.analyzeInput(input)

	// Return input unchanged (this is a learning layer, not a compression layer)
	return input, 0
}

// analyzeInput scans for patterns and updates error statistics.
func (el *EngramLearner) analyzeInput(input string) {
	el.mu.Lock()
	defer el.mu.Unlock()

	for _, classifier := range defaultClassifiers {
		if classifier.Pattern.MatchString(input) {
			pattern := classifier.Pattern.String()
			if ep, exists := el.patterns[pattern]; exists {
				ep.Count++
				ep.LastSeen = time.Now()
			} else {
				el.patterns[pattern] = &ErrorPattern{
					Pattern:     pattern,
					Count:       1,
					FirstSeen:   time.Now(),
					LastSeen:    time.Now(),
					SampleInput: truncateString(input, 200),
				}
			}

			// Generate rule if pattern is frequent enough
			if el.patterns[pattern].Count >= 3 {
				el.generateRule(classifier, input)
			}
		}
	}
}

// generateRule creates a new rule from a classifier and evidence.
func (el *EngramLearner) generateRule(classifier ErrorClassifier, input string) {
	ruleID := hashString(classifier.ID + input[:min(len(input), 100)])

	// Check if rule already exists
	for _, r := range el.rules {
		if r.ID == ruleID {
			return
		}
	}

	rule := EngramRule{
		ID:           ruleID,
		Name:         classifier.Name,
		Pattern:      classifier.Pattern.String(),
		Type:         classifier.RuleType,
		Severity:     classifier.Severity,
		Confidence:   0.7,
		CreatedAt:    time.Now(),
		AppliedCount: 0,
		SuccessCount: 0,
		Evidence: []Evidence{
			{
				Timestamp:   time.Now(),
				InputHash:   hashString(input),
				Description: fmt.Sprintf("Pattern detected: %s", classifier.Description),
				Context:     extractContext(input, classifier.Pattern),
			},
		},
	}

	el.rules = append(el.rules, rule)
	el.SaveRules()
}

// GetRules returns all learned rules.
func (el *EngramLearner) GetRules() []EngramRule {
	el.mu.RLock()
	defer el.mu.RUnlock()

	result := make([]EngramRule, len(el.rules))
	copy(result, el.rules)
	return result
}

// GetRulesForContent returns applicable rules for given content.
func (el *EngramLearner) GetRulesForContent(content string) []EngramRule {
	el.mu.RLock()
	defer el.mu.RUnlock()

	var applicable []EngramRule
	for _, rule := range el.rules {
		if matched, _ := regexp.MatchString(rule.Pattern, content); matched {
			applicable = append(applicable, rule)
		}
	}
	return applicable
}

// RecordSuccess records a successful application of a rule.
func (el *EngramLearner) RecordSuccess(ruleID string) {
	el.mu.Lock()
	defer el.mu.Unlock()

	for i := range el.rules {
		if el.rules[i].ID == ruleID {
			el.rules[i].AppliedCount++
			el.rules[i].SuccessCount++
			// Increase confidence on success
			if el.rules[i].Confidence < 0.95 {
				el.rules[i].Confidence += 0.05
			}
			break
		}
	}
}

// RecordFailure records a failed application of a rule.
func (el *EngramLearner) RecordFailure(ruleID string) {
	el.mu.Lock()
	defer el.mu.Unlock()

	for i := range el.rules {
		if el.rules[i].ID == ruleID {
			el.rules[i].AppliedCount++
			// Decrease confidence on failure
			if el.rules[i].Confidence > 0.3 {
				el.rules[i].Confidence -= 0.1
			}
			break
		}
	}
}

// SaveRules persists rules to disk.
func (el *EngramLearner) SaveRules() error {
	el.mu.RLock()
	defer el.mu.RUnlock()

	if err := os.MkdirAll(filepath.Dir(el.storagePath), 0700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(el.rules, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(el.storagePath, data, 0600)
}

// LoadRules loads rules from disk.
func (el *EngramLearner) LoadRules() error {
	el.mu.Lock()
	defer el.mu.Unlock()

	data, err := os.ReadFile(el.storagePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No rules file yet
		}
		return err
	}

	return json.Unmarshal(data, &el.rules)
}

// GetStats returns learning statistics.
func (el *EngramLearner) GetStats() map[string]interface{} {
	el.mu.RLock()
	defer el.mu.RUnlock()

	return map[string]interface{}{
		"rules_learned":    len(el.rules),
		"patterns_tracked": len(el.patterns),
		"classifiers":      len(defaultClassifiers),
		"storage_path":     el.storagePath,
	}
}

// Helper functions

func hashString(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])[:16]
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func extractContext(input string, pattern *regexp.Regexp) string {
	loc := pattern.FindStringIndex(input)
	if loc == nil {
		return ""
	}
	start := max(0, loc[0]-50)
	end := min(len(input), loc[1]+50)
	return input[start:end]
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Compile-time check
var _ Filter = (*EngramLearner)(nil)
