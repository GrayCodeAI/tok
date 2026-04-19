// Package discover provides command discovery and auto-rewrite functionality.
// It implements transparent tok command rewriting for supported shells and agents.
package discover

import (
	"regexp"
	"strings"
	"sync"

	"github.com/lakshmanpatel/tok/internal/telemetry"
)

// rewriteCache caches command rewrite results to avoid reprocessing
var rewriteCache = struct {
	sync.RWMutex
	data map[string]rewriteCacheEntry
}{
	data: make(map[string]rewriteCacheEntry),
}

type rewriteCacheEntry struct {
	rewritten string
	changed   bool
}

// Cache hit stats for monitoring
var cacheStats = struct {
	sync.RWMutex
	hits   int64
	misses int64
}{}

// CommandPattern defines a command pattern for rewriting
type CommandPattern struct {
	Name        string
	Pattern     *regexp.Regexp
	Rewrite     string
	Description string
	Priority    int // Higher priority patterns are checked first
}

// SupportLevel describes how well tok handles a command family.
type SupportLevel string

const (
	SupportOptimized   SupportLevel = "optimized"
	SupportPassthrough SupportLevel = "passthrough"
	SupportUnsupported SupportLevel = "unsupported"
)

// Common rewrite patterns for tok auto-rewrite
var rewritePatterns = []CommandPattern{
	// Test runners - high priority
	{Name: "cargo test", Pattern: regexp.MustCompile(`^cargo\s+test`), Rewrite: "tok cargo test", Description: "Rust tests", Priority: 100},
	{Name: "go test", Pattern: regexp.MustCompile(`^go\s+test`), Rewrite: "tok go test", Description: "Go tests", Priority: 100},
	{Name: "npm test", Pattern: regexp.MustCompile(`^npm\s+test`), Rewrite: "tok npm test", Description: "npm tests", Priority: 100},
	{Name: "pnpm test", Pattern: regexp.MustCompile(`^pnpm\s+test`), Rewrite: "tok pnpm test", Description: "pnpm tests", Priority: 100},
	{Name: "pytest", Pattern: regexp.MustCompile(`^pytest`), Rewrite: "tok pytest", Description: "Python tests", Priority: 100},
	{Name: "vitest", Pattern: regexp.MustCompile(`^vitest|^npx\s+vitest`), Rewrite: "tok vitest", Description: "Vitest tests", Priority: 100},
	{Name: "jest", Pattern: regexp.MustCompile(`^jest|^npx\s+jest`), Rewrite: "tok jest", Description: "Jest tests", Priority: 100},
	{Name: "playwright", Pattern: regexp.MustCompile(`^playwright\s+test|^npx\s+playwright\s+test`), Rewrite: "tok playwright", Description: "Playwright tests", Priority: 100},
	{Name: "rspec", Pattern: regexp.MustCompile(`^rspec`), Rewrite: "tok rspec", Description: "RSpec tests", Priority: 100},
	{Name: "rake test", Pattern: regexp.MustCompile(`^rake\s+test`), Rewrite: "tok rake test", Description: "Rake tests", Priority: 100},

	// Build commands
	{Name: "cargo build", Pattern: regexp.MustCompile(`^cargo\s+build`), Rewrite: "tok cargo build", Description: "Rust build", Priority: 90},
	{Name: "cargo clippy", Pattern: regexp.MustCompile(`^cargo\s+clippy`), Rewrite: "tok cargo clippy", Description: "Rust lint", Priority: 90},
	{Name: "npm run build", Pattern: regexp.MustCompile(`^npm\s+run\s+build`), Rewrite: "tok err npm run build", Description: "npm build", Priority: 90},
	{Name: "pnpm build", Pattern: regexp.MustCompile(`^pnpm\s+(run\s+)?build`), Rewrite: "tok err pnpm build", Description: "pnpm build", Priority: 90},
	{Name: "tsc", Pattern: regexp.MustCompile(`^tsc`), Rewrite: "tok tsc", Description: "TypeScript compiler", Priority: 90},
	{Name: "next build", Pattern: regexp.MustCompile(`^next\s+build`), Rewrite: "tok next build", Description: "Next.js build", Priority: 90},
	{Name: "golangci-lint", Pattern: regexp.MustCompile(`^golangci-lint`), Rewrite: "tok golangci-lint", Description: "Go linter", Priority: 90},
	{Name: "ruff", Pattern: regexp.MustCompile(`^ruff\s+(check|format)`), Rewrite: "tok ruff", Description: "Python linter", Priority: 90},

	// Git commands
	{Name: "git status", Pattern: regexp.MustCompile(`^git\s+status`), Rewrite: "tok git status", Description: "Git status", Priority: 80},
	{Name: "git log", Pattern: regexp.MustCompile(`^git\s+log`), Rewrite: "tok git log", Description: "Git log", Priority: 80},
	{Name: "git diff", Pattern: regexp.MustCompile(`^git\s+diff`), Rewrite: "tok git diff", Description: "Git diff", Priority: 80},
	{Name: "git add", Pattern: regexp.MustCompile(`^git\s+add`), Rewrite: "tok git add", Description: "Git add", Priority: 80},
	{Name: "git commit", Pattern: regexp.MustCompile(`^git\s+commit`), Rewrite: "tok git commit", Description: "Git commit", Priority: 80},
	{Name: "git push", Pattern: regexp.MustCompile(`^git\s+push`), Rewrite: "tok git push", Description: "Git push", Priority: 80},
	{Name: "git pull", Pattern: regexp.MustCompile(`^git\s+pull`), Rewrite: "tok git pull", Description: "Git pull", Priority: 80},

	// GitHub CLI
	{Name: "gh pr", Pattern: regexp.MustCompile(`^gh\s+pr`), Rewrite: "tok gh pr", Description: "GitHub PR", Priority: 80},
	{Name: "gh issue", Pattern: regexp.MustCompile(`^gh\s+issue`), Rewrite: "tok gh issue", Description: "GitHub issue", Priority: 80},
	{Name: "gh run", Pattern: regexp.MustCompile(`^gh\s+run`), Rewrite: "tok gh run", Description: "GitHub Actions", Priority: 80},

	// Docker/Kubernetes
	{Name: "docker ps", Pattern: regexp.MustCompile(`^docker\s+ps`), Rewrite: "tok docker ps", Description: "Docker containers", Priority: 70},
	{Name: "docker images", Pattern: regexp.MustCompile(`^docker\s+images`), Rewrite: "tok docker images", Description: "Docker images", Priority: 70},
	{Name: "docker logs", Pattern: regexp.MustCompile(`^docker\s+(logs|compose\s+logs)`), Rewrite: "tok docker logs", Description: "Docker logs", Priority: 70},
	{Name: "kubectl", Pattern: regexp.MustCompile(`^kubectl\s+(get|logs|describe)`), Rewrite: "tok kubectl", Description: "Kubernetes", Priority: 70},

	// Package managers
	{Name: "npm ls", Pattern: regexp.MustCompile(`^npm\s+ls`), Rewrite: "tok npm ls", Description: "npm list", Priority: 60},
	{Name: "pnpm list", Pattern: regexp.MustCompile(`^pnpm\s+list`), Rewrite: "tok pnpm list", Description: "pnpm list", Priority: 60},
	{Name: "pip list", Pattern: regexp.MustCompile(`^pip\s+list`), Rewrite: "tok pip list", Description: "pip list", Priority: 60},
	{Name: "bundle install", Pattern: regexp.MustCompile(`^bundle\s+install`), Rewrite: "tok bundle install", Description: "Bundle install", Priority: 60},

	// System commands
	{Name: "ls", Pattern: regexp.MustCompile(`^ls\b`), Rewrite: "tok ls", Description: "List directory", Priority: 50},
	{Name: "tree", Pattern: regexp.MustCompile(`^tree`), Rewrite: "tok tree", Description: "Directory tree", Priority: 50},
	{Name: "cat", Pattern: regexp.MustCompile(`^cat\s`), Rewrite: "tok read", Description: "Read file", Priority: 50},
	{Name: "grep", Pattern: regexp.MustCompile(`^grep\s`), Rewrite: "tok grep", Description: "Search files", Priority: 50},
	{Name: "find", Pattern: regexp.MustCompile(`^find\s`), Rewrite: "tok find", Description: "Find files", Priority: 50},
	{Name: "wc", Pattern: regexp.MustCompile(`^wc\s`), Rewrite: "tok wc", Description: "Word count", Priority: 50},
	{Name: "env", Pattern: regexp.MustCompile(`^env$`), Rewrite: "tok env", Description: "Environment variables", Priority: 50},
}

