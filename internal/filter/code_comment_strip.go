package filter

import (
	"regexp"
)

// commentFallbackRe is used when no language-specific pattern is found.
var commentFallbackRe = regexp.MustCompile(`(?m)^//.*$|/\*[\s\S]*?\*/|^#.*$`)

// CommentPatternsMap maps languages to their comment regex patterns
var CommentPatternsMap = map[Language]*regexp.Regexp{
	LangRust:       regexp.MustCompile(`(?m)^//.*$|/\*[\s\S]*?\*/`),
	LangGo:         regexp.MustCompile(`(?m)^//.*$|/\*[\s\S]*?\*/`),
	LangPython:     regexp.MustCompile(`(?m)^#.*$|"""[\s\S]*?"""|'''[\s\S]*?'''`),
	LangJavaScript: regexp.MustCompile(`(?m)^//.*$|/\*[\s\S]*?\*/`),
	LangTypeScript: regexp.MustCompile(`(?m)^//.*$|/\*[\s\S]*?\*/`),
	LangJava:       regexp.MustCompile(`(?m)^//.*$|/\*[\s\S]*?\*/`),
	LangC:          regexp.MustCompile(`(?m)^//.*$|/\*[\s\S]*?\*/`),
	LangCpp:        regexp.MustCompile(`(?m)^//.*$|/\*[\s\S]*?\*/`),
	LangRuby:       regexp.MustCompile(`(?m)^#.*$`),
	LangShell:      regexp.MustCompile(`(?m)^#.*$`),
	LangSQL:        regexp.MustCompile(`(?m)^--.*$|/\*[\s\S]*?\*/`),
}

// CommentFilter strips comments from code.
type CommentFilter struct {
	patterns map[Language]*regexp.Regexp
}

// newCommentFilter creates a new comment filter.
func newCommentFilter() *CommentFilter {
	return &CommentFilter{
		patterns: CommentPatternsMap,
	}
}

// Name returns the filter name.
func (f *CommentFilter) Name() string {
	return "comment"
}

// Apply strips comments and returns token savings.
func (f *CommentFilter) Apply(input string, mode Mode) (string, int) {
	lang := DetectLanguageFromInput(input)

	pattern, ok := f.patterns[lang]
	if !ok {
		pattern = commentFallbackRe
	}

	original := len(input)
	output := pattern.ReplaceAllString(input, "")

	bytesSaved := original - len(output)
	tokensSaved := bytesSaved / 4

	return output, tokensSaved
}
