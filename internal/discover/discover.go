// Package discover provides command discovery and auto-rewrite functionality.
// This package implements RTK-style command rewriting for transparent tokman integration.
package discover

import (
	"regexp"
	"strings"
	"sync"

	"github.com/GrayCodeAI/tokman/internal/telemetry"
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

// Common rewrite patterns for RTK-style auto-rewrite
var rewritePatterns = []CommandPattern{
	// Test runners - high priority
	{Name: "cargo test", Pattern: regexp.MustCompile(`^cargo\s+test`), Rewrite: "tokman cargo test", Description: "Rust tests", Priority: 100},
	{Name: "go test", Pattern: regexp.MustCompile(`^go\s+test`), Rewrite: "tokman go test", Description: "Go tests", Priority: 100},
	{Name: "npm test", Pattern: regexp.MustCompile(`^npm\s+test`), Rewrite: "tokman npm test", Description: "npm tests", Priority: 100},
	{Name: "pnpm test", Pattern: regexp.MustCompile(`^pnpm\s+test`), Rewrite: "tokman pnpm test", Description: "pnpm tests", Priority: 100},
	{Name: "pytest", Pattern: regexp.MustCompile(`^pytest`), Rewrite: "tokman pytest", Description: "Python tests", Priority: 100},
	{Name: "vitest", Pattern: regexp.MustCompile(`^vitest|^npx\s+vitest`), Rewrite: "tokman vitest", Description: "Vitest tests", Priority: 100},
	{Name: "jest", Pattern: regexp.MustCompile(`^jest|^npx\s+jest`), Rewrite: "tokman jest", Description: "Jest tests", Priority: 100},
	{Name: "playwright", Pattern: regexp.MustCompile(`^playwright\s+test|^npx\s+playwright\s+test`), Rewrite: "tokman playwright", Description: "Playwright tests", Priority: 100},
	{Name: "rspec", Pattern: regexp.MustCompile(`^rspec`), Rewrite: "tokman rspec", Description: "RSpec tests", Priority: 100},
	{Name: "rake test", Pattern: regexp.MustCompile(`^rake\s+test`), Rewrite: "tokman rake test", Description: "Rake tests", Priority: 100},

	// Build commands
	{Name: "cargo build", Pattern: regexp.MustCompile(`^cargo\s+build`), Rewrite: "tokman cargo build", Description: "Rust build", Priority: 90},
	{Name: "cargo clippy", Pattern: regexp.MustCompile(`^cargo\s+clippy`), Rewrite: "tokman cargo clippy", Description: "Rust lint", Priority: 90},
	{Name: "npm run build", Pattern: regexp.MustCompile(`^npm\s+run\s+build`), Rewrite: "tokman err npm run build", Description: "npm build", Priority: 90},
	{Name: "pnpm build", Pattern: regexp.MustCompile(`^pnpm\s+(run\s+)?build`), Rewrite: "tokman err pnpm build", Description: "pnpm build", Priority: 90},
	{Name: "tsc", Pattern: regexp.MustCompile(`^tsc`), Rewrite: "tokman tsc", Description: "TypeScript compiler", Priority: 90},
	{Name: "next build", Pattern: regexp.MustCompile(`^next\s+build`), Rewrite: "tokman next build", Description: "Next.js build", Priority: 90},
	{Name: "golangci-lint", Pattern: regexp.MustCompile(`^golangci-lint`), Rewrite: "tokman golangci-lint", Description: "Go linter", Priority: 90},
	{Name: "ruff", Pattern: regexp.MustCompile(`^ruff\s+(check|format)`), Rewrite: "tokman ruff", Description: "Python linter", Priority: 90},

	// Git commands
	{Name: "git status", Pattern: regexp.MustCompile(`^git\s+status`), Rewrite: "tokman git status", Description: "Git status", Priority: 80},
	{Name: "git log", Pattern: regexp.MustCompile(`^git\s+log`), Rewrite: "tokman git log", Description: "Git log", Priority: 80},
	{Name: "git diff", Pattern: regexp.MustCompile(`^git\s+diff`), Rewrite: "tokman git diff", Description: "Git diff", Priority: 80},
	{Name: "git add", Pattern: regexp.MustCompile(`^git\s+add`), Rewrite: "tokman git add", Description: "Git add", Priority: 80},
	{Name: "git commit", Pattern: regexp.MustCompile(`^git\s+commit`), Rewrite: "tokman git commit", Description: "Git commit", Priority: 80},
	{Name: "git push", Pattern: regexp.MustCompile(`^git\s+push`), Rewrite: "tokman git push", Description: "Git push", Priority: 80},
	{Name: "git pull", Pattern: regexp.MustCompile(`^git\s+pull`), Rewrite: "tokman git pull", Description: "Git pull", Priority: 80},

	// GitHub CLI
	{Name: "gh pr", Pattern: regexp.MustCompile(`^gh\s+pr`), Rewrite: "tokman gh pr", Description: "GitHub PR", Priority: 80},
	{Name: "gh issue", Pattern: regexp.MustCompile(`^gh\s+issue`), Rewrite: "tokman gh issue", Description: "GitHub issue", Priority: 80},
	{Name: "gh run", Pattern: regexp.MustCompile(`^gh\s+run`), Rewrite: "tokman gh run", Description: "GitHub Actions", Priority: 80},

	// Docker/Kubernetes
	{Name: "docker ps", Pattern: regexp.MustCompile(`^docker\s+ps`), Rewrite: "tokman docker ps", Description: "Docker containers", Priority: 70},
	{Name: "docker images", Pattern: regexp.MustCompile(`^docker\s+images`), Rewrite: "tokman docker images", Description: "Docker images", Priority: 70},
	{Name: "docker logs", Pattern: regexp.MustCompile(`^docker\s+(logs|compose\s+logs)`), Rewrite: "tokman docker logs", Description: "Docker logs", Priority: 70},
	{Name: "kubectl", Pattern: regexp.MustCompile(`^kubectl\s+(get|logs|describe)`), Rewrite: "tokman kubectl", Description: "Kubernetes", Priority: 70},

	// Package managers
	{Name: "npm ls", Pattern: regexp.MustCompile(`^npm\s+ls`), Rewrite: "tokman npm ls", Description: "npm list", Priority: 60},
	{Name: "pnpm list", Pattern: regexp.MustCompile(`^pnpm\s+list`), Rewrite: "tokman pnpm list", Description: "pnpm list", Priority: 60},
	{Name: "pip list", Pattern: regexp.MustCompile(`^pip\s+list`), Rewrite: "tokman pip list", Description: "pip list", Priority: 60},
	{Name: "bundle install", Pattern: regexp.MustCompile(`^bundle\s+install`), Rewrite: "tokman bundle install", Description: "Bundle install", Priority: 60},

	// System commands
	{Name: "ls", Pattern: regexp.MustCompile(`^ls\b`), Rewrite: "tokman ls", Description: "List directory", Priority: 50},
	{Name: "tree", Pattern: regexp.MustCompile(`^tree`), Rewrite: "tokman tree", Description: "Directory tree", Priority: 50},
	{Name: "cat", Pattern: regexp.MustCompile(`^cat\s`), Rewrite: "tokman read", Description: "Read file", Priority: 50},
	{Name: "grep", Pattern: regexp.MustCompile(`^grep\s`), Rewrite: "tokman grep", Description: "Search files", Priority: 50},
	{Name: "find", Pattern: regexp.MustCompile(`^find\s`), Rewrite: "tokman find", Description: "Find files", Priority: 50},
	{Name: "wc", Pattern: regexp.MustCompile(`^wc\s`), Rewrite: "tokman wc", Description: "Word count", Priority: 50},
	{Name: "env", Pattern: regexp.MustCompile(`^env$`), Rewrite: "tokman env", Description: "Environment variables", Priority: 50},
}

// RewriteOptions provides options for command rewriting
type RewriteOptions struct {
	// DisableTestRunner disables automatic test-runner detection
	DisableTestRunner bool
	// PreferExplicit prefers explicit tokman commands over test-runner
	PreferExplicit bool
	// DisableCache disables caching for this rewrite
	DisableCache bool
}

// RewriteCommand rewrites a command using tokman equivalents (RTK-style).
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

	// Skip if already starts with tokman
	if strings.HasPrefix(trimmed, "tokman ") {
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

			return rewritten, true
		}
	}

	// Try test-runner auto-detection for unknown test commands
	if !options.DisableTestRunner {
		if testCmd := detectTestRunner(trimmed); testCmd != "" {
			rewritten := "tokman test-runner " + trimmed

			// Cache the result
			if !options.DisableCache {
				rewriteCache.Lock()
				rewriteCache.data[trimmed] = rewriteCacheEntry{rewritten, true}
				rewriteCache.Unlock()
			}

			telemetry.TrackRewrite(cmd, rewritten, true)
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
	return "tokman test-runner " + cmd
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
	if cmd == "" {
		return false
	}

	trimmed := strings.TrimSpace(cmd)
	if strings.HasPrefix(trimmed, "tokman ") {
		return false
	}

	for _, pattern := range rewritePatterns {
		if pattern.Pattern.MatchString(trimmed) {
			return true
		}
	}

	return false
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
