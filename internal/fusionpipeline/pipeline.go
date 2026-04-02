package fusionpipeline

import (
	"strings"
)

type FusionStage interface {
	Name() string
	ShouldApply(input string) bool
	Apply(input string) (string, int)
}

type FusionResult struct {
	OriginalTokens   int `json:"original_tokens"`
	CompressedTokens int `json:"compressed_tokens"`
	Savings          int `json:"savings"`
	StagesRun        int `json:"stages_run"`
	StagesSkipped    int `json:"stages_skipped"`
}

type FusionEngine struct {
	stages []FusionStage
}

func NewFusionEngine() *FusionEngine {
	e := &FusionEngine{}
	e.stages = []FusionStage{
		NewQuantumLockStage(),
		NewCortexStage(),
		NewPhotonStage(),
		NewRLEStage(),
		NewSemanticDedupStage(),
		NewIonizerStage(),
		NewLogCrunchStage(),
		NewSearchCrunchStage(),
		NewDiffCrunchStage(),
		NewStructuralCollapseStage(),
		NewNeurosyntaxStage(),
		NewNexusStage(),
		NewTokenOptStage(),
		NewAbbrevStage(),
	}
	return e
}

func (e *FusionEngine) Compress(input string) (string, *FusionResult) {
	result := &FusionResult{
		OriginalTokens: len(input) / 4,
	}
	current := input
	stagesRun := 0
	stagesSkipped := 0

	for _, stage := range e.stages {
		if stage.ShouldApply(current) {
			compressed, _ := stage.Apply(current)
			if len(compressed) < len(current) {
				current = compressed
			}
			stagesRun++
		} else {
			stagesSkipped++
		}
	}

	result.CompressedTokens = len(current) / 4
	result.Savings = result.OriginalTokens - result.CompressedTokens
	result.StagesRun = stagesRun
	result.StagesSkipped = stagesSkipped
	return current, result
}

type QuantumLockStage struct{}

func NewQuantumLockStage() *QuantumLockStage { return &QuantumLockStage{} }

func (s *QuantumLockStage) Name() string { return "quantum_lock" }

func (s *QuantumLockStage) ShouldApply(input string) bool {
	return strings.Contains(input, "system") || strings.Contains(input, "<system>")
}

func (s *QuantumLockStage) Apply(input string) (string, int) {
	lines := strings.Split(input, "\n")
	var filtered []string
	seen := make(map[string]bool)
	for _, line := range lines {
		key := strings.TrimSpace(strings.ToLower(line))
		if key == "" || seen[key] {
			continue
		}
		seen[key] = true
		filtered = append(filtered, line)
	}
	output := strings.Join(filtered, "\n")
	return output, len(input) - len(output)
}

type CortexStage struct{}

func NewCortexStage() *CortexStage { return &CortexStage{} }

func (s *CortexStage) Name() string { return "cortex" }

func (s *CortexStage) ShouldApply(input string) bool {
	return len(input) > 100
}

func (s *CortexStage) Apply(input string) (string, int) {
	lines := strings.Split(input, "\n")
	var filtered []string
	for _, line := range lines {
		if len(strings.TrimSpace(line)) > 0 {
			filtered = append(filtered, line)
		}
	}
	output := strings.Join(filtered, "\n")
	return output, len(input) - len(output)
}

type PhotonStage struct{}

func NewPhotonStage() *PhotonStage { return &PhotonStage{} }

func (s *PhotonStage) Name() string { return "photon" }

func (s *PhotonStage) ShouldApply(input string) bool {
	return strings.Contains(input, "base64") || strings.Contains(input, "data:image")
}

func (s *PhotonStage) Apply(input string) (string, int) {
	lines := strings.Split(input, "\n")
	var filtered []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if len(trimmed) > 200 && isBase64(trimmed) {
			filtered = append(filtered, "[base64_data: "+trimmed[:20]+"...]")
			continue
		}
		filtered = append(filtered, line)
	}
	output := strings.Join(filtered, "\n")
	return output, len(input) - len(output)
}

