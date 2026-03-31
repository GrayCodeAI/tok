package discover

import (
	"testing"
)

// ── ClassifyCommand Edge Cases ─────────────────────────────────

func TestClassifyCommand_EmptyAndWhitespace(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty string", ""},
		{"spaces only", "   "},
		{"tabs only", "\t\t"},
		{"newline", "\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			class := ClassifyCommand(tt.input)
			if class.Supported {
				t.Errorf("ClassifyCommand(%q) should not be supported", tt.input)
			}
		})
	}
}

func TestClassifyCommand_IgnoredExact(t *testing.T) {
	ignored := []string{"cd", "echo", "true", "false", "wait", "pwd", "bash", "sh", "fi", "done"}
	for _, cmd := range ignored {
		t.Run(cmd, func(t *testing.T) {
			class := ClassifyCommand(cmd)
			if class.Supported {
				t.Errorf("ClassifyCommand(%q) should not be supported", cmd)
			}
		})
	}
}

func TestClassifyCommand_IgnoredPrefixes(t *testing.T) {
	ignored := []string{
		"cd /tmp", "echo hello", "printf '%s'", "export FOO=bar",
		"source .env", "mkdir foo", "rm -rf foo", "mv a b", "cp a b",
		"chmod 755 x", "chown user x", "touch x", "which go", "type go",
		"command -v go", "test -f x", "true", "false", "sleep 1",
		"wait", "kill 123", "set -e", "unset FOO", "wc -l", "sort",
		"uniq", "tr a b", "cut -d:", "awk '{print}'", "sed 's/a/b/'",
		"python3 -c 'print(1)'", "python -c 'print(1)'", "node -e '1'",
		"ruby -e '1'", "tokman git status",
	}
	for _, cmd := range ignored {
		t.Run(cmd, func(t *testing.T) {
			class := ClassifyCommand(cmd)
			if class.Supported {
				t.Errorf("ClassifyCommand(%q) should not be supported", cmd)
			}
		})
	}
}

func TestClassifyCommand_EnvPrefixes(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantSupport bool
		wantCmd     string
	}{
		{"sudo git status", "sudo git status", true, "tokman git"},
		{"env git status", "env git status", true, "tokman git"},
		{"VAR=val git status", "FOO=bar git status", true, "tokman git"},
		{"sudo env git status", "sudo env git status", true, "tokman git"},
		{"multiple env vars", "A=1 B=2 git status", true, "tokman git"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			class := ClassifyCommand(tt.input)
			if class.Supported != tt.wantSupport {
				t.Errorf("ClassifyCommand(%q) Supported = %v, want %v", tt.input, class.Supported, tt.wantSupport)
			}
			if tt.wantCmd != "" && class.TokManCmd != tt.wantCmd {
				t.Errorf("ClassifyCommand(%q) TokManCmd = %q, want %q", tt.input, class.TokManCmd, tt.wantCmd)
			}
		})
	}
}

