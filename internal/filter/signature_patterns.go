package filter

import "regexp"

// SignaturePatterns for function/type signature detection
var SignaturePatterns = []*regexp.Regexp{
	regexp.MustCompile(`^(pub\s+)?(async\s+)?fn\s+\w+`),
	regexp.MustCompile(`^(pub\s+)?struct\s+\w+`),
	regexp.MustCompile(`^(pub\s+)?enum\s+\w+`),
	regexp.MustCompile(`^(pub\s+)?trait\s+\w+`),
	regexp.MustCompile(`^(pub\s+)?type\s+\w+`),
	regexp.MustCompile(`^impl\s+`),
	regexp.MustCompile(`^func\s+(\([^)]+\)\s+)?\w+`),
	regexp.MustCompile(`^type\s+\w+\s+(struct|interface)`),
	regexp.MustCompile(`^type\s+\w+\s+\w+`),
	regexp.MustCompile(`^def\s+\w+`),
	regexp.MustCompile(`^async\s+def\s+\w+`),
	regexp.MustCompile(`^class\s+\w+`),
	regexp.MustCompile(`^function\s+\w+`),
	regexp.MustCompile(`^(export\s+)?(async\s+)?function\s*\w*`),
	regexp.MustCompile(`^(export\s+)?(default\s+)?class\s+\w+`),
	regexp.MustCompile(`^(export\s+)?const\s+\w+\s*=\s*(async\s+)?\([^)]*\)\s*=>`),
	regexp.MustCompile(`^interface\s+\w+`),
	regexp.MustCompile(`^type\s+\w+\s*=`),
	regexp.MustCompile(`^(public|private|protected)?\s*(static\s+)?(class|interface|enum)\s+\w+`),
	regexp.MustCompile(`^(public|private|protected)?\s*(static\s+)?(async\s+)?\w+\s+\w+\s*\(`),
}

// ImportPatterns for various languages
var ImportPatterns = []*regexp.Regexp{
	regexp.MustCompile(`^use\s+`),
	regexp.MustCompile(`^import\s+`),
	regexp.MustCompile(`^from\s+\S+\s+import`),
	regexp.MustCompile(`^require\(`),
	regexp.MustCompile(`^import\s*\(`),
	regexp.MustCompile(`^import\s+"`),
	regexp.MustCompile(`#include\s*<`),
	regexp.MustCompile(`#include\s*"`),
	regexp.MustCompile(`^package\s+`),
}

// BlockDelimiters for brace tracking
var BlockDelimiters = map[rune]rune{
	'{': '}',
	'[': ']',
	'(': ')',
}
