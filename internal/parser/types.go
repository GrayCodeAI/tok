package parser

import "fmt"

// TestStatus represents the outcome of a test case.
type TestStatus string

const (
	TestPass  TestStatus = "pass"
	TestFail  TestStatus = "fail"
	TestSkip  TestStatus = "skip"
	TestError TestStatus = "error"
	TestPanic TestStatus = "panic"
	TestFlaky TestStatus = "flaky"
)

// TestCase represents a single test case result.
type TestCase struct {
	Name     string     `json:"name"`
	Package  string     `json:"package,omitempty"`
	Status   TestStatus `json:"status"`
	Duration float64    `json:"duration_ms"`
	Message  string     `json:"message,omitempty"`
	File     string     `json:"file,omitempty"`
	Line     int        `json:"line,omitempty"`
}

// TestResult represents the aggregated result of a test run.
type TestResult struct {
	Total     int        `json:"total"`
	Passed    int        `json:"passed"`
	Failed    int        `json:"failed"`
	Skipped   int        `json:"skipped"`
	Errors    int        `json:"errors"`
	Duration  float64    `json:"duration_ms"`
	Cases     []TestCase `json:"cases,omitempty"`
	RawOutput string     `json:"raw_output,omitempty"`
	Tool      string     `json:"tool"`
}

// Summary returns a human-readable summary of test results.
func (r *TestResult) Summary() string {
	status := "PASS"
	if r.Failed > 0 || r.Errors > 0 {
		status = "FAIL"
	}
	return fmt.Sprintf("%s: %d tests, %d passed, %d failed, %d skipped, %d errors (%.0fms)",
		status, r.Total, r.Passed, r.Failed, r.Skipped, r.Errors, r.Duration)
}

// IsSuccess returns true if all tests passed.
func (r *TestResult) IsSuccess() bool {
	return r.Failed == 0 && r.Errors == 0
}

// LintSeverity represents the severity of a lint issue.
type LintSeverity string

const (
	LintError   LintSeverity = "error"
	LintWarning LintSeverity = "warning"
	LintInfo    LintSeverity = "info"
	LintHint    LintSeverity = "hint"
)

// LintIssue represents a single lint finding.
type LintIssue struct {
	File     string       `json:"file"`
	Line     int          `json:"line"`
	Column   int          `json:"column,omitempty"`
	EndLine  int          `json:"end_line,omitempty"`
	EndCol   int          `json:"end_column,omitempty"`
	Severity LintSeverity `json:"severity"`
	Rule     string       `json:"rule"`
	Message  string       `json:"message"`
	Fix      string       `json:"fix,omitempty"`
}

// LintResult represents the aggregated result of a lint run.
type LintResult struct {
	Total     int         `json:"total"`
	Errors    int         `json:"errors"`
	Warnings  int         `json:"warnings"`
	Infos     int         `json:"infos"`
	Fixable   int         `json:"fixable"`
	Duration  float64     `json:"duration_ms"`
	Issues    []LintIssue `json:"issues,omitempty"`
	RawOutput string      `json:"raw_output,omitempty"`
	Tool      string      `json:"tool"`
}

// Summary returns a human-readable summary of lint results.
func (r *LintResult) Summary() string {
	status := "OK"
	if r.Errors > 0 {
		status = "ERRORS"
	} else if r.Warnings > 0 {
		status = "WARNINGS"
	}
	return fmt.Sprintf("%s: %d issues (%d errors, %d warnings, %d info), %d fixable (%.0fms)",
		status, r.Total, r.Errors, r.Warnings, r.Infos, r.Fixable, r.Duration)
}

// IsClean returns true if no issues were found.
func (r *LintResult) IsClean() bool {
	return r.Total == 0
}

// BuildStatus represents the outcome of a build.
type BuildStatus string

const (
	BuildSuccess BuildStatus = "success"
	BuildFailed  BuildStatus = "failed"
	BuildWarning BuildStatus = "warning"
)

// BuildIssue represents a compiler error or warning.
type BuildIssue struct {
	File     string       `json:"file"`
	Line     int          `json:"line"`
	Column   int          `json:"column,omitempty"`
	Severity LintSeverity `json:"severity"`
	Message  string       `json:"message"`
	Code     string       `json:"code,omitempty"`
}

// BuildResult represents the result of a build command.
type BuildResult struct {
	Status    BuildStatus  `json:"status"`
	Errors    int          `json:"errors"`
	Warnings  int          `json:"warnings"`
	Duration  float64      `json:"duration_ms"`
	Issues    []BuildIssue `json:"issues,omitempty"`
	RawOutput string       `json:"raw_output,omitempty"`
	Tool      string       `json:"tool"`
}

// IsSuccess returns true if the build succeeded.
func (r *BuildResult) IsSuccess() bool {
	return r.Status == BuildSuccess
}

// DepInfo represents a single dependency.
type DepInfo struct {
	Name     string `json:"name"`
	Version  string `json:"version"`
	Latest   string `json:"latest,omitempty"`
	Outdated bool   `json:"outdated"`
	Direct   bool   `json:"direct"`
}

// DepResult represents the result of a dependency listing.
type DepResult struct {
	Total      int       `json:"total"`
	Outdated   int       `json:"outdated"`
	Direct     int       `json:"direct"`
	Transitive int       `json:"transitive"`
	Deps       []DepInfo `json:"deps,omitempty"`
	Tool       string    `json:"tool"`
}
