package core

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	out "github.com/lakshmanpatel/tok/internal/output"

	"github.com/spf13/cobra"

	"github.com/lakshmanpatel/tok/internal/audit"
	"github.com/lakshmanpatel/tok/internal/commands/registry"
	"github.com/lakshmanpatel/tok/internal/commands/shared"
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
	Long: `Runs tok's optimization audit engine:
- waste detector framework
- context overhead audit
- quality score
- checkpoint policy recommendations
- drift snapshots and comparisons
- optional HTML dashboard`,
	Annotations: map[string]string{
		"tok:skip_integrity": "true",
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
		out.Global().Printf("Snapshot saved: %s\n", path)
	}

	if auditHTMLPath != "" {
		if err := audit.RenderHTML(auditHTMLPath, report); err != nil {
			return fmt.Errorf("render html: %w", err)
		}
		out.Global().Printf("Dashboard written: %s\n", auditHTMLPath)
	}

	if auditJSON {
		b, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			return err
		}
		out.Global().Println(string(b))
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
		out.Global().Println(string(b))
		return nil
	}

	out.Global().Println("tok Drift Validation")
	out.Global().Println("=======================")
	out.Global().Printf("Base:      %s\n", compare.BaseName)
	out.Global().Printf("Candidate: %s\n", compare.CandidateName)
	out.Global().Printf("Saved Tokens Delta:  %+d\n", compare.DeltaSavedTokens)
	out.Global().Printf("Reduction Delta:     %+0.2f%%\n", compare.DeltaReductionPct)
	out.Global().Printf("Quality Delta:       %+0.2f\n", compare.DeltaQualityScore)
	out.Global().Printf("Parse Failure Delta: %+d\n", compare.DeltaParseFailures)
	out.Global().Printf("Drift Changed:       %v\n", compare.DriftChanged)
	out.Global().Printf("Verdict:             %s\n", compare.Verdict)
	return nil
}

func printAuditReport(r *audit.Report) {
	out.Global().Println("tok Optimization Audit")
	out.Global().Println("=========================")
	out.Global().Printf("Window: %d days\n", r.Days)
	out.Global().Printf("Commands: %d\n", r.Summary.CommandCount)
	out.Global().Printf("Original Tokens: %d\n", r.Summary.Original)
	out.Global().Printf("Filtered Tokens: %d\n", r.Summary.Filtered)
	out.Global().Printf("Saved Tokens: %d\n", r.Summary.Saved)
	out.Global().Printf("Reduction: %.2f%%\n", r.Summary.ReductionPct)
	out.Global().Printf("Quality: %.1f (%s)\n", r.Quality.Score, r.Quality.Band)
	out.Global().Printf("Budget Controller: %s\n", r.BudgetController.RecommendedMode)
	out.Global().Printf("Anchor Retention: %s (keep-rate %.1f%%)\n", r.AnchorRetention.Grade, r.AnchorRetention.EstimatedKeepRate)
	if r.DriftFingerprint != "" {
		out.Global().Printf("Drift Fingerprint: %s\n", r.DriftFingerprint[:16])
	}
	out.Global().Println()

	out.Global().Println("Waste Findings")
	out.Global().Println("--------------")
	if len(r.WasteFindings) == 0 {
		out.Global().Println("No major waste findings.")
	} else {
		for _, f := range r.WasteFindings {
			out.Global().Printf("[%s] %s | waste=%d tokens (~$%.4f)\n", f.Severity, f.Description, f.EstimatedWaste, f.EstimatedWasteD)
			out.Global().Printf("  Fix: %s\n", f.Recommendation)
		}
	}
	out.Global().Println()

	out.Global().Println("Checkpoint Policy")
	out.Global().Println("-----------------")
	out.Global().Printf("Recommended Triggers: %v\n", r.CheckpointPolicy.RecommendedTriggers)
	for _, note := range r.CheckpointPolicy.Notes {
		out.Global().Printf("- %s\n", note)
	}
	out.Global().Println()

	out.Global().Println("Top Layers")
	out.Global().Println("----------")
	if len(r.TopLayers) == 0 {
		out.Global().Println("No layer data yet.")
	} else {
		for _, l := range r.TopLayers {
			out.Global().Printf("%s | total=%d avg=%.1f calls=%d\n", l.LayerName, l.TotalSaved, l.AvgSaved, l.CallCount)
		}
	}
	out.Global().Println()

	out.Global().Println("Costly Prompts")
	out.Global().Println("-------------")
	if len(r.CostlyPrompts) == 0 {
		out.Global().Println("No costly prompt data yet.")
	} else {
		for _, cp := range r.CostlyPrompts {
			out.Global().Printf("%s | count=%d original=%d est=$%.4f\n", cp.Command, cp.Count, cp.Original, cp.EstimatedUS)
		}
	}
	out.Global().Println()

	out.Global().Println("Intent Profiles")
	out.Global().Println("---------------")
	if len(r.IntentProfiles) == 0 {
		out.Global().Println("No intent profile data yet.")
	} else {
		for _, ip := range r.IntentProfiles {
			out.Global().Printf("%s | commands=%d reduction=%.1f%%\n", ip.Intent, ip.Commands, ip.ReductionPct)
		}
	}
	out.Global().Println()

	out.Global().Println("Agent Budgets")
	out.Global().Println("-------------")
	if len(r.AgentBudgets) == 0 {
		out.Global().Println("No agent budget data yet.")
	} else {
		for _, a := range r.AgentBudgets {
			out.Global().Printf("%s | share=%.1f%% reduction=%.1f%% cost=$%.4f\n", a.Agent, a.BudgetShare, a.ReductionPct, a.EstimatedUS)
		}
	}
	out.Global().Println()

	out.Global().Println("Recommendations")
	out.Global().Println("---------------")
	for _, rec := range r.Recommendations {
		out.Global().Printf("- %s\n", rec)
	}
}