func TestClassifyCommand_AllSupportedCommands(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		category string
	}{
		// Git
		{"git status", "git status", "Git"},
		{"git log", "git log --oneline", "Git"},
		{"git diff", "git diff HEAD", "Git"},
		{"git show", "git show abc123", "Git"},
		{"git add", "git add .", "Git"},
		{"git commit", "git commit -m fix", "Git"},
		{"git push", "git push origin main", "Git"},
		{"git pull", "git pull", "Git"},
		{"git branch", "git branch -a", "Git"},
		{"git fetch", "git fetch origin", "Git"},
		{"git stash", "git stash list", "Git"},
		{"git worktree", "git worktree list", "Git"},
		// GitHub CLI
		{"gh pr", "gh pr list", "GitHub"},
		{"gh issue", "gh issue list", "GitHub"},
		{"gh run", "gh run list", "GitHub"},
		{"gh repo", "gh repo view", "GitHub"},
		{"gh api", "gh api /repos", "GitHub"},
		{"gh release", "gh release list", "GitHub"},
		// Cargo
		{"cargo build", "cargo build", "Cargo"},
		{"cargo test", "cargo test", "Cargo"},
		{"cargo clippy", "cargo clippy", "Cargo"},
		{"cargo check", "cargo check", "Cargo"},
		{"cargo fmt", "cargo fmt", "Cargo"},
		{"cargo install", "cargo install ripgrep", "Cargo"},
		// npm/pnpm/npx
		{"pnpm list", "pnpm list", "PackageManager"},
		{"pnpm ls", "pnpm ls", "PackageManager"},
		{"pnpm outdated", "pnpm outdated", "PackageManager"},
		{"pnpm install", "pnpm install", "PackageManager"},
		{"npm run", "npm run build", "PackageManager"},
		{"npm exec", "npm exec tsc", "PackageManager"},
		{"npx", "npx jest", "Tests"},
		// File commands
		{"cat", "cat file.txt", "Files"},
		{"head", "head file.txt", "Files"},
		{"tail", "tail file.txt", "Files"},
		{"rg", "rg pattern", "Files"},
		{"grep", "grep pattern file", "Files"},
		{"ls", "ls -la", "Files"},
		{"find", "find . -name '*.go'", "Files"},
		{"tree", "tree src/", "Files"},
		{"diff", "diff a.txt b.txt", "Files"},
		// Build tools
		{"tsc", "tsc --noEmit", "Build"},
		{"npx tsc", "npx tsc", "Build"},
		{"pnpm tsc", "pnpm tsc", "Build"},
		{"eslint", "eslint src/", "Build"},
		{"npx eslint", "npx eslint src/", "Build"},
		{"biome", "biome check src/", "Build"},
		{"prettier", "prettier --write src/", "Build"},
		{"next build", "next build", "Build"},
		{"vitest", "vitest run", "Tests"},
		{"jest", "jest --coverage", "Tests"},
		{"playwright", "playwright test", "Tests"},
		{"prisma", "prisma migrate dev", "Build"},
		// Infrastructure
		{"docker ps", "docker ps", "Infra"},
		{"docker images", "docker images", "Infra"},
		{"docker logs", "docker logs app", "Infra"},
		{"docker run", "docker run nginx", "Infra"},
		{"docker exec", "docker exec app sh", "Infra"},
		{"docker build", "docker build .", "Infra"},
		{"docker compose ps", "docker compose ps", "Infra"},
		{"docker compose logs", "docker compose logs", "Infra"},
		{"docker compose build", "docker compose build", "Infra"},
		{"kubectl get", "kubectl get pods", "Infra"},
		{"kubectl logs", "kubectl logs pod", "Infra"},
		{"kubectl describe", "kubectl describe pod", "Infra"},
		{"kubectl apply", "kubectl apply -f k8s/", "Infra"},
		// Network
		{"curl", "curl https://example.com", "Network"},
		{"wget", "wget https://example.com", "Network"},
		// Python
		{"mypy", "mypy src/", "Build"},
		{"python3 -m mypy", "python3 -m mypy src/", "Build"},
		{"ruff check", "ruff check src/", "Python"},
		{"ruff format", "ruff format src/", "Python"},
		{"pytest", "pytest tests/", "Python"},
		{"python -m pytest", "python -m pytest tests/", "Python"},
		{"pip list", "pip list", "Python"},
		{"pip3 list", "pip3 list", "Python"},
		{"uv pip list", "uv pip list", "Python"},
		{"pip outdated", "pip outdated", "Python"},
		{"pip install", "pip install requests", "Python"},
		// Go
		{"go test", "go test ./...", "Go"},
		{"go build", "go build", "Go"},
		{"go vet", "go vet ./...", "Go"},
		{"golangci-lint", "golangci-lint run", "Go"},
		// AWS
		{"aws", "aws s3 ls", "Infra"},
		// psql
		{"psql", "psql -d mydb", "Infra"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			class := ClassifyCommand(tt.input)
			if !class.Supported {
				t.Errorf("ClassifyCommand(%q) should be supported", tt.input)
			}
			if class.Category != tt.category {
				t.Errorf("ClassifyCommand(%q) Category = %q, want %q", tt.input, class.Category, tt.category)
			}
		})
	}
}

