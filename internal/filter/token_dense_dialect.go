package filter

import (
	"fmt"
	"strconv"
	"strings"
)

// TokenDenseDialect implements symbol shorthand for compact LLM communication.
// Inspired by lean-ctx's Token Dense Dialect (TDD).
// Replaces common programming terms with Unicode symbols for 8-25% extra savings.
type TokenDenseDialect struct {
	config  TDDConfig
	symbols map[string]string
	reverse map[string]string
}

// TDDConfig holds configuration for Token Dense Dialect.
type TDDConfig struct {
	Enabled         bool
	MinSavings      int
	MaxReplacements int
}

// DefaultTDDConfig returns default TDD configuration.
func DefaultTDDConfig() TDDConfig {
	return TDDConfig{
		Enabled:         true,
		MinSavings:      2,
		MaxReplacements: 100,
	}
}

// NewTokenDenseDialect creates a new TDD encoder.
func NewTokenDenseDialect(cfg TDDConfig) *TokenDenseDialect {
	return &TokenDenseDialect{
		config:  cfg,
		symbols: buildSymbolMap(),
		reverse: buildReverseMap(),
	}
}

// Encode replaces common terms with Unicode symbols.
func (tdd *TokenDenseDialect) Encode(input string) (string, int) {
	if !tdd.config.Enabled {
		return input, 0
	}

	result := input
	count := 0
	for term, symbol := range tdd.symbols {
		if count >= tdd.config.MaxReplacements {
			break
		}
		if strings.Contains(result, term) {
			newResult := strings.ReplaceAll(result, term, symbol)
			saved := len(term) - len(symbol)
			if saved >= tdd.config.MinSavings {
				result = newResult
				count++
			}
		}
	}
	return result, count
}

// Decode restores original terms from symbols.
func (tdd *TokenDenseDialect) Decode(input string) string {
	result := input
	for symbol, term := range tdd.reverse {
		result = strings.ReplaceAll(result, symbol, term)
	}
	return result
}

