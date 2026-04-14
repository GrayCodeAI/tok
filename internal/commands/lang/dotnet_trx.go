package lang

import (
	"encoding/xml"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tokman/internal/commands/registry"
)

var dotnetTrxCmd = &cobra.Command{
	Use:   "trx [file.trx]",
	Short: "Parse and display TRX test results compactly",
	Long: `Parse Visual Studio Test Results (TRX) files and display a compact summary.

Shows test summary, failed tests with details, and timing information.
Useful for CI/CD pipelines and local test runs.

Examples:
  tokman dotnet trx TestResults.trx
  tokman dotnet trx --file TestResults.trx --format json
  cat TestResults.trx | tokman dotnet trx`,
	RunE: runDotnetTrx,
}

var (
	trxFile   string
	trxFormat string
)

func init() {
	registry.Add(func() { registry.Register(dotnetTrxCmd) })
	dotnetTrxCmd.Flags().StringVarP(&trxFile, "file", "f", "", "TRX file path")
	dotnetTrxCmd.Flags().StringVarP(&trxFormat, "format", "o", "text", "Output format: text, json, csv")
}

// TRX XML structures
type TestRun struct {
	XMLName   xml.Name  `xml:"TestRun"`
	Id        string    `xml:"id,attr"`
	Name      string    `xml:"name,attr"`
	RunUser   string    `xml:"runUser,attr"`
	Times     Times     `xml:"Times"`
	Results   Results   `xml:"Results"`
	TestDefs  TestDefs  `xml:"TestDefinitions"`
	ResultSum ResultSum `xml:"ResultSummary"`
}

type Times struct {
	Creation string `xml:"creation,attr"`
	Queuing  string `xml:"queuing,attr"`
	Start    string `xml:"start,attr"`
	Finish   string `xml:"finish,attr"`
}

type Results struct {
	UnitTestResults []UnitTestResult `xml:"UnitTestResult"`
}

type UnitTestResult struct {
	TestId       string  `xml:"testId,attr"`
	TestName     string  `xml:"testName,attr"`
	ComputerName string  `xml:"computerName,attr"`
	Duration     string  `xml:"duration,attr"`
	StartTime    string  `xml:"startTime,attr"`
	EndTime      string  `xml:"endTime,attr"`
	Outcome      string  `xml:"outcome,attr"`
	Output       *Output `xml:"Output"`
}

type Output struct {
	StdOut    string     `xml:"StdOut"`
	StdErr    string     `xml:"StdErr"`
	ErrorInfo *ErrorInfo `xml:"ErrorInfo"`
}

type ErrorInfo struct {
	Message    string `xml:"Message"`
	StackTrace string `xml:"StackTrace"`
}

type TestDefs struct {
	UnitTests []UnitTest `xml:"UnitTest"`
}

type UnitTest struct {
	Id         string     `xml:"id,attr"`
	Name       string     `xml:"name,attr"`
	Storage    string     `xml:"storage,attr"`
	Execution  Execution  `xml:"Execution"`
	TestMethod TestMethod `xml:"TestMethod"`
}

type Execution struct {
	Id string `xml:"id,attr"`
}

type TestMethod struct {
	ClassName string `xml:"className,attr"`
	Name      string `xml:"name,attr"`
}

type ResultSum struct {
	Outcome  string    `xml:"outcome,attr"`
	Counters Counters  `xml:"Counters"`
	RunInfo  []RunInfo `xml:"RunInfos>RunInfo"`
}

type Counters struct {
	Total               int `xml:"total,attr"`
	Executed            int `xml:"executed,attr"`
	Passed              int `xml:"passed,attr"`
	Failed              int `xml:"failed,attr"`
	Error               int `xml:"error,attr"`
	Timeout             int `xml:"timeout,attr"`
	Aborted             int `xml:"aborted,attr"`
	Inconclusive        int `xml:"inconclusive,attr"`
	PassedButRunAborted int `xml:"passedButRunAborted,attr"`
	NotRunnable         int `xml:"notRunnable,attr"`
	NotExecuted         int `xml:"notExecuted,attr"`
	Disconnected        int `xml:"disconnected,attr"`
	Warning             int `xml:"warning,attr"`
	Completed           int `xml:"completed,attr"`
	InProgress          int `xml:"inProgress,attr"`
	Pending             int `xml:"pending,attr"`
}

