package lang

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// TRXFile represents a Visual Studio Test Results XML file.
type TRXFile struct {
	XMLName   xml.Name `xml:"TestRun"`
	ID        string   `xml:"id,attr"`
	Name      string   `xml:"name,attr"`
	RunConfig struct {
		XMLName xml.Name `xml:"RunConfiguration"`
	} `xml:"TestSettings"`
	Times    TRXTimes   `xml:"Times"`
	Results  TRXResults `xml:"Results"`
	TestDefs struct {
		UnitTests []TRXUnitTest `xml:"UnitTest"`
	} `xml:"TestDefinitions"`
}

// TRXTimes holds start and finish timestamps.
type TRXTimes struct {
	XMLName  xml.Name `xml:"Times"`
	Creation string   `xml:"creation,attr"`
	Queuing  string   `xml:"queuing,attr"`
	Start    string   `xml:"start,attr"`
	Finish   string   `xml:"finish,attr"`
}

// TRXResults holds unit test results.
type TRXResults struct {
	XMLName     xml.Name        `xml:"Results"`
	UnitResults []TRXUnitResult `xml:"UnitTestResult"`
}

// TRXUnitResult holds a single test result.
type TRXUnitResult struct {
	XMLName   xml.Name  `xml:"UnitTestResult"`
	TestID    string    `xml:"testId,attr"`
	TestName  string    `xml:"testName,attr"`
	Outcome   string    `xml:"outcome,attr"`
	Duration  string    `xml:"duration,attr"`
	StartTime string    `xml:"startTime,attr"`
	EndTime   string    `xml:"endTime,attr"`
	Output    TRXOutput `xml:"Output"`
}

// TRXOutput holds test output (error messages, stdout).
type TRXOutput struct {
	XMLName   xml.Name     `xml:"Output"`
	ErrorInfo TRXErrorInfo `xml:"ErrorInfo"`
	StdOut    string       `xml:"StdOut"`
}

// TRXErrorInfo holds error message and stack trace.
type TRXErrorInfo struct {
	XMLName    xml.Name `xml:"ErrorInfo"`
	Message    string   `xml:"Message"`
	StackTrace string   `xml:"StackTrace"`
}

// TRXUnitTest holds test definition metadata.
type TRXUnitTest struct {
	XMLName xml.Name `xml:"UnitTest"`
	ID      string   `xml:"id,attr"`
	Name    string   `xml:"name,attr"`
	Storage string   `xml:"storage,attr"`
}

// TestSummary holds aggregated test results.
type TestSummary struct {
	Total        int
	Passed       int
	Failed       int
	Skipped      int
	FailedTests  []FailedTest
	Duration     time.Duration
	DurationText string
}

// FailedTest holds details about a failed test.
type FailedTest struct {
	Name    string
	Message string
	Stack   string
}

// ParseTRXFile parses a .trx test results file and returns a summary.
func ParseTRXFile(path string) (*TestSummary, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read TRX file: %w", err)
	}
	return ParseTRXContent(string(data))
}

// ParseTRXContent parses TRX XML content and returns a summary.
func ParseTRXContent(content string) (*TestSummary, error) {
	var trx TRXFile
	if err := xml.Unmarshal([]byte(content), &trx); err != nil {
		return nil, fmt.Errorf("failed to parse TRX XML: %w", err)
	}

	summary := &TestSummary{}

	// Build test name map
	testNames := make(map[string]string)
	for _, ut := range trx.TestDefs.UnitTests {
		testNames[ut.ID] = ut.Name
	}

	// Process results
	for _, result := range trx.Results.UnitResults {
		testName := result.TestName
		if name, ok := testNames[result.TestID]; ok && name != "" {
			testName = name
		}

		summary.Total++

		switch strings.ToLower(result.Outcome) {
		case "passed":
			summary.Passed++
		case "failed":
			summary.Failed++
			msg := result.Output.ErrorInfo.Message
			stack := result.Output.ErrorInfo.StackTrace
			// Truncate stack to first 5 lines
			stackLines := strings.Split(stack, "\n")
			if len(stackLines) > 5 {
				stack = strings.Join(stackLines[:5], "\n") + "\n..."
			}
			summary.FailedTests = append(summary.FailedTests, FailedTest{
				Name:    testName,
				Message: msg,
				Stack:   stack,
			})
		case "notexecuted", "skipped":
			summary.Skipped++
		}

		// Parse duration
		if result.Duration != "" {
			d, err := parseTRXDuration(result.Duration)
			if err == nil {
				summary.Duration += d
			}
		}
	}

	// Calculate overall duration from Times
	if trx.Times.Start != "" && trx.Times.Finish != "" {
		start, err1 := time.Parse(time.RFC3339, trx.Times.Start)
		finish, err2 := time.Parse(time.RFC3339, trx.Times.Finish)
		if err1 == nil && err2 == nil {
			summary.Duration = finish.Sub(start)
		}
	}

	summary.DurationText = formatDuration(summary.Duration)
	return summary, nil
}

