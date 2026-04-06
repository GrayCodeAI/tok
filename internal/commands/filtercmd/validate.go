package filtercmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tokman/internal/commands/registry"
	"github.com/GrayCodeAI/tokman/internal/config"
	"github.com/GrayCodeAI/tokman/internal/toml"
)

var (
	validateAll     bool
	validateVerbose bool
)

var validateCmd = &cobra.Command{
	Use:   "validate [filter-file]",
	Short: "Validate TOML filter files",
	Long: `Validate TOML filter files for syntax errors and configuration issues.

Checks performed:
  - TOML syntax validation
  - Schema version compatibility
  - Regex pattern validity
  - Required field presence
  - Test case validity (if tests are defined)
  - Conflicting configuration detection

Usage:
  tokman filter validate                    # Validate all filters
  tokman filter validate git.toml           # Validate specific file
  tokman filter validate --all              # Validate all (builtin + user)
  tokman filter validate -v                 # Verbose output`,
	RunE: runValidate,
}

func init() {
	validateCmd.Flags().BoolVarP(&validateAll, "all", "a", false, "validate all filters (builtin + user)")
	validateCmd.Flags().BoolVarP(&validateVerbose, "verbose", "v", false, "verbose output")
	registry.Add(func() { registry.Register(validateCmd) })
}

func runValidate(cmd *cobra.Command, args []string) error {
	var filesToValidate []string

	// Determine which files to validate
	if len(args) > 0 {
		// Validate specific file
		filePath := args[0]
		if !filepath.IsAbs(filePath) {
			// Try to find it in filter directories
			found, err := findFilterFile(filePath)
			if err != nil {
				return err
			}
			filePath = found
		}
		filesToValidate = []string{filePath}
	} else {
		// Validate all filters
		files, err := findAllFilterFiles(validateAll)
		if err != nil {
			return fmt.Errorf("failed to find filter files: %w", err)
		}
		filesToValidate = files
	}

	if len(filesToValidate) == 0 {
		fmt.Println("No filter files found to validate")
		return nil
	}

	// Print header
	cyan := color.New(color.FgCyan).SprintFunc()
	fmt.Printf("%s\n", cyan(fmt.Sprintf("Validating %d filter file(s)...", len(filesToValidate))))
	fmt.Println()

	// Validate each file
	results := make([]ValidationResult, 0, len(filesToValidate))
	for _, file := range filesToValidate {
		result := validateFilterFile(file)
		results = append(results, result)
	}

	// Print results
	printValidationResults(results)

	// Exit with error if any validation failed
	hasErrors := false
	for _, result := range results {
		if !result.Valid {
			hasErrors = true
			break
		}
	}

	if hasErrors {
		os.Exit(1)
	}

	return nil
}

type ValidationResult struct {
	FilePath   string
	Valid      bool
	Errors     []string
	Warnings   []string
	FilterName string
	TestCount  int
}

func validateFilterFile(path string) ValidationResult {
	result := ValidationResult{
		FilePath: path,
		Valid:    true,
		Errors:   []string{},
		Warnings: []string{},
	}

	// Try to parse the file
	parser := toml.NewParser()
	filter, err := parser.ParseFile(path)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to parse: %v", err))
		return result
	}

	// Check schema version
	if filter.SchemaVersion != toml.SchemaVersion {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("Schema version mismatch: got %d, expected %d",
				filter.SchemaVersion, toml.SchemaVersion))
	}

	// Validate each filter rule
	for name, rule := range filter.Filters {
		result.FilterName = name

		// Validate match_command regex
		if rule.MatchCommand == "" {
			result.Errors = append(result.Errors,
				fmt.Sprintf("[%s] missing match_command", name))
			result.Valid = false
		} else {
			if _, err := regexp.Compile(rule.MatchCommand); err != nil {
				result.Errors = append(result.Errors,
					fmt.Sprintf("[%s] invalid match_command regex: %v", name, err))
				result.Valid = false
			}
		}

		// Validate strip_lines_matching patterns
		for i, pattern := range rule.StripLinesMatching {
			if _, err := regexp.Compile(pattern); err != nil {
				result.Errors = append(result.Errors,
					fmt.Sprintf("[%s] invalid strip_lines_matching[%d] regex: %v", name, i, err))
				result.Valid = false
			}
		}

		// Validate keep_lines_matching patterns
		for i, pattern := range rule.KeepLinesMatching {
			if _, err := regexp.Compile(pattern); err != nil {
				result.Errors = append(result.Errors,
					fmt.Sprintf("[%s] invalid keep_lines_matching[%d] regex: %v", name, i, err))
				result.Valid = false
			}
		}

		// Check for conflicts
		if rule.Head > 0 && rule.Tail > 0 {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("[%s] both head and tail specified (head takes precedence)", name))
		}

		if len(rule.StripLinesMatching) > 0 && len(rule.KeepLinesMatching) > 0 {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("[%s] both strip and keep patterns specified (keep is applied after strip)", name))
		}

		// Check for reasonable limits
		if rule.MaxLines > 1000 {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("[%s] max_lines is very large (%d), may not reduce tokens effectively", name, rule.MaxLines))
		}
	}

	// Validate tests
	tests, err := toml.ParseTests(filter.RawContent)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to parse tests: %v", err))
		result.Valid = false
	} else {
		result.TestCount = tests.TotalTests()

		// Warn if no tests
		if result.TestCount == 0 {
			result.Warnings = append(result.Warnings, "No tests defined (consider adding [[tests.filtername]] sections)")
		}

		// Validate each test
		for filterName, testList := range tests.Tests {
			// Check that filter exists
			if _, ok := filter.Filters[filterName]; !ok {
				result.Errors = append(result.Errors,
					fmt.Sprintf("Tests defined for unknown filter '%s'", filterName))
				result.Valid = false
			}

			// Validate test structure
			for i, test := range testList {
				if test.Name == "" {
					result.Errors = append(result.Errors,
						fmt.Sprintf("[%s] test %d missing 'name' field", filterName, i))
					result.Valid = false
				}
				if test.Input == "" {
					result.Warnings = append(result.Warnings,
						fmt.Sprintf("[%s/%s] test has empty input", filterName, test.Name))
				}
			}
		}
	}

	return result
}

