// Package tokftest provides declarative filter testing based on tokf's framework.
// Test Format Specification (TOML):
//
//	name = "Test name"
//	description = "What this test checks"
//	filter = "filter_name"  # or filter_file = "path.toml"
//
//	[[input]]
//	name = "case1"
//	source = "input text"
//
//	[[input.file]]
//	path = "test.txt"
//	content = "file content"
//
//	[[expect]]
//	name = "case1"
//	output = "expected output"
//	contains = ["must", "contain"]
//	excludes = ["must", "not have"]
//	tokens.lt = 100  # less than 100 tokens
//	saved.pct = 50   # 50% savings
//
//	[[expect.match]]
//	pattern = "regex"
//	count = 2  # exactly 2 matches
package tokftest

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/GrayCodeAI/tokman/internal/filter"
)

// TestSpec represents a declarative filter test specification.
type TestSpec struct {
	Name        string                 `toml:"name"`
	Description string                 `toml:"description"`
	Filter      string                 `toml:"filter,omitempty"`
	FilterFile  string                 `toml:"filter_file,omitempty"`
	Mode        string                 `toml:"mode,omitempty"`
	Config      map[string]interface{} `toml:"config,omitempty"`
	Inputs      []InputCase            `toml:"input"`
	Expects     []ExpectCase           `toml:"expect"`
	Fixtures    []Fixture              `toml:"fixture,omitempty"`
	Skip        bool                   `toml:"skip,omitempty"`
	Only        bool                   `toml:"only,omitempty"`
	Tags        []string               `toml:"tags,omitempty"`
}

// InputCase represents a test input case.
type InputCase struct {
	Name    string     `toml:"name"`
	Source  string     `toml:"source,omitempty"`
	Files   []TestFile `toml:"file,omitempty"`
	Command string     `toml:"command,omitempty"`
}

// TestFile represents a file fixture.
type TestFile struct {
	Path    string `toml:"path"`
	Content string `toml:"content"`
}

// ExpectCase represents expected output for a test case.
type ExpectCase struct {
	Name     string         `toml:"name"`
	Output   string         `toml:"output,omitempty"`
	Contains []string       `toml:"contains,omitempty"`
	Excludes []string       `toml:"excludes,omitempty"`
	Tokens   *TokenExpect   `toml:"tokens,omitempty"`
	Saved    *SavedExpect   `toml:"saved,omitempty"`
	Matches  []MatchExpect  `toml:"match,omitempty"`
	Error    *ErrorExpect   `toml:"error,omitempty"`
	Quality  *QualityExpect `toml:"quality,omitempty"`
}

// TokenExpect represents token count expectations.
type TokenExpect struct {
	LT  *int `toml:"lt,omitempty"`  // less than
	LTE *int `toml:"lte,omitempty"` // less than or equal
	GT  *int `toml:"gt,omitempty"`  // greater than
	GTE *int `toml:"gte,omitempty"` // greater than or equal
	EQ  *int `toml:"eq,omitempty"`  // equal
}

// SavedExpect represents savings expectations.
type SavedExpect struct {
	Count *int     `toml:"count,omitempty"`
	Pct   *float64 `toml:"pct,omitempty"`   // percentage
	Ratio *float64 `toml:"ratio,omitempty"` // compression ratio
}

// MatchExpect represents regex match expectations.
type MatchExpect struct {
	Pattern string `toml:"pattern"`
	Count   *int   `toml:"count,omitempty"`
	Min     *int   `toml:"min,omitempty"`
	Max     *int   `toml:"max,omitempty"`
}

// ErrorExpect represents error expectations.
type ErrorExpect struct {
	ShouldError bool   `toml:"should_error,omitempty"`
	Contains    string `toml:"contains,omitempty"`
}

// QualityExpect represents quality expectations.
type QualityExpect struct {
	MinScore       *float64 `toml:"min_score,omitempty"`
	PreserveErrors bool     `toml:"preserve_errors,omitempty"`
	PreserveURLs   bool     `toml:"preserve_urls,omitempty"`
}

// Fixture represents a reusable test fixture.
type Fixture struct {
	Name    string `toml:"name"`
	Content string `toml:"content"`
	File    string `toml:"file,omitempty"`
}

// TestResult represents the result of running a test.
type TestResult struct {
	Spec     *TestSpec
	Passed   bool
	Duration int64 // milliseconds
	Cases    []CaseResult
	Errors   []string
	Skipped  bool
}

// CaseResult represents the result of a single test case.
type CaseResult struct {
	Name     string
	Passed   bool
	Input    string
	Output   string
	Tokens   int
	Saved    int
	Errors   []string
	Duration int64 // milliseconds
}

// Parser parses test specifications.
type Parser struct {
	fixtureDir string
}

