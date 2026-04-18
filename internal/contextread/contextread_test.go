package contextread

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildSignaturesAndLineNumbers(t *testing.T) {
	dataHome := t.TempDir()
	t.Setenv("XDG_DATA_HOME", dataHome)

	content := strings.Join([]string{
		"package main",
		"",
		"// comment",
		"type User struct{}",
		"func main() {",
		`  fmt.Println("hello")`,
		"}",
	}, "\n")

	out, orig, filt, err := Build("main.go", content, "", Options{
		Mode:        "signatures",
		LineNumbers: true,
	})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if orig <= 0 || filt <= 0 {
		t.Fatalf("unexpected token counts: orig=%d filt=%d", orig, filt)
	}
	if !strings.Contains(out, "package main") || !strings.Contains(out, "type User struct{}") {
		t.Fatalf("signature output missing expected lines:\n%s", out)
	}
	if !strings.Contains(out, "   1 |") {
		t.Fatalf("expected line numbers in output:\n%s", out)
	}
}

func TestBuildDeltaUsesSnapshots(t *testing.T) {
	dataHome := t.TempDir()
	t.Setenv("XDG_DATA_HOME", dataHome)

	path := filepath.Join(t.TempDir(), "main.go")
	initial := "package main\nfunc main() {}\n"
	if _, _, _, err := Build(path, initial, "", Options{Mode: "auto", SaveSnapshot: true}); err != nil {
		t.Fatalf("initial Build() error = %v", err)
	}

	updated := "package main\nfunc main() {\n  println(\"changed\")\n}\n"
	out, _, _, err := Build(path, updated, "", Options{Mode: "delta", SaveSnapshot: false})
	if err != nil {
		t.Fatalf("delta Build() error = %v", err)
	}
	if !strings.Contains(out, "# Delta for") || !strings.Contains(out, "~ ") {
		t.Fatalf("unexpected delta output:\n%s", out)
	}
}

func TestTrackedCommandPatterns(t *testing.T) {
	patterns := TrackedCommandPatterns()
	if len(patterns) == 0 {
		t.Fatal("TrackedCommandPatterns() should not be empty")
	}
	if got := TrackedCommandPatternsForKind("read"); len(got) == 0 {
		t.Fatal("TrackedCommandPatternsForKind(read) should not be empty")
	}
}

func TestBuildSavesSnapshot(t *testing.T) {
	dataHome := t.TempDir()
	t.Setenv("XDG_DATA_HOME", dataHome)

	path := filepath.Join(t.TempDir(), "sample.txt")
	if _, _, _, err := Build(path, "hello\nworld\n", "", Options{SaveSnapshot: true}); err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if _, err := os.Stat(snapshotPath(path)); err != nil {
		t.Fatalf("snapshot not written: %v", err)
	}
}
