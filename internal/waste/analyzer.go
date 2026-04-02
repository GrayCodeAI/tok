package waste

import (
	"regexp"
	"strings"
)

type WasteType string

const (
	WasteWhitespace WasteType = "whitespace_bloat"
	WasteFiller     WasteType = "filler"
	WasteRedundant  WasteType = "redundant_instructions"
	WasteOutput     WasteType = "output_underutilization"
)

type WasteFinding struct {
	Type        WasteType `json:"type"`
	Description string    `json:"description"`
	Lines       []int     `json:"lines,omitempty"`
	Savings     int       `json:"estimated_savings_tokens"`
}

type WasteReport struct {
	TotalTokens     int            `json:"total_tokens"`
	WasteTokens     int            `json:"waste_tokens"`
	WasteScore      float64        `json:"waste_score"`
	Findings        []WasteFinding `json:"findings"`
	Recommendations []string       `json:"recommendations"`
}

type WhitespaceBloatDetector struct {
	trailingSpaceRe     *regexp.Regexp
	excessiveNewlinesRe *regexp.Regexp
	leadingSpaceRe      *regexp.Regexp
}

func NewWhitespaceBloatDetector() *WhitespaceBloatDetector {
	return &WhitespaceBloatDetector{
		trailingSpaceRe:     regexp.MustCompile(`[ \t]+$`),
		excessiveNewlinesRe: regexp.MustCompile(`\n{3,}`),
		leadingSpaceRe:      regexp.MustCompile(`^[ \t]{4,}`),
	}
}

func (d *WhitespaceBloatDetector) Detect(input string) []WasteFinding {
	var findings []WasteFinding
	lines := strings.Split(input, "\n")
	var affectedLines []int
	totalSavings := 0

	for i, line := range lines {
		if d.trailingSpaceRe.MatchString(line) {
			affectedLines = append(affectedLines, i+1)
			totalSavings += len(d.trailingSpaceRe.FindAllString(line, -1))
		}
	}

	matches := d.excessiveNewlinesRe.FindAllStringIndex(input, -1)
	if len(matches) > 0 {
		for _, m := range matches {
			extraNewlines := strings.Count(input[m[0]:m[1]], "\n") - 2
			totalSavings += extraNewlines
		}
	}

	if len(affectedLines) > 0 || len(matches) > 0 {
		findings = append(findings, WasteFinding{
			Type:        WasteWhitespace,
			Description: "Excessive whitespace detected",
			Lines:       affectedLines,
			Savings:     totalSavings,
		})
	}

	return findings
}

type FillerDetector struct {
	fillerPatterns []*regexp.Regexp
}

var defaultFillerPatterns = []string{
	`(?i)\b(here is the|here's the|the following|as mentioned|as stated)\b`,
	`(?i)\b(based on the|according to|in order to|due to the)\b`,
	`(?i)\b(it is important to|please note that|note that|keep in mind)\b`,
	`(?i)\b(let me|I will|I'll|I can|I should|I would)\b`,
	`(?i)\b(of course|certainly|absolutely|sure|okay|alright)\b`,
	`(?i)\b(however|therefore|furthermore|moreover|additionally)\b`,
	`(?i)\b(in conclusion|to summarize|in summary|overall)\b`,
	`(?i)\bas you can see|as shown|as demonstrated`,
}

func NewFillerDetector() *FillerDetector {
	fd := &FillerDetector{}
	for _, pattern := range defaultFillerPatterns {
		fd.fillerPatterns = append(fd.fillerPatterns, regexp.MustCompile(pattern))
	}
	return fd
}

func (d *FillerDetector) Detect(input string) []WasteFinding {
	var findings []WasteFinding
	lines := strings.Split(input, "\n")
	var affectedLines []int
	totalSavings := 0

	for i, line := range lines {
		for _, re := range d.fillerPatterns {
			if re.MatchString(line) {
				affectedLines = append(affectedLines, i+1)
				totalSavings += len(re.FindAllString(line, -1)) * 2
				break
			}
		}
	}

	if len(affectedLines) > 0 {
		findings = append(findings, WasteFinding{
			Type:        WasteFiller,
			Description: "Filler words and phrases detected",
			Lines:       affectedLines,
			Savings:     totalSavings,
		})
	}

	return findings
}

