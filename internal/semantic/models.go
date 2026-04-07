package semantic

import "time"

// SemanticModel represents the configuration for semantic compression.
type SemanticModel struct {
	Name          string
	Model         string  // Ollama model name (e.g., "mistral", "neural-chat")
	Endpoint      string  // Ollama API endpoint
	Temperature   float32 // 0-1: lower = more deterministic
	TopK          int     // Top-K sampling
	TopP          float32 // Nucleus sampling parameter
	MaxTokens     int     // Max tokens in response
	Timeout       int     // Request timeout in seconds
	ContextWindow int     // Model's context window size
}

// SemanticAnalysis represents analysis of code semantics.
type SemanticAnalysis struct {
	ID             string
	Input          string
	InputTokens    int
	Summary        string
	SummaryTokens  int
	KeyPatterns    []string
	HasBoilerplate bool
	BoilerplateRatio float64 // 0-1: percentage of boilerplate
	Confidence     float64   // 0-1: confidence in analysis
	ExecTimeMs     int64
	Model          string
	CreatedAt      time.Time
}

// CodePattern represents identified code patterns.
type CodePattern struct {
	ID       string
	Name     string
	Regex    string
	Weight   float64 // 0-1: importance weight
	Examples []string
}

// CompressionStrategy represents a strategy for semantic compression.
type CompressionStrategy string

const (
	StrategyRemoveBoilerplate   CompressionStrategy = "remove_boilerplate"
	StrategyExtractKeyFunctions CompressionStrategy = "extract_key_functions"
	StrategyExtractComments     CompressionStrategy = "extract_comments"
	StrategySimplifyStructures  CompressionStrategy = "simplify_structures"
	StrategyIdentifyPatterns    CompressionStrategy = "identify_patterns"
)

// PromptTemplate represents a prompt template for semantic analysis.
type PromptTemplate struct {
	Name        string
	Description string
	Prompt      string
	Examples    []string
}

// DefaultPrompts defines default prompts for semantic analysis.
var DefaultPrompts = map[string]*PromptTemplate{
	"summarize": {
		Name:        "Summarize Code",
		Description: "Summarize the key functionality of code",
		Prompt: `Summarize the following code in 2-3 sentences focusing on its main purpose:

%s

Summary:`,
		Examples: []string{
			"# Function that validates email addresses and checks if domain exists",
			"# HTTP handler that authenticates requests and logs access",
		},
	},
	"extract_patterns": {
		Name:        "Extract Code Patterns",
		Description: "Identify recurring patterns and boilerplate",
		Prompt: `Identify recurring patterns and boilerplate code in:

%s

List each pattern found:`,
		Examples: []string{
			"1. Error checking pattern (if err != nil ...)",
			"2. Function parameter validation",
		},
	},
	"analyze_boilerplate": {
		Name:        "Analyze Boilerplate",
		Description: "Estimate percentage of boilerplate vs meaningful code",
		Prompt: `Estimate what percentage of this code is boilerplate vs meaningful logic (0-100):

%s

Boilerplate percentage: __% (explanation below)`,
		Examples: []string{
			"45% (mostly import statements, function signatures, and error handling patterns)",
			"60% (significant auto-generated code, repeated validation logic)",
		},
	},
	"extract_summary": {
		Name:        "Extract Summary",
		Description: "Extract a concise summary preserving only critical information",
		Prompt: `Extract ONLY the critical semantic information from this code, removing all boilerplate and comments:

%s

Critical code (preserving structure):`,
		Examples: []string{
			"// Critical business logic only, no error handling or validation",
		},
	},
}

// SemanticContext represents full semantic analysis results.
type SemanticContext struct {
	ID                string
	CodeInput         string
	InputTokens       int
	Summary           string
	KeyFunctions      []string
	KeyVariables      []string
	ApiCalls          []string
	Dependencies      []string
	BoilerplateRatio  float64
	CompressedOutput  string
	CompressionRatio  float64
	CriticalPatterns  []string
	Model             string
	Confidence        float64
	ProcessingTimeMs  int64
	CreatedAt         time.Time
}
