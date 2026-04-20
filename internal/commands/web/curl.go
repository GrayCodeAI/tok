package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	out "github.com/lakshmanpatel/tok/internal/output"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/lakshmanpatel/tok/internal/commands/registry"
	"github.com/lakshmanpatel/tok/internal/commands/shared"
	"github.com/lakshmanpatel/tok/internal/filter"
	"github.com/lakshmanpatel/tok/internal/tracking"
)

var curlCmd = &cobra.Command{
	Use:   "curl [flags...] <URL>",
	Short: "curl with auto-JSON detection and formatting",
	Long: `Execute curl commands with intelligent output formatting.

Auto-detects JSON responses and formats them for readability.
Preserves standard curl behavior while adding token-efficient output.

Examples:
  tok curl https://api.example.com/users
  tok curl -H "Authorization: Bearer token" https://api.example.com/data
  tok curl -X POST -d '{"key":"value"}' https://api.example.com/create`,
	DisableFlagParsing: true,
	RunE:               runCurl,
}

func init() {
	registry.Add(func() { registry.Register(curlCmd) })
}

func runCurl(cmd *cobra.Command, args []string) error {
	timer := tracking.Start()

	if len(args) == 0 {
		return fmt.Errorf("curl requires a URL")
	}

	// Build curl command with all provided arguments
	curlArgs := append([]string{"-s", "-w", "\n%{content_type}\n%{http_code}"}, args...)
	execCmd := exec.Command("curl", curlArgs...)

	output, err := execCmd.CombinedOutput()
	raw := string(output)

	// Parse output to separate body from metadata
	body, contentType, statusCode := parseCurlOutput(raw)

	// Format output based on content type
	filtered := formatCurlOutput(body, contentType, statusCode)

	out.Global().Println(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track(fmt.Sprintf("curl %s", strings.Join(args, " ")), "tok curl", originalTokens, filteredTokens)

	shared.PrintTokenSavings(originalTokens, filteredTokens)

	// Return error if curl failed or returned non-2xx status
	if err != nil || (statusCode != "" && !strings.HasPrefix(statusCode, "2")) {
		if statusCode != "" {
			return fmt.Errorf("HTTP %s", statusCode)
		}
		return err
	}

	return nil
}

func parseCurlOutput(output string) (body, contentType, statusCode string) {
	lines := strings.Split(output, "\n")

	if len(lines) >= 2 {
		// Last two lines are content-type and status code
		statusCode = strings.TrimSpace(lines[len(lines)-1])
		contentType = strings.TrimSpace(lines[len(lines)-2])
		body = strings.Join(lines[:len(lines)-2], "\n")
	} else {
		body = output
	}

	return body, contentType, statusCode
}

func formatCurlOutput(body, contentType, statusCode string) string {
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()

	var result strings.Builder

	// Show status code
	if statusCode != "" {
		if strings.HasPrefix(statusCode, "2") {
			result.WriteString(fmt.Sprintf("%s HTTP %s\n", green("OK"), statusCode))
		} else if strings.HasPrefix(statusCode, "4") || strings.HasPrefix(statusCode, "5") {
			result.WriteString(fmt.Sprintf("%s HTTP %s\n", red("FAIL"), statusCode))
		} else {
			result.WriteString(fmt.Sprintf("HTTP %s\n", statusCode))
		}
	}

	// Auto-detect and format JSON
	trimmedBody := strings.TrimSpace(body)
	if isJSON(trimmedBody) || strings.Contains(contentType, "application/json") {
		formatted := formatJSON(trimmedBody)
		if formatted != "" {
			result.WriteString(cyan("JSON Response:\n"))
			result.WriteString(formatted)
			return result.String()
		}
	}

	// Return body as-is for non-JSON
	if body != "" {
		result.WriteString(body)
	}

	return result.String()
}

func isJSON(s string) bool {
	s = strings.TrimSpace(s)
	return (strings.HasPrefix(s, "{") && strings.HasSuffix(s, "}")) ||
		(strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]"))
}

func formatJSON(jsonStr string) string {
	var data interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		// Not valid JSON, return as-is
		return jsonStr
	}

	// Pretty print with limited depth
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false)

	if err := encoder.Encode(data); err != nil {
		return jsonStr
	}

	return buf.String()
}
