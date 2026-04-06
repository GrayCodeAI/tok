package analysis

// truncateStr truncates a string to max characters, adding "..." if truncated.
// The total output length (including "...") will not exceed max.
func truncateStr(s string, max int) string {
	if len(s) <= max {
		return s
	}
	if max <= 3 {
		return s[:max]
	}
	return s[:max-3] + "..."
}

// truncate is an alias for truncateStr for backward compatibility.
func truncate(s string, max int) string {
	return truncateStr(s, max)
}
