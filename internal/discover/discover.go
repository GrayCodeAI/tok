// Package discover provides command discovery functionality (stub implementation).
// NOTE: This is a stub package. The full implementation was removed as dead code.
// These stub functions maintain API compatibility.
package discover

// RewriteCommand rewrites a command using tokman equivalents (stub).
func RewriteCommand(cmd string, opts interface{}) (string, bool) {
	return cmd, false
}

// DetectCommand detects if a command is known (stub).
func DetectCommand(cmd string) bool {
	return false
}

// KnownCommands returns list of known commands (stub).
func KnownCommands() []string {
	return []string{}
}