// RewriteOptions provides options for command rewriting
type RewriteOptions struct {
	// DisableTestRunner disables automatic test-runner detection
	DisableTestRunner bool
	// PreferExplicit prefers explicit tok commands over test-runner
	PreferExplicit bool
	// DisableCache disables caching for this rewrite
	DisableCache bool
	// SkipTelemetry disables rewrite telemetry emission
	SkipTelemetry bool
}

// RewriteCommand rewrites a command using tok equivalents.
// Returns the rewritten command and true if a rewrite occurred.
// Results are cached for performance.
func RewriteCommand(cmd string, opts interface{}) (string, bool) {
	if cmd == "" {
		return cmd, false
	}

	// Parse options
	options := &RewriteOptions{}
	if o, ok := opts.(*RewriteOptions); ok {
		options = o
	}

	trimmed := strings.TrimSpace(cmd)

	// Skip if already starts with tok
	if strings.HasPrefix(trimmed, "tok ") {
		return cmd, false
	}

	// Check cache (unless disabled)
	if !options.DisableCache {
		rewriteCache.RLock()
		if entry, ok := rewriteCache.data[trimmed]; ok {
			rewriteCache.RUnlock()
			cacheStats.Lock()
			cacheStats.hits++
			cacheStats.Unlock()
			return entry.rewritten, entry.changed
		}
		rewriteCache.RUnlock()
	}

	// Update cache miss stats
	cacheStats.Lock()
	cacheStats.misses++
	cacheStats.Unlock()

	// Check against patterns
	for _, pattern := range rewritePatterns {
		if pattern.Pattern.MatchString(trimmed) {
			rewritten := pattern.Pattern.ReplaceAllString(trimmed, pattern.Rewrite)

			// Handle test commands specially for test-runner integration
			if !options.DisableTestRunner && isTestCommand(trimmed) && !options.PreferExplicit {
				rewritten = rewriteWithTestRunner(trimmed)
			}

			// Cache the result
			if !options.DisableCache {
				rewriteCache.Lock()
				rewriteCache.data[trimmed] = rewriteCacheEntry{rewritten, true}
				rewriteCache.Unlock()
			}

			if !options.SkipTelemetry {
				telemetry.TrackRewrite(cmd, rewritten, isTestCommand(trimmed))
			}
			return rewritten, true
		}
	}

	// Try test-runner auto-detection for unknown test commands
	if !options.DisableTestRunner {
		if testCmd := detectTestRunner(trimmed); testCmd != "" {
			rewritten := "tok test-runner " + trimmed

			// Cache the result
			if !options.DisableCache {
				rewriteCache.Lock()
				rewriteCache.data[trimmed] = rewriteCacheEntry{rewritten, true}
				rewriteCache.Unlock()
			}

			if !options.SkipTelemetry {
				telemetry.TrackRewrite(cmd, rewritten, true)
			}
			return rewritten, true
		}
	}

	// Cache negative result (no rewrite)
	if !options.DisableCache {
		rewriteCache.Lock()
		rewriteCache.data[trimmed] = rewriteCacheEntry{cmd, false}
		rewriteCache.Unlock()
	}

	return cmd, false
}

