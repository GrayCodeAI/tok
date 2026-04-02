package headless

import (
	"encoding/json"
	"fmt"
	"strings"
)

type HeadlessMode struct {
	outputFormat string
	verbose      bool
	metrics      *HeadlessMetrics
}

type HeadlessMetrics struct {
	CommandsRun     int     `json:"commands_run"`
	TokensProcessed int     `json:"tokens_processed"`
	TokensSaved     int     `json:"tokens_saved"`
	SavingsPct      float64 `json:"savings_pct"`
	AvgLatency      float64 `json:"avg_latency_ms"`
}

func NewHeadlessMode() *HeadlessMode {
	return &HeadlessMode{
		outputFormat: "json",
		metrics:      &HeadlessMetrics{},
	}
}

func (h *HeadlessMode) SetFormat(format string) {
	h.outputFormat = format
}

func (h *HeadlessMode) SetVerbose(v bool) {
	h.verbose = v
}

func (h *HeadlessMode) Record(command string, inputTokens, outputTokens, savedTokens int) {
	h.metrics.CommandsRun++
	h.metrics.TokensProcessed += inputTokens
	h.metrics.TokensSaved += savedTokens
	if h.metrics.TokensProcessed > 0 {
		h.metrics.SavingsPct = float64(h.metrics.TokensSaved) / float64(h.metrics.TokensProcessed) * 100
	}
}

func (h *HeadlessMode) Report() string {
	switch h.outputFormat {
	case "json":
		data, _ := json.MarshalIndent(h.metrics, "", "  ")
		return string(data)
	case "text":
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("Commands Run: %d\n", h.metrics.CommandsRun))
		sb.WriteString(fmt.Sprintf("Tokens Processed: %d\n", h.metrics.TokensProcessed))
		sb.WriteString(fmt.Sprintf("Tokens Saved: %d\n", h.metrics.TokensSaved))
		sb.WriteString(fmt.Sprintf("Savings: %.1f%%\n", h.metrics.SavingsPct))
		return sb.String()
	default:
		return ""
	}
}

func (h *HeadlessMode) Metrics() *HeadlessMetrics {
	return h.metrics
}

func (h *HeadlessMode) Reset() {
	h.metrics = &HeadlessMetrics{}
}
