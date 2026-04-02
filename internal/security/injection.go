// Package security provides prompt injection detection.
package security

import (
	"regexp"
	"sort"
	"strings"
)

// InjectionType represents the type of injection detected.
type InjectionType string

const (
	InjectionSystemPrompt   InjectionType = "system_prompt"
	InjectionJailbreak      InjectionType = "jailbreak"
	InjectionDataExtraction InjectionType = "data_extraction"
	InjectionIndirect       InjectionType = "indirect"
	InjectionContextShift   InjectionType = "context_shift"
	InjectionRolePlay       InjectionType = "role_play"
)

// InjectionPattern represents a detected injection attempt.
type InjectionPattern struct {
	Type        InjectionType `json:"type"`
	Confidence  float64       `json:"confidence"`
	MatchedText string        `json:"matched_text"`
	Position    int           `json:"position"`
	Severity    string        `json:"severity"` // low, medium, high, critical
}

// InjectionDetector detects prompt injection attempts.
type InjectionDetector struct {
	patterns map[InjectionType][]*regexp.Regexp
	enabled  map[InjectionType]bool
	scores   map[InjectionType]float64
}

// NewInjectionDetector creates a new injection detector.
func NewInjectionDetector() *InjectionDetector {
	d := &InjectionDetector{
		patterns: make(map[InjectionType][]*regexp.Regexp),
		enabled:  make(map[InjectionType]bool),
		scores:   make(map[InjectionType]float64),
	}
	d.initPatterns()
	return d
}

// Enable enables a detection type.
func (d *InjectionDetector) Enable(injectionType InjectionType) {
	d.enabled[injectionType] = true
}

// Disable disables a detection type.
func (d *InjectionDetector) Disable(injectionType InjectionType) {
	d.enabled[injectionType] = false
}

// SetScoreThreshold sets the confidence threshold.
func (d *InjectionDetector) SetScoreThreshold(injectionType InjectionType, score float64) {
	d.scores[injectionType] = score
}

func (d *InjectionDetector) initPatterns() {
	// System prompt injection patterns
	d.patterns[InjectionSystemPrompt] = []*regexp.Regexp{
		regexp.MustCompile(`(?i)(?:system|assistant|ai)\s*(?:prompt|instruction|directive)`),
		regexp.MustCompile(`(?i)(?:ignore|disregard)\s*(?:previous|prior|above)\s*(?:instruction|prompt)`),
		regexp.MustCompile(`(?i)new\s*(?:system|assistant)\s*(?:prompt|instruction)`),
		regexp.MustCompile(`(?i)(?:you\s+are\s+now|from\s+now\s+on\s+you\s+are)`),
	}
	d.enabled[InjectionSystemPrompt] = true
	d.scores[InjectionSystemPrompt] = 0.7

	// Jailbreak patterns
	d.patterns[InjectionJailbreak] = []*regexp.Regexp{
		regexp.MustCompile(`(?i)(?:DAN|do\s+anything\s+now)`),
		regexp.MustCompile(`(?i)(?:jailbreak|mode\s*:\s*unrestricted)`),
		regexp.MustCompile(`(?i)(?:developer\s*mode|debug\s*mode)`),
		regexp.MustCompile(`(?i)(?:no\s+restriction|unfiltered|uncensored)`),
		regexp.MustCompile(`(?i)(?:pretend|simulate|act\s+as\s+if)\s+(?:you(?:'re|\s+are)?\s+)?(?:not\s+)?(?:bound|restricted|limited)`),
		regexp.MustCompile(`(?i)(?:ignore\s+ethical|ignore\s+moral|ignore\s+safety)`),
	}
	d.enabled[InjectionJailbreak] = true
	d.scores[InjectionJailbreak] = 0.8

	// Data extraction patterns
	d.patterns[InjectionDataExtraction] = []*regexp.Regexp{
		regexp.MustCompile(`(?i)(?:repeat\s+after\s+me|repeat\s+the\s+above|repeat\s+your)`),
		regexp.MustCompile(`(?i)(?:output\s+(?:all|everything)|show\s+(?:all|everything))`),
		regexp.MustCompile(`(?i)(?:copy|paste|reproduce)\s+(?:the\s+)?(?:prompt|instruction|system)`),
		regexp.MustCompile(`(?i)(?:what\s+(?:was|were)|reveal)\s+(?:your|the)\s+(?:instruction|prompt|system)`),
	}
	d.enabled[InjectionDataExtraction] = true
	d.scores[InjectionDataExtraction] = 0.6

	// Indirect injection patterns
	d.patterns[InjectionIndirect] = []*regexp.Regexp{
		regexp.MustCompile(`(?i)(?:summarize|translate|process)\s+(?:the\s+following|this):\s*(?:[^\n]*\n)*\s*(?:instruction|prompt)`),
		regexp.MustCompile(`(?i)(?:document|text|content)\s*:\s*.*?\n\s*(?:ignore|disregard)`),
		regexp.MustCompile(`(?i)---\s*begin\s*(?:document|text|content)\s*---.*?---\s*end\s*---`),
	}
	d.enabled[InjectionIndirect] = true
	d.scores[InjectionIndirect] = 0.5

	// Context shift patterns
	d.patterns[InjectionContextShift] = []*regexp.Regexp{
		regexp.MustCompile(`(?i)(?:let's\s+change|switching\s+to|moving\s+to)\s+(?:topic|subject|context)`),
		regexp.MustCompile(`(?i)(?:forget\s+(?:about|that)|never\s+mind\s+(?:that|the)\s+above)`),
		regexp.MustCompile(`(?i)(?:instead|rather),?\s*(?:let's|we\s+should)`),
	}
	d.enabled[InjectionContextShift] = true
	d.scores[InjectionContextShift] = 0.4

	// Role play patterns
	d.patterns[InjectionRolePlay] = []*regexp.Regexp{
		regexp.MustCompile(`(?i)(?:pretend|act|behave)\s+(?:to\s+be|as\s+if|like\s+you(?:'re|\s+are))`),
		regexp.MustCompile(`(?i)(?:roleplay|role-play)\s+as\s+`),
		regexp.MustCompile(`(?i)(?:you\s+are\s+now|from\s+now\s+on)\s+(?:an?|the)\s+`),
	}
	d.enabled[InjectionRolePlay] = true
	d.scores[InjectionRolePlay] = 0.3
}