type RunInfo struct {
	ComputerName string `xml:"computerName,attr"`
	Outcome      string `xml:"outcome,attr"`
	Timestamp    string `xml:"timestamp,attr"`
	Text         string `xml:"Text"`
}

func runDotnetTrx(cmd *cobra.Command, args []string) error {
	var data []byte
	var err error

	// Read from file or stdin
	if trxFile != "" {
		data, err = os.ReadFile(trxFile)
		if err != nil {
			return fmt.Errorf("failed to read TRX file: %w", err)
		}
	} else if len(args) > 0 {
		data, err = os.ReadFile(args[0])
		if err != nil {
			return fmt.Errorf("failed to read TRX file: %w", err)
		}
	} else {
		// Read from stdin
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			data, err = os.ReadFile("/dev/stdin")
			if err != nil {
				return fmt.Errorf("failed to read from stdin: %w", err)
			}
		} else {
			return fmt.Errorf("no TRX file specified (use --file or provide as argument)")
		}
	}

	// Parse TRX XML
	var testRun TestRun
	if err := xml.Unmarshal(data, &testRun); err != nil {
		return fmt.Errorf("failed to parse TRX XML: %w", err)
	}

	// Format output
	switch trxFormat {
	case "json":
		return outputTrxJSON(&testRun)
	case "csv":
		return outputTrxCSV(&testRun)
	default:
		return outputTrxText(&testRun)
	}
}

func outputTrxText(testRun *TestRun) error {
	counters := testRun.ResultSum.Counters

	// Summary header
	fmt.Println("╔════════════════════════════════════════╗")
	fmt.Println("║        TRX Test Results Summary        ║")
	fmt.Println("╚════════════════════════════════════════╝")
	fmt.Println()

	// Test counts
	fmt.Printf("📊 Total Tests: %d\n", counters.Total)
	fmt.Printf("   ✅ Passed:   %d\n", counters.Passed)
	fmt.Printf("   ❌ Failed:   %d\n", counters.Failed)
	fmt.Printf("   ⚠️  Skipped:  %d\n", counters.NotExecuted)
	fmt.Printf("   ⏱️  Duration: %s\n", formatTrxDuration(testRun.Times.Start, testRun.Times.Finish))
	fmt.Println()

	// Failed tests
	if counters.Failed > 0 {
		fmt.Println("═══════════════════════════════════════════")
		fmt.Println("FAILED TESTS:")
		fmt.Println("═══════════════════════════════════════════")

		failCount := 0
		for _, result := range testRun.Results.UnitTestResults {
			if result.Outcome == "Failed" {
				failCount++
				fmt.Printf("\n%d. ❌ %s\n", failCount, result.TestName)
				fmt.Printf("   Duration: %s\n", result.Duration)

				if result.Output != nil && result.Output.ErrorInfo != nil {
					msg := strings.TrimSpace(result.Output.ErrorInfo.Message)
					if msg != "" {
						// Truncate long messages
						lines := strings.Split(msg, "\n")
						if len(lines) > 3 {
							msg = strings.Join(lines[:3], "\n") + "\n   ..."
						}
						fmt.Printf("   Error: %s\n", msg)
					}

					stack := strings.TrimSpace(result.Output.ErrorInfo.StackTrace)
					if stack != "" {
						// Show first line of stack trace
						lines := strings.Split(stack, "\n")
						if len(lines) > 0 {
							fmt.Printf("   Stack: %s\n", strings.TrimSpace(lines[0]))
						}
					}
				}
			}
		}
		fmt.Println()
	}

	// Run info / warnings
	if len(testRun.ResultSum.RunInfo) > 0 {
		fmt.Println("═══════════════════════════════════════════")
		fmt.Println("RUN INFO:")
		fmt.Println("═══════════════════════════════════════════")
		for _, info := range testRun.ResultSum.RunInfo {
			fmt.Printf("  [%s] %s\n", info.Outcome, info.Text)
		}
		fmt.Println()
	}

	// Final outcome
	outcome := testRun.ResultSum.Outcome
	if outcome == "Completed" && counters.Failed == 0 {
		fmt.Println("✅ All tests passed!")
	} else if counters.Failed > 0 {
		fmt.Printf("❌ %d test(s) failed\n", counters.Failed)
	} else {
		fmt.Printf("⚠️  Outcome: %s\n", outcome)
	}

	return nil
}