func parseTRXDuration(s string) (time.Duration, error) {
	// TRX duration format: HH:MM:SS.ssssss
	parts := strings.Split(s, ":")
	if len(parts) != 3 {
		return 0, fmt.Errorf("invalid duration format: %s", s)
	}
	hours := 0
	minutes := 0
	seconds := 0.0
	fmt.Sscanf(parts[0], "%d", &hours)
	fmt.Sscanf(parts[1], "%d", &minutes)
	fmt.Sscanf(parts[2], "%f", &seconds)
	totalSeconds := float64(hours)*3600 + float64(minutes)*60 + seconds
	return time.Duration(totalSeconds * float64(time.Second)), nil
}

func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%d ms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1f s", d.Seconds())
	}
	return fmt.Sprintf("%dm %.1fs", int(d.Minutes()), d.Seconds()-float64(int(d.Minutes()))*60)
}

// BinlogIssue represents a build error or warning from MSBuild output.
type BinlogIssue struct {
	Code    string
	File    string
	Line    int
	Column  int
	Kind    string // "error" or "warning"
	Message string
}

// BuildSummary holds aggregated build results.
type BuildSummary struct {
	Succeeded    bool
	ProjectCount int
	Errors       []BinlogIssue
	Warnings     []BinlogIssue
	DurationText string
}

var (
	issueRegex        = regexp.MustCompile(`(?m)^\s*([^\r\n:(]+)\((\d+),(\d+)\):\s*(error|warning)\s*(?:([A-Za-z]+\d+)\s*:\s*)?(.*)$`)
	buildSummaryRegex = regexp.MustCompile(`(?mi)^\s*(\d+)\s+(warning|error)\(s\)`)
	durationRegex     = regexp.MustCompile(`(?m)^\s*Time Elapsed\s+(.+)$`)
	testResultRegex   = regexp.MustCompile(`(?m)(?:Passed!|Failed!)\s*-\s*Failed:\s*(\d+),\s*Passed:\s*(\d+),\s*Skipped:\s*(\d+),\s*Total:\s*(\d+),\s*Duration:\s*([^\r\n-]+)`)
	failedTestRegex   = regexp.MustCompile(`(?m)^\s*Failed\s+([^\r\n\[]+)\s+\[[^\]\r\n]+\]\s*$`)
)

// ParseBuildOutput parses MSBuild/dotnet build output and extracts errors/warnings.
func ParseBuildOutput(output string) *BuildSummary {
	summary := &BuildSummary{Succeeded: true}
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		// Check for issues with file:line:column format
		matches := issueRegex.FindStringSubmatch(line)
		if len(matches) == 7 {
			issue := BinlogIssue{
				File:    strings.TrimSpace(matches[1]),
				Line:    parseInt(matches[2]),
				Column:  parseInt(matches[3]),
				Kind:    matches[4],
				Code:    matches[5],
				Message: strings.TrimSpace(matches[6]),
			}
			// Truncate long file paths
			if len(issue.File) > 80 {
				issue.File = "..." + issue.File[len(issue.File)-77:]
			}
			if issue.Kind == "error" {
				summary.Errors = append(summary.Errors, issue)
				summary.Succeeded = false
			} else {
				summary.Warnings = append(summary.Warnings, issue)
			}
			continue
		}

		// Check for duration
		if m := durationRegex.FindStringSubmatch(line); len(m) > 1 {
			summary.DurationText = strings.TrimSpace(m[1])
		}

		// Count projects (lines ending with .csproj, .fsproj, .vbproj)
		if strings.HasSuffix(strings.TrimSpace(line), ".csproj") ||
			strings.HasSuffix(strings.TrimSpace(line), ".fsproj") ||
			strings.HasSuffix(strings.TrimSpace(line), ".vbproj") {
			summary.ProjectCount++
		}
	}

	return summary
}

