// Package yamlfilter provides YAML-based filter definitions.
// 
// YAML filters allow users to define command filters declaratively
// without code changes, similar to Snip's approach.
//
// Example filter:
//   name: go-test
//   match: "^go test"
//   pipeline:
//     - strip_lines_matching:
//         - "^=== RUN"
//         - "^--- PASS"
//     - max_lines: 5
//
package yamlfilter

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// Filter represents a single YAML filter definition.
type Filter struct {
	Name         string   `yaml:"name"`
	Description  string   `yaml:"description,omitempty"`
	Match        string   `yaml:"match"`
	StripLines   []string `yaml:"strip_lines_matching,omitempty"`
	KeepLines    []string `yaml:"keep_lines_matching,omitempty"`
	MaxLines     int      `yaml:"max_lines,omitempty"`
	Prefix       string   `yaml:"prefix,omitempty"`
	Suffix       string   `yaml:"suffix,omitempty"`
	GroupBy      string   `yaml:"group_by,omitempty"`
	Deduplicate  bool     `yaml:"deduplicate,omitempty"`
	MinLines     int      `yaml:"min_lines,omitempty"` // Only apply if output >= this many lines
}

// FilterList represents a YAML file containing multiple filters.
type FilterList struct {
	Version  int      `yaml:"schema_version"`
	Filters  []Filter `yaml:"filters"`
}

// Loader loads and manages YAML filters.
type Loader struct {
	filters []Filter
	regexp  map[string]*regexp.Regexp // Compiled regex cache
}

// New creates a new YAML filter loader.
func New() *Loader {
	return &Loader{
		regexp: make(map[string]*regexp.Regexp),
	}
}

// LoadFromFile loads filters from a single YAML file.
func (l *Loader) LoadFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	var list FilterList
	if err := yaml.Unmarshal(data, &list); err != nil {
		return fmt.Errorf("unmarshal YAML: %w", err)
	}

	// Validate schema version
	if list.Version != 0 && list.Version != 1 {
		return fmt.Errorf("unsupported schema version: %d", list.Version)
	}

	// Compile and cache regex patterns
	for _, f := range list.Filters {
		if f.Match != "" {
			if _, err := regexp.Compile(f.Match); err != nil {
				return fmt.Errorf("invalid match regex %q: %w", f.Match, err)
			}
		}
		for _, p := range f.StripLines {
			if _, err := regexp.Compile(p); err != nil {
				return fmt.Errorf("invalid strip_lines regex %q: %w", p, err)
			}
		}
		for _, p := range f.KeepLines {
			if _, err := regexp.Compile(p); err != nil {
				return fmt.Errorf("invalid keep_lines regex %q: %w", p, err)
			}
		}
	}

	l.filters = append(l.filters, list.Filters...)
	return nil
}

// LoadFromDir loads all YAML filter files from a directory.
func (l *Loader) LoadFromDir(dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".yml") || strings.HasSuffix(path, ".yaml") {
			return l.LoadFromFile(path)
		}
		return nil
	})
}

// Match finds a filter that matches the given command.
func (l *Loader) Match(command string) *Filter {
	for _, f := range l.filters {
		if f.MatchCommand(command) {
			return &f
		}
	}
	return nil
}

// Apply applies a filter to the given output text.
func (l *Loader) Apply(filter *Filter, output string) string {
	lines := strings.Split(output, "\n")

	// Check minimum lines
	if filter.MinLines > 0 && len(lines) < filter.MinLines {
		return output // Too few lines, don't apply
	}

	// Collect lines to keep
	var kept []string
	for _, line := range lines {
		// Check if line should be stripped
		if l.shouldStrip(filter, line) {
			continue
		}
		kept = append(kept, line)
	}

	// Apply max lines
	if filter.MaxLines > 0 && len(kept) > filter.MaxLines {
		half := filter.MaxLines / 2
		// Keep first half and last half
		remaining := len(kept) - filter.MaxLines
		lastStart := half + remaining
		if lastStart > len(kept) {
			lastStart = len(kept)
		}
		
		middle := "..."
		if filter.Prefix != "" {
			middle = filter.Prefix + "\n...\n"
		} else if filter.Suffix != "" {
			middle = "...\n" + filter.Suffix + "\n"
		} else {
			middle = fmt.Sprintf("\n... [%d lines truncated] ...\n", remaining)
		}
		
		if half >= 0 && len(kept) > half && lastStart < len(kept) {
			kept = append(kept[:half], append([]string{middle}, kept[lastStart:]...)...)
		} else if len(kept) > filter.MaxLines {
			kept = kept[:filter.MaxLines]
		}
	}

	result := strings.Join(kept, "\n")

	// Add prefix/suffix if set and not already added
	if filter.Prefix != "" && !strings.HasPrefix(result, filter.Prefix) {
		result = filter.Prefix + "\n" + result
	}
	if filter.Suffix != "" && !strings.HasSuffix(result, filter.Suffix) {
		result = result + "\n" + filter.Suffix
	}

	return result
}

func (l *Loader) shouldStrip(filter *Filter, line string) bool {
	trimmed := strings.TrimSpace(line)
	
	// Check keep patterns first (if any keep patterns, non-matching are stripped)
	if len(filter.KeepLines) > 0 {
		for _, pattern := range filter.KeepLines {
			re := l.getOrCompile(pattern)
			if re.MatchString(trimmed) {
				return false // Keep this line
			}
		}
		return true // Stripped because it doesn't match any keep pattern
	}

	// Check strip patterns
	for _, pattern := range filter.StripLines {
		re := l.getOrCompile(pattern)
		if re.MatchString(trimmed) {
			return true // Strip this line
		}
	}

	return false
}

func (l *Loader) getOrCompile(pattern string) *regexp.Regexp {
	if re, ok := l.regexp[pattern]; ok {
		return re
	}
	re, _ := regexp.Compile(pattern)
	if re != nil {
		l.regexp[pattern] = re
	}
	return re
}

// MatchCommand checks if the filter matches the given command string.
func (f *Filter) MatchCommand(command string) bool {
	if f.Match == "" {
		return false
	}
	
	// Try to find compiled regex in loader's cache first (not available here)
	re := regexp.MustCompilePOSIX(f.Match)
	if re == nil {
		re = regexp.MustCompile(regexp.QuoteMeta(f.Match))
	}
	return re.MatchString(command)
}

// List returns all loaded filters.
func (l *Loader) List() []Filter {
	return l.filters
}

// Count returns the number of loaded filters.
func (l *Loader) Count() int {
	return len(l.filters)
}
