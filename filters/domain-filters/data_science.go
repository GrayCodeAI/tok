package domain_filters

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/GrayCodeAI/tokman/internal/filter"
)

// DataScienceFilter removes boilerplate from ML/data science logs
type DataScienceFilter struct {
	mode filter.Mode
}

// Name returns the filter name.
func (dsf *DataScienceFilter) Name() string {
	return "data_science"
}

// Apply compresses data science output
func (dsf *DataScienceFilter) Apply(input string, mode filter.Mode) (string, int) {
	original := len(input)
	dsf.mode = mode

	output := input

	// Remove verbose training logs
	output = removeTrainingVerbosity(output)

	// Remove duplicate model outputs
	output = deduplicateModelOutputs(output)

	// Compress tensor/array dumps
	output = compressTensorDumps(output)

	// Remove progress bars
	output = removeProgressBars(output)

	// Remove matplotlib figure output
	output = removeMatplotlibOutput(output)

	// Compress large dataframe displays
	output = compressDataFrames(output)

	// Remove sklearn warnings (keep critical only)
	output = filterScikitWarnings(output)

	// Compress numpy array displays
	output = compressNumpyArrays(output)

	saved := original - len(output)
	if saved < 0 {
		saved = 0
	}

	return output, saved
}

func removeTrainingVerbosity(text string) string {
	// Remove per-batch loss output
	batchRegex := regexp.MustCompile(`Epoch \d+/\d+.*?loss: [\d.]+.*?\n`)
	text = batchRegex.ReplaceAllString(text, "")

	// Remove learning rate schedulings
	lrRegex := regexp.MustCompile(`Setting learning rate to \d+\.\d+\.\n`)
	text = lrRegex.ReplaceAllString(text, "")

	return text
}

func deduplicateModelOutputs(text string) string {
	// If same output appears 3+ times, keep only first and mark
	lines := strings.Split(text, "\n")
	seen := make(map[string]int)
	var result []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			result = append(result, line)
			continue
		}

		if seen[trimmed] == 0 {
			result = append(result, line)
		} else if seen[trimmed] == 1 {
			result = append(result, "... (repeated "+fmt.Sprintf("%d", seen[trimmed])+" times) ...")
		}
		seen[trimmed]++
	}

	return strings.Join(result, "\n")
}

func compressTensorDumps(text string) string {
	// Replace full tensor dumps with shape info
	tensorRegex := regexp.MustCompile(`tensor\(\[\[[\d\s.,\-e\+\[\]]+\]\]\)`)
	matches := tensorRegex.FindAllString(text, -1)

	for _, match := range matches {
		// Extract shape
		lines := strings.Count(match, "[")
		cols := 0
		if strings.Contains(match, ",") {
			// Count cols from first row
			firstRowEnd := strings.Index(match, "]")
			if firstRowEnd > 0 {
				firstRow := match[1:firstRowEnd]
				cols = strings.Count(firstRow, ",") + 1
			}
		}
		replacement := "[tensor shape: " + string(rune(lines)) + "x" + string(rune(cols)) + "]"
		text = strings.Replace(text, match, replacement, 1)
	}

	return text
}

func removeProgressBars(text string) string {
	// Remove tqdm/progress bar output
	progressRegex := regexp.MustCompile(`\d+%\|[▏▎▍▌▋▊▉█]+\|\s+\d+/\d+.*?\r`)
	text = progressRegex.ReplaceAllString(text, "")

	return text
}

func removeMatplotlibOutput(text string) string {
	// Remove matplotlib figure data
	figRegex := regexp.MustCompile(`<Figure[\s\S]*?>\n`)
	text = figRegex.ReplaceAllString(text, "[Figure output]\n")

	return text
}

func compressDataFrames(text string) string {
	// Compress pandas DataFrame displays
	dfRegex := regexp.MustCompile(`(?s)\s+\d+\s+[\d\.]+\s+[\d\.]+.*?(?:rows|rows omitted)`)
	matches := dfRegex.FindAllString(text, -1)

	for _, match := range matches {
		rowCount := strings.Count(match, "\n")
		replacement := "[DataFrame: " + string(rune(rowCount)) + " rows]"
		text = strings.Replace(text, match, replacement, 1)
	}

	return text
}

func filterScikitWarnings(text string) string {
	// Keep only critical sklearn warnings
	lines := strings.Split(text, "\n")
	var result []string

	for _, line := range lines {
		// Keep convergence warnings, drop deprecations unless critical
		if strings.Contains(line, "ConvergenceWarning") ||
			strings.Contains(line, "did not converge") ||
			strings.Contains(line, "nan") {
			result = append(result, line)
		} else if !strings.Contains(line, "DeprecationWarning") &&
			!strings.Contains(line, "FutureWarning") {
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
}

func compressNumpyArrays(text string) string {
	// Replace numpy array dumps with summaries
	arrayRegex := regexp.MustCompile(`array\(\[[\d\s.,\-\[\]e\+]+\]\)`)
	matches := arrayRegex.FindAllString(text, -1)

	for _, match := range matches {
		dims := strings.Count(match, "[") - 1
		replacement := "[ndarray " + string(rune(dims)) + "D]"
		text = strings.Replace(text, match, replacement, 1)
	}

	return text
}
