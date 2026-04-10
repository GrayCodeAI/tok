package filter

import (
	"fmt"
	"strings"
)

// PhotonFilter detects and compresses base64-encoded images.
// Inspired by claw-compactor's Photon stage.
type PhotonFilter struct {
	threshold1MB int
	threshold2MB int
}

// NewPhotonFilter creates a new Photon filter with default configuration.
func NewPhotonFilter() *PhotonFilter {
	return &PhotonFilter{
		threshold1MB: 1 * 1024 * 1024,
		threshold2MB: 2 * 1024 * 1024,
	}
}

// Name returns the filter name for the pipeline.
func (f *PhotonFilter) Name() string { return "0_photon" }

// Apply implements the Filter interface for Photon image compression.
func (f *PhotonFilter) Apply(input string, mode Mode) (string, int) {
	return f.processWithMode(input, mode)
}

// processWithMode detects and compresses base64 images in content.
func (f *PhotonFilter) processWithMode(content string, mode Mode) (string, int) {
	if mode == ModeNone {
		return content, 0
	}

	// Find base64 data URIs
	dataURIPrefixes := []string{
		"data:image/png;base64,",
		"data:image/jpeg;base64,",
		"data:image/gif;base64,",
		"data:image/webp;base64,",
		"data:image/svg+xml;base64,",
	}

	result := content
	saved := 0

	for _, prefix := range dataURIPrefixes {
		for {
			idx := strings.Index(result, prefix)
			if idx == -1 {
				break
			}

			// Find end of base64 data
			start := idx + len(prefix)
			end := start
			for end < len(result) && result[end] != '"' && result[end] != '\'' && result[end] != ')' && result[end] != ' ' && result[end] != ',' {
				end++
			}

			if end > start {
				base64Data := result[start:end]
				decodedLen := len(base64Data) * 3 / 4
				format := strings.TrimPrefix(prefix, "data:image/")
				format = strings.TrimSuffix(format, ";base64,")

				replacement := fmt.Sprintf("[image:%s size=%d]", format, decodedLen)
				result = result[:idx] + replacement + result[end:]
				saved += decodedLen/4 - len(replacement)
			} else {
				break
			}
		}
	}

	return result, saved
}

// LogCrunch folds repeated log lines with occurrence counts.
// Inspired by claw-compactor's LogCrunch stage.
type LogCrunch struct {
	config LogCrunchConfig
}

// LogCrunchConfig holds configuration for LogCrunch.
type LogCrunchConfig struct {
	Enabled        bool
	MinRepetitions int
	AlwaysPreserve []string
}

// DefaultLogCrunchConfig returns default LogCrunch configuration.
func DefaultLogCrunchConfig() LogCrunchConfig {
	return LogCrunchConfig{
		Enabled:        true,
		MinRepetitions: 3,
		AlwaysPreserve: []string{"ERROR", "FATAL", "CRITICAL", "PANIC"},
	}
}

// NewLogCrunch creates a new LogCrunch filter.
func NewLogCrunch(cfg LogCrunchConfig) *LogCrunch {
	return &LogCrunch{config: cfg}
}

// Process folds repeated log lines.
func (lc *LogCrunch) Process(content string) (string, int) {
	if !lc.config.Enabled {
		return content, 0
	}

	lines := strings.Split(content, "\n")
	if len(lines) < lc.config.MinRepetitions {
		return content, 0
	}

	var result []string
	i := 0
	for i < len(lines) {
		line := lines[i]
		if lc.shouldPreserve(line) {
			result = append(result, line)
			i++
			continue
		}

		count := 1
		for j := i + 1; j < len(lines); j++ {
			if lines[j] == line {
				count++
			} else {
				break
			}
		}

		if count >= lc.config.MinRepetitions {
			result = append(result, fmt.Sprintf("%s (repeated %d times)", line, count))
		} else {
			for k := 0; k < count; k++ {
				result = append(result, line)
			}
		}
		i += count
	}

	return strings.Join(result, "\n"), len(lines) - len(result)
}

func (lc *LogCrunch) shouldPreserve(line string) bool {
	for _, keyword := range lc.config.AlwaysPreserve {
		if strings.Contains(strings.ToUpper(line), keyword) {
			return true
		}
	}
	return false
}

// DiffCrunch folds unchanged context lines in unified diffs.
// Inspired by claw-compactor's DiffCrunch stage.
type DiffCrunch struct {
	config DiffCrunchConfig
}

// DiffCrunchConfig holds configuration for DiffCrunch.
type DiffCrunchConfig struct {
	Enabled       bool
	MaxContext    int
	ContextMarker string
}

// DefaultDiffCrunchConfig returns default DiffCrunch configuration.
func DefaultDiffCrunchConfig() DiffCrunchConfig {
	return DiffCrunchConfig{
		Enabled:       true,
		MaxContext:    3,
		ContextMarker: "... (%d unchanged lines folded)",
	}
}

