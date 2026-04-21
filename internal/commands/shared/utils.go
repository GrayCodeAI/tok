package shared

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	out "github.com/GrayCodeAI/tok/internal/output"
)

// OutputType represents the type of command output.
type OutputType int

const (
	OutputTypeTest OutputType = iota
	OutputTypeBuild
	OutputTypeLog
	OutputTypeList
	OutputTypeJSON
	OutputTypeGeneric
)

// ShortenPath shortens a file path for display.
func ShortenPath(path string) string {
	parts := strings.Split(path, string(filepath.Separator))
	if len(parts) <= 4 {
		return path
	}
	return filepath.Join(parts[0], "...", parts[len(parts)-2], parts[len(parts)-1])
}

// TruncateLine truncates a line to maxLen characters.
func TruncateLine(line string, maxLen int) string {
	if len(line) <= maxLen {
		return line
	}
	return line[:maxLen-3] + "..."
}

// Truncate truncates a string to maxLen characters including the "..." suffix.
func Truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return "..."
	}
	return s[:maxLen-3] + "..."
}

// TryJSONSchema generates a JSON schema from a JSON string.
func TryJSONSchema(jsonStr string, maxDepth int) string {
	var v any
	if err := json.Unmarshal([]byte(jsonStr), &v); err != nil {
		return ""
	}
	return generateSchemaFromJSON(v, 0, maxDepth)
}

func generateSchemaFromJSON(v any, depth, maxDepth int) string {
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
		elemType := generateSchemaFromJSON(val[0], depth+1, maxDepth)
		return fmt.Sprintf("[%s, ...]", elemType)
	case map[string]any:
		if len(val) == 0 {
			return "{}"
		}
		var parts []string
		for k, v := range val {
			schema := generateSchemaFromJSON(v, depth+1, maxDepth)
			parts = append(parts, fmt.Sprintf("%s: %s", k, schema))
		}
		indent := strings.Repeat("  ", depth)
		return fmt.Sprintf("{\n%s  %s\n%s}", indent, strings.Join(parts, ",\n"+indent+"  "), indent)
	default:
		return fmt.Sprintf("%T", v)
	}
}

// PrintTokenSavings prints token savings info to stderr when in verbose mode.
func PrintTokenSavings(originalTokens, filteredTokens int) {
	if !IsQuietMode() && IsVerbose() {
		out.Global().Errorf("Tokens saved: %d\n", originalTokens-filteredTokens)
	}
}

const maxArgLength = 4096
const maxArgsCount = 256

// SanitizeArgs validates and sanitizes command arguments.
// Returns error if arguments contain dangerous patterns.
func SanitizeArgs(args []string) error {
	if len(args) > maxArgsCount {
		return fmt.Errorf("too many arguments: %d (max %d)", len(args), maxArgsCount)
	}

	for i, arg := range args {
		if len(arg) > maxArgLength {
			return fmt.Errorf("argument %d exceeds max length %d", i, maxArgLength)
		}
		if strings.ContainsRune(arg, '\x00') {
			return fmt.Errorf("argument %d contains null byte", i)
		}
	}
	return nil
}