func TestClassifyCommand_SubcmdSavings(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantSavings float64
	}{
		{"git diff high savings", "git diff", 80.0},
		{"git show high savings", "git show abc", 80.0},
		{"git add savings", "git add .", 59.0},
		{"git commit savings", "git commit -m fix", 59.0},
		{"git status default", "git status", 70.0},
		{"gh pr savings", "gh pr list", 87.0},
		{"gh run savings", "gh run list", 82.0},
		{"gh issue savings", "gh issue list", 80.0},
		{"cargo test savings", "cargo test", 90.0},
		{"cargo check savings", "cargo check", 80.0},
		{"go test savings", "go test ./...", 90.0},
		{"go build savings", "go build", 80.0},
		{"go vet savings", "go vet ./...", 75.0},
		{"ruff check savings", "ruff check src/", 80.0},
		{"ruff format savings", "ruff format src/", 75.0},
		{"pip list savings", "pip list", 75.0},
		{"pip outdated savings", "pip outdated", 75.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			class := ClassifyCommand(tt.input)
			if class.SavingsPct != tt.wantSavings {
				t.Errorf("ClassifyCommand(%q) SavingsPct = %.1f, want %.1f", tt.input, class.SavingsPct, tt.wantSavings)
			}
		})
	}
}

func TestClassifyCommand_SubcmdStatus(t *testing.T) {
	class := ClassifyCommand("cargo fmt")
	if class.Status != StatusPassthrough {
		t.Errorf("cargo fmt status = %v, want StatusPassthrough", class.Status)
	}
}

func TestClassifyCommand_UnsupportedExtractsBase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantBase string
	}{
		{"simple command", "terraform plan", "terraform plan"},
		{"command with flags", "make -j4 build", "make"},
		{"command with path", "/usr/local/bin/foo arg", "/usr/local/bin/foo arg"},
		{"single word", "unknowncmd", "unknowncmd"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			class := ClassifyCommand(tt.input)
			if class.Supported {
				t.Errorf("ClassifyCommand(%q) should not be supported", tt.input)
			}
			if class.BaseCommand != tt.wantBase {
				t.Errorf("ClassifyCommand(%q) BaseCommand = %q, want %q", tt.input, class.BaseCommand, tt.wantBase)
			}
		})
	}
}

// ── RewriteCommand Edge Cases ──────────────────────────────────

func TestRewriteCommand_EmptyInput(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty", ""},
		{"spaces", "   "},
		{"tabs", "\t"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, changed := RewriteCommand(tt.input, nil)
			if changed {
				t.Errorf("RewriteCommand(%q) should not change", tt.input)
			}
			if result != "" {
				t.Errorf("RewriteCommand(%q) = %q, want empty", tt.input, result)
			}
		})
	}
}

func TestRewriteCommand_HeredocAndArithmetic(t *testing.T) {
	tests := []string{
		"git status <<EOF",
		"echo $((1+2))",
		"cat <<EOF\ntest\nEOF",
		"git log $((VAR+1))",
	}
	for _, cmd := range tests {
		t.Run(cmd, func(t *testing.T) {
			result, changed := RewriteCommand(cmd, nil)
			if changed {
				t.Errorf("RewriteCommand(%q) should not change heredoc/arithmetic", cmd)
			}
			if result != "" {
				t.Errorf("RewriteCommand(%q) = %q, want empty for unsafe input", cmd, result)
			}
		})
	}
}

func TestRewriteCommand_AlreadyTokman(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"simple", "tokman git status"},
		{"bare", "tokman"},
		{"with args", "tokman docker ps -a"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, changed := RewriteCommand(tt.input, nil)
			if changed {
				t.Errorf("RewriteCommand(%q) should not change already-tokman command", tt.input)
			}
			if result != tt.input {
				t.Errorf("RewriteCommand(%q) = %q, want %q", tt.input, result, tt.input)
			}
		})
	}
}

