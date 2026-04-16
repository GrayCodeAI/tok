package discover

import (
	"fmt"
	"testing"
)

func BenchmarkRewriteCommand(b *testing.B) {
	tests := []string{
		"git status",
		"cargo test",
		"npm test",
		"go test ./...",
		"docker ps -a",
		"ls -la",
		"pytest -v",
		"unknown-command",
		"tokman git status",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, cmd := range tests {
			RewriteCommand(cmd, nil)
		}
	}
}

func BenchmarkRewriteCommandWithOptions(b *testing.B) {
	opts := &RewriteOptions{
		DisableTestRunner: true,
		PreferExplicit:    true,
	}

	tests := []string{
		"cargo test",
		"npm test",
		"go test ./...",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, cmd := range tests {
			RewriteCommand(cmd, opts)
		}
	}
}

func BenchmarkDetectCommand(b *testing.B) {
	tests := []string{
		"git status",
		"cargo test",
		"npm test",
		"unknown-command",
		"tokman git status",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, cmd := range tests {
			DetectCommand(cmd)
		}
	}
}

func BenchmarkIsTestCommand(b *testing.B) {
	tests := []string{
		"cargo test",
		"go test ./...",
		"npm test",
		"pytest",
		"vitest run",
		"jest",
		"rspec",
		"rake test",
		"playwright test",
		"git status",
		"docker ps",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, cmd := range tests {
			isTestCommand(cmd)
		}
	}
}

func BenchmarkShouldRewriteFile(b *testing.B) {
	tests := []string{
		"foo_test.go",
		"foo_spec.rb",
		"foo.test.js",
		"foo.spec.js",
		"test_foo.py",
		"FooTest.java",
		"foo_test.rs",
		"foo.go",
		"foo.js",
		"README.md",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, file := range tests {
			ShouldRewriteFile(file)
		}
	}
}

func BenchmarkKnownCommands(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		KnownCommands()
	}
}

func BenchmarkRewriteCommandParallel(b *testing.B) {
	cmds := []string{
		"git status",
		"cargo test",
		"npm test",
		"go test ./...",
	}

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			RewriteCommand(cmds[i%len(cmds)], nil)
			i++
		}
	})
}

func BenchmarkRewriteCommandWithCaching(b *testing.B) {
	// Simulate caching by pre-warming
	cache := make(map[string]struct {
		rewritten string
		changed   bool
	})

	cmds := []string{
		"git status",
		"cargo test",
		"npm test",
		"go test ./...",
	}

	// Pre-populate cache
	for _, cmd := range cmds {
		rewritten, changed := RewriteCommand(cmd, nil)
		cache[cmd] = struct {
			rewritten string
			changed   bool
		}{rewritten, changed}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cmd := cmds[i%len(cmds)]
		if cached, ok := cache[cmd]; ok {
			_ = cached.rewritten
			_ = cached.changed
		} else {
			RewriteCommand(cmd, nil)
		}
	}
}

func BenchmarkRewriteDifferentPatterns(b *testing.B) {
	patterns := []struct {
		name string
		cmd  string
	}{
		{"git", "git status"},
		{"cargo", "cargo test"},
		{"npm", "npm test"},
		{"docker", "docker ps"},
		{"ls", "ls -la"},
		{"pytest", "pytest"},
		{"unknown", "unknown-cmd"},
	}

	for _, p := range patterns {
		b.Run(p.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				RewriteCommand(p.cmd, nil)
			}
		})
	}
}

func BenchmarkRewriteCommandMemoryAllocation(b *testing.B) {
	cmd := "cargo test --verbose --release --features full"

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RewriteCommand(cmd, nil)
	}
}

func BenchmarkDetectCommandWithManyChecks(b *testing.B) {
	// Test performance when checking many commands
	cmds := make([]string, 100)
	for i := 0; i < 100; i++ {
		if i%2 == 0 {
			cmds[i] = fmt.Sprintf("git command-%d", i)
		} else {
			cmds[i] = fmt.Sprintf("unknown-cmd-%d", i)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, cmd := range cmds {
			DetectCommand(cmd)
		}
	}
}
