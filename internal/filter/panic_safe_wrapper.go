package filter

import (
	"fmt"
	"os"
	"runtime/debug"
)

// SafeFilter wraps a filter with nil checks and panic recovery.
// Prevents nil pointer dereferences from crashing the pipeline.
type SafeFilter struct {
	filter Filter
	name   string
}

// NewSafeFilter creates a nil-safe filter wrapper.
func NewSafeFilter(filter Filter, name string) *SafeFilter {
	return &SafeFilter{
		filter: filter,
		name:   name,
	}
}

// Apply safely applies the filter with nil checks and recovery.
func (sf *SafeFilter) Apply(input string, mode Mode) (output string, saved int) {
	// Nil check
	if sf.filter == nil {
		return input, 0
	}

	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "tok: filter %q panicked: %v\n%s\n", sf.name, r, debug.Stack())
			output = input
			saved = 0
		}
	}()

	return sf.filter.Apply(input, mode)
}

// Name returns the filter name.
func (sf *SafeFilter) Name() string {
	return sf.name
}

// IsNil reports whether the wrapped filter is nil.
func (sf *SafeFilter) IsNil() bool {
	return sf.filter == nil
}

// SafeFilterFunc is a convenience function for creating safe filters.
func SafeFilterFunc(filter Filter, name string) Filter {
	if filter == nil {
		return &SafeFilter{filter: nil, name: name}
	}
	return &SafeFilter{filter: filter, name: name}
}