func outputTrxJSON(testRun *TestRun) error {
	counters := testRun.ResultSum.Counters

	type jsonResult struct {
		TestName string `json:"testName"`
		Outcome  string `json:"outcome"`
		Duration string `json:"duration"`
		Error    string `json:"error,omitempty"`
	}

	var results []jsonResult
	for _, r := range testRun.Results.UnitTestResults {
		jr := jsonResult{
			TestName: r.TestName,
			Outcome:  r.Outcome,
			Duration: r.Duration,
		}
		if r.Output != nil && r.Output.ErrorInfo != nil {
			jr.Error = r.Output.ErrorInfo.Message
		}
		results = append(results, jr)
	}

	type jsonOutput struct {
		Total    int          `json:"total"`
		Passed   int          `json:"passed"`
		Failed   int          `json:"failed"`
		Skipped  int          `json:"skipped"`
		Duration string       `json:"duration"`
		Outcome  string       `json:"outcome"`
		Results  []jsonResult `json:"results"`
	}

	output := jsonOutput{
		Total:    counters.Total,
		Passed:   counters.Passed,
		Failed:   counters.Failed,
		Skipped:  counters.NotExecuted,
		Duration: formatTrxDuration(testRun.Times.Start, testRun.Times.Finish),
		Outcome:  testRun.ResultSum.Outcome,
		Results:  results,
	}

	// Use simple JSON encoding
	jsonBytes, err := xml.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(jsonBytes))
	return nil
}

func outputTrxCSV(testRun *TestRun) error {
	fmt.Println("TestName,Outcome,Duration,Error")
	for _, r := range testRun.Results.UnitTestResults {
		errorMsg := ""
		if r.Output != nil && r.Output.ErrorInfo != nil {
			errorMsg = strings.ReplaceAll(r.Output.ErrorInfo.Message, "\n", " ")
			errorMsg = strings.ReplaceAll(errorMsg, "\"", "\"\"")
		}
		fmt.Printf("\"%s\",%s,%s,\"%s\"\n", r.TestName, r.Outcome, r.Duration, errorMsg)
	}
	return nil
}

func formatTrxDuration(start, finish string) string {
	if start == "" || finish == "" {
		return "unknown"
	}

	// Parse timestamps
	startTime, err1 := time.Parse(time.RFC3339, start)
	finishTime, err2 := time.Parse(time.RFC3339, finish)

	if err1 != nil || err2 != nil {
		// Try alternative format
		startTime, err1 = time.Parse("2006-01-02T15:04:05.9999999-07:00", start)
		finishTime, err2 = time.Parse("2006-01-02T15:04:05.9999999-07:00", finish)
		if err1 != nil || err2 != nil {
			return "unknown"
		}
	}

	duration := finishTime.Sub(startTime)
	if duration < time.Second {
		return fmt.Sprintf("%dms", duration.Milliseconds())
	} else if duration < time.Minute {
		return fmt.Sprintf("%.1fs", duration.Seconds())
	} else {
		return fmt.Sprintf("%.1fm", duration.Minutes())
	}
}
