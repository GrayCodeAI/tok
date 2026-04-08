// Package pipeline provides pre-processing and post-processing stages.
package pipeline

import (
	"fmt"
	"strings"

	"github.com/GrayCodeAI/tokman/internal/cortex"
	"github.com/GrayCodeAI/tokman/internal/security"
	"github.com/GrayCodeAI/tokman/internal/simd"
)

// PreProcessor handles pre-processing before compression.
type PreProcessor struct {
	cortex          *cortex.GateRegistry
	scanner         *security.Scanner
	stripANSI       bool
	normalize       bool
	detectPII       bool
	detectInjection bool
}

// NewPreProcessor creates a new pre-processor.
func NewPreProcessor() *PreProcessor {
	return &PreProcessor{
		cortex:          cortex.NewGateRegistry(),
		scanner:         security.NewScanner(),
		stripANSI:       true,
		normalize:       true,
		detectPII:       true,
		detectInjection: true,
	}
}

// PreProcessResult contains pre-processing results.
type PreProcessResult struct {
	Content       string             `json:"content"`
	ContentType   string             `json:"content_type"`
	Language      string             `json:"language"`
	SecurityScan  []security.Finding `json:"security_scan,omitempty"`
	ANSIRemoved   bool               `json:"ansi_removed"`
	AppliedGates  []string           `json:"applied_gates,omitempty"`
	OriginalSize  int                `json:"original_size"`
	ProcessedSize int                `json:"processed_size"`
}

// Process runs pre-processing.
func (p *PreProcessor) Process(content string) *PreProcessResult {
	result := &PreProcessResult{
		Content:      content,
		OriginalSize: len(content),
	}

	// 1. Security scanning
	if p.detectPII || p.detectInjection {
		findings := p.scanner.Scan(content)
		result.SecurityScan = findings
		// Redact secrets if found
		for _, finding := range findings {
			if finding.Severity == "critical" {
				content = security.RedactPII(content)
				break
			}
		}
	}

	// 2. Strip ANSI codes
	if p.stripANSI {
		if simd.HasANSI(content) {
			content = simd.StripANSI(content)
			result.ANSIRemoved = true
		}
	}

	// 3. Normalize whitespace
	if p.normalize {
		content = normalizeWhitespace(content)
	}

	// 4. Content detection
	detection := p.cortex.Analyze(content)
	result.ContentType = detection.ContentType
	result.Language = detection.Language

	// 5. Get applicable gates
	result.AppliedGates = p.cortex.GetApplicableGates(content)

	result.Content = content
	result.ProcessedSize = len(content)

	return result
}

// SetOptions sets pre-processing options.
func (p *PreProcessor) SetOptions(opts PreProcessOptions) {
	p.stripANSI = opts.StripANSI
	p.normalize = opts.Normalize
	p.detectPII = opts.DetectPII
	p.detectInjection = opts.DetectInjection
}

// PreProcessOptions provides pre-processing options.
type PreProcessOptions struct {
	StripANSI       bool
	Normalize       bool
	DetectPII       bool
	DetectInjection bool
}

// PostProcessor handles post-processing after compression.
type PostProcessor struct {
	addMarkers   bool
	addStats     bool
	verifyOutput bool
}

// NewPostProcessor creates a new post-processor.
func NewPostProcessor() *PostProcessor {
	return &PostProcessor{
		addMarkers:   true,
		addStats:     true,
		verifyOutput: false,
	}
}

// PostProcessResult contains post-processing results.
type PostProcessResult struct {
	Content        string  `json:"content"`
	OriginalSize   int     `json:"original_size"`
	CompressedSize int     `json:"compressed_size"`
	TokensSaved    int     `json:"tokens_saved"`
	ReductionPct   float64 `json:"reduction_percent"`
	Marker         string  `json:"marker,omitempty"`
	Verified       bool    `json:"verified"`
}

// Process runs post-processing.
func (p *PostProcessor) Process(original, compressed string, stats CompressionStats) *PostProcessResult {
	result := &PostProcessResult{
		Content:        compressed,
		OriginalSize:   stats.OriginalSize,
		CompressedSize: stats.CompressedSize,
		TokensSaved:    stats.TokensSaved,
		ReductionPct:   stats.ReductionPercent(),
	}

	// Add compression marker
	if p.addMarkers && stats.TokensSaved > 0 {
		result.Marker = formatCompressionMarker(stats)
		result.Content = compressed + "\n" + result.Marker
	}

	// Add stats if enabled
	if p.addStats {
		statsLine := formatStatsLine(stats)
		result.Content += "\n" + statsLine
	}

	return result
}

// CompressionStats contains compression statistics.
type CompressionStats struct {
	OriginalSize   int
	CompressedSize int
	TokensSaved    int
	Algorithm      string
}

// ReductionPercent returns the reduction percentage.
func (s CompressionStats) ReductionPercent() float64 {
	if s.OriginalSize == 0 {
		return 0
	}
	return float64(s.OriginalSize-s.CompressedSize) / float64(s.OriginalSize) * 100
}

// SetOptions sets post-processing options.
func (p *PostProcessor) SetOptions(opts PostProcessOptions) {
	p.addMarkers = opts.AddMarkers
	p.addStats = opts.AddStats
	p.verifyOutput = opts.VerifyOutput
}

// PostProcessOptions provides post-processing options.
type PostProcessOptions struct {
	AddMarkers   bool
	AddStats     bool
	VerifyOutput bool
}

// Pipeline coordinates pre and post processing.
type Pipeline struct {
	pre  *PreProcessor
	post *PostProcessor
}

// NewPipeline creates a new processing pipeline.
func NewPipeline() *Pipeline {
	return &Pipeline{
		pre:  NewPreProcessor(),
		post: NewPostProcessor(),
	}
}

// Process runs the full pipeline.
func (p *Pipeline) Process(content string, compressFunc func(string) (string, int)) (*PipelineResult, error) {
	// Pre-processing
	preResult := p.pre.Process(content)

	// Compression
	compressed, saved := compressFunc(preResult.Content)

	stats := CompressionStats{
		OriginalSize:   preResult.ProcessedSize,
		CompressedSize: len(compressed),
		TokensSaved:    saved,
	}

	// Post-processing
	postResult := p.post.Process(preResult.Content, compressed, stats)

	return &PipelineResult{
		PreProcess:  preResult,
		PostProcess: postResult,
		FinalOutput: postResult.Content,
	}, nil
}

// PipelineResult contains the full pipeline results.
type PipelineResult struct {
	PreProcess  *PreProcessResult  `json:"pre_process"`
	PostProcess *PostProcessResult `json:"post_process"`
	FinalOutput string             `json:"final_output"`
}

// Helper functions

func normalizeWhitespace(content string) string {
	// Replace tabs with spaces
	content = strings.ReplaceAll(content, "\t", "    ")

	// Remove trailing whitespace
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " \t")
	}

	return strings.Join(lines, "\n")
}

func formatCompressionMarker(stats CompressionStats) string {
	return fmt.Sprintf("[Compressed: %.1f%% reduction, %d tokens saved]",
		stats.ReductionPercent(), stats.TokensSaved)
}

func formatStatsLine(stats CompressionStats) string {
	return fmt.Sprintf("# Stats: original=%d compressed=%d saved=%d ratio=%.2f",
		stats.OriginalSize, stats.CompressedSize, stats.TokensSaved,
		float64(stats.CompressedSize)/float64(stats.OriginalSize))
}
