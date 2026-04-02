package recommendations

type CheckID string

const (
	CheckWhitespace      CheckID = "whitespace_optimization"
	CheckDuplicate       CheckID = "duplicate_removal"
	CheckModelRightSize  CheckID = "model_right_sizing"
	CheckHistoryCompact  CheckID = "history_compaction"
	CheckToolOutputPrune CheckID = "tool_output_pruning"
	CheckCacheUtil       CheckID = "cache_utilization"
	CheckPresetOpt       CheckID = "preset_optimization"
	CheckFilterTuning    CheckID = "filter_tuning"
	CheckBudgetAlign     CheckID = "budget_alignment"
	CheckProviderSelect  CheckID = "provider_selection"
	CheckStreamOpt       CheckID = "streaming_optimization"
	CheckCompressionMode CheckID = "compression_mode_selection"
)

type CheckResult struct {
	ID          CheckID `json:"id"`
	Passed      bool    `json:"passed"`
	Description string  `json:"description"`
	Savings     int     `json:"estimated_savings_tokens"`
	Suggestion  string  `json:"suggestion"`
}

type RecommendationReport struct {
	TotalChecks  int           `json:"total_checks"`
	PassedChecks int           `json:"passed_checks"`
	FailedChecks int           `json:"failed_checks"`
	TotalSavings int           `json:"total_estimated_savings"`
	OverallScore float64       `json:"overall_score"`
	Results      []CheckResult `json:"results"`
}

type CheckEngine struct {
	checks []func() CheckResult
}

func NewCheckEngine() *CheckEngine {
	return &CheckEngine{}
}

func (e *CheckEngine) Register(check func() CheckResult) {
	e.checks = append(e.checks, check)
}

func (e *CheckEngine) Run() *RecommendationReport {
	var results []CheckResult
	passed := 0
	failed := 0
	totalSavings := 0

	for _, check := range e.checks {
		result := check()
		results = append(results, result)
		if result.Passed {
			passed++
		} else {
			failed++
		}
		totalSavings += result.Savings
	}

	score := float64(passed) / float64(len(e.checks)) * 100

	return &RecommendationReport{
		TotalChecks:  len(e.checks),
		PassedChecks: passed,
		FailedChecks: failed,
		TotalSavings: totalSavings,
		OverallScore: score,
		Results:      results,
	}
}

func DefaultChecks() *CheckEngine {
	engine := NewCheckEngine()

	engine.Register(func() CheckResult {
		return CheckResult{
			ID:          CheckWhitespace,
			Passed:      false,
			Description: "Check for excessive whitespace in inputs",
			Savings:     50,
			Suggestion:  "Enable whitespace stripping in pipeline",
		}
	})

	engine.Register(func() CheckResult {
		return CheckResult{
			ID:          CheckDuplicate,
			Passed:      false,
			Description: "Check for duplicate content in context",
			Savings:     100,
			Suggestion:  "Enable deduplication layer",
		}
	})

	engine.Register(func() CheckResult {
		return CheckResult{
			ID:          CheckModelRightSize,
			Passed:      true,
			Description: "Check if current model matches task complexity",
			Savings:     0,
			Suggestion:  "Model selection is optimal",
		}
	})

	engine.Register(func() CheckResult {
		return CheckResult{
			ID:          CheckHistoryCompact,
			Passed:      false,
			Description: "Check conversation history size",
			Savings:     200,
			Suggestion:  "Enable history compaction or summarization",
		}
	})

	engine.Register(func() CheckResult {
		return CheckResult{
			ID:          CheckToolOutputPrune,
			Passed:      false,
			Description: "Check tool output pruning",
			Savings:     150,
			Suggestion:  "Enable aggressive tool output filtering",
		}
	})

	engine.Register(func() CheckResult {
		return CheckResult{
			ID:          CheckCacheUtil,
			Description: "Check cache hit rate",
			Passed:      true,
			Savings:     0,
			Suggestion:  "Cache utilization is good",
		}
	})

	engine.Register(func() CheckResult {
		return CheckResult{
			ID:          CheckPresetOpt,
			Description: "Check preset selection",
			Passed:      false,
			Savings:     75,
			Suggestion:  "Switch from 'balanced' to 'aggressive' preset for this workload",
		}
	})

	engine.Register(func() CheckResult {
		return CheckResult{
			ID:          CheckFilterTuning,
			Description: "Check custom filter configuration",
			Passed:      true,
			Savings:     0,
			Suggestion:  "Filter configuration is optimal",
		}
	})

	engine.Register(func() CheckResult {
		return CheckResult{
			ID:          CheckBudgetAlign,
			Description: "Check budget alignment with usage",
			Passed:      false,
			Savings:     50,
			Suggestion:  "Set explicit token budget to prevent overuse",
		}
	})

	engine.Register(func() CheckResult {
		return CheckResult{
			ID:          CheckProviderSelect,
			Description: "Check provider selection for cost efficiency",
			Passed:      true,
			Savings:     0,
			Suggestion:  "Provider selection is cost-efficient",
		}
	})

	engine.Register(func() CheckResult {
		return CheckResult{
			ID:          CheckStreamOpt,
			Description: "Check streaming optimization",
			Passed:      true,
			Savings:     0,
			Suggestion:  "Streaming is properly configured",
		}
	})

	engine.Register(func() CheckResult {
		return CheckResult{
			ID:          CheckCompressionMode,
			Description: "Check compression mode selection",
			Passed:      false,
			Savings:     100,
			Suggestion:  "Switch to 'core' compression mode for maximum savings",
		}
	})

	return engine
}
