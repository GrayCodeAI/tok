package filter

import (
	"fmt"
	"strings"

	"github.com/BurntSushi/toml"
)

// TOMLPipelineConfig holds TOML-based pipeline configuration.
type TOMLPipelineConfig struct {
	Pipeline PipelineSection `toml:"pipeline"`
	Layers   LayersSection   `toml:"layers"`
	Safety   SafetySection   `toml:"safety"`
}

// PipelineSection holds pipeline configuration.
type PipelineSection struct {
	Mode        string `toml:"mode"`
	Budget      int    `toml:"budget"`
	QueryIntent string `toml:"query_intent"`
	LLMEnabled  bool   `toml:"llm_enabled"`
}

// LayersSection holds layer enable/disable configuration.
type LayersSection struct {
	Entropy        bool `toml:"entropy"`
	Perplexity     bool `toml:"perplexity"`
	GoalDriven     bool `toml:"goal_driven"`
	AST            bool `toml:"ast"`
	Contrastive    bool `toml:"contrastive"`
	Ngram          bool `toml:"ngram"`
	Evaluator      bool `toml:"evaluator"`
	Gist           bool `toml:"gist"`
	Hierarchical   bool `toml:"hierarchical"`
	Compaction     bool `toml:"compaction"`
	Attribution    bool `toml:"attribution"`
	H2O            bool `toml:"h2o"`
	AttentionSink  bool `toml:"attention_sink"`
	MetaToken      bool `toml:"meta_token"`
	SemanticChunk  bool `toml:"semantic_chunk"`
	SketchStore    bool `toml:"sketch_store"`
	LazyPruner     bool `toml:"lazy_pruner"`
	SemanticAnchor bool `toml:"semantic_anchor"`
	AgentMemory    bool `toml:"agent_memory"`
	TFIDF          bool `toml:"tfidf"`
	Symbolic       bool `toml:"symbolic"`
	PhraseGroup    bool `toml:"phrase_group"`
	Numerical      bool `toml:"numerical"`
	DynamicRatio   bool `toml:"dynamic_ratio"`
	TOON           bool `toml:"toon"`
	TDD            bool `toml:"tdd"`
}

// SafetySection holds safety configuration.
type SafetySection struct {
	CheckFilterSafety bool `toml:"check_filter_safety"`
	MaxFilterSize     int  `toml:"max_filter_size"`
	AllowRemote       bool `toml:"allow_remote"`
}

// LoadPipelineFromTOML loads pipeline configuration from TOML.
func LoadPipelineFromTOML(path string) (PipelineConfig, error) {
	var cfg TOMLPipelineConfig
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return PipelineConfig{}, err
	}
	return tomlToPipelineConfig(cfg), nil
}

func tomlToPipelineConfig(t TOMLPipelineConfig) PipelineConfig {
	pc := PipelineConfig{
		Budget:      t.Pipeline.Budget,
		QueryIntent: t.Pipeline.QueryIntent,
		LLMEnabled:  t.Pipeline.LLMEnabled,
	}
	switch t.Pipeline.Mode {
	case "minimal":
		pc.Mode = ModeMinimal
	case "aggressive":
		pc.Mode = ModeAggressive
	default:
		pc.Mode = ModeNone
	}
	l := t.Layers
	pc.EnableEntropy = l.Entropy
	pc.EnablePerplexity = l.Perplexity
	pc.EnableGoalDriven = l.GoalDriven
	pc.EnableAST = l.AST
	pc.EnableContrastive = l.Contrastive
	pc.NgramEnabled = l.Ngram
	pc.EnableEvaluator = l.Evaluator
	pc.EnableGist = l.Gist
	pc.EnableHierarchical = l.Hierarchical
	pc.EnableCompaction = l.Compaction
	pc.EnableAttribution = l.Attribution
	pc.EnableH2O = l.H2O
	pc.EnableAttentionSink = l.AttentionSink
	pc.EnableMetaToken = l.MetaToken
	pc.EnableSemanticChunk = l.SemanticChunk
	pc.EnableSketchStore = l.SketchStore
	pc.EnableLazyPruner = l.LazyPruner
	pc.EnableSemanticAnchor = l.SemanticAnchor
	pc.EnableAgentMemory = l.AgentMemory
	pc.EnableTFIDF = l.TFIDF
	pc.EnableSymbolicCompress = l.Symbolic
	pc.EnablePhraseGrouping = l.PhraseGroup
	pc.EnableNumericalQuant = l.Numerical
	pc.EnableDynamicRatio = l.DynamicRatio
	return pc
}

