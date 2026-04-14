package filter

import (
	"regexp"
	"strings"
)

// SmartTruncate truncates content while preserving function signatures.
// This is useful for code files where you want to keep the API surface
// but remove implementation details.
func SmartTruncate(content string, maxLines int, lang Language) string {
	if maxLines <= 0 {
		return content
	}

	lines := strings.Split(content, "\n")
	if len(lines) <= maxLines {
		return content
	}

	// Extract signatures based on language
	signatures := extractSignatures(content, lang)

	// Calculate how many lines to keep from the beginning and end
	headLines := maxLines / 2
	tailLines := maxLines - headLines

	// Build result
	var result []string

	// Add head lines
	result = append(result, lines[:headLines]...)

	// Add signature block if we have signatures
	if len(signatures) > 0 {
		result = append(result, "")
		result = append(result, "// ... (implementation omitted, signatures preserved) ...")
		result = append(result, "")
		for _, sig := range signatures {
			result = append(result, sig)
		}
		result = append(result, "")
		result = append(result, "// ...")
		result = append(result, "")
	} else {
		result = append(result, "")
		result = append(result, "// ... (truncated)")
		result = append(result, "")
	}

	// Add tail lines
	result = append(result, lines[len(lines)-tailLines:]...)

	return strings.Join(result, "\n")
}

// extractSignatures extracts function/method signatures from code
func extractSignatures(content string, lang Language) []string {
	switch lang {
	case LangRust:
		return extractRustSignatures(content)
	case LangGo:
		return extractGoSignatures(content)
	case LangPython:
		return extractPythonSignatures(content)
	case LangJavaScript, LangTypeScript:
		return extractJSSignatures(content)
	case LangJava:
		return extractJavaSignatures(content)
	case LangRuby:
		return extractRubySignatures(content)
	default:
		return extractGenericSignatures(content)
	}
}

func extractRustSignatures(content string) []string {
	var signatures []string

	// Pattern for function signatures
	fnRe := regexp.MustCompile(`(?m)^\s*(pub\s+)?(?:async\s+)?fn\s+(\w+)\s*\([^)]*\)(?:\s*->\s*[^{]+)?`)
	traitRe := regexp.MustCompile(`(?m)^\s*(pub\s+)?trait\s+(\w+)`)
	structRe := regexp.MustCompile(`(?m)^\s*(pub\s+)?struct\s+(\w+)`)
	implRe := regexp.MustCompile(`(?m)^\s*impl\s+[^\s{]+`)

	for _, match := range fnRe.FindAllString(content, -1) {
		signatures = append(signatures, strings.TrimSpace(match)+";")
	}
	for _, match := range traitRe.FindAllString(content, -1) {
		signatures = append(signatures, strings.TrimSpace(match)+" { ... }")
	}
	for _, match := range structRe.FindAllString(content, -1) {
		signatures = append(signatures, strings.TrimSpace(match)+" { ... }")
	}
	for _, match := range implRe.FindAllString(content, -1) {
		signatures = append(signatures, strings.TrimSpace(match)+" { ... }")
	}

	return signatures
}

func extractGoSignatures(content string) []string {
	var signatures []string

	// Pattern for function signatures
	fnRe := regexp.MustCompile(`(?m)^\s*func\s+(?:\([^)]+\)\s+)?(\w+)\s*\([^)]*\)(?:\s*\([^)]*\))?(?:\s*\w+)?`)
	typeRe := regexp.MustCompile(`(?m)^\s*type\s+(\w+)\s+(?:struct|interface)`)

	for _, match := range fnRe.FindAllString(content, -1) {
		signatures = append(signatures, strings.TrimSpace(match))
	}
	for _, match := range typeRe.FindAllString(content, -1) {
		signatures = append(signatures, strings.TrimSpace(match)+" { ... }")
	}

	return signatures
}

func extractPythonSignatures(content string) []string {
	var signatures []string

	// Pattern for function and class definitions
	fnRe := regexp.MustCompile(`(?m)^\s*(?:async\s+)?def\s+(\w+)\s*\([^)]*\)(?:\s*->\s*[^:]+)?:`)
	classRe := regexp.MustCompile(`(?m)^\s*class\s+(\w+)(?:\s*\([^)]*\))?:`)

	for _, match := range fnRe.FindAllString(content, -1) {
		sig := strings.TrimSpace(match)
		sig = strings.TrimSuffix(sig, ":")
		signatures = append(signatures, sig+"...")
	}
	for _, match := range classRe.FindAllString(content, -1) {
		sig := strings.TrimSpace(match)
		sig = strings.TrimSuffix(sig, ":")
		signatures = append(signatures, sig+": ...")
	}

	return signatures
}