// NewDiffCrunch creates a new DiffCrunch filter.
func NewDiffCrunch(cfg DiffCrunchConfig) *DiffCrunch {
	return &DiffCrunch{config: cfg}
}

// Process folds unchanged context lines in diffs.
func (dc *DiffCrunch) Process(content string) (string, int) {
	if !dc.config.Enabled {
		return content, 0
	}

	lines := strings.Split(content, "\n")
	var result []string
	contextCount := 0

	for _, line := range lines {
		if strings.HasPrefix(line, " ") {
			contextCount++
			if contextCount <= dc.config.MaxContext {
				result = append(result, line)
			}
		} else {
			if contextCount > dc.config.MaxContext {
				folded := contextCount - dc.config.MaxContext
				result = append(result, fmt.Sprintf(dc.config.ContextMarker, folded))
			}
			result = append(result, line)
			contextCount = 0
		}
	}

	if contextCount > dc.config.MaxContext {
		folded := contextCount - dc.config.MaxContext
		result = append(result, fmt.Sprintf(dc.config.ContextMarker, folded))
	}

	return strings.Join(result, "\n"), len(lines) - len(result)
}

// StructuralCollapse merges import blocks and collapses repeated patterns.
// Inspired by claw-compactor's StructuralCollapse stage.
type StructuralCollapse struct {
	config StructuralCollapseConfig
}

// StructuralCollapseConfig holds configuration.
type StructuralCollapseConfig struct {
	Enabled         bool
	CollapseImports bool
	CollapseAsserts bool
	MaxRepeated     int
}

// DefaultStructuralCollapseConfig returns default configuration.
func DefaultStructuralCollapseConfig() StructuralCollapseConfig {
	return StructuralCollapseConfig{
		Enabled:         true,
		CollapseImports: true,
		CollapseAsserts: true,
		MaxRepeated:     3,
	}
}

// NewStructuralCollapse creates a new StructuralCollapse filter.
func NewStructuralCollapse(cfg StructuralCollapseConfig) *StructuralCollapse {
	return &StructuralCollapse{config: cfg}
}

// Process collapses structural patterns.
func (sc *StructuralCollapse) Process(content string) (string, int) {
	if !sc.config.Enabled {
		return content, 0
	}

	lines := strings.Split(content, "\n")
	var result []string
	i := 0

	for i < len(lines) {
		line := lines[i]

		// Collapse import blocks
		if sc.config.CollapseImports && (strings.HasPrefix(line, "import ") || strings.HasPrefix(line, "#include") || strings.HasPrefix(line, "require ")) {
			imports := []string{line}
			for j := i + 1; j < len(lines); j++ {
				l := strings.TrimSpace(lines[j])
				if strings.HasPrefix(l, "\"") || strings.HasPrefix(l, "<") || strings.HasPrefix(l, "'") || l == "" {
					if l != "" {
						imports = append(imports, lines[j])
					}
					i = j + 1
				} else {
					break
				}
			}
			if len(imports) > sc.config.MaxRepeated {
				result = append(result, imports[0])
				result = append(result, fmt.Sprintf("... (%d more imports)", len(imports)-1))
				if imports[len(imports)-1] != "" {
					result = append(result, imports[len(imports)-1])
				}
			} else {
				result = append(result, imports...)
			}
			continue
		}

		result = append(result, line)
		i++
	}

	return strings.Join(result, "\n"), len(lines) - len(result)
}

// DictionaryEncoding implements auto-learned codebook substitution.
// Inspired by claw-compactor's dictionary encoding.
type DictionaryEncoding struct {
	codebook map[string]string
	reverse  map[string]string
	counter  int
}

// NewDictionaryEncoding creates a new dictionary encoder.
func NewDictionaryEncoding() *DictionaryEncoding {
	return &DictionaryEncoding{
		codebook: make(map[string]string),
		reverse:  make(map[string]string),
	}
}

// Encode replaces frequent patterns with dictionary references.
func (de *DictionaryEncoding) Encode(content string) (string, int) {
	// Find frequent patterns
	freq := make(map[string]int)
	words := strings.Fields(content)
	for _, w := range words {
		if len(w) > 3 {
			freq[w]++
		}
	}

	// Build codebook for frequent patterns
	for word, count := range freq {
		if count >= 3 && len(word) > 5 {
			if _, ok := de.codebook[word]; !ok {
				de.counter++
				symbol := fmt.Sprintf("$%02X", de.counter)
				de.codebook[word] = symbol
				de.reverse[symbol] = word
			}
		}
	}

	// Apply substitutions
	result := content
	saved := 0
	for word, symbol := range de.codebook {
		count := strings.Count(result, word)
		if count > 0 {
			result = strings.ReplaceAll(result, word, symbol)
			saved += count * (len(word) - len(symbol))
		}
	}

	return result, saved
}

// Decode restores original content from dictionary references.
func (de *DictionaryEncoding) Decode(content string) string {
	result := content
	for symbol, word := range de.reverse {
		result = strings.ReplaceAll(result, symbol, word)
	}
	return result
}