// TemplatePipe implements template pipe chains for filter output processing.
type TemplatePipe struct {
	operations []PipeOp
}

// PipeOp represents a single pipe operation.
type PipeOp struct {
	Type string
	Args []string
}

// NewTemplatePipe creates a new template pipe from a pipe chain string.
func NewTemplatePipe(chain string) *TemplatePipe {
	var ops []PipeOp
	parts := strings.Split(chain, "|")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		fields := strings.Fields(part)
		if len(fields) > 0 {
			ops = append(ops, PipeOp{Type: fields[0], Args: fields[1:]})
		}
	}
	return &TemplatePipe{operations: ops}
}

// Process applies the pipe chain to input.
func (tp *TemplatePipe) Process(input string) string {
	result := input
	for _, op := range tp.operations {
		switch op.Type {
		case "join":
			sep := "\n"
			if len(op.Args) > 0 {
				sep = op.Args[0]
			}
			result = strings.Join(strings.Fields(result), sep)
		case "truncate":
			n := 100
			if len(op.Args) > 0 {
				fmt.Sscanf(op.Args[0], "%d", &n)
			}
			if len(result) > n {
				result = result[:n] + "..."
			}
		case "lines":
			lines := strings.Split(result, "\n")
			if len(op.Args) > 0 {
				n := 0
				fmt.Sscanf(op.Args[0], "%d", &n)
				if n > 0 && len(lines) > n {
					lines = lines[:n]
				}
			}
			result = strings.Join(lines, "\n")
		case "keep":
			pattern := ""
			if len(op.Args) > 0 {
				pattern = op.Args[0]
			}
			var kept []string
			for _, line := range strings.Split(result, "\n") {
				if strings.Contains(line, pattern) {
					kept = append(kept, line)
				}
			}
			result = strings.Join(kept, "\n")
		case "where":
			pattern := ""
			if len(op.Args) > 0 {
				pattern = op.Args[0]
			}
			var kept []string
			for _, line := range strings.Split(result, "\n") {
				if !strings.Contains(line, pattern) {
					kept = append(kept, line)
				}
			}
			result = strings.Join(kept, "\n")
		case "each":
			if len(op.Args) > 1 {
				var transformed []string
				for _, line := range strings.Split(result, "\n") {
					transformed = append(transformed, strings.ReplaceAll(line, op.Args[0], op.Args[1]))
				}
				result = strings.Join(transformed, "\n")
			}
		}
	}
	return result
}

// JSONPathExtract extracts values from JSON using simple path notation.
func JSONPathExtract(jsonStr, path string) string {
	parts := strings.Split(path, ".")
	current := jsonStr
	for _, part := range parts {
		key := `"` + part + `":`
		idx := strings.Index(current, key)
		if idx == -1 {
			return ""
		}
		start := idx + len(key)
		if current[start] == '"' {
			start++
			end := strings.Index(current[start:], `"`)
			if end == -1 {
				return ""
			}
			current = current[start : start+end]
		} else {
			end := start
			for end < len(current) && current[end] != ',' && current[end] != '}' && current[end] != ']' {
				end++
			}
			current = strings.TrimSpace(current[start:end])
		}
	}
	return current
}
