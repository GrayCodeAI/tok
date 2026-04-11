package filter

import (
	"regexp"
	"strings"
)

// ContentProfile auto-detects the best compression profile based on output content.
func ContentProfile(input string) Profile {
	lines := strings.Split(input, "\n")
	if len(lines) == 0 {
		return ProfileFast
	}

	sample := input
	if len(sample) > 4096 {
		sample = sample[:4096]
	}
	lower := strings.ToLower(sample)

	// Score each profile
	scores := map[Profile]int{
		ProfileCode:     0,
		ProfileLog:      0,
		ProfileChat:     0,
		ProfileBalanced: 0,
	}

	// --- Code detection ---
	codePatterns := []string{
		"diff --git", "@@ -", "+func ", "-func ", "package ",
		"import (", "export default", "module.exports",
		"class ", "def ", "fn ", "pub fn ", "impl ",
		"const ", "let ", "var ", "struct ", "enum ",
		"```go", "```ts", "```js", "```py", "```rs",
		"```java", "```cpp", "```c", "```ruby",
		"error[E", "warning[E", "panic(",
		"stack trace:", "goroutine ", "[signal",
	}
	for _, p := range codePatterns {
		if strings.Contains(lower, p) {
			scores[ProfileCode] += 2
		}
	}

	// Code file extensions in output
	extRe := regexp.MustCompile(`\.(go|rs|py|ts|js|java|cpp|c|rb|toml|yaml|yml|json|xml|html|css|sql|sh|bash|zsh)`)
	if len(extRe.FindAllString(sample, -1)) > 3 {
		scores[ProfileCode] += 3
	}

	// --- Log detection ---
	logPatterns := []string{
		"[info]", "[warn]", "[error]", "[debug]", "[trace]",
		"level=info", "level=warn", "level=error",
		"timestamp=", "ts=", "time=",
		"pid=", "tid=", "thread=",
		"2025-", "2026-", // date prefixes
		"passed", "failed", "skipped", "pending",
		"test session starts", "test session ends",
		"collected ", "assert ", "ok   ", "FAIL",
		"npm warn", "npm err", "pip install",
		"building ", "compiled ", "bundled ",
	}
	for _, p := range logPatterns {
		if strings.Contains(lower, p) {
			scores[ProfileLog] += 2
		}
	}

	// Repetitive lines (characteristic of logs)
	lineFreq := make(map[string]int)
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			lineFreq[trimmed]++
		}
	}
	dupCount := 0
	for _, count := range lineFreq {
		if count > 2 {
			dupCount++
		}
	}
	if dupCount > 3 {
		scores[ProfileLog] += 3
	}

	// --- Chat/conversation detection ---
	chatPatterns := []string{
		"user:", "assistant:", "human:", "ai:",
		"turn ", "session", "conversation",
		">>> ", "<<< ",
		"q:", "a:", "question:", "answer:",
	}
	for _, p := range chatPatterns {
		if strings.Contains(lower, p) {
			scores[ProfileChat] += 2
		}
	}

	// --- Determine winner ---
	bestProfile := ProfileBalanced
	bestScore := 0
	for profile, score := range scores {
		if score > bestScore {
			bestScore = score
			bestProfile = profile
		}
	}

	// If no strong signal, use balanced
	if bestScore < 3 {
		return ProfileBalanced
	}

	return bestProfile
}

// AutoProcess detects content type and applies the optimal profile.
func AutoProcess(input string, mode Mode) (string, int) {
	profile := ContentProfile(input)
	return ApplyProfile(input, mode, profile)
}
