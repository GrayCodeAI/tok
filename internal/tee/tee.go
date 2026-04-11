// Package tee provides tee functionality for capturing command output (stub implementation).
// NOTE: This is a stub package. The full implementation was removed as dead code.
// These stub functions maintain API compatibility.
package tee

import "fmt"

// WriteAndHint writes content to a tee file and returns a hint message (stub).
func WriteAndHint(content string, commandSlug string, exitCode int) string {
	// Stub: no-op, just return a message
	return fmt.Sprintf("[tee] output captured for %s (exit code %d)", commandSlug, exitCode)
}