// NewParser creates a new test parser.
func NewParser(fixtureDir string) *Parser {
	return &Parser{fixtureDir: fixtureDir}
}

// ParseFile parses a test specification from a TOML file.
func (p *Parser) ParseFile(path string) (*TestSpec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read test file: %w", err)
	}

	var spec TestSpec
	if err := toml.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("failed to parse TOML: %w", err)
	}

	// Set default name from filename if not specified
	if spec.Name == "" {
		spec.Name = filepath.Base(path)
	}

	// Load fixtures if referenced
	for i, fixture := range spec.Fixtures {
		if fixture.File != "" {
			content, err := p.loadFixture(fixture.File)
			if err != nil {
				return nil, fmt.Errorf("failed to load fixture %s: %w", fixture.File, err)
			}
			spec.Fixtures[i].Content = content
		}
	}

	// Validate
	if err := p.validate(&spec); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return &spec, nil
}

// ParseDir parses all test files in a directory.
func (p *Parser) ParseDir(dir string) ([]*TestSpec, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var specs []*TestSpec
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.HasSuffix(entry.Name(), ".toml") {
			continue
		}
		if strings.HasSuffix(entry.Name(), ".fixture.toml") {
			continue
		}

		spec, err := p.ParseFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("failed to parse %s: %w", entry.Name(), err)
		}
		specs = append(specs, spec)
	}

	return specs, nil
}

func (p *Parser) loadFixture(name string) (string, error) {
	// Try fixture directory first
	if p.fixtureDir != "" {
		path := filepath.Join(p.fixtureDir, name)
		if data, err := os.ReadFile(path); err == nil {
			return string(data), nil
		}
	}

	// Try current directory
	if data, err := os.ReadFile(name); err == nil {
		return string(data), nil
	}

	return "", fmt.Errorf("fixture not found: %s", name)
}

func (p *Parser) validate(spec *TestSpec) error {
	if spec.Filter == "" && spec.FilterFile == "" {
		return fmt.Errorf("either 'filter' or 'filter_file' must be specified")
	}

	if len(spec.Inputs) == 0 {
		return fmt.Errorf("at least one input case is required")
	}

	if len(spec.Expects) == 0 {
		return fmt.Errorf("at least one expect case is required")
	}

	// Validate input/expect name matching
	inputNames := make(map[string]bool)
	for _, in := range spec.Inputs {
		if in.Name == "" {
			return fmt.Errorf("input case must have a name")
		}
		inputNames[in.Name] = true
	}

	for _, exp := range spec.Expects {
		if exp.Name == "" {
			return fmt.Errorf("expect case must have a name")
		}
		if !inputNames[exp.Name] {
			return fmt.Errorf("expect case '%s' has no matching input", exp.Name)
		}
	}

	return nil
}

// Runner executes test specifications.
type Runner struct {
	filterLoader FilterLoader
}

// FilterLoader loads filter configurations.
type FilterLoader interface {
	Load(name string) (*filter.Engine, error)
	LoadFromFile(path string) (*filter.Engine, error)
}

// NewRunner creates a new test runner.
func NewRunner(loader FilterLoader) *Runner {
	return &Runner{filterLoader: loader}
}

// Run executes a test specification.
func (r *Runner) Run(spec *TestSpec) *TestResult {
	result := &TestResult{
		Spec:   spec,
		Passed: true,
		Cases:  make([]CaseResult, 0, len(spec.Inputs)),
	}

	if spec.Skip {
		result.Skipped = true
		return result
	}

	// Load filter
	engine, err := r.loadFilter(spec)
	if err != nil {
		result.Passed = false
		result.Errors = append(result.Errors, fmt.Sprintf("failed to load filter: %v", err))
		return result
	}

	// Build input/expect map
	expectMap := make(map[string]ExpectCase)
	for _, exp := range spec.Expects {
		expectMap[exp.Name] = exp
	}

	// Run each input case
	for _, input := range spec.Inputs {
		expect, ok := expectMap[input.Name]
		if !ok {
			result.Errors = append(result.Errors, fmt.Sprintf("no expect for input '%s'", input.Name))
			result.Passed = false
			continue
		}

		caseResult := r.runCase(input, expect, engine)
		result.Cases = append(result.Cases, caseResult)
		if !caseResult.Passed {
			result.Passed = false
		}
	}

	return result
}

func (r *Runner) loadFilter(spec *TestSpec) (*filter.Engine, error) {
	if spec.FilterFile != "" {
		return r.filterLoader.LoadFromFile(spec.FilterFile)
	}

	mode := filter.ModeMinimal
	if spec.Mode == "aggressive" {
		mode = filter.ModeAggressive
	}

	return filter.NewEngine(mode), nil
}