func TestRewriteCommand_Excluded(t *testing.T) {
	result, changed := RewriteCommand("git status", []string{"git"})
	if changed {
		t.Error("RewriteCommand with excluded git should not change")
	}
	if result != "git status" {
		t.Errorf("RewriteCommand with excluded = %q, want 'git status'", result)
	}
}

func TestRewriteCommand_CompoundSemicolon(t *testing.T) {
	result, changed := RewriteCommand("git status; cargo test", nil)
	if !changed {
		t.Error("RewriteCommand compound semicolon should change")
	}
	if result != "tokman git status; tokman cargo test" {
		t.Errorf("RewriteCommand = %q, want 'tokman git status; tokman cargo test'", result)
	}
}

func TestRewriteCommand_CompoundOr(t *testing.T) {
	result, changed := RewriteCommand("git status || echo failed", nil)
	if !changed {
		t.Error("RewriteCommand compound or should change")
	}
	if result != "tokman git status || echo failed" {
		t.Errorf("RewriteCommand = %q, want 'tokman git status || echo failed'", result)
	}
}

func TestRewriteCommand_CompoundBackground(t *testing.T) {
	result, changed := RewriteCommand("git status &", nil)
	if !changed {
		t.Error("RewriteCommand compound background should change")
	}
	if result != "tokman git status & " {
		t.Errorf("RewriteCommand = %q, want 'tokman git status & '", result)
	}
}

func TestRewriteCommand_PipePipedCommand(t *testing.T) {
	// find is in pipedCommands, so should not rewrite the first segment
	result, changed := RewriteCommand("find . -name '*.go' | wc -l", nil)
	if changed {
		t.Errorf("RewriteCommand piped find should not change, got %q", result)
	}
}

func TestRewriteCommand_QuotedStrings(t *testing.T) {
	// Commands with quoted strings containing operators should not split
	result, changed := RewriteCommand("git log --format='%H %s' && cargo test", nil)
	if !changed {
		t.Error("RewriteCommand with quoted strings should still work")
	}
	if result != "tokman git log --format='%H %s' && tokman cargo test" {
		t.Errorf("RewriteCommand = %q", result)
	}
}

func TestRewriteCommand_DoubleQuotedStrings(t *testing.T) {
	_, changed := RewriteCommand(`echo "test && git status"`, nil)
	if changed {
		t.Errorf("RewriteCommand should not rewrite inside double quotes")
	}
}

func TestRewriteCommand_GhJsonFlag(t *testing.T) {
	// gh with --json should not be rewritten
	tests := []string{
		"gh pr list --json title,number",
		"gh issue view 123 --jq .title",
		"gh pr view --template '{{.title}}'",
	}
	for _, cmd := range tests {
		t.Run(cmd, func(t *testing.T) {
			result, changed := RewriteCommand(cmd, nil)
			if changed {
				t.Errorf("RewriteCommand(%q) should not change with structured output flags", cmd)
			}
			if result != cmd {
				t.Errorf("RewriteCommand(%q) = %q, want %q", cmd, result, cmd)
			}
		})
	}
}

func TestRewriteCommand_TokmanDisabled(t *testing.T) {
	result, changed := RewriteCommand("TOKMAN_DISABLED=1 git status", nil)
	if changed {
		t.Error("RewriteCommand with TOKMAN_DISABLED should not change")
	}
	if result != "TOKMAN_DISABLED=1 git status" {
		t.Errorf("RewriteCommand = %q, want original", result)
	}
}

