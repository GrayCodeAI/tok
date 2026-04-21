package output

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	out "github.com/GrayCodeAI/tok/internal/output"

	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tok/internal/commands/registry"
	"github.com/GrayCodeAI/tok/internal/commands/shared"
	"github.com/GrayCodeAI/tok/internal/filter"
	"github.com/GrayCodeAI/tok/internal/tracking"
)

var (
	jsonDepth    int
	jsonKeysOnly bool
)

var jsonCmd = &cobra.Command{
	Use:   "json <file>",
	Short: "Show JSON structure without values",
	Long: `Show JSON structure/schema without actual values.

Useful for understanding large JSON responses without consuming tokens on values.

Examples:
  tok json response.json
  tok json response.json --depth 3
  tok json response.json --keys-only`,
	Args: cobra.ExactArgs(1),
	RunE: runJSON,
}

func init() {
	registry.Add(func() {
		jsonCmd.Flags().IntVarP(&jsonDepth, "depth", "d", 5, "Max depth to show")
		jsonCmd.Flags().BoolVar(&jsonKeysOnly, "keys-only", false, "Show only top-level keys")
		registry.Register(jsonCmd)
	})
}

func runJSON(cmd *cobra.Command, args []string) error {
	timer := tracking.Start()

	filePath := args[0]
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	var filtered string
	if jsonKeysOnly {
		filtered = generateKeysOnly(v)
	} else {
		filtered = generateSchema(v, 0, jsonDepth)
	}

	out.Global().Println(filtered)

	originalTokens := filter.EstimateTokens(string(data))
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track(filePath, "tok json", originalTokens, filteredTokens)

	shared.PrintTokenSavings(originalTokens, filteredTokens)

	return nil
}

func generateSchema(v any, depth, maxDepth int) string {
	if depth > maxDepth {
		return "..."
	}

	switch val := v.(type) {
	case nil:
		return "null"
	case bool:
		return "bool"
	case float64:
		return "number"
	case string:
		return "string"
	case []any:
		if len(val) == 0 {
			return "[]"
		}
		elemType := generateSchema(val[0], depth+1, maxDepth)
		return fmt.Sprintf("[%s, ...]", elemType)
	case map[string]any:
		if len(val) == 0 {
			return "{}"
		}
		var parts []string
		for k, v := range val {
			schema := generateSchema(v, depth+1, maxDepth)
			parts = append(parts, fmt.Sprintf("%s: %s", k, schema))
		}
		indent := strings.Repeat("  ", depth)
		return fmt.Sprintf("{\n%s  %s\n%s}", indent, strings.Join(parts, ",\n"+indent+"  "), indent)
	default:
		return fmt.Sprintf("%T", v)
	}
}

func generateKeysOnly(v any) string {
	switch val := v.(type) {
	case map[string]any:
		var keys []string
		for k := range val {
			keys = append(keys, k)
		}
		return fmt.Sprintf("Keys (%d): %s", len(keys), strings.Join(keys, ", "))
	case []any:
		return fmt.Sprintf("Array with %d element(s)", len(val))
	default:
		return generateSchema(v, 0, 0)
	}
}