// ParseTestOutput parses dotnet test output and extracts test results.
func ParseTestOutput(output string) *TestSummary {
	summary := &TestSummary{}

	// Try NUnit/xUnit style: "Failed! - Failed: 1, Passed: 5, ..."
	if m := testResultRegex.FindStringSubmatch(output); len(m) == 6 {
		summary.Failed = parseInt(m[1])
		summary.Passed = parseInt(m[2])
		summary.Skipped = parseInt(m[3])
		summary.Total = parseInt(m[4])
		summary.DurationText = strings.TrimSpace(m[5])
	}

	// Extract failed test names
	for _, m := range failedTestRegex.FindAllStringSubmatch(output, -1) {
		if len(m) > 1 {
			name := strings.TrimSpace(m[1])
			// Find the error message after the failed test line
			idx := strings.Index(output, m[0])
			if idx >= 0 {
				rest := output[idx+len(m[0]):]
				restLines := strings.SplitN(rest, "\n", 10)
				var msgLines []string
				for _, l := range restLines {
					l = strings.TrimSpace(l)
					if l == "" || strings.HasPrefix(l, "Failed ") || strings.HasPrefix(l, "Passed ") {
						break
					}
					msgLines = append(msgLines, l)
				}
				msg := strings.Join(msgLines, "\n")
				if len(msg) > 200 {
					msg = msg[:200] + "..."
				}
				summary.FailedTests = append(summary.FailedTests, FailedTest{
					Name:    name,
					Message: msg,
				})
			}
		}
	}

	return summary
}

// FormatTestSummary formats a test summary for display.
func FormatTestSummary(s *TestSummary) string {
	status := "PASS"
	if s.Failed > 0 {
		status = "FAIL"
	}
	parts := []string{
		fmt.Sprintf("%s: %d tests", status, s.Total),
		fmt.Sprintf("%d passed", s.Passed),
		fmt.Sprintf("%d failed", s.Failed),
		fmt.Sprintf("%d skipped", s.Skipped),
	}
	if s.DurationText != "" {
		parts = append(parts, s.DurationText)
	}

	result := strings.Join(parts, ", ")

	if len(s.FailedTests) > 0 {
		result += "\n\nFailed tests:"
		for i, ft := range s.FailedTests {
			if i >= 10 {
				result += fmt.Sprintf("\n  ... and %d more", len(s.FailedTests)-10)
				break
			}
			result += fmt.Sprintf("\n  ✗ %s", ft.Name)
			if ft.Message != "" {
				// Show first line of error
				firstLine := strings.Split(ft.Message, "\n")[0]
				if len(firstLine) > 80 {
					firstLine = firstLine[:77] + "..."
				}
				result += fmt.Sprintf("\n    %s", firstLine)
			}
		}
	}

	return result
}

// FormatBuildSummary formats a build summary for display.
func FormatBuildSummary(s *BuildSummary) string {
	status := "SUCCEEDED"
	if !s.Succeeded {
		status = "FAILED"
	}
	parts := []string{
		fmt.Sprintf("Build %s", status),
		fmt.Sprintf("%d errors", len(s.Errors)),
		fmt.Sprintf("%d warnings", len(s.Warnings)),
	}
	if s.ProjectCount > 0 {
		parts = append(parts, fmt.Sprintf("%d projects", s.ProjectCount))
	}
	if s.DurationText != "" {
		parts = append(parts, s.DurationText)
	}

	result := strings.Join(parts, ", ")

	if len(s.Errors) > 0 {
		result += "\n\nErrors:"
		for i, e := range s.Errors {
			if i >= 10 {
				result += fmt.Sprintf("\n  ... and %d more", len(s.Errors)-10)
				break
			}
			loc := e.File
			if e.Line > 0 {
				loc = fmt.Sprintf("%s(%d,%d)", filepath.Base(e.File), e.Line, e.Column)
			}
			code := ""
			if e.Code != "" {
				code = e.Code + ": "
			}
			result += fmt.Sprintf("\n  %s: %s%s", loc, code, e.Message)
		}
	}

	return result
}

func parseInt(s string) int {
	var n int
	fmt.Sscanf(s, "%d", &n)
	return n
}
