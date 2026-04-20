package core

import (
	"encoding/json"
	"fmt"
	out "github.com/lakshmanpatel/tok/internal/output"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/lakshmanpatel/tok/internal/commands/registry"
	"github.com/lakshmanpatel/tok/internal/config"
)

var (
	learnShowRules     bool
	learnReset         bool
	learnExport        string
	learnImport        string
	learnStats         bool
	learnWriteRules    bool
	learnRulesPath     string
	learnMinConfidence float64
	learnMinOccurs     int
)

var learnCmd = &cobra.Command{
	Use:   "learn",
	Short: "ML-based command corrections and suggestions",
	Long: `Learn from command history to provide intelligent corrections.

tok analyzes your command patterns and common mistakes to suggest
corrections and optimizations. All learning is local and private.

Features:
  • Auto-correct common typos in commands
  • Suggest optimized alternatives
  • Learn from manual corrections
  • Export/import learned rules

Examples:
  tok learn --rules       # Show learned rules
  tok learn --stats       # Show learning statistics
  tok learn --reset       # Reset all learned data
  tok learn --export rules.json  # Export learned rules
  tok learn --write-rules        # Write rules to ~/.claude/rules/cli-corrections.md
  tok learn --write-rules --min-confidence 0.7 --min-occurrences 3`,
	Annotations: map[string]string{
		"tok:skip_integrity": "true",
	},
	RunE: runLearn,
}

func init() {
	registry.Add(func() { registry.Register(learnCmd) })
	learnCmd.Flags().BoolVar(&learnShowRules, "rules", false, "Show learned correction rules")
	learnCmd.Flags().BoolVar(&learnReset, "reset", false, "Reset all learned data")
	learnCmd.Flags().StringVar(&learnExport, "export", "", "Export learned rules to file")
	learnCmd.Flags().StringVar(&learnImport, "import", "", "Import learned rules from file")
	learnCmd.Flags().BoolVar(&learnStats, "stats", false, "Show learning statistics")
	learnCmd.Flags().BoolVar(&learnWriteRules, "write-rules", false, "Write rules as markdown to ~/.claude/rules/cli-corrections.md (or --rules-path)")
	learnCmd.Flags().StringVar(&learnRulesPath, "rules-path", "", "Override output path for --write-rules")
	learnCmd.Flags().Float64Var(&learnMinConfidence, "min-confidence", 0.6, "Minimum confidence threshold for --write-rules")
	learnCmd.Flags().IntVar(&learnMinOccurs, "min-occurrences", 2, "Minimum usage count for --write-rules")
}

func runLearn(cmd *cobra.Command, args []string) error {
	// Default to showing stats if no flags provided
	if !learnShowRules && !learnReset && learnExport == "" && learnImport == "" && !learnStats && !learnWriteRules {
		learnStats = true
	}

	if learnWriteRules {
		return writeRulesMarkdown(learnRulesPath)
	}

	if learnReset {
		return resetLearnedData()
	}

	if learnExport != "" {
		return exportLearnedRules(learnExport)
	}

	if learnImport != "" {
		return importLearnedRules(learnImport)
	}

	if learnShowRules {
		return showLearnedRules()
	}

	if learnStats {
		return showLearningStats()
	}

	return nil
}

func learnDataPath() string {
	return filepath.Join(config.DataPath(), "learn")
}

type learnedRule struct {
	Pattern    string    `json:"pattern"`
	Correction string    `json:"correction"`
	Confidence float64   `json:"confidence"`
	UsageCount int       `json:"usage_count"`
	LastUsed   time.Time `json:"last_used"`
	CreatedAt  time.Time `json:"created_at"`
}

type learningStatistics struct {
	TotalRules       int       `json:"total_rules"`
	TotalCorrections int       `json:"total_corrections"`
	LastUpdated      time.Time `json:"last_updated"`
	DataSize         int64     `json:"data_size_bytes"`
}

func loadLearnedRules() ([]learnedRule, error) {
	path := filepath.Join(learnDataPath(), "rules.json")

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []learnedRule{}, nil
		}
		return nil, err
	}

	var rules []learnedRule
	if err := json.Unmarshal(data, &rules); err != nil {
		return nil, err
	}

	return rules, nil
}

