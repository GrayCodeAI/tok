package discover

import (
	"testing"
)

func TestRewriteCommand(t *testing.T) {
	tests := []struct {
		name     string
		cmd      string
		opts     interface{}
		wantCmd  string
		wantBool bool
	}{
		// Test commands that should be rewritten
		{
			name:     "git status",
			cmd:      "git status",
			opts:     nil,
			wantCmd:  "tokman git status",
			wantBool: true,
		},
		{
			name:     "cargo test",
			cmd:      "cargo test",
			opts:     nil,
			wantCmd:  "tokman test-runner cargo test",
			wantBool: true,
		},
		{
			name:     "npm test",
			cmd:      "npm test",
			opts:     nil,
			wantCmd:  "tokman test-runner npm test",
			wantBool: true,
		},
		{
			name:     "go test",
			cmd:      "go test ./...",
			opts:     nil,
			wantCmd:  "tokman test-runner go test ./...",
			wantBool: true,
		},
		{
			name:     "pytest",
			cmd:      "pytest -v",
			opts:     nil,
			wantCmd:  "tokman test-runner pytest -v",
			wantBool: true,
		},
		{
			name:     "docker ps",
			cmd:      "docker ps -a",
			opts:     nil,
			wantCmd:  "tokman docker ps -a",
			wantBool: true,
		},
		{
			name:     "ls command",
			cmd:      "ls -la",
			opts:     nil,
			wantCmd:  "tokman ls -la",
			wantBool: true,
		},
		// Commands that should NOT be rewritten
		{
			name:     "already tokman prefixed",
			cmd:      "tokman git status",
			opts:     nil,
			wantCmd:  "tokman git status",
			wantBool: false,
		},
		{
			name:     "unknown command",
			cmd:      "some-random-command",
			opts:     nil,
			wantCmd:  "some-random-command",
			wantBool: false,
		},
		{
			name:     "empty command",
			cmd:      "",
			opts:     nil,
			wantCmd:  "",
			wantBool: false,
		},
		{
			name:     "git without subcommand",
			cmd:      "git",
			opts:     nil,
			wantCmd:  "git",
			wantBool: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCmd, gotBool := RewriteCommand(tt.cmd, tt.opts)
			if gotCmd != tt.wantCmd {
				t.Errorf("RewriteCommand() gotCmd = %v, want %v", gotCmd, tt.wantCmd)
			}
			if gotBool != tt.wantBool {
				t.Errorf("RewriteCommand() gotBool = %v, want %v", gotBool, tt.wantBool)
			}
		})
	}
}

func TestRewriteCommandWithOptions(t *testing.T) {
	// Clear cache before running tests
	ClearCache()

	tests := []struct {
		name     string
		cmd      string
		opts     *RewriteOptions
		wantCmd  string
		wantBool bool
	}{
		{
			name: "disable test runner",
			cmd:  "cargo test",
			opts: &RewriteOptions{DisableTestRunner: true},
			// When test runner is disabled, it should use explicit tokman cargo test
			wantCmd:  "tokman cargo test",
			wantBool: true,
		},
		{
			name:     "prefer explicit",
			cmd:      "npm test",
			opts:     &RewriteOptions{PreferExplicit: true},
			wantCmd:  "tokman npm test",
			wantBool: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear cache for each test to avoid cross-test pollution
			ClearCache()

			gotCmd, gotBool := RewriteCommand(tt.cmd, tt.opts)
			if gotCmd != tt.wantCmd {
				t.Errorf("RewriteCommand() gotCmd = %v, want %v", gotCmd, tt.wantCmd)
			}
			if gotBool != tt.wantBool {
				t.Errorf("RewriteCommand() gotBool = %v, want %v", gotBool, tt.wantBool)
			}
		})
	}
}