func extractJSSignatures(content string) []string {
	var signatures []string

	// Pattern for function declarations
	fnRe := regexp.MustCompile(`(?m)^\s*(?:export\s+)?(?:async\s+)?function\s+(\w+)\s*\([^)]*\)`)
	arrowRe := regexp.MustCompile(`(?m)^\s*(?:export\s+)?const\s+(\w+)\s*=\s*(?:async\s*)?\([^)]*\)\s*=>`)
	classRe := regexp.MustCompile(`(?m)^\s*(?:export\s+)?class\s+(\w+)`)
	interfaceRe := regexp.MustCompile(`(?m)^\s*(?:export\s+)?interface\s+(\w+)`)

	for _, match := range fnRe.FindAllString(content, -1) {
		signatures = append(signatures, strings.TrimSpace(match)+" { ... }")
	}
	for _, match := range arrowRe.FindAllString(content, -1) {
		signatures = append(signatures, strings.TrimSpace(match)+" { ... }")
	}
	for _, match := range classRe.FindAllString(content, -1) {
		signatures = append(signatures, strings.TrimSpace(match)+" { ... }")
	}
	for _, match := range interfaceRe.FindAllString(content, -1) {
		signatures = append(signatures, strings.TrimSpace(match)+" { ... }")
	}

	return signatures
}

func extractJavaSignatures(content string) []string {
	var signatures []string

	// Pattern for method signatures
	methodRe := regexp.MustCompile(`(?m)^\s*(?:public|private|protected)?\s*(?:static\s+)?(?:final\s+)?(?:[\w<>]+\s+)?(\w+)\s*\([^)]*\)(?:\s*throws\s+\w+)?\s*\{`)
	classRe := regexp.MustCompile(`(?m)^\s*(?:public\s+)?(?:abstract\s+)?(?:final\s+)?class\s+(\w+)`)
	interfaceRe := regexp.MustCompile(`(?m)^\s*(?:public\s+)?interface\s+(\w+)`)

	for _, match := range methodRe.FindAllString(content, -1) {
		sig := strings.TrimSpace(match)
		sig = strings.TrimSuffix(sig, "{")
		signatures = append(signatures, sig+"{ ... }")
	}
	for _, match := range classRe.FindAllString(content, -1) {
		signatures = append(signatures, strings.TrimSpace(match)+" { ... }")
	}
	for _, match := range interfaceRe.FindAllString(content, -1) {
		signatures = append(signatures, strings.TrimSpace(match)+" { ... }")
	}

	return signatures
}

func extractRubySignatures(content string) []string {
	var signatures []string

	// Pattern for method and class definitions
	methodRe := regexp.MustCompile(`(?m)^\s*def\s+(\w+)(?:\s*\([^)]*\))?`)
	classRe := regexp.MustCompile(`(?m)^\s*class\s+(\w+)`)
	moduleRe := regexp.MustCompile(`(?m)^\s*module\s+(\w+)`)

	for _, match := range methodRe.FindAllString(content, -1) {
		signatures = append(signatures, strings.TrimSpace(match)+" ... end")
	}
	for _, match := range classRe.FindAllString(content, -1) {
		signatures = append(signatures, strings.TrimSpace(match)+" ... end")
	}
	for _, match := range moduleRe.FindAllString(content, -1) {
		signatures = append(signatures, strings.TrimSpace(match)+" ... end")
	}

	return signatures
}

func extractGenericSignatures(content string) []string {
	// Generic signature extraction - look for common patterns
	var signatures []string

	// Try to match function-like patterns
	fnRe := regexp.MustCompile(`(?m)^\s*(?:function|func|def|fn|method)\s+(\w+)\s*\([^)]*\)`)

	for _, match := range fnRe.FindAllString(content, -1) {
		signatures = append(signatures, strings.TrimSpace(match)+" { ... }")
	}

	return signatures
}