func saveLearnedRules(rules []learnedRule) error {
	path := filepath.Join(learnDataPath(), "rules.json")

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(rules, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func showLearnedRules() error {
	rules, err := loadLearnedRules()
	if err != nil {
		return fmt.Errorf("failed to load rules: %w", err)
	}

	cyan := color.New(color.FgCyan).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	out.Global().Println()
	out.Global().Println(cyan("Learned Correction Rules"))
	out.Global().Println(strings.Repeat("═", 60))

	if len(rules) == 0 {
		out.Global().Println("No learned rules yet.")
		out.Global().Println("tok will learn from your command patterns over time.")
		return nil
	}

	for i, rule := range rules {
		confidenceColor := color.New(color.FgGreen)
		if rule.Confidence < 0.5 {
			confidenceColor = color.New(color.FgYellow)
		}
		if rule.Confidence < 0.3 {
			confidenceColor = color.New(color.FgRed)
		}

		out.Global().Printf("\n%d. %s → %s\n", i+1, yellow(rule.Pattern), rule.Correction)
		out.Global().Printf("   Confidence: %s  Uses: %d  Last: %s\n",
			confidenceColor.Sprintf("%.0f%%", rule.Confidence*100),
			rule.UsageCount,
			rule.LastUsed.Format("2006-01-02"))
	}

	out.Global().Printf("\nTotal: %d rule(s)\n", len(rules))
	return nil
}

func showLearningStats() error {
	rules, err := loadLearnedRules()
	if err != nil {
		return fmt.Errorf("failed to load rules: %w", err)
	}

	cyan := color.New(color.FgCyan).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()

	out.Global().Println()
	out.Global().Println(cyan("tok Learning Statistics"))
	out.Global().Println(strings.Repeat("═", 40))

	out.Global().Printf("Learned rules:     %s\n", green(len(rules)))

	var totalCorrections int
	var avgConfidence float64
	for _, rule := range rules {
		totalCorrections += rule.UsageCount
		avgConfidence += rule.Confidence
	}

	if len(rules) > 0 {
		avgConfidence /= float64(len(rules))
		out.Global().Printf("Total corrections: %s\n", green(totalCorrections))
		out.Global().Printf("Avg confidence:    %.0f%%\n", avgConfidence*100)
	}

	// Show data size
	dataSize := getDataSize(learnDataPath())
	if dataSize > 0 {
		out.Global().Printf("Data size:         %d KB\n", dataSize/1024)
	}

	out.Global().Println()
	out.Global().Println("Learning features:")
	out.Global().Println("  • Common typo correction")
	out.Global().Println("  • Command optimization suggestions")
	out.Global().Println("  • Pattern recognition from history")

	return nil
}

func resetLearnedData() error {
	path := learnDataPath()

	// Confirm before reset
	out.Global().Print("Reset all learned data? This cannot be undone [y/N]: ")
	var response string
	fmt.Scanln(&response)

	if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
		out.Global().Println("Reset canceled.")
		return nil
	}

	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("failed to reset data: %w", err)
	}

	green := color.New(color.FgGreen).SprintFunc()
	out.Global().Printf("%s All learned data has been reset\n", green("✓"))

	return nil
}

func exportLearnedRules(outputPath string) error {
	rules, err := loadLearnedRules()
	if err != nil {
		return fmt.Errorf("failed to load rules: %w", err)
	}

	export := map[string]interface{}{
		"exported_at": time.Now().UTC().Format(time.RFC3339),
		"version":     "1.0",
		"rules":       rules,
	}

	data, err := json.MarshalIndent(export, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal rules: %w", err)
	}

	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write export: %w", err)
	}

	green := color.New(color.FgGreen).SprintFunc()
	out.Global().Printf("%s Exported %d rule(s) to %s\n", green("✓"), len(rules), outputPath)

	return nil
}