// ClassifyCommand returns tok support level and the recommended tok equivalent when known.
func ClassifyCommand(cmd string) (string, SupportLevel) {
	rewritten, changed := RewriteCommand(cmd, &RewriteOptions{
		DisableCache:      true,
		DisableTestRunner: false,
		SkipTelemetry:     true,
	})
	if changed && rewritten != cmd {
		return rewritten, SupportOptimized
	}
	if rewritten, ok := detectPassthroughCommand(cmd); ok {
		return rewritten, SupportPassthrough
	}
	return cmd, SupportUnsupported
}

// GetCacheStats returns cache hit/miss statistics
func GetCacheStats() (hits, misses int64) {
	cacheStats.RLock()
	defer cacheStats.RUnlock()
	return cacheStats.hits, cacheStats.misses
}

// ClearCache clears the rewrite cache
func ClearCache() {
	rewriteCache.Lock()
	rewriteCache.data = make(map[string]rewriteCacheEntry)
	rewriteCache.Unlock()

	cacheStats.Lock()
	cacheStats.hits = 0
	cacheStats.misses = 0
	cacheStats.Unlock()
}

// isTestCommand checks if a command is a known test command
func isTestCommand(cmd string) bool {
	testPatterns := []*regexp.Regexp{
		regexp.MustCompile(`^(cargo|go|npm|pnpm|yarn)\s+test`),
		regexp.MustCompile(`^pytest`),
		regexp.MustCompile(`^rspec`),
		regexp.MustCompile(`^vitest`),
		regexp.MustCompile(`^jest`),
		regexp.MustCompile(`^rake\s+test`),
		regexp.MustCompile(`^playwright\s+test`),
	}
	for _, p := range testPatterns {
		if p.MatchString(cmd) {
			return true
		}
	}
	return false
}

