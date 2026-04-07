package domain_filters

import (
	"regexp"
	"strings"

	"github.com/GrayCodeAI/tokman/internal/filter"
)

// DevOpsFilter removes boilerplate from DevOps/infrastructure logs
type DevOpsFilter struct {
	mode filter.Mode
}

// Name returns the filter name.
func (dof *DevOpsFilter) Name() string {
	return "devops"
}

// Apply compresses DevOps output
func (dof *DevOpsFilter) Apply(input string, mode filter.Mode) (string, int) {
	original := len(input)
	dof.mode = mode

	output := input

	// Remove verbose Kubernetes logs
	output = compressKubernetesOutput(output)

	// Compress Terraform plans
	output = compressTerraformPlan(output)

	// Remove Docker build verbosity
	output = compressDockerBuild(output)

	// Simplify Helm output
	output = compressHelmOutput(output)

	// Remove cloud provider logs
	output = removeProviderVerbosity(output)

	// Compress log timestamps
	output = compressLogTimestamps(output)

	// Remove polling output
	output = removePollingOutput(output)

	saved := original - len(output)
	if saved < 0 {
		saved = 0
	}

	return output, saved
}

func compressKubernetesOutput(text string) string {
	// Remove per-pod startup logs
	podRegex := regexp.MustCompile(`pod/\w+\s+\d+/\d+.*?Running\n`)
	text = podRegex.ReplaceAllString(text, "")

	// Compress kubectl get output
	getRegex := regexp.MustCompile(`(?m)^(\w+)\s+\d+/\d+\s+[\w-]+\s+[\w-]+\s+\d+\d+h\n`)
	text = getRegex.ReplaceAllString(text, "$1: running\n")

	// Remove duplicate status messages
	text = deduplicateMessages(text)

	return text
}

func compressTerraformPlan(text string) string {
	// Compress detailed attribute changes
	attrRegex := regexp.MustCompile(`(?m)^\s+-\s+\w+\s+= .*$\n?`)
	text = attrRegex.ReplaceAllString(text, "[attributes changes]\n")

	// Keep only summary
	summaryRegex := regexp.MustCompile(`Plan: \d+ to add, \d+ to change, \d+ to destroy`)
	if matches := summaryRegex.FindString(text); matches != "" {
		// Find the plan section
		if idx := strings.Index(text, "Plan:"); idx >= 0 {
			return text[idx:]
		}
	}

	return text
}

func compressDockerBuild(text string) string {
	// Remove layer-by-layer output
	layerRegex := regexp.MustCompile(`Step \d+/\d+ : .*?\n.*?---> [\w]+\n`)
	matches := layerRegex.FindAllString(text, -1)

	if len(matches) > 5 {
		// If many steps, compress intermediate ones
		text = "Building (multiple steps)...\n"
		if lastIdx := strings.LastIndex(text, "Step"); lastIdx > 0 {
			text += text[lastIdx:]
		}
	}

	// Remove verbose pull messages
	pullRegex := regexp.MustCompile(`Pulling from \w+.*?Pulling fs layer.*?Download complete\n`)
	text = pullRegex.ReplaceAllString(text, "[Pulling base image]\n")

	return text
}

func compressHelmOutput(text string) string {
	// Compress Helm release status
	statusRegex := regexp.MustCompile(`(?m)^STATUS: \w+$`)
	text = statusRegex.ReplaceAllString(text, "")

	// Remove metadata noise
	metaRegex := regexp.MustCompile(`(?m)^(CHART|APP VERSION|NAMESPACE): .*$`)
	text = metaRegex.ReplaceAllString(text, "")

	return text
}

func removeProviderVerbosity(text string) string {
	// Remove AWS/GCP/Azure debug logs
	providerRegex := regexp.MustCompile(`(?i)(DEBUG|TRACE).*?(aws|gcp|azure|azure|google|amazon|microsoft).*?\n`)
	text = providerRegex.ReplaceAllString(text, "")

	// Remove IAM policy dumps
	policyRegex := regexp.MustCompile(`"Version": "2012-10-17"[\s\S]*?"Statement":.*?\]`)
	text = policyRegex.ReplaceAllString(text, "[IAM Policy]")

	return text
}

func compressLogTimestamps(text string) string {
	// Remove millisecond precision from timestamps
	tsRegex := regexp.MustCompile(`(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2})\.\d+Z`)
	text = tsRegex.ReplaceAllString(text, "$1Z")

	// Remove duplicate nearby timestamps
	lines := strings.Split(text, "\n")
	var result []string
	var lastTS string

	for _, line := range lines {
		tsMatch := regexp.MustCompile(`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}`).FindString(line)
		if tsMatch == lastTS && tsMatch != "" {
			// Skip duplicate timestamp lines
			continue
		}
		if tsMatch != "" {
			lastTS = tsMatch
		}
		result = append(result, line)
	}

	return strings.Join(result, "\n")
}

func removePollingOutput(text string) string {
	// Remove status check polling lines (same message repeated)
	lines := strings.Split(text, "\n")
	var result []string
	var lastLine string
	repeatCount := 0

	for _, line := range lines {
		if line == lastLine {
			repeatCount++
			if repeatCount > 2 {
				continue // Skip after 3rd repeat
			}
		} else {
			repeatCount = 0
		}
		result = append(result, line)
		lastLine = line
	}

	return strings.Join(result, "\n")
}

func deduplicateMessages(text string) string {
	lines := strings.Split(text, "\n")
	seen := make(map[string]bool)
	var result []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			result = append(result, line)
		} else if !seen[trimmed] {
			result = append(result, line)
			seen[trimmed] = true
		}
	}

	return strings.Join(result, "\n")
}