func isBase64(s string) bool {
	for _, c := range s {
		if !((c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '+' || c == '/' || c == '=') {
			return false
		}
	}
	return true
}

type RLEStage struct{}

func NewRLEStage() *RLEStage { return &RLEStage{} }

func (s *RLEStage) Name() string { return "rle" }

func (s *RLEStage) ShouldApply(input string) bool {
	return strings.Contains(input, ".") || strings.Contains(input, "/")
}

func (s *RLEStage) Apply(input string) (string, int) {
	lines := strings.Split(input, "\n")
	var filtered []string
	seen := make(map[string]int)
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if len(trimmed) > 0 {
			count := seen[trimmed]
			if count >= 3 {
				if count == 3 {
					filtered = append(filtered, "[repeated "+trimmed[:min(30, len(trimmed))]+"]")
				}
				continue
			}
			seen[trimmed] = count + 1
			filtered = append(filtered, line)
		}
	}
	output := strings.Join(filtered, "\n")
	return output, len(input) - len(output)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

type SemanticDedupStage struct{}

func NewSemanticDedupStage() *SemanticDedupStage { return &SemanticDedupStage{} }

func (s *SemanticDedupStage) Name() string { return "semantic_dedup" }

func (s *SemanticDedupStage) ShouldApply(input string) bool {
	return len(strings.Split(input, "\n")) > 10
}

func (s *SemanticDedupStage) Apply(input string) (string, int) {
	lines := strings.Split(input, "\n")
	var filtered []string
	seenHashes := make(map[uint64]bool)
	for _, line := range lines {
		h := fnvHash(strings.TrimSpace(strings.ToLower(line)))
		if !seenHashes[h] && len(strings.TrimSpace(line)) > 0 {
			seenHashes[h] = true
			filtered = append(filtered, line)
		}
	}
	output := strings.Join(filtered, "\n")
	return output, len(input) - len(output)
}

func fnvHash(s string) uint64 {
	h := uint64(14695981039346656037)
	for _, c := range s {
		h ^= uint64(c)
		h *= 1099511628211
	}
	return h
}

type IonizerStage struct{}

func NewIonizerStage() *IonizerStage { return &IonizerStage{} }

func (s *IonizerStage) Name() string { return "ionizer" }

func (s *IonizerStage) ShouldApply(input string) bool {
	return strings.Count(input, "[") > 5 || strings.Count(input, "{") > 5
}

func (s *IonizerStage) Apply(input string) (string, int) {
	lines := strings.Split(input, "\n")
	var filtered []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if len(trimmed) > 200 && (strings.Contains(trimmed, "[") || strings.Contains(trimmed, "{")) {
			filtered = append(filtered, trimmed[:min(200, len(trimmed))]+"...")
			continue
		}
		filtered = append(filtered, line)
	}
	output := strings.Join(filtered, "\n")
	return output, len(input) - len(output)
}

type LogCrunchStage struct{}

func NewLogCrunchStage() *LogCrunchStage { return &LogCrunchStage{} }

func (s *LogCrunchStage) Name() string { return "log_crunch" }

func (s *LogCrunchStage) ShouldApply(input string) bool {
	logMarkers := []string{"INFO:", "DEBUG:", "WARN:", "ERROR:", "TRACE:", "[INFO]", "[DEBUG]"}
	for _, m := range logMarkers {
		if strings.Contains(input, m) {
			return true
		}
	}
	return false
}

func (s *LogCrunchStage) Apply(input string) (string, int) {
	lines := strings.Split(input, "\n")
	var filtered []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, "DEBUG") || strings.Contains(trimmed, "TRACE") {
			continue
		}
		filtered = append(filtered, line)
	}
	output := strings.Join(filtered, "\n")
	return output, len(input) - len(output)
}

type SearchCrunchStage struct{}

func NewSearchCrunchStage() *SearchCrunchStage { return &SearchCrunchStage{} }

func (s *SearchCrunchStage) Name() string { return "search_crunch" }

func (s *SearchCrunchStage) ShouldApply(input string) bool {
	return strings.Contains(input, "grep") || strings.Contains(input, "find") || strings.Contains(input, "rg")
}

func (s *SearchCrunchStage) Apply(input string) (string, int) {
	lines := strings.Split(input, "\n")
	if len(lines) > 50 {
		head := lines[:5]
		tail := lines[len(lines)-5:]
		omitted := len(lines) - 10
		result := append(head, "..."+strings.Join([]string{""}, "")+""+string(rune(omitted))+" lines omitted...")
		result = append(result, tail...)
		output := strings.Join(result, "\n")
		return output, len(input) - len(output)
	}
	return input, 0
}

type DiffCrunchStage struct{}

func NewDiffCrunchStage() *DiffCrunchStage { return &DiffCrunchStage{} }

func (s *DiffCrunchStage) Name() string { return "diff_crunch" }

