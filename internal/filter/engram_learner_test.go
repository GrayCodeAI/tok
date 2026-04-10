package filter

import (
	"path/filepath"
	"regexp"
	"testing"
)

func TestEngramLearner_New(t *testing.T) {
	el := NewEngramLearner()
	if el == nil {
		t.Fatal("expected non-nil EngramLearner")
	}
	if !el.enabled {
		t.Error("expected learner to be enabled")
	}
	if el.patterns == nil {
		t.Error("expected patterns map to be initialized")
	}
}

func TestEngramLearner_Name(t *testing.T) {
	el := NewEngramLearner()
	if el.Name() != "engram_learner" {
		t.Errorf("expected name 'engram_learner', got '%s'", el.Name())
	}
}

func TestEngramLearner_Apply(t *testing.T) {
	el := NewEngramLearner()

	// Test with mode None - should pass through unchanged
	input := "test content"
	output, saved := el.Apply(input, ModeNone)
	if output != input {
		t.Error("expected unchanged output with ModeNone")
	}
	if saved != 0 {
		t.Error("expected 0 savings with ModeNone")
	}

	// Test with disabled learner
	el.enabled = false
	output, saved = el.Apply(input, ModeMinimal)
	if output != input {
		t.Error("expected unchanged output when disabled")
	}
	if saved != 0 {
		t.Error("expected 0 savings when disabled")
	}
}

func TestEngramLearner_AnalyzePatterns(t *testing.T) {
	el := NewEngramLearner()

	// Test with stack trace pattern
	input := `goroutine 1 [running]:
main.main()
	/home/user/project/main.go:10 +0x39`
	el.Apply(input, ModeMinimal)

	// Check that patterns were detected
	el.mu.RLock()
	patternCount := len(el.patterns)
	el.mu.RUnlock()

	if patternCount == 0 {
		t.Error("expected patterns to be detected in stack trace")
	}
}

func TestEngramLearner_GenerateRule(t *testing.T) {
	el := NewEngramLearner()
	
	// Create a test classifier
	classifier := ErrorClassifier{
		ID:       "test",
		Name:     "Test Pattern",
		Pattern:  regexp.MustCompile("test"),
		RuleType: RuleTypePreserve,
		Severity: SeverityHigh,
	}

	// Use temp directory to avoid file I/O issues
	tmpDir := t.TempDir()
	el.storagePath = filepath.Join(tmpDir, "test_rules.json")

	// Manually trigger rule generation (generateRule handles its own locking)
	el.generateRule(classifier, "test input content")

	// Check that rule was created
	rules := el.GetRules()
	if len(rules) == 0 {
		t.Error("expected rule to be generated")
	}
}

func TestEngramLearner_GetRules(t *testing.T) {
	el := NewEngramLearner()

	// Add a test rule
	el.mu.Lock()
	el.rules = append(el.rules, EngramRule{
		ID:         "test-rule",
		Name:       "Test Rule",
		Pattern:    "test",
		Type:       RuleTypePreserve,
		Severity:   SeverityHigh,
		Confidence: 0.8,
	})
	el.mu.Unlock()

	rules := el.GetRules()
	if len(rules) != 1 {
		t.Errorf("expected 1 rule, got %d", len(rules))
	}
	if rules[0].ID != "test-rule" {
		t.Error("expected rule ID to match")
	}
}

func TestEngramLearner_GetRulesForContent(t *testing.T) {
	el := NewEngramLearner()

	// Add a test rule
	el.mu.Lock()
	el.rules = append(el.rules, EngramRule{
		ID:      "stack-trace-rule",
		Name:    "Stack Trace Rule",
		Pattern: "goroutine.*\\[running\\]",
		Type:    RuleTypePreserve,
	})
	el.mu.Unlock()

	// Test matching content
	content := "goroutine 1 [running]:\nmain.main()"
	applicable := el.GetRulesForContent(content)
	if len(applicable) == 0 {
		t.Error("expected matching rules for stack trace content")
	}

	// Test non-matching content
	content = "hello world"
	applicable = el.GetRulesForContent(content)
	if len(applicable) != 0 {
		t.Error("expected no matching rules for plain content")
	}
}

func TestEngramLearner_RecordSuccess(t *testing.T) {
	el := NewEngramLearner()

	// Add a test rule
	el.mu.Lock()
	el.rules = append(el.rules, EngramRule{
		ID:           "test-rule",
		Confidence:   0.7,
		AppliedCount: 0,
		SuccessCount: 0,
	})
	el.mu.Unlock()

	// Record success
	el.RecordSuccess("test-rule")

	// Check that confidence increased
	el.mu.RLock()
	confidence := el.rules[0].Confidence
	el.mu.RUnlock()

	if confidence <= 0.7 {
		t.Error("expected confidence to increase after success")
	}
}

