package system

import (
	"sort"
	"strconv"
	"strings"
	"time"

	out "github.com/GrayCodeAI/tok/internal/output"

	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tok/internal/commands/registry"
	"github.com/GrayCodeAI/tok/internal/filter"
)

var (
	benchLines   int
	benchProfile string
	benchMode    string
)

var benchCmd = &cobra.Command{
	Use:   "pipeline-bench",
	Short: "Run stage-by-stage pipeline benchmark report",
	Long:  "Runs tok pipeline on synthetic mixed content and prints per-stage token savings and time.",
	RunE:  runBenchmark,
}

func init() {
	benchCmd.Flags().IntVar(&benchLines, "lines", 1200, "number of synthetic lines")
	benchCmd.Flags().StringVar(&benchProfile, "profile", "adaptive", "compression profile: surface|trim|extract|core|adaptive|code|log|thread")
	benchCmd.Flags().StringVar(&benchMode, "mode", "minimal", "compression mode: minimal|aggressive")
	registry.Add(func() { registry.Register(benchCmd) })
}

func runBenchmark(cmd *cobra.Command, args []string) error {
	mode := filter.ModeMinimal
	if strings.EqualFold(benchMode, "aggressive") {
		mode = filter.ModeAggressive
	}

	profile := filter.Tier(strings.ToLower(strings.TrimSpace(benchProfile)))
	cfg := filter.TierConfig(profile, mode)
	cfg.EnableDiffAdapt = true
	cfg.EnableEPiC = true
	cfg.EnableSSDP = true
	cfg.EnableAgentOCR = true
	cfg.EnableS2MAD = true
	cfg.EnableACON = true
	cfg.EnableLatentCollab = true
	cfg.EnableGraphCoT = true
	cfg.EnableRoleBudget = true
	cfg.EnableSWEAdaptive = true
	cfg.EnableAgentOCRHist = true
	cfg.EnablePlanBudget = true
	cfg.EnableLightMem = true
	cfg.EnablePathShorten = true
	cfg.EnableJSONSampler = true
	cfg.EnableContextCrunch = true
	cfg.EnableSearchCrunch = true
	cfg.EnableStructColl = true

	input := syntheticBenchmarkInput(benchLines)
	p := filter.NewPipelineCoordinator(cfg)
	start := time.Now()
	_, stats, err := p.Process(input)
	if err != nil {
		return err
	}
	totalDur := time.Since(start)

	out.Global().Printf("tok Pipeline Benchmark\n")
	out.Global().Printf("Profile=%s Mode=%s Lines=%d\n\n", profile, mode, benchLines)
	fusion := filter.ClawFusionStageCoverage()
	out.Global().Printf("Fusion Coverage: %d/14 stages mapped\n\n", len(fusion))
	out.Global().Printf("%-26s %-10s %-12s\n", "Stage", "Saved", "Time")
	out.Global().Printf("%-26s %-10s %-12s\n", strings.Repeat("-", 26), strings.Repeat("-", 10), strings.Repeat("-", 12))

	keys := sortedLayerKeys(stats.LayerStats)
	for _, k := range keys {
		st := stats.LayerStats[k]
		if st.TokensSaved == 0 {
			continue
		}
		out.Global().Printf("%-26s %-10d %-12s\n", k, st.TokensSaved, formatDuration(st.Duration))
	}

	out.Global().Printf("\nOriginal: %d tokens\n", stats.OriginalTokens)
	out.Global().Printf("Final:    %d tokens\n", stats.FinalTokens)
	out.Global().Printf("Saved:    %d tokens (%.1f%%)\n", stats.TotalSaved, stats.ReductionPercent)
	out.Global().Printf("Total:    %s\n", totalDur.Round(time.Millisecond))
	return nil
}

func syntheticBenchmarkInput(lines int) string {
	if lines < 80 {
		lines = 80
	}
	out := make([]string, 0, lines)
	out = append(out,
		"Planner: investigate flaky migration and auth panic",
		"ERROR: nil pointer at internal/services/payment/handler/process.go:88",
		"diff --git a/internal/services/payment/handler/process.go b/internal/services/payment/handler/process.go",
		"@@ -84,7 +84,9 @@",
	)
	for i := 0; i < lines-20; i++ {
		switch {
		case i%11 == 0:
			out = append(out, "{\"id\": "+strconv.Itoa(i)+", \"path\": \"/api/v1/orders\", \"status\": \"ok\"},")
		case i%9 == 0:
			out = append(out, "INFO request completed path=/api/v1/orders duration=45ms user_id=123")
		case i%7 == 0:
			out = append(out, "Executor: apply migration in internal/services/payment/handler/process.go and rerun tests")
		case i%5 == 0:
			out = append(out, "therefore add nil guard before decode and return explicit error")
		default:
			out = append(out, "noise filler line for benchmark payload generation")
		}
	}
	out = append(out,
		"path: internal/services/payment/handler/process.go",
		"path: internal/services/payment/handler/process.go",
		"WARN: retrying migration",
		"WARN: retrying migration",
		"Result: go test ./... passed",
	)
	return strings.Join(out, "\n")
}

func sortedLayerKeys(m map[string]filter.LayerStat) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		ni, oi := stageOrder(keys[i])
		nj, oj := stageOrder(keys[j])
		if oi && oj && ni != nj {
			return ni < nj
		}
		return keys[i] < keys[j]
	})
	return keys
}

func stageOrder(k string) (int, bool) {
	parts := strings.SplitN(k, "_", 2)
	if len(parts) == 0 {
		return 0, false
	}
	n, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, false
	}
	return n, true
}

func formatDuration(ns int64) string {
	if ns <= 0 {
		return "0ms"
	}
	d := time.Duration(ns)
	if d < time.Millisecond {
		return d.Round(time.Microsecond).String()
	}
	return d.Round(time.Millisecond).String()
}