// rewriteWithTestRunner rewrites a test command to use test-runner
func rewriteWithTestRunner(cmd string) string {
	return "tok test-runner " + cmd
}

// detectTestRunner attempts to detect the appropriate test runner for a command
func detectTestRunner(cmd string) string {
	// Check for test-related keywords
	testKeywords := []string{"test", "spec", "jest", "mocha", "ava", "tap"}
	cmdLower := strings.ToLower(cmd)

	for _, kw := range testKeywords {
		if strings.Contains(cmdLower, kw) {
			// This looks like a test command
			return cmd
		}
	}

	return ""
}

// DetectCommand detects if a command is known and can be rewritten.
func DetectCommand(cmd string) bool {
	_, level := ClassifyCommand(cmd)
	return level != SupportUnsupported
}

// KnownCommands returns list of known command patterns.
func KnownCommands() []string {
	var names []string
	for _, p := range rewritePatterns {
		names = append(names, p.Name)
	}
	return names
}

// ShouldRewriteFile checks if a file should trigger rewrite detection
func ShouldRewriteFile(filename string) bool {
	testSuffixes := []string{
		"_test.go", "_spec.rb", ".test.js", ".spec.js",
		"_test.py", "Test.java", "_test.rs",
	}
	for _, suffix := range testSuffixes {
		if strings.HasSuffix(filename, suffix) {
			return true
		}
	}
	// Python files starting with test_
	if strings.HasPrefix(filename, "test_") && strings.HasSuffix(filename, ".py") {
		return true
	}
	return false
}

func detectPassthroughCommand(cmd string) (string, bool) {
	fields := strings.Fields(strings.TrimSpace(cmd))
	if len(fields) < 2 {
		return "", false
	}

	if !isKnownPassthroughRoot(fields[0]) {
		return "", false
	}
	if fields[0] == "git" && isExplicitGitOptimization(fields[1]) {
		return "", false
	}
	return "tok " + strings.TrimSpace(cmd), true
}

func isKnownPassthroughRoot(root string) bool {
	switch root {
	case "git", "gh", "docker", "kubectl", "aws", "npm", "pnpm", "npx", "pip", "go",
		"cargo", "ruff", "terraform", "helm", "ansible", "make", "gradle", "mvn", "mix",
		"bundle", "rake", "rspec", "next", "prisma", "tsc", "curl", "wget", "jq",
		"ls", "grep", "find", "tree", "wc", "df", "du", "mise", "just":
		return true
	default:
		return false
	}
}

func isExplicitGitOptimization(subcommand string) bool {
	switch subcommand {
	case "status", "log", "diff", "add", "commit", "push", "pull", "show", "branch", "fetch", "stash", "worktree":
		return true
	default:
		return false
	}
}