func TestEngramLearner_RecordFailure(t *testing.T) {
	el := NewEngramLearner()

	// Add a test rule
	el.mu.Lock()
	el.rules = append(el.rules, EngramRule{
		ID:           "test-rule",
		Confidence:   0.8,
		AppliedCount: 0,
		SuccessCount: 0,
	})
	el.mu.Unlock()

	// Record failure
	el.RecordFailure("test-rule")

	// Check that confidence decreased
	el.mu.RLock()
	confidence := el.rules[0].Confidence
	el.mu.RUnlock()

	if confidence >= 0.8 {
		t.Error("expected confidence to decrease after failure")
	}
}

func TestEngramLearner_SaveAndLoadRules(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()

	el := NewEngramLearner()
	el.storagePath = filepath.Join(tmpDir, "test_rules.json")

	// Add test rules
	el.rules = append(el.rules, EngramRule{
		ID:         "rule-1",
		Name:       "Test Rule 1",
		Pattern:    "pattern1",
		Type:       RuleTypePreserve,
		Severity:   SeverityHigh,
		Confidence: 0.8,
	})

	// Save rules
	err := el.SaveRules()
	if err != nil {
		t.Fatalf("failed to save rules: %v", err)
	}

	// Create new learner and load
	el2 := NewEngramLearner()
	el2.storagePath = el.storagePath
	err = el2.LoadRules()
	if err != nil {
		t.Fatalf("failed to load rules: %v", err)
	}

	// Verify loaded rules
	if len(el2.rules) != 1 {
		t.Errorf("expected 1 rule loaded, got %d", len(el2.rules))
	}
	if el2.rules[0].ID != "rule-1" {
		t.Error("expected rule ID to match after load")
	}
}

func TestEngramLearner_GetStats(t *testing.T) {
	el := NewEngramLearner()

	// Add test data
	el.mu.Lock()
	el.rules = append(el.rules, EngramRule{ID: "rule-1"})
	el.rules = append(el.rules, EngramRule{ID: "rule-2"})
	el.patterns["pattern-1"] = &ErrorPattern{}
	el.patterns["pattern-2"] = &ErrorPattern{}
	el.mu.Unlock()

	stats := el.GetStats()

	rulesLearned, ok := stats["rules_learned"].(int)
	if !ok || rulesLearned != 2 {
		t.Errorf("expected 2 rules learned, got %v", stats["rules_learned"])
	}

	patternsTracked, ok := stats["patterns_tracked"].(int)
	if !ok || patternsTracked != 2 {
		t.Errorf("expected 2 patterns tracked, got %v", stats["patterns_tracked"])
	}
}

func TestEngramLearner_DefaultClassifiers(t *testing.T) {
	// Test that default classifiers are defined
	if len(defaultClassifiers) == 0 {
		t.Fatal("expected default classifiers to be defined")
	}

	// Test each classifier has required fields
	for _, c := range defaultClassifiers {
		if c.ID == "" {
			t.Error("expected classifier to have ID")
		}
		if c.Name == "" {
			t.Error("expected classifier to have Name")
		}
		if c.Pattern == nil {
			t.Error("expected classifier to have Pattern")
		}
	}
}

func TestEngramLearner_ClassifierMatching(t *testing.T) {
	tests := []struct {
		name        string
		classifier  string
		input       string
		shouldMatch bool
	}{
		{
			name:        "stack_trace",
			classifier:  "stack_trace_loss",
			input:       "goroutine 1 [running]:\nmain.main()\n\t/home/user/main.go:10",
			shouldMatch: true,
		},
		{
			name:        "error_code",
			classifier:  "error_code_drop",
			input:       "Error: connection refused\nExit code: 1",
			shouldMatch: true,
		},
		{
			name:        "json_schema",
			classifier:  "json_schema_broken",
			input:       `{"key": "value", "number": 123}`,
			shouldMatch: true,
		},
		{
			name:        "no_match",
			classifier:  "stack_trace_loss",
			input:       "hello world",
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Find classifier
			var classifier *ErrorClassifier
			for i := range defaultClassifiers {
				if defaultClassifiers[i].ID == tt.classifier {
					classifier = &defaultClassifiers[i]
					break
				}
			}
			if classifier == nil {
				t.Fatalf("classifier %s not found", tt.classifier)
			}

			matches := classifier.Pattern.MatchString(tt.input)
			if matches != tt.shouldMatch {
				t.Errorf("expected match=%v, got match=%v for input: %s",
					tt.shouldMatch, matches, tt.input)
			}
		})
	}
}


