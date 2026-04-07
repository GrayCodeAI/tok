package yamlfilter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMatchCommand(t *testing.T) {
	tests := []struct {
		match   string
		command string
		want    bool
	}{
		{`^go test`, "go test ./...", true},
		{`^go test`, "go build ./...", false},
		{`^git (status|diff)`, "git status", true},
		{`^git (status|diff)`, "git diff", true},
		{`^git (status|diff)`, "git log", false},
		{`^npm`, "npm install", true},
		{"", "anything", false},
	}

	for _, tt := range tests {
		f := Filter{Match: tt.match}
		got := f.MatchCommand(tt.command)
		if got != tt.want {
			t.Errorf("MatchCommand(%s, %s) = %v, want %v", tt.match, tt.command, got, tt.want)
		}
	}
}

func TestApply(t *testing.T) {
	loader := New()

	// Test strip_lines_matching
	filter := &Filter{
		Name:       "test-strip",
		Match:      "^go test",
		StripLines: []string{`^=== RUN`, `^--- PASS`},
	}

	input := `=== RUN   TestFoo
--- PASS: TestFoo (0.01s)
ok      pkg     0.5s`

	expected := `ok      pkg     0.5s`

	output := loader.Apply(filter, input)
	if output != expected {
		t.Errorf("apply strip:\ngot:\n%q\nwant:\n%q", output, expected)
	}
}

func TestApplyMaxLines(t *testing.T) {
	loader := New()

	filter := &Filter{
		Name:     "test-maxlines",
		Match:    "^cat",
		MaxLines: 5,
	}

	// Create 20 lines of input
	var input string
	for i := 0; i < 20; i++ {
		input += "line " + string(rune('A'+i%26)) + "\n"
	}

	output := loader.Apply(filter, input)
	lines := 0
	for _, c := range output {
		if c == '\n' {
			lines++
		}
	}

	// Should have around 5 lines plus truncation message
	if lines > 10 {
		t.Errorf("max_lines not applied: got %d lines, expected ~5", lines)
	}
}

func TestApplyKeepLines(t *testing.T) {
	loader := New()

	filter := &Filter{
		Name:      "test-keep",
		Match:     "^npm",
		KeepLines: []string{`ERROR`, `WARNING`},
	}

	input := `npm install
Downloading package...
Extracting...
npm WARNING deprecated pkg@1.0
Installing modules...
ERROR ENOENT: missing file
Done.`

	output := loader.Apply(filter, input)
	if !containsLine(output, "npm WARNING") {
		t.Error("expected WARNING line to be kept")
	}
	if !containsLine(output, "ERROR ENOENT") {
		t.Error("expected ERROR line to be kept")
	}
}

func TestApplyMinLines(t *testing.T) {
	loader := New()

	filter := &Filter{
		Name:     "test-minlines",
		Match:    "^docker",
		MinLines: 5,
	}

	// Short input should pass through unchanged
	short := "line1\nline2"
	output := loader.Apply(filter, short)
	if output != short {
		t.Error("short input should not be modified")
	}

	// Long input should be processed
	long := "line1\nline2\nline3\nline4\nline5\nline6"
	output = loader.Apply(filter, long)
	// Should be processed (no strip/keep/maxlines, so same output)
	if output != long {
		t.Error("long input should be returned when no transformations match")
	}
}

func TestApplyPrefixSuffix(t *testing.T) {
	loader := New()

	filter := &Filter{
		Name:   "test-prefix-suffix",
		Match:  "^curl",
		Prefix: "[filtered output]",
	}

	input := "response body"
	output := loader.Apply(filter, input)
	if !containsLine(output, "[filtered output]") {
		t.Error("expected prefix to be added")
	}
}

func TestLoadFromFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "yaml-test-*")
	if err != nil {
		t.Fatalf("create temp: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	yamlContent := `schema_version: 1
filters:
  - name: go-test
    match: "^go test"
    strip_lines_matching:
      - "^=== RUN"
      - "^--- PASS"
    max_lines: 10
    description: "Compress go test output"
  - name: npm-install
    match: "^npm install"
    strip_lines_matching:
      - "^Downloading"
      - "^Extracting"
`

	filePath := filepath.Join(tmpDir, "test-filters.yaml")
	if err := os.WriteFile(filePath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	loader := New()
	if err := loader.LoadFromFile(filePath); err != nil {
		t.Fatalf("load file: %v", err)
	}

	if loader.Count() != 2 {
		t.Errorf("count = %d, want 2", loader.Count())
	}

	// Test matching
	f := loader.Match("go test ./...")
	if f == nil {
		t.Fatal("expected match for 'go test ./...'")
	}
	if f.Name != "go-test" {
		t.Errorf("name = %s, want go-test", f.Name)
	}
	if len(f.StripLines) != 2 {
		t.Errorf("strip_lines = %d, want 2", len(f.StripLines))
	}
}

func TestLoadFromDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "yaml-dir-*")
	if err != nil {
		t.Fatalf("create temp: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create multiple YAML files
	files := map[string]string{
		"go.yaml":      "schema_version: 1\nfilters:\n  - name: go1\n    match: \"^go\"\n",
		"npm.yml":      "schema_version: 1\nfilters:\n  - name: npm1\n    match: \"^npm\"\n",
		"not-filter.txt": "this is not a filter",
	}

	for name, content := range files {
		if err := os.WriteFile(filepath.Join(tmpDir, name), []byte(content), 0644); err != nil {
			t.Fatalf("write file: %v", err)
		}
	}

	loader := New()
	if err := loader.LoadFromDir(tmpDir); err != nil {
		t.Fatalf("load from dir: %v", err)
	}

	if loader.Count() != 2 {
		t.Errorf("count = %d, want 2", loader.Count())
	}
}

func TestInvalidYAML(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "yaml-invalid-*")
	if err != nil {
		t.Fatalf("create temp: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	filePath := filepath.Join(tmpDir, "bad.yaml")
	if err := os.WriteFile(filePath, []byte("invalid: yaml: {"), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	loader := New()
	err = loader.LoadFromFile(filePath)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestMatchNone(t *testing.T) {
	loader := New()

	// Add a filter
	f, _ := os.CreateTemp("", "test-*.yaml")
	defer f.Close()
	defer os.Remove(f.Name())

	f.WriteString("schema_version: 1\nfilters:\n  - name: go\n    match: \"^go\"\n")
	f.Close()

	loader.LoadFromFile(f.Name())

	// Match something that doesn't exist
	result := loader.Match("cargo build")
	if result != nil {
		t.Error("expected nil for no match")
	}
}

func containsLine(output, substr string) bool {
	for _, line := range strings.Split(output, "\n") {
		if len(line) >= len(substr) {
			for i := 0; i <= len(line)-len(substr); i++ {
				if line[i:i+len(substr)] == substr {
					return true
				}
			}
		}
	}
	return false
}
