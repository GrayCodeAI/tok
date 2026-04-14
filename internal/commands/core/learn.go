package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tokman/internal/commands/registry"
	"github.com/GrayCodeAI/tokman/internal/config"
)

var (
	learnShowRules bool
	learnReset     bool
	learnExport    string
	learnImport    string
	learnStats     bool
)

var learnCmd = &cobra.Command{
	Use:   "learn",
	Short: "ML-based command corrections and suggestions",
	Long: `Learn from command history to provide intelligent corrections.

TokMan analyzes your command patterns and common mistakes to suggest
corrections and optimizations. All learning is local and private.

Features:
  • Auto-correct common typos in commands
  • Suggest optimized alternatives
  • Learn from manual corrections
  • Export/import learned rules

Examples:
  tokman learn --rules       # Show learned rules
  tokman learn --stats       # Show learning statistics
  tokman learn --reset       # Reset all learned data
  tokman learn --export rules.json  # Export learned rules`,
	Annotations: map[string]string{
		"tokman:skip_integrity": "true",
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
}

func runLearn(cmd *cobra.Command, args []string) error {
	// Default to showing stats if no flags provided
	if !learnShowRules && !learnReset && learnExport == "" && learnImport == "" && !learnStats {
		learnStats = true
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

	fmt.Println()
	fmt.Println(cyan("Learned Correction Rules"))
	fmt.Println(strings.Repeat("═", 60))

	if len(rules) == 0 {
		fmt.Println("No learned rules yet.")
		fmt.Println("TokMan will learn from your command patterns over time.")
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

		fmt.Printf("\n%d. %s → %s\n", i+1, yellow(rule.Pattern), rule.Correction)
		fmt.Printf("   Confidence: %s  Uses: %d  Last: %s\n",
			confidenceColor.Sprintf("%.0f%%", rule.Confidence*100),
			rule.UsageCount,
			rule.LastUsed.Format("2006-01-02"))
	}

	fmt.Printf("\nTotal: %d rule(s)\n", len(rules))
	return nil
}

func showLearningStats() error {
	rules, err := loadLearnedRules()
	if err != nil {
		return fmt.Errorf("failed to load rules: %w", err)
	}

	cyan := color.New(color.FgCyan).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()

	fmt.Println()
	fmt.Println(cyan("TokMan Learning Statistics"))
	fmt.Println(strings.Repeat("═", 40))

	fmt.Printf("Learned rules:     %s\n", green(len(rules)))

	var totalCorrections int
	var avgConfidence float64
	for _, rule := range rules {
		totalCorrections += rule.UsageCount
		avgConfidence += rule.Confidence
	}

	if len(rules) > 0 {
		avgConfidence /= float64(len(rules))
		fmt.Printf("Total corrections: %s\n", green(totalCorrections))
		fmt.Printf("Avg confidence:    %.0f%%\n", avgConfidence*100)
	}

	// Show data size
	dataSize := getDataSize(learnDataPath())
	if dataSize > 0 {
		fmt.Printf("Data size:         %d KB\n", dataSize/1024)
	}

	fmt.Println()
	fmt.Println("Learning features:")
	fmt.Println("  • Common typo correction")
	fmt.Println("  • Command optimization suggestions")
	fmt.Println("  • Pattern recognition from history")

	return nil
}

func resetLearnedData() error {
	path := learnDataPath()

	// Confirm before reset
	fmt.Print("Reset all learned data? This cannot be undone [y/N]: ")
	var response string
	fmt.Scanln(&response)

	if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
		fmt.Println("Reset canceled.")
		return nil
	}

	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("failed to reset data: %w", err)
	}

	green := color.New(color.FgGreen).SprintFunc()
	fmt.Printf("%s All learned data has been reset\n", green("✓"))

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
	fmt.Printf("%s Exported %d rule(s) to %s\n", green("✓"), len(rules), outputPath)

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

	// Simple merge - could be smarter about duplicates
	merged := append(existingRules, importData.Rules...)

	if err := saveLearnedRules(merged); err != nil {
		return fmt.Errorf("failed to save rules: %w", err)
	}

	green := color.New(color.FgGreen).SprintFunc()
	fmt.Printf("%s Imported %d rule(s)\n", green("✓"), len(importData.Rules))

	return nil
}

func getDataSize(path string) int64 {
	var size int64
	filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size
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
