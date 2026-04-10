package core

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tokman/internal/audit"
	"github.com/GrayCodeAI/tokman/internal/commands/registry"
	"github.com/GrayCodeAI/tokman/internal/commands/shared"
)

var (
	auditDays     int
	auditJSON     bool
	auditHTMLPath string
	auditSnapshot string
	auditCompareA string
	auditCompareB string
)

var auditCmd = &cobra.Command{
	Use:   "audit",
	Short: "Run optimization audit with waste, quality, context, and checkpoint analysis",
	Long: `Runs TokMan's optimization audit engine:
- waste detector framework
- context overhead audit
- quality score
- checkpoint policy recommendations
- drift snapshots and comparisons
- optional HTML dashboard`,
	Annotations: map[string]string{
		"tokman:skip_integrity": "true",
	},
	RunE: runAudit,
}

func init() {
	auditCmd.Flags().IntVar(&auditDays, "days", 30, "window to analyze in days")
	auditCmd.Flags().BoolVar(&auditJSON, "json", false, "output JSON")
	auditCmd.Flags().StringVar(&auditHTMLPath, "html", "", "write HTML dashboard to this path")
	auditCmd.Flags().StringVar(&auditSnapshot, "snapshot", "", "save snapshot with this name")
	auditCmd.Flags().StringVar(&auditCompareA, "compare-base", "", "base snapshot file for drift comparison")
	auditCmd.Flags().StringVar(&auditCompareB, "compare-candidate", "", "candidate snapshot file for drift comparison")

	registry.Add(func() { registry.Register(auditCmd) })
}

func runAudit(cmd *cobra.Command, args []string) error {
	if auditCompareA != "" || auditCompareB != "" {
		return runAuditCompare()
	}

	tracker, err := shared.OpenTracker()
	if err != nil {
		return fmt.Errorf("open tracker: %w", err)
	}
	defer tracker.Close()

	report, err := audit.GenerateWithOptions(tracker, auditDays, audit.GenerateOptions{
		ConfigPath: shared.GetConfigPath(),
	})
	if err != nil {
		return fmt.Errorf("generate audit: %w", err)
	}

	if auditSnapshot != "" {
		snapDir := filepath.Join(shared.GetConfigDir(), "audit", "snapshots")
		path, err := audit.SaveSnapshot(snapDir, auditSnapshot, report)
		if err != nil {
			return fmt.Errorf("save snapshot: %w", err)
		}
		fmt.Printf("Snapshot saved: %s\n", path)
	}

	if auditHTMLPath != "" {
		if err := audit.RenderHTML(auditHTMLPath, report); err != nil {
			return fmt.Errorf("render html: %w", err)
		}
		fmt.Printf("Dashboard written: %s\n", auditHTMLPath)
	}

	if auditJSON {
		b, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(b))
		return nil
	}

	printAuditReport(report)
	return nil
}

func runAuditCompare() error {
	if auditCompareA == "" || auditCompareB == "" {
		return fmt.Errorf("--compare-base and --compare-candidate are both required")
	}
	base, err := audit.LoadSnapshot(auditCompareA)
	if err != nil {
		return fmt.Errorf("load base snapshot: %w", err)
	}
	candidate, err := audit.LoadSnapshot(auditCompareB)
	if err != nil {
		return fmt.Errorf("load candidate snapshot: %w", err)
	}
	compare := audit.Compare(base, candidate)

	if auditJSON {
		b, err := json.MarshalIndent(compare, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(b))
		return nil
	}

	fmt.Println("TokMan Drift Validation")
	fmt.Println("=======================")
	fmt.Printf("Base:      %s\n", compare.BaseName)
	fmt.Printf("Candidate: %s\n", compare.CandidateName)
	fmt.Printf("Saved Tokens Delta:  %+d\n", compare.DeltaSavedTokens)
	fmt.Printf("Reduction Delta:     %+0.2f%%\n", compare.DeltaReductionPct)
	fmt.Printf("Quality Delta:       %+0.2f\n", compare.DeltaQualityScore)
	fmt.Printf("Parse Failure Delta: %+d\n", compare.DeltaParseFailures)
	fmt.Printf("Drift Changed:       %v\n", compare.DriftChanged)
	fmt.Printf("Verdict:             %s\n", compare.Verdict)
	return nil
}