func (s *DiffCrunchStage) ShouldApply(input string) bool {
	return strings.Contains(input, "diff") || strings.Contains(input, "+++")
}

func (s *DiffCrunchStage) Apply(input string) (string, int) {
	lines := strings.Split(input, "\n")
	var filtered []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "@@") || strings.HasPrefix(trimmed, "+++") || strings.HasPrefix(trimmed, "---") {
			filtered = append(filtered, line)
		} else if strings.HasPrefix(trimmed, "+") || strings.HasPrefix(trimmed, "-") {
			filtered = append(filtered, line)
		}
	}
	output := strings.Join(filtered, "\n")
	return output, len(input) - len(output)
}

type StructuralCollapseStage struct{}

func NewStructuralCollapseStage() *StructuralCollapseStage { return &StructuralCollapseStage{} }

func (s *StructuralCollapseStage) Name() string { return "structural_collapse" }

func (s *StructuralCollapseStage) ShouldApply(input string) bool {
	return strings.Count(input, "{") > 3
}

func (s *StructuralCollapseStage) Apply(input string) (string, int) {
	lines := strings.Split(input, "\n")
	var filtered []string
	depth := 0
	for _, line := range lines {
		depth += strings.Count(line, "{") - strings.Count(line, "}")
		if depth > 2 && strings.TrimSpace(line) != "" {
			continue
		}
		filtered = append(filtered, line)
	}
	output := strings.Join(filtered, "\n")
	return output, len(input) - len(output)
}

type NeurosyntaxStage struct{}

func NewNeurosyntaxStage() *NeurosyntaxStage { return &NeurosyntaxStage{} }

func (s *NeurosyntaxStage) Name() string { return "neurosyntax" }

func (s *NeurosyntaxStage) ShouldApply(input string) bool {
	langMarkers := []string{"func ", "def ", "class ", "import ", "package ", "public class"}
	for _, m := range langMarkers {
		if strings.Contains(input, m) {
			return true
		}
	}
	return false
}

func (s *NeurosyntaxStage) Apply(input string) (string, int) {
	lines := strings.Split(input, "\n")
	var filtered []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "/*") || strings.HasPrefix(trimmed, "*") {
			continue
		}
		filtered = append(filtered, line)
	}
	output := strings.Join(filtered, "\n")
	return output, len(input) - len(output)
}

type NexusStage struct{}

func NewNexusStage() *NexusStage { return &NexusStage{} }

func (s *NexusStage) Name() string { return "nexus" }

func (s *NexusStage) ShouldApply(input string) bool {
	return len(input) > 5000
}

func (s *NexusStage) Apply(input string) (string, int) {
	words := strings.Fields(input)
	if len(words) > 200 {
		half := len(words) / 2
		output := strings.Join(words[:half], " ")
		output += "\n...[truncated]"
		return output, len(input) - len(output)
	}
	return input, 0
}

type TokenOptStage struct{}

func NewTokenOptStage() *TokenOptStage { return &TokenOptStage{} }

func (s *TokenOptStage) Name() string { return "token_opt" }

func (s *TokenOptStage) ShouldApply(input string) bool {
	return strings.ContainsAny(input, "\t ") || strings.Contains(input, "\n\n")
}

func (s *TokenOptStage) Apply(input string) (string, int) {
	output := strings.ReplaceAll(input, "\t", "  ")
	for strings.Contains(output, "    ") {
		output = strings.ReplaceAll(output, "    ", "  ")
	}
	for strings.Contains(output, "\n\n\n") {
		output = strings.ReplaceAll(output, "\n\n\n", "\n\n")
	}
	return output, len(input) - len(output)
}

type AbbrevStage struct{}

func NewAbbrevStage() *AbbrevStage { return &AbbrevStage{} }

func (s *AbbrevStage) Name() string { return "abbrev" }

func (s *AbbrevStage) ShouldApply(input string) bool {
	return len(input) > 1000
}

func (s *AbbrevStage) Apply(input string) (string, int) {
	abbrevs := map[string]string{
		"function": "fn", "variable": "var", "constant": "const",
		"import": "imp", "export": "exp", "return": "ret",
		"parameter": "param", "argument": "arg", "interface": "iface",
		"implementation": "impl", "configuration": "config",
	}
	output := input
	for full, abbr := range abbrevs {
		output = strings.ReplaceAll(output, " "+full+" ", " "+abbr+" ")
	}
	return output, len(input) - len(output)
}