type RedundantInstructionDetector struct{}

func NewRedundantInstructionDetector() *RedundantInstructionDetector {
	return &RedundantInstructionDetector{}
}

func (d *RedundantInstructionDetector) Detect(input string) []WasteFinding {
	var findings []WasteFinding
	lines := strings.Split(input, "\n")
	seen := make(map[string]int)
	var redundantLines []int
	totalSavings := 0

	for i, line := range lines {
		normalized := strings.TrimSpace(strings.ToLower(line))
		if len(normalized) < 20 {
			continue
		}
		if prevLine, exists := seen[normalized]; exists {
			redundantLines = append(redundantLines, i+1)
			totalSavings += len(normalized) / 4
			seen[normalized] = prevLine
		} else {
			seen[normalized] = i + 1
		}
	}

	if len(redundantLines) > 0 {
		findings = append(findings, WasteFinding{
			Type:        WasteRedundant,
			Description: "Redundant or duplicate instructions detected",
			Lines:       redundantLines,
			Savings:     totalSavings,
		})
	}

	return findings
}

type OutputUtilizationTracker struct {
}

func NewOutputUtilizationTracker() *OutputUtilizationTracker {
	return &OutputUtilizationTracker{}
}

func (t *OutputUtilizationTracker) Track(inputTokens int, outputTokens int) WasteFinding {
	if inputTokens == 0 {
		return WasteFinding{}
	}
	utilization := float64(outputTokens) / float64(inputTokens) * 100
	if utilization < 30 {
		return WasteFinding{
			Type:        WasteOutput,
			Description: "Low output utilization - input much larger than output",
			Savings:     (inputTokens - outputTokens*3) / 4,
		}
	}
	return WasteFinding{}
}

type WasteScoreCalculator struct{}

func NewWasteScoreCalculator() *WasteScoreCalculator {
	return &WasteScoreCalculator{}
}

func (c *WasteScoreCalculator) Calculate(totalTokens int, findings []WasteFinding) float64 {
	if totalTokens == 0 {
		return 0
	}
	totalWaste := 0
	for _, f := range findings {
		totalWaste += f.Savings
	}
	score := float64(totalWaste) / float64(totalTokens) * 100
	if score > 100 {
		score = 100
	}
	return score
}

type WasteAnalyzer struct {
	whitespace  *WhitespaceBloatDetector
	filler      *FillerDetector
	redundant   *RedundantInstructionDetector
	utilization *OutputUtilizationTracker
	calculator  *WasteScoreCalculator
}

func NewWasteAnalyzer() *WasteAnalyzer {
	return &WasteAnalyzer{
		whitespace:  NewWhitespaceBloatDetector(),
		filler:      NewFillerDetector(),
		redundant:   NewRedundantInstructionDetector(),
		utilization: NewOutputUtilizationTracker(),
		calculator:  NewWasteScoreCalculator(),
	}
}

func (a *WasteAnalyzer) Analyze(input string) *WasteReport {
	var allFindings []WasteFinding
	allFindings = append(allFindings, a.whitespace.Detect(input)...)
	allFindings = append(allFindings, a.filler.Detect(input)...)
	allFindings = append(allFindings, a.redundant.Detect(input)...)

	totalTokens := len(input) / 4
	totalWaste := 0
	for _, f := range allFindings {
		totalWaste += f.Savings
	}

	wasteScore := a.calculator.Calculate(totalTokens, allFindings)

	var recommendations []string
	for _, f := range allFindings {
		switch f.Type {
		case WasteWhitespace:
			recommendations = append(recommendations, "Remove excessive whitespace to save tokens")
		case WasteFiller:
			recommendations = append(recommendations, "Strip filler words and phrases")
		case WasteRedundant:
			recommendations = append(recommendations, "Remove duplicate instructions")
		case WasteOutput:
			recommendations = append(recommendations, "Reduce input context - output utilization is low")
		}
	}

	return &WasteReport{
		TotalTokens:     totalTokens,
		WasteTokens:     totalWaste,
		WasteScore:      wasteScore,
		Findings:        allFindings,
		Recommendations: recommendations,
	}
}