func printAuditReport(r *audit.Report) {
	fmt.Println("TokMan Optimization Audit")
	fmt.Println("=========================")
	fmt.Printf("Window: %d days\n", r.Days)
	fmt.Printf("Commands: %d\n", r.Summary.CommandCount)
	fmt.Printf("Original Tokens: %d\n", r.Summary.Original)
	fmt.Printf("Filtered Tokens: %d\n", r.Summary.Filtered)
	fmt.Printf("Saved Tokens: %d\n", r.Summary.Saved)
	fmt.Printf("Reduction: %.2f%%\n", r.Summary.ReductionPct)
	fmt.Printf("Quality: %.1f (%s)\n", r.Quality.Score, r.Quality.Band)
	fmt.Printf("Budget Controller: %s\n", r.BudgetController.RecommendedMode)
	fmt.Printf("Anchor Retention: %s (keep-rate %.1f%%)\n", r.AnchorRetention.Grade, r.AnchorRetention.EstimatedKeepRate)
	if r.DriftFingerprint != "" {
		fmt.Printf("Drift Fingerprint: %s\n", r.DriftFingerprint[:16])
	}
	fmt.Println()

	fmt.Println("Waste Findings")
	fmt.Println("--------------")
	if len(r.WasteFindings) == 0 {
		fmt.Println("No major waste findings.")
	} else {
		for _, f := range r.WasteFindings {
			fmt.Printf("[%s] %s | waste=%d tokens (~$%.4f)\n", f.Severity, f.Description, f.EstimatedWaste, f.EstimatedWasteD)
			fmt.Printf("  Fix: %s\n", f.Recommendation)
		}
	}
	fmt.Println()

	fmt.Println("Checkpoint Policy")
	fmt.Println("-----------------")
	fmt.Printf("Recommended Triggers: %v\n", r.CheckpointPolicy.RecommendedTriggers)
	for _, note := range r.CheckpointPolicy.Notes {
		fmt.Printf("- %s\n", note)
	}
	fmt.Println()

	fmt.Println("Top Layers")
	fmt.Println("----------")
	if len(r.TopLayers) == 0 {
		fmt.Println("No layer data yet.")
	} else {
		for _, l := range r.TopLayers {
			fmt.Printf("%s | total=%d avg=%.1f calls=%d\n", l.LayerName, l.TotalSaved, l.AvgSaved, l.CallCount)
		}
	}
	fmt.Println()

	fmt.Println("Costly Prompts")
	fmt.Println("-------------")
	if len(r.CostlyPrompts) == 0 {
		fmt.Println("No costly prompt data yet.")
	} else {
		for _, cp := range r.CostlyPrompts {
			fmt.Printf("%s | count=%d original=%d est=$%.4f\n", cp.Command, cp.Count, cp.Original, cp.EstimatedUS)
		}
	}
	fmt.Println()

	fmt.Println("Intent Profiles")
	fmt.Println("---------------")
	if len(r.IntentProfiles) == 0 {
		fmt.Println("No intent profile data yet.")
	} else {
		for _, ip := range r.IntentProfiles {
			fmt.Printf("%s | commands=%d reduction=%.1f%%\n", ip.Intent, ip.Commands, ip.ReductionPct)
		}
	}
	fmt.Println()

	fmt.Println("Agent Budgets")
	fmt.Println("-------------")
	if len(r.AgentBudgets) == 0 {
		fmt.Println("No agent budget data yet.")
	} else {
		for _, a := range r.AgentBudgets {
			fmt.Printf("%s | share=%.1f%% reduction=%.1f%% cost=$%.4f\n", a.Agent, a.BudgetShare, a.ReductionPct, a.EstimatedUS)
		}
	}
	fmt.Println()

	fmt.Println("Recommendations")
	fmt.Println("---------------")
	for _, rec := range r.Recommendations {
		fmt.Printf("- %s\n", rec)
	}
}
