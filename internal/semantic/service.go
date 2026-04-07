package semantic

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"
)

// SemanticService provides semantic code analysis using local LLMs.
type SemanticService struct {
	model  *SemanticModel
	logger *slog.Logger
	// llmClient would go here for Ollama integration
}

// NewSemanticService creates a new semantic service.
func NewSemanticService(model *SemanticModel, logger *slog.Logger) *SemanticService {
	if logger == nil {
		logger = slog.Default()
	}
	return &SemanticService{
		model:  model,
		logger: logger,
	}
}

// Summarize analyzes code and returns a semantic summary.
func (ss *SemanticService) Summarize(ctx context.Context, code string) (*SemanticAnalysis, error) {
	if code == "" {
		return nil, fmt.Errorf("empty code input")
	}

	start := time.Now()

	// For now, return a placeholder analysis
	// In production, this would call Ollama API
	analysis := &SemanticAnalysis{
		ID:               generateID(),
		Input:            code,
		InputTokens:      estimateTokens(code),
		Summary:          "Code analysis pending LLM integration",
		SummaryTokens:    10,
		HasBoilerplate:   detectBoilerplate(code) > 0.3,
		BoilerplateRatio: detectBoilerplate(code),
		Confidence:       0.8,
		ExecTimeMs:       int64(time.Since(start).Milliseconds()),
		Model:            ss.model.Model,
		CreatedAt:        time.Now(),
	}

	ss.logger.Debug("semantic analysis completed",
		"input_tokens", analysis.InputTokens,
		"summary_tokens", analysis.SummaryTokens,
		"boilerplate_ratio", analysis.BoilerplateRatio,
		"exec_time_ms", analysis.ExecTimeMs,
	)

	return analysis, nil
}

// ExtractPatterns identifies code patterns and extracts key information.
func (ss *SemanticService) ExtractPatterns(ctx context.Context, code string) (*SemanticContext, error) {
	if code == "" {
		return nil, fmt.Errorf("empty code input")
	}

	start := time.Now()

	boilerplateRatio := detectBoilerplate(code)
	compressionRatio := 1.0 - boilerplateRatio

	context := &SemanticContext{
		ID:               generateID(),
		CodeInput:        code,
		InputTokens:      estimateTokens(code),
		Summary:          extractSummary(code),
		KeyFunctions:     extractFunctions(code),
		KeyVariables:     extractVariables(code),
		ApiCalls:         extractApiCalls(code),
		Dependencies:     extractDependencies(code),
		BoilerplateRatio: boilerplateRatio,
		CompressedOutput: compressCode(code, boilerplateRatio),
		CompressionRatio: compressionRatio,
		CriticalPatterns: extractCriticalPatterns(code),
		Model:            ss.model.Model,
		Confidence:       0.85,
		ProcessingTimeMs: int64(time.Since(start).Milliseconds()),
		CreatedAt:        time.Now(),
	}

	ss.logger.Info("pattern extraction completed",
		"functions_found", len(context.KeyFunctions),
		"patterns_found", len(context.CriticalPatterns),
		"compression_ratio", context.CompressionRatio,
	)

	return context, nil
}

// CompressCodeSemantically compresses code using semantic understanding.
func (ss *SemanticService) CompressCodeSemantically(ctx context.Context, code string, targetRatio float64) (string, error) {
	if code == "" {
		return "", fmt.Errorf("empty code input")
	}

	boilerplateRatio := detectBoilerplate(code)

	if boilerplateRatio < targetRatio {
		// Can't compress beyond boilerplate ratio
		ss.logger.Warn("target compression impossible",
			"target_ratio", targetRatio,
			"boilerplate_ratio", boilerplateRatio,
		)
		return code, nil
	}

	compressed := compressCode(code, targetRatio)
	return compressed, nil
}

// IdentifyBoilerplate identifies boilerplate code sections.
func (ss *SemanticService) IdentifyBoilerplate(code string) map[string]float64 {
	patterns := map[string]float64{
		"import_statements":  detectImports(code),
		"error_handling":     detectErrorHandling(code),
		"type_definitions":   detectTypes(code),
		"validation_logic":   detectValidation(code),
		"logging_statements": detectLogging(code),
	}
	return patterns
}

// Helper functions

func detectBoilerplate(code string) float64 {
	totalLines := strings.Count(code, "\n") + 1
	if totalLines == 0 {
		return 0
	}

	boilerplateLines := 0.0

	boilerplateLines += detectImports(code) * float64(totalLines)
	boilerplateLines += detectErrorHandling(code) * float64(totalLines)
	boilerplateLines += detectTypes(code) * float64(totalLines)
	boilerplateLines += detectValidation(code) * float64(totalLines)

	return boilerplateLines / float64(totalLines)
}