func (r *Runner) runCase(input InputCase, expect ExpectCase, engine *filter.Engine) CaseResult {
	result := CaseResult{
		Name:   input.Name,
		Input:  input.Source,
		Passed: true,
	}

	// Process input
	output, saved := engine.Process(input.Source)
	result.Output = output
	result.Saved = saved
	result.Tokens = filter.EstimateTokens(output)

	// Check exact output match
	if expect.Output != "" && output != expect.Output {
		result.Passed = false
		result.Errors = append(result.Errors, fmt.Sprintf("output mismatch:\ngot:\n%s\nexpected:\n%s", output, expect.Output))
	}

	// Check contains
	for _, s := range expect.Contains {
		if !strings.Contains(output, s) {
			result.Passed = false
			result.Errors = append(result.Errors, fmt.Sprintf("output missing expected: %q", s))
		}
	}

	// Check excludes
	for _, s := range expect.Excludes {
		if strings.Contains(output, s) {
			result.Passed = false
			result.Errors = append(result.Errors, fmt.Sprintf("output contains unexpected: %q", s))
		}
	}

	// Check token expectations
	if expect.Tokens != nil {
		if err := r.checkTokens(result.Tokens, expect.Tokens); err != nil {
			result.Passed = false
			result.Errors = append(result.Errors, err.Error())
		}
	}

	// Check savings expectations
	if expect.Saved != nil {
		if err := r.checkSaved(result.Saved, len(input.Source), expect.Saved); err != nil {
			result.Passed = false
			result.Errors = append(result.Errors, err.Error())
		}
	}

	// Check regex matches
	for _, match := range expect.Matches {
		if err := r.checkMatch(output, match); err != nil {
			result.Passed = false
			result.Errors = append(result.Errors, err.Error())
		}
	}

	// Check quality
	if expect.Quality != nil {
		if err := r.checkQuality(input.Source, output, expect.Quality); err != nil {
			result.Passed = false
			result.Errors = append(result.Errors, err.Error())
		}
	}

	return result
}

func (r *Runner) checkTokens(actual int, expect *TokenExpect) error {
	if expect.LT != nil && actual >= *expect.LT {
		return fmt.Errorf("expected tokens < %d, got %d", *expect.LT, actual)
	}
	if expect.LTE != nil && actual > *expect.LTE {
		return fmt.Errorf("expected tokens <= %d, got %d", *expect.LTE, actual)
	}
	if expect.GT != nil && actual <= *expect.GT {
		return fmt.Errorf("expected tokens > %d, got %d", *expect.GT, actual)
	}
	if expect.GTE != nil && actual < *expect.GTE {
		return fmt.Errorf("expected tokens >= %d, got %d", *expect.GTE, actual)
	}
	if expect.EQ != nil && actual != *expect.EQ {
		return fmt.Errorf("expected tokens = %d, got %d", *expect.EQ, actual)
	}
	return nil
}

func (r *Runner) checkSaved(saved, originalLen int, expect *SavedExpect) error {
	if expect.Count != nil && saved != *expect.Count {
		return fmt.Errorf("expected saved = %d, got %d", *expect.Count, saved)
	}
	if expect.Pct != nil {
		pct := float64(saved) / float64(originalLen) * 100
		if pct < *expect.Pct-0.1 || pct > *expect.Pct+0.1 {
			return fmt.Errorf("expected saved pct = %.1f%%, got %.1f%%", *expect.Pct, pct)
		}
	}
	if expect.Ratio != nil {
		ratio := float64(saved) / float64(originalLen)
		if ratio < *expect.Ratio-0.01 || ratio > *expect.Ratio+0.01 {
			return fmt.Errorf("expected ratio = %.2f, got %.2f", *expect.Ratio, ratio)
		}
	}
	return nil
}

func (r *Runner) checkMatch(output string, expect MatchExpect) error {
	re, err := regexp.Compile(expect.Pattern)
	if err != nil {
		return fmt.Errorf("invalid regex pattern: %w", err)
	}

	matches := re.FindAllString(output, -1)
	count := len(matches)

	if expect.Count != nil && count != *expect.Count {
		return fmt.Errorf("expected %d matches for pattern %q, got %d", *expect.Count, expect.Pattern, count)
	}
	if expect.Min != nil && count < *expect.Min {
		return fmt.Errorf("expected at least %d matches for pattern %q, got %d", *expect.Min, expect.Pattern, count)
	}
	if expect.Max != nil && count > *expect.Max {
		return fmt.Errorf("expected at most %d matches for pattern %q, got %d", *expect.Max, expect.Pattern, count)
	}
	return nil
}

func (r *Runner) checkQuality(original, output string, expect *QualityExpect) error {
	equiv := filter.NewSemanticEquivalence()
	report := equiv.Check(original, output)

	if expect.MinScore != nil && report.Score < *expect.MinScore {
		return fmt.Errorf("quality score %.2f below minimum %.2f", report.Score, *expect.MinScore)
	}
	return nil
}