func printValidationResults(results []ValidationResult) {
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	totalValid := 0
	totalErrors := 0
	totalWarnings := 0

	for _, result := range results {
		// Print file name
		fileName := filepath.Base(result.FilePath)
		
		if result.Valid {
			totalValid++
			fmt.Printf("%s %s", green("✓"), fileName)
			if result.FilterName != "" {
				fmt.Printf(" [%s]", result.FilterName)
			}
			if result.TestCount > 0 {
				fmt.Printf(" (%d test%s)", result.TestCount, pluralize(result.TestCount))
			}
			fmt.Println()
		} else {
			fmt.Printf("%s %s", red("✗"), fileName)
			if result.FilterName != "" {
				fmt.Printf(" [%s]", result.FilterName)
			}
			fmt.Println()
		}

		// Print errors
		for _, err := range result.Errors {
			fmt.Printf("  %s %s\n", red("✗"), err)
			totalErrors++
		}

		// Print warnings
		for _, warning := range result.Warnings {
			fmt.Printf("  %s %s\n", yellow("⚠"), warning)
			totalWarnings++
		}

		if len(result.Errors) > 0 || len(result.Warnings) > 0 {
			fmt.Println()
		}
	}

	// Print summary
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	if totalErrors == 0 {
		fmt.Printf("%s\n", green(fmt.Sprintf("All %d filter(s) valid!", len(results))))
	} else {
		fmt.Printf("%s\n", red(fmt.Sprintf("%d filter(s) with errors", len(results)-totalValid)))
	}

	if totalWarnings > 0 {
		fmt.Printf("%s\n", yellow(fmt.Sprintf("%d warning(s)", totalWarnings)))
	}
}

func findFilterFile(name string) (string, error) {
	// Check builtin filters
	builtinPath := filepath.Join("internal/toml/builtin", name)
	if _, err := os.Stat(builtinPath); err == nil {
		return builtinPath, nil
	}

	// Check user config
	configPath := config.ConfigPath()
	filtersPath := filepath.Join(filepath.Dir(configPath), "filters", name)
	if _, err := os.Stat(filtersPath); err == nil {
		return filtersPath, nil
	}

	// Try as absolute path
	if _, err := os.Stat(name); err == nil {
		return name, nil
	}

	return "", fmt.Errorf("filter file not found: %s", name)
}

func findAllFilterFiles(includeUser bool) ([]string, error) {
	var files []string

	// Always include builtin filters
	builtinPath := "internal/toml/builtin"
	if _, err := os.Stat(builtinPath); err == nil {
		matches, err := filepath.Glob(filepath.Join(builtinPath, "*.toml"))
		if err != nil {
			return nil, err
		}
		files = append(files, matches...)
	}

	// Include user filters if requested
	if includeUser {
		configPath := config.ConfigPath()
		filtersPath := filepath.Join(filepath.Dir(configPath), "filters")
		if _, err := os.Stat(filtersPath); err == nil {
			matches, err := filepath.Glob(filepath.Join(filtersPath, "*.toml"))
			if err != nil {
				return nil, err
			}
			files = append(files, matches...)
		}
	}

	return files, nil
}

func pluralize(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}
