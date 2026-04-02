package asteng

import (
	"strings"
)

type Signature struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Line    int    `json:"line"`
	Content string `json:"content"`
}

type ASTEngine struct {
	languagePatterns map[string][]SignaturePattern
}

type SignaturePattern struct {
	Type    string
	Pattern string
}

func NewASTEngine() *ASTEngine {
	return &ASTEngine{
		languagePatterns: map[string][]SignaturePattern{
			"go": {
				{Type: "function", Pattern: "func "},
				{Type: "type", Pattern: "type "},
				{Type: "interface", Pattern: "interface {"},
				{Type: "struct", Pattern: "struct {"},
				{Type: "import", Pattern: "import "},
			},
			"python": {
				{Type: "function", Pattern: "def "},
				{Type: "class", Pattern: "class "},
				{Type: "import", Pattern: "import "},
				{Type: "decorator", Pattern: "@"},
			},
			"javascript": {
				{Type: "function", Pattern: "function "},
				{Type: "arrow", Pattern: "=>"},
				{Type: "class", Pattern: "class "},
				{Type: "export", Pattern: "export "},
				{Type: "import", Pattern: "import "},
			},
			"typescript": {
				{Type: "function", Pattern: "function "},
				{Type: "interface", Pattern: "interface "},
				{Type: "type", Pattern: "type "},
				{Type: "class", Pattern: "class "},
				{Type: "export", Pattern: "export "},
				{Type: "import", Pattern: "import "},
			},
			"rust": {
				{Type: "function", Pattern: "fn "},
				{Type: "struct", Pattern: "struct "},
				{Type: "enum", Pattern: "enum "},
				{Type: "impl", Pattern: "impl "},
				{Type: "trait", Pattern: "trait "},
				{Type: "use", Pattern: "use "},
			},
			"java": {
				{Type: "class", Pattern: "class "},
				{Type: "interface", Pattern: "interface "},
				{Type: "method", Pattern: "public "},
				{Type: "import", Pattern: "import "},
			},
			"ruby": {
				{Type: "method", Pattern: "def "},
				{Type: "class", Pattern: "class "},
				{Type: "module", Pattern: "module "},
				{Type: "require", Pattern: "require "},
			},
			"c": {
				{Type: "function", Pattern: "("},
				{Type: "include", Pattern: "#include"},
				{Type: "define", Pattern: "#define"},
				{Type: "typedef", Pattern: "typedef"},
			},
			"cpp": {
				{Type: "class", Pattern: "class "},
				{Type: "function", Pattern: "("},
				{Type: "include", Pattern: "#include"},
				{Type: "namespace", Pattern: "namespace "},
				{Type: "template", Pattern: "template"},
			},
			"sql": {
				{Type: "select", Pattern: "SELECT"},
				{Type: "create", Pattern: "CREATE"},
				{Type: "insert", Pattern: "INSERT"},
				{Type: "update", Pattern: "UPDATE"},
			},
			"shell": {
				{Type: "function", Pattern: "function "},
				{Type: "alias", Pattern: "alias "},
				{Type: "export", Pattern: "export "},
			},
			"markdown": {
				{Type: "heading", Pattern: "#"},
			},
			"json": {
				{Type: "key", Pattern: "\""},
			},
			"yaml": {
				{Type: "key", Pattern: ":"},
			},
			"terraform": {
				{Type: "resource", Pattern: "resource"},
				{Type: "data", Pattern: "data"},
				{Type: "variable", Pattern: "variable"},
				{Type: "module", Pattern: "module"},
			},
		},
	}
}

func (e *ASTEngine) ExtractSignatures(content, language string) []Signature {
	patterns, ok := e.languagePatterns[language]
	if !ok {
		patterns = e.languagePatterns["go"]
	}

	var signatures []Signature
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		for _, p := range patterns {
			if strings.Contains(trimmed, p.Pattern) {
				signatures = append(signatures, Signature{
					Name:    extractName(trimmed),
					Type:    p.Type,
					Line:    i + 1,
					Content: trimmed,
				})
				break
			}
		}
	}
	return signatures
}

func (e *ASTEngine) DetectLanguage(content string) string {
	scores := make(map[string]int)
	for lang := range e.languagePatterns {
		scores[lang] = 0
	}

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		for lang, patterns := range e.languagePatterns {
			for _, p := range patterns {
				if strings.Contains(trimmed, p.Pattern) {
					scores[lang]++
				}
			}
		}
	}

	bestLang := "unknown"
	bestScore := 0
	for lang, score := range scores {
		if score > bestScore {
			bestScore = score
			bestLang = lang
		}
	}
	return bestLang
}

func (e *ASTEngine) FileMap(content, language string) string {
	signatures := e.ExtractSignatures(content, language)
	var sb strings.Builder
	sb.WriteString("File Map (" + language + "):\n")
	for _, sig := range signatures {
		sb.WriteString("  " + sig.Type + ": " + sig.Name + " (line " + string(rune(sig.Line)) + ")\n")
	}
	return sb.String()
}

func extractName(line string) string {
	parts := strings.Fields(line)
	if len(parts) > 1 {
		return parts[1]
	}
	return line
}
