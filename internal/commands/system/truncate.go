package system

// truncate truncates a string to maxLen characters, adding "..." if truncated.
// The total output length (including "...") will not exceed maxLen.
// If maxLen <= 3 and the string is longer, returns "...".
func truncate(s string, maxLen int) string {
	if s == "" || len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return "..."
	}
	return s[:maxLen-3] + "..."
}