func importLearnedRules(inputPath string) error {
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var importData struct {
		Version string        `json:"version"`
		Rules   []learnedRule `json:"rules"`
	}

	if err := json.Unmarshal(data, &importData); err != nil {
		return fmt.Errorf("failed to parse file: %w", err)
	}

	// Merge with existing rules
	existingRules, err := loadLearnedRules()
	if err != nil {
		return fmt.Errorf("failed to load existing rules: %w", err)
	}

	// Merge by Pattern: sum UsageCount, keep highest confidence, latest LastUsed.
	byPattern := make(map[string]learnedRule, len(existingRules)+len(importData.Rules))
	order := make([]string, 0, len(existingRules)+len(importData.Rules))
	upsert := func(r learnedRule) {
		existing, ok := byPattern[r.Pattern]
		if !ok {
			byPattern[r.Pattern] = r
			order = append(order, r.Pattern)
			return
		}
		existing.UsageCount += r.UsageCount
		if r.Confidence > existing.Confidence {
			existing.Confidence = r.Confidence
			existing.Correction = r.Correction
		}
		if r.LastUsed.After(existing.LastUsed) {
			existing.LastUsed = r.LastUsed
		}
		if !r.CreatedAt.IsZero() && (existing.CreatedAt.IsZero() || r.CreatedAt.Before(existing.CreatedAt)) {
			existing.CreatedAt = r.CreatedAt
		}
		byPattern[r.Pattern] = existing
	}
	for _, r := range existingRules {
		upsert(r)
	}
	imported := 0
	for _, r := range importData.Rules {
		if _, dup := byPattern[r.Pattern]; !dup {
			imported++
		}
		upsert(r)
	}
	merged := make([]learnedRule, 0, len(order))
	for _, p := range order {
		merged = append(merged, byPattern[p])
	}

	if err := saveLearnedRules(merged); err != nil {
		return fmt.Errorf("failed to save rules: %w", err)
	}

	green := color.New(color.FgGreen).SprintFunc()
	out.Global().Printf("%s Imported %d new rule(s), merged %d duplicate(s)\n",
		green("✓"), imported, len(importData.Rules)-imported)

	return nil
}

func getDataSize(path string) int64 {
	var size int64
	_ = filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		size += info.Size()
		return nil
	})
	return size
}

// writeRulesMarkdown exports learned rules as markdown to the given path.
// If path is "default" or "" it writes to ~/.claude/rules/cli-corrections.md.
// Filters by --min-confidence and --min-occurrences thresholds.
func writeRulesMarkdown(target string) error {
	if target == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("cannot resolve home: %w", err)
		}
		target = filepath.Join(home, ".claude", "rules", "cli-corrections.md")
	}

	rules, err := loadLearnedRules()
	if err != nil {
		return fmt.Errorf("load rules: %w", err)
	}

	kept := rules[:0]
	for _, r := range rules {
		if r.Confidence >= learnMinConfidence && r.UsageCount >= learnMinOccurs {
			kept = append(kept, r)
		}
	}

	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}

	var b strings.Builder
	fmt.Fprintf(&b, "# CLI Corrections (tok learn)\n\n")
	fmt.Fprintf(&b, "_Generated: %s_\n", time.Now().UTC().Format(time.RFC3339))
	fmt.Fprintf(&b, "_Filters: confidence ≥ %.2f, occurrences ≥ %d_\n\n",
		learnMinConfidence, learnMinOccurs)

	if len(kept) == 0 {
		b.WriteString("No rules meet the threshold yet.\n")
	} else {
		b.WriteString("| Pattern | Correction | Confidence | Uses |\n")
		b.WriteString("|---------|-----------|-----------|------|\n")
		for _, r := range kept {
			fmt.Fprintf(&b, "| `%s` | `%s` | %.0f%% | %d |\n",
				escapeMD(r.Pattern), escapeMD(r.Correction),
				r.Confidence*100, r.UsageCount)
		}
	}

	if err := os.WriteFile(target, []byte(b.String()), 0644); err != nil {
		return fmt.Errorf("write: %w", err)
	}

	green := color.New(color.FgGreen).SprintFunc()
	out.Global().Printf("%s Wrote %d rule(s) to %s\n", green("✓"), len(kept), target)
	return nil
}

func escapeMD(s string) string {
	return strings.NewReplacer("|", "\\|", "`", "'").Replace(s)
}

// FindCorrection looks for a learned correction for the given input
// This function is called by the command discovery system
func FindCorrection(input string) (string, bool) {
	rules, err := loadLearnedRules()
	if err != nil {
		return "", false
	}

	for _, rule := range rules {
		if strings.EqualFold(rule.Pattern, input) && rule.Confidence > 0.5 {
			return rule.Correction, true
		}
	}

	return "", false
}