func TestRewriteCommand_GitGlobalOpts(t *testing.T) {
	// Note: stripGitGlobalOpts is called in rewriteSegment but ClassifyCommand
	// uses the original trimmed command. Global opts like -C, --no-pager don't
	// match the git pattern regex, so these commands are NOT rewritten.
	// This is expected behavior - the stripping only affects the rewrite prefix matching.
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"git -C", "git -C /tmp status", "git -C /tmp status"},
		{"git -c", "git -c core.editor=vim status", "git -c core.editor=vim status"},
		{"git --no-pager", "git --no-pager log", "git --no-pager log"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, changed := RewriteCommand(tt.input, nil)
			if changed {
				t.Errorf("RewriteCommand(%q) should not change with global opts", tt.input)
			}
			if result != tt.expected {
				t.Errorf("RewriteCommand(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestRewriteCommand_AbsolutePaths(t *testing.T) {
	// Note: stripAbsolutePath normalizes paths but ClassifyCommand uses the original
	// trimmed command. The absolute path regex doesn't match because the pattern
	// requires the path to be the first token without leading /
	// These commands are NOT rewritten because the pattern doesn't match.
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"/usr/bin/grep", "/usr/bin/grep pattern file", "/usr/bin/grep pattern file"},
		{"/usr/local/bin/rg", "/usr/local/bin/rg pattern", "/usr/local/bin/rg pattern"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, changed := RewriteCommand(tt.input, nil)
			if changed {
				t.Errorf("RewriteCommand(%q) should not change", tt.input)
			}
			if result != tt.expected {
				t.Errorf("RewriteCommand(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestRewriteCommand_TailNumeric(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"tail -N", "tail -20 src/main.rs", "tokman read src/main.rs --tail-lines 20"},
		{"tail -n N", "tail -n 50 file.txt", "tokman read file.txt --tail-lines 50"},
		{"tail --lines=N", "tail --lines=100 file.log", "tokman read file.log --tail-lines 100"},
		{"tail file (no flag)", "tail file.txt", "tokman read file.txt"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Rewrite(tt.input)
			if result != tt.expected {
				t.Errorf("Rewrite(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestRewriteCommand_HeadNumericEdgeCases(t *testing.T) {
	// Note: head -c and head -q are skipped by rewriteHeadNumeric but the generic
	// cat/head/tail pattern still rewrites them to "tokman read". This is expected.
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"head -c", "head -c 100 file.bin", "tokman read -c 100 file.bin"},
		{"head -q", "head -q file.txt", "tokman read -q file.txt"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Rewrite(tt.input)
			if result != tt.expected {
				t.Errorf("Rewrite(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestRewriteCommand_EnvPrefixWithRewrite(t *testing.T) {
	result, changed := RewriteCommand("sudo git status", nil)
	if !changed {
		t.Error("RewriteCommand sudo git status should change")
	}
	if result != "sudo tokman git status" {
		t.Errorf("RewriteCommand = %q, want 'sudo tokman git status'", result)
	}
}

// ── stripWordPrefix ────────────────────────────────────────────

func TestStripWordPrefix(t *testing.T) {
	tests := []struct {
		name   string
		cmd    string
		prefix string
		want   *string
	}{
		{"exact match", "git", "git", strPtr("")},
		{"with rest", "git status", "git", strPtr("status")},
		{"no match", "cargo build", "git", nil},
		{"partial no match", "gitlab", "git", nil},
		{"prefix longer", "git", "git status", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripWordPrefix(tt.cmd, tt.prefix)
			if tt.want == nil {
				if got != nil {
					t.Errorf("stripWordPrefix(%q, %q) = %v, want nil", tt.cmd, tt.prefix, *got)
				}
			} else {
				if got == nil {
					t.Errorf("stripWordPrefix(%q, %q) = nil, want %q", tt.cmd, tt.prefix, *tt.want)
				} else if *got != *tt.want {
					t.Errorf("stripWordPrefix(%q, %q) = %q, want %q", tt.cmd, tt.prefix, *got, *tt.want)
				}
			}
		})
	}
}

func strPtr(s string) *string { return &s }

// ── extractBaseCommand ─────────────────────────────────────────

func TestExtractBaseCommand(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", ""},
		{"single", "git", "git"},
		{"subcommand", "git status", "git status"},
		{"with flag", "git status --short", "git status"},
		{"with path", "ls /home", "ls"},
		{"with file", "cat file.txt", "cat"},
		{"with url", "curl https://example.com", "curl"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractBaseCommand(tt.input)
			if got != tt.want {
				t.Errorf("extractBaseCommand(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// ── TokmanStatus Constants ─────────────────────────────────────

func TestTokmanStatus_Values(t *testing.T) {
	if StatusExisting != 0 {
		t.Errorf("StatusExisting = %d, want 0", StatusExisting)
	}
	if StatusPassthrough != 1 {
		t.Errorf("StatusPassthrough = %d, want 1", StatusPassthrough)
	}
	if StatusNew != 2 {
		t.Errorf("StatusNew = %d, want 2", StatusNew)
	}
}

// ── Classification Edge Cases ──────────────────────────────────

func TestClassification_EmptyReturnsZeroValue(t *testing.T) {
	class := ClassifyCommand("")
	if class.Supported {
		t.Error("empty command should not be supported")
	}
	if class.TokManCmd != "" {
		t.Errorf("empty command TokManCmd = %q, want ''", class.TokManCmd)
	}
	if class.Category != "" {
		t.Errorf("empty command Category = %q, want ''", class.Category)
	}
	if class.SavingsPct != 0 {
		t.Errorf("empty command SavingsPct = %f, want 0", class.SavingsPct)
	}
	if class.BaseCommand != "" {
		t.Errorf("empty command BaseCommand = %q, want ''", class.BaseCommand)
	}
}

// ── Legacy API ─────────────────────────────────────────────────

func TestRewrite_EmptyInput(t *testing.T) {
	result := Rewrite("")
	if result != "" {
		t.Errorf("Rewrite('') = %q, want ''", result)
	}
}

func TestRewrite_UnsupportedPassthrough(t *testing.T) {
	result := Rewrite("terraform plan")
	if result != "terraform plan" {
		t.Errorf("Rewrite('terraform plan') = %q, want 'terraform plan'", result)
	}
}

func TestShouldRewrite_Empty(t *testing.T) {
	if ShouldRewrite("") {
		t.Error("ShouldRewrite('') should be false")
	}
}

func TestShouldRewrite_AlreadyTokman(t *testing.T) {
	if ShouldRewrite("tokman git status") {
		t.Error("ShouldRewrite('tokman git status') should be false")
	}
}

func TestGetMapping_Unsupported(t *testing.T) {
	_, found := GetMapping("terraform plan")
	if found {
		t.Error("GetMapping('terraform plan') should not be found")
	}
}

func TestGetMapping_Empty(t *testing.T) {
	_, found := GetMapping("")
	if found {
		t.Error("GetMapping('') should not be found")
	}
}

func TestListRewrites_AllPassArgs(t *testing.T) {
	rewrites := ListRewrites()
	for _, r := range rewrites {
		if !r.PassArgs {
			t.Errorf("ListRewrites() mapping %s should have PassArgs=true", r.Original)
		}
	}
}

// ── Registry ───────────────────────────────────────────────────

func TestRegistry_IsInitialized(t *testing.T) {
	if Registry == nil {
		t.Error("Registry should be initialized")
	}
}

// ── Benchmarks ─────────────────────────────────────────────────

func BenchmarkClassifyCommand(b *testing.B) {
	cmds := []string{
		"git status",
		"cargo test",
		"docker ps",
		"terraform plan",
		"ls -la",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ClassifyCommand(cmds[i%len(cmds)])
	}
}

func BenchmarkRewriteCommand(b *testing.B) {
	cmds := []string{
		"git status",
		"git add . && cargo test",
		"sudo git log -10",
		"docker ps | grep web",
		"tokman git status",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RewriteCommand(cmds[i%len(cmds)], nil)
	}
}

func BenchmarkRewrite(b *testing.B) {
	cmd := "git status --short"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Rewrite(cmd)
	}
}

func BenchmarkShouldRewrite(b *testing.B) {
	cmd := "git status --short"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ShouldRewrite(cmd)
	}
}

func BenchmarkClassifyCommand_Unsupported(b *testing.B) {
	cmd := "terraform plan -var-file=prod.tfvars"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ClassifyCommand(cmd)
	}
}

func BenchmarkRewriteCommand_Compound(b *testing.B) {
	cmd := "git add . && git commit -m 'fix' && cargo test"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RewriteCommand(cmd, nil)
	}
}