func TestDetectCommand(t *testing.T) {
	tests := []struct {
		name string
		cmd  string
		want bool
	}{
		{
			name: "known command - git status",
			cmd:  "git status",
			want: true,
		},
		{
			name: "known command - cargo test",
			cmd:  "cargo test",
			want: true,
		},
		{
			name: "known command - docker ps",
			cmd:  "docker ps",
			want: true,
		},
		{
			name: "passthrough command - terraform plan",
			cmd:  "terraform plan",
			want: true,
		},
		{
			name: "unknown command",
			cmd:  "unknown-cmd",
			want: false,
		},
		{
			name: "empty command",
			cmd:  "",
			want: false,
		},
		{
			name: "already tokman prefixed",
			cmd:  "tokman git status",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DetectCommand(tt.cmd); got != tt.want {
				t.Errorf("DetectCommand() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClassifyCommand(t *testing.T) {
	tests := []struct {
		name      string
		cmd       string
		wantCmd   string
		wantLevel SupportLevel
	}{
		{
			name:      "optimized command",
			cmd:       "git status",
			wantCmd:   "tokman git status",
			wantLevel: SupportOptimized,
		},
		{
			name:      "passthrough command",
			cmd:       "terraform plan",
			wantCmd:   "tokman terraform plan",
			wantLevel: SupportPassthrough,
		},
		{
			name:      "unsupported command",
			cmd:       "unknown-cmd foo",
			wantCmd:   "unknown-cmd foo",
			wantLevel: SupportUnsupported,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCmd, gotLevel := ClassifyCommand(tt.cmd)
			if gotCmd != tt.wantCmd {
				t.Fatalf("ClassifyCommand() cmd = %q, want %q", gotCmd, tt.wantCmd)
			}
			if gotLevel != tt.wantLevel {
				t.Fatalf("ClassifyCommand() level = %q, want %q", gotLevel, tt.wantLevel)
			}
		})
	}
}

func TestKnownCommands(t *testing.T) {
	commands := KnownCommands()

	if commands == nil {
		t.Error("KnownCommands() returned nil")
	}

	// Should have many commands now
	if len(commands) == 0 {
		t.Error("KnownCommands() returned empty slice")
	}

	// Check for some expected commands
	expected := map[string]bool{
		"git status": false,
		"cargo test": false,
		"docker ps":  false,
	}

	for _, cmd := range commands {
		if _, ok := expected[cmd]; ok {
			expected[cmd] = true
		}
	}

	for cmd, found := range expected {
		if !found {
			t.Errorf("Expected command %q not found in KnownCommands()", cmd)
		}
	}
}

func TestIsTestCommand(t *testing.T) {
	tests := []struct {
		name string
		cmd  string
		want bool
	}{
		{"cargo test", "cargo test", true},
		{"go test", "go test ./...", true},
		{"npm test", "npm test", true},
		{"pytest", "pytest -v", true},
		{"vitest", "vitest run", true},
		{"jest", "jest --watch", true},
		{"rspec", "rspec spec/", true},
		{"rake test", "rake test", true},
		{"playwright", "playwright test", true},
		{"git status", "git status", false},
		{"docker ps", "docker ps", false},
		{"ls", "ls -la", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isTestCommand(tt.cmd); got != tt.want {
				t.Errorf("isTestCommand(%q) = %v, want %v", tt.cmd, got, tt.want)
			}
		})
	}
}

func TestShouldRewriteFile(t *testing.T) {
	tests := []struct {
		name string
		file string
		want bool
	}{
		{"Go test file", "foo_test.go", true},
		{"Ruby spec", "foo_spec.rb", true},
		{"JS test", "foo.test.js", true},
		{"JS spec", "foo.spec.js", true},
		{"Python test", "test_foo.py", true},
		{"Java test", "FooTest.java", true},
		{"Rust test", "foo_test.rs", true},
		{"Regular Go file", "foo.go", false},
		{"Regular JS file", "foo.js", false},
		{"README", "README.md", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ShouldRewriteFile(tt.file); got != tt.want {
				t.Errorf("ShouldRewriteFile(%q) = %v, want %v", tt.file, got, tt.want)
			}
		})
	}
}

func TestRewriteCommandCaching(t *testing.T) {
	// Clear cache before test
	ClearCache()

	// First call should be a cache miss
	_, _ = RewriteCommand("git status", nil)

	hits, misses := GetCacheStats()
	if hits != 0 {
		t.Errorf("Expected 0 hits after first call, got %d", hits)
	}
	if misses != 1 {
		t.Errorf("Expected 1 miss after first call, got %d", misses)
	}

	// Second call should be a cache hit
	_, _ = RewriteCommand("git status", nil)

	hits, misses = GetCacheStats()
	if hits != 1 {
		t.Errorf("Expected 1 hit after second call, got %d", hits)
	}
	if misses != 1 {
		t.Errorf("Expected 1 miss (unchanged), got %d", misses)
	}

	// Clear cache
	ClearCache()
	hits, misses = GetCacheStats()
	if hits != 0 || misses != 0 {
		t.Errorf("Expected 0 hits and 0 misses after clear, got %d hits, %d misses", hits, misses)
	}
}

func TestRewriteCommandDisableCache(t *testing.T) {
	// Clear cache
	ClearCache()

	opts := &RewriteOptions{DisableCache: true}

	// Multiple calls with cache disabled should all be misses
	for i := 0; i < 5; i++ {
		_, _ = RewriteCommand("git status", opts)
	}

	hits, misses := GetCacheStats()
	if hits != 0 {
		t.Errorf("Expected 0 hits with cache disabled, got %d", hits)
	}
	if misses != 5 {
		t.Errorf("Expected 5 misses with cache disabled, got %d", misses)
	}
}

func TestRewriteCommandCacheConsistency(t *testing.T) {
	// Clear cache
	ClearCache()

	// Rewrite a command
	rewritten1, changed1 := RewriteCommand("cargo test", nil)

	// Same command should return cached result
	rewritten2, changed2 := RewriteCommand("cargo test", nil)

	if rewritten1 != rewritten2 {
		t.Errorf("Cached result mismatch: %q vs %q", rewritten1, rewritten2)
	}
	if changed1 != changed2 {
		t.Errorf("Cached changed status mismatch: %v vs %v", changed1, changed2)
	}

	// Check cache was used
	hits, _ := GetCacheStats()
	if hits < 1 {
		t.Errorf("Expected at least 1 cache hit, got %d", hits)
	}
}