// Detect scans content for injection attempts.
func (d *InjectionDetector) Detect(content string) []InjectionPattern {
	var detections []InjectionPattern

	for injectionType, patterns := range d.patterns {
		if !d.enabled[injectionType] {
			continue
		}

		for _, re := range patterns {
			matches := re.FindAllStringIndex(content, -1)
			for _, match := range matches {
				if len(match) >= 2 {
					matchedText := content[match[0]:match[1]]
					confidence := d.calculateConfidence(injectionType, matchedText, content)

					if confidence >= d.scores[injectionType] {
						detections = append(detections, InjectionPattern{
							Type:        injectionType,
							Confidence:  confidence,
							MatchedText: matchedText,
							Position:    match[0],
							Severity:    d.getSeverity(injectionType, confidence),
						})
					}
				}
			}
		}
	}

	return detections
}

// IsSafe checks if content appears safe (no injection detected).
func (d *InjectionDetector) IsSafe(content string) bool {
	detections := d.Detect(content)
	return len(detections) == 0
}

// ScanResult contains detailed scan results.
type ScanResult struct {
	Safe       bool               `json:"safe"`
	Score      float64            `json:"score"`      // 0.0 = safe, 1.0 = definitely injection
	Severity   string             `json:"severity"`   // none, low, medium, high, critical
	Detections []InjectionPattern `json:"detections"`
}

// Scan performs a detailed scan.
func (d *InjectionDetector) Scan(content string) ScanResult {
	detections := d.Detect(content)

	if len(detections) == 0 {
		return ScanResult{
			Safe:     true,
			Score:    0.0,
			Severity: "none",
		}
	}

	// Calculate overall score
	maxScore := 0.0
	maxSeverity := "low"
	for _, det := range detections {
		if det.Confidence > maxScore {
			maxScore = det.Confidence
		}
		if severityRank(det.Severity) > severityRank(maxSeverity) {
			maxSeverity = det.Severity
		}
	}

	return ScanResult{
		Safe:       false,
		Score:      maxScore,
		Severity:   maxSeverity,
		Detections: detections,
	}
}

// calculateConfidence calculates detection confidence.
func (d *InjectionDetector) calculateConfidence(injectionType InjectionType, match, content string) float64 {
	baseScore := 0.5

	// Adjust based on match characteristics
	matchLen := len(match)
	contentLen := len(content)

	// Longer matches relative to content are more confident
	if contentLen > 0 {
		ratio := float64(matchLen) / float64(contentLen)
		if ratio > 0.5 {
			baseScore += 0.2
		} else if ratio > 0.2 {
			baseScore += 0.1
		}
	}

	// Multiple indicators increase confidence
	lowerMatch := strings.ToLower(match)
	switch injectionType {
	case InjectionSystemPrompt:
		if strings.Contains(lowerMatch, "ignore") && strings.Contains(lowerMatch, "instruction") {
			baseScore += 0.2
		}
	case InjectionJailbreak:
		if strings.Contains(lowerMatch, "dan") || strings.Contains(lowerMatch, "unrestricted") {
			baseScore += 0.2
		}
	case InjectionDataExtraction:
		if strings.Contains(lowerMatch, "repeat") && strings.Contains(lowerMatch, "prompt") {
			baseScore += 0.2
		}
	}

	// Cap at 1.0
	if baseScore > 1.0 {
		baseScore = 1.0
	}

	return baseScore
}

// getSeverity returns severity based on type and confidence.
func (d *InjectionDetector) getSeverity(injectionType InjectionType, confidence float64) string {
	switch injectionType {
	case InjectionJailbreak:
		if confidence > 0.9 {
			return "critical"
		}
		if confidence > 0.7 {
			return "high"
		}
		return "medium"
	case InjectionSystemPrompt:
		if confidence > 0.8 {
			return "high"
		}
		if confidence > 0.6 {
			return "medium"
		}
		return "low"
	case InjectionDataExtraction:
		if confidence > 0.8 {
			return "medium"
		}
		return "low"
	default:
		if confidence > 0.8 {
			return "medium"
		}
		return "low"
	}
}

// severityRank returns numeric rank for severity comparison.
func severityRank(severity string) int {
	switch severity {
	case "critical":
		return 4
	case "high":
		return 3
	case "medium":
		return 2
	case "low":
		return 1
	default:
		return 0
	}
}

// Sanitize removes potentially dangerous content.
func (d *InjectionDetector) Sanitize(content string) string {
	detections := d.Detect(content)
	if len(detections) == 0 {
		return content
	}

	// Sort by position descending
	sort.Slice(detections, func(i, j int) bool {
		return detections[i].Position > detections[j].Position
	})

	result := content
	for _, det := range detections {
		start := det.Position
		end := det.Position + len(det.MatchedText)
		if start < len(result) && end <= len(result) {
			result = result[:start] + "[FILTERED]" + result[end:]
		}
	}

	return result
}

// AddPattern adds a custom detection pattern.
func (d *InjectionDetector) AddPattern(injectionType InjectionType, pattern string) error {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}
	d.patterns[injectionType] = append(d.patterns[injectionType], re)
	return nil
}
