package system

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tokman/internal/commands/registry"
	"github.com/GrayCodeAI/tokman/internal/commands/shared"
	"github.com/GrayCodeAI/tokman/internal/filter"
	"github.com/GrayCodeAI/tokman/internal/tracking"
)

var (
	wgetOutputFile string
)

// wgetCmd represents the wget command
var wgetCmd = &cobra.Command{
	Use:   "wget [URL]",
	Short: "Download with compact output (strips progress bars)",
	Long: `Download files using wget with token-optimized output.

Strips progress bars, download statistics, and other verbose output
while preserving the essential download result information.

Examples:
  tokman wget https://example.com/file.txt
  tokman wget -O output.txt https://example.com/file.txt`,
	Args: cobra.ExactArgs(1),
	RunE: runWget,
}

func init() {
	registry.Add(func() { registry.Register(wgetCmd) })
	wgetCmd.Flags().StringVarP(&wgetOutputFile, "output-document", "O", "", "Output file (-O - for stdout)")
}

func runWget(cmd *cobra.Command, args []string) error {
	url := args[0]
	timer := tracking.Start()

	if shared.Verbose > 0 {
		fmt.Fprintf(os.Stderr, "Downloading: %s\n", url)
	}

	// Build wget arguments
	wgetArgs := []string{url}
	if wgetOutputFile != "" {
		wgetArgs = append([]string{"-O", wgetOutputFile}, url)
	}

	// Run wget
	execCmd := exec.Command("wget", wgetArgs...)
	output, err := execCmd.CombinedOutput()
	raw := string(output)

	// Filter output
	filtered := filterWgetOutput(raw)
	fmt.Println(filtered)

	// Track metrics
	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track(fmt.Sprintf("wget %s", url), "tokman wget", originalTokens, filteredTokens)

	return err
}

func filterWgetOutput(raw string) string {
	lines := strings.Split(raw, "\n")
	var result []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Skip progress bar lines
		if strings.Contains(line, "%") && strings.Contains(line, "[") {
			continue
		}
		if strings.Contains(line, "..........") {
			continue
		}
		if strings.Contains(line, " saved") && strings.Contains(line, "bytes") {
			// Keep the saved summary line
			result = append(result, line)
			continue
		}

		// Skip download progress lines
		if strings.HasPrefix(line, "HTTP") {
			continue
		}
		if strings.Contains(line, "Connecting to") || strings.Contains(line, "Resolving") {
			continue
		}

		result = append(result, line)
	}

	if len(result) == 0 {
		return "Download completed"
	}

	return strings.Join(result, "\n")
}