func buildSymbolMap() map[string]string {
	return map[string]string{
		"function":    "fn",
		"const":       "κ",
		"let":         "λ",
		"var":         "ν",
		"return":      "↩",
		"import":      "→",
		"export":      "←",
		"default":     "∂",
		"interface":   "∩",
		"implements":  "⊨",
		"extends":     "⊃",
		"typeof":      "τ",
		"instanceof":  "∈",
		"undefined":   "∅",
		"null":        "∅",
		"true":        "⊤",
		"false":       "⊥",
		"async":       "⚡",
		"await":       "⏳",
		"promise":     "ℙ",
		"string":      "σ",
		"number":      "ν",
		"boolean":     "β",
		"object":      "Ω",
		"array":       "Α",
		"error":       "ε",
		"throw":       "↑",
		"catch":       "⊂",
		"finally":     "∎",
		"try":         "⊢",
		"if":          "ι",
		"else":        "ε",
		"for":         "φ",
		"while":       "ω",
		"switch":      "σ",
		"case":        "κ",
		"break":       "⊘",
		"continue":    "↻",
		"new":         "Ν",
		"this":        "τ",
		"self":        "σ",
		"super":       "Σ",
		"class":       "ℂ",
		"static":      "⊙",
		"public":      "⊕",
		"private":     "⊖",
		"protected":   "⊗",
		"abstract":    "α",
		"virtual":     "ν",
		"override":    "⊕",
		"readonly":    "⊘",
		"optional":    "ο",
		"required":    "ρ",
		"nullable":    "η",
		"non-null":    "ν",
		"any":         "∀",
		"unknown":     "∃",
		"never":       "⊥",
		"void":        "∅",
		"tuple":       "Τ",
		"enum":        "Ε",
		"type":        "τ",
		"alias":       "α",
		"generic":     "Γ",
		"parameter":   "π",
		"argument":    "α",
		"variable":    "ν",
		"constant":    "κ",
		"literal":     "λ",
		"template":    "Τ",
		"interpolate": "⊕",
		"concat":      "⊕",
		"slice":       "⊂",
		"splice":      "⊗",
		"map":         "Μ",
		"filter":      "Φ",
		"reduce":      "Ρ",
		"forEach":     "∀",
		"some":        "∃",
		"every":       "∀",
		"find":        "∋",
		"includes":    "∈",
		"indexOf":     "ι",
		"length":      "ℓ",
		"size":        "σ",
		"count":       "κ",
		"sum":         "Σ",
		"average":     "μ",
		"min":         "⊥",
		"max":         "⊤",
		"sort":        "σ",
		"reverse":     "ρ",
		"join":        "⊕",
		"split":       "⊘",
		"trim":        "τ",
		"replace":     "ρ",
		"match":       "μ",
		"search":      "σ",
		"exec":        "ε",
		"compile":     "ℂ",
		"parse":       "π",
		"stringify":   "σ",
		"serialize":   "σ",
		"deserialize": "δ",
		"encode":      "ε",
		"decode":      "δ",
		"compress":    "κ",
		"decompress":  "δ",
		"encrypt":     "ε",
		"decrypt":     "δ",
		"hash":        "ℎ",
		"sign":        "σ",
		"validate":    "ν",
		"sanitize":    "σ",
		"escape":      "ε",
		"unescape":    "υ",
		"normalize":   "ν",
		"canonical":   "κ",
		"transform":   "τ",
		"convert":     "κ",
		"cast":        "κ",
		"coerce":      "κ",
		"infer":       "ι",
		"resolve":     "ρ",
		"reject":      "ρ",
		"handle":      "ℎ",
		"process":     "π",
		"execute":     "ε",
		"run":         "ρ",
		"start":       "σ",
		"stop":        "σ",
		"pause":       "π",
		"resume":      "ρ",
		"cancel":      "κ",
		"abort":       "α",
		"retry":       "ρ",
		"repeat":      "ρ",
		"loop":        "λ",
		"iterate":     "ι",
		"traverse":    "τ",
		"walk":        "ω",
		"visit":       "ν",
		"enter":       "ε",
		"exit":        "ε",
		"push":        "⊕",
		"pop":         "⊖",
		"shift":       "σ",
		"unshift":     "υ",
		"insert":      "ι",
		"delete":      "δ",
		"update":      "υ",
		"create":      "κ",
		"read":        "ρ",
		"write":       "ω",
		"load":        "λ",
		"save":        "σ",
		"fetch":       "φ",
		"send":        "σ",
		"receive":     "ρ",
		"emit":        "ε",
		"listen":      "λ",
		"on":          "⊕",
		"off":         "⊖",
		"once":        "1",
		"bind":        "β",
		"unbind":      "υ",
		"attach":      "α",
		"detach":      "δ",
		"connect":     "κ",
		"disconnect":  "δ",
		"mount":       "μ",
		"unmount":     "υ",
		"render":      "ρ",
		"paint":       "π",
		"draw":        "δ",
		"layout":      "λ",
		"measure":     "μ",
		"compute":     "κ",
		"calculate":   "κ",
		"evaluate":    "ε",
		"assess":      "α",
		"check":       "κ",
		"assert":      "α",
		"expect":      "ε",
		"assertion":   "α",
		"spec":        "σ",
		"suite":       "σ",
		"describe":    "δ",
		"it":          "ι",
		"before":      "β",
		"after":       "α",
		"setup":       "σ",
		"teardown":    "τ",
		"fixture":     "φ",
		"mock":        "μ",
		"stub":        "σ",
		"spy":         "σ",
		"fake":        "φ",
		"dummy":       "δ",
		"placeholder": "π",
	}
}

func buildReverseMap() map[string]string {
	m := make(map[string]string)
	for term, symbol := range buildSymbolMap() {
		m[symbol] = term
	}
	return m
}

// TDDStats holds encoding statistics.
type TDDStats struct {
	Replacements  int
	OriginalLen   int
	CompressedLen int
	SavingsPct    float64
}

// EncodeWithStats encodes and returns statistics.
func (tdd *TokenDenseDialect) EncodeWithStats(input string) (string, TDDStats) {
	encoded, count := tdd.Encode(input)
	stats := TDDStats{
		Replacements:  count,
		OriginalLen:   len(input),
		CompressedLen: len(encoded),
	}
	if len(input) > 0 {
		stats.SavingsPct = float64(len(input)-len(encoded)) / float64(len(input)) * 100
	}
	return encoded, stats
}

// FormatTDDStats returns a human-readable stats string.
func FormatTDDStats(stats TDDStats) string {
	return fmt.Sprintf("TDD: %d replacements, %d→%d chars (%.1f%% saved)",
		stats.Replacements, stats.OriginalLen, stats.CompressedLen, stats.SavingsPct)
}

// Itoa converts int to string.
func Itoa(n int) string {
	return strconv.Itoa(n)
}

// Ftoa converts float to string with precision.
func Ftoa(f float64, prec int) string {
	return strconv.FormatFloat(f, 'f', prec, 64)
}