func detectImports(code string) float64 {
	importCount := strings.Count(code, "import ") + strings.Count(code, "require(")
	totalLines := strings.Count(code, "\n") + 1
	if totalLines == 0 {
		return 0
	}
	return float64(importCount) / float64(totalLines) * 0.5
}

func detectErrorHandling(code string) float64 {
	errorCount := strings.Count(code, "error") + strings.Count(code, "Error") + strings.Count(code, "try") + strings.Count(code, "catch")
	totalLines := strings.Count(code, "\n") + 1
	if totalLines == 0 {
		return 0
	}
	return float64(errorCount) / float64(totalLines) * 0.3
}

func detectTypes(code string) float64 {
	typeCount := strings.Count(code, "type ") + strings.Count(code, "interface ") + strings.Count(code, "struct ")
	totalLines := strings.Count(code, "\n") + 1
	if totalLines == 0 {
		return 0
	}
	return float64(typeCount) / float64(totalLines) * 0.2
}

func detectValidation(code string) float64 {
	validationKeywords := []string{"validate", "check", "assert", "if ", "len(", "nil"}
	count := 0
	for _, kw := range validationKeywords {
		count += strings.Count(code, kw)
	}
	totalLines := strings.Count(code, "\n") + 1
	if totalLines == 0 {
		return 0
	}
	return float64(count) / float64(totalLines) * 0.15
}

func detectLogging(code string) float64 {
	logCount := strings.Count(code, "log") + strings.Count(code, "print") + strings.Count(code, "fmt.Print")
	totalLines := strings.Count(code, "\n") + 1
	if totalLines == 0 {
		return 0
	}
	return float64(logCount) / float64(totalLines) * 0.1
}

func extractSummary(code string) string {
	// Extract first meaningful comment or function signature
	lines := strings.Split(code, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "\"\"\"") {
			return trimmed
		}
		if strings.Contains(trimmed, "func ") || strings.Contains(trimmed, "def ") {
			return trimmed
		}
	}
	return "Code analysis"
}

func extractFunctions(code string) []string {
	funcRegex := regexp.MustCompile(`(?:func|def|function)\s+(\w+)`)
	matches := funcRegex.FindAllStringSubmatch(code, -1)

	var functions []string
	for _, match := range matches {
		if len(match) > 1 {
			functions = append(functions, match[1])
		}
	}

	return functions
}

func extractVariables(code string) []string {
	varRegex := regexp.MustCompile(`(?:var|let|const)\s+(\w+)`)
	matches := varRegex.FindAllStringSubmatch(code, -1)

	var variables []string
	for _, match := range matches {
		if len(match) > 1 {
			variables = append(variables, match[1])
		}
	}

	return variables
}

func extractApiCalls(code string) []string {
	apiRegex := regexp.MustCompile(`(?:http\.|request\.|fetch\(|urllib|requests\.)`)
	if apiRegex.MatchString(code) {
		return []string{"external_api_calls"}
	}
	return []string{}
}

func extractDependencies(code string) []string {
	depRegex := regexp.MustCompile(`(?:import|require|from)\s+"?([^"\s]+)"?`)
	matches := depRegex.FindAllStringSubmatch(code, -1)

	var deps []string
	for _, match := range matches {
		if len(match) > 1 {
			deps = append(deps, match[1])
		}
	}

	return deps
}

func extractCriticalPatterns(code string) []string {
	patterns := []string{}

	if strings.Contains(code, "SELECT") || strings.Contains(code, "INSERT") {
		patterns = append(patterns, "database_operations")
	}
	if strings.Contains(code, "http") || strings.Contains(code, "request") {
		patterns = append(patterns, "http_operations")
	}
	if strings.Contains(code, "for ") || strings.Contains(code, "while ") {
		patterns = append(patterns, "loops")
	}
	if strings.Contains(code, "if ") || strings.Contains(code, "switch ") {
		patterns = append(patterns, "conditionals")
	}

	return patterns
}

func compressCode(code string, targetRatio float64) string {
	// Remove boilerplate lines
	lines := strings.Split(code, "\n")
	var compressed []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Keep non-boilerplate lines
		if !strings.HasPrefix(trimmed, "import ") &&
			!strings.HasPrefix(trimmed, "require(") &&
			!strings.HasPrefix(trimmed, "//") &&
			trimmed != "" {
			compressed = append(compressed, line)
		}
	}

	return strings.Join(compressed, "\n")
}

func estimateTokens(code string) int {
	// Rough estimate: ~4 characters per token
	return len(code) / 4
}

func generateID() string {
	return fmt.Sprintf("sem_%d", time.Now().UnixNano())
}
