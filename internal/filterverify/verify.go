package filterverify

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type FilterTestCase struct {
	ID         string  `json:"id"`
	FilterName string  `json:"filter_name"`
	Input      string  `json:"input"`
	Expected   string  `json:"expected"`
	MaxTokens  int     `json:"max_tokens"`
	MinSavings float64 `json:"min_savings"`
}

type FilterTestResult struct {
	TestCaseID string        `json:"test_case_id"`
	Passed     bool          `json:"passed"`
	Output     string        `json:"output"`
	Savings    float64       `json:"savings"`
	Tokens     int           `json:"tokens"`
	Duration   time.Duration `json:"duration"`
	Error      string        `json:"error,omitempty"`
}

type FilterVerifier struct {
	db    *sql.DB
	tests map[string]*FilterTestCase
}

func NewFilterVerifier(db *sql.DB) *FilterVerifier {
	return &FilterVerifier{
		db:    db,
		tests: make(map[string]*FilterTestCase),
	}
}

func (v *FilterVerifier) Init() error {
	query := `
	CREATE TABLE IF NOT EXISTS filter_test_cases (
		id TEXT PRIMARY KEY,
		filter_name TEXT NOT NULL,
		input TEXT NOT NULL,
		expected TEXT,
		max_tokens INTEGER,
		min_savings REAL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err := v.db.Exec(query)
	return err
}

func (v *FilterVerifier) AddTest(tc *FilterTestCase) error {
	v.tests[tc.ID] = tc
	data, _ := json.Marshal(tc)
	_, err := v.db.Exec(`
		INSERT OR REPLACE INTO filter_test_cases (id, filter_name, input, expected, max_tokens, min_savings)
		VALUES (?, ?, ?, ?, ?, ?)
	`, tc.ID, tc.FilterName, tc.Input, tc.Expected, tc.MaxTokens, tc.MinSavings)
	if err != nil {
		return err
	}
	_ = data
	return nil
}

func (v *FilterVerifier) RunTests(filterName string, filterFunc func(string) string) []FilterTestResult {
	var results []FilterTestResult
	for _, tc := range v.tests {
		if tc.FilterName != filterName {
			continue
		}

		start := time.Now()
		output := filterFunc(tc.Input)
		duration := time.Since(start)

		tokens := len(output) / 4
		origTokens := len(tc.Input) / 4
		savings := 0.0
		if origTokens > 0 {
			savings = float64(origTokens-tokens) / float64(origTokens) * 100
		}

		passed := true
		var errMsg string

		if tc.MaxTokens > 0 && tokens > tc.MaxTokens {
			passed = false
			errMsg = fmt.Sprintf("output %d tokens exceeds max %d", tokens, tc.MaxTokens)
		}
		if tc.MinSavings > 0 && savings < tc.MinSavings {
			passed = false
			errMsg = fmt.Sprintf("savings %.1f%% below minimum %.1f%%", savings, tc.MinSavings)
		}
		if tc.Expected != "" && !strings.Contains(output, tc.Expected) {
			passed = false
			errMsg = "expected output not found"
		}

		results = append(results, FilterTestResult{
			TestCaseID: tc.ID,
			Passed:     passed,
			Output:     output,
			Savings:    savings,
			Tokens:     tokens,
			Duration:   duration,
			Error:      errMsg,
		})
	}
	return results
}

func (v *FilterVerifier) RunAllTests(filterFuncs map[string]func(string) string) map[string][]FilterTestResult {
	results := make(map[string][]FilterTestResult)
	for name, fn := range filterFuncs {
		results[name] = v.RunTests(name, fn)
	}
	return results
}

func (v *FilterVerifier) GetTests(filterName string) []*FilterTestCase {
	var result []*FilterTestCase
	for _, tc := range v.tests {
		if filterName == "" || tc.FilterName == filterName {
			result = append(result, tc)
		}
	}
	return result
}
