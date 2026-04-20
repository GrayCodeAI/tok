// Package filtergolden runs every fixture under tests/filters/<case>/
// through the real production TOML-filter engine and diffs against
// expected.txt. This catches silent regressions where a filter change
// drops a line the model actually needed — the kind of quality break
// unit tests with synthetic inputs miss.
//
// Layout:
//
//	tests/filters/<case>/
//	    cmd.txt        — the shell command the fixture represents
//	    input.txt      — raw command output (what the agent would see today)
//	    expected.txt   — the filter's output (what we want the agent to see)
//
// Filter selection: the harness loads every *.toml under internal/toml/builtin/
// and filters/, then MatchesCommand picks the rule whose match_command regex
// accepts cmd.txt. This mirrors production runtime behavior exactly.
//
// Updating fixtures: if a filter intentionally changes output, regenerate
// with  `go test ./tests/filters -update`.
package filtergolden

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lakshmanpatel/tok/internal/filter"
	"github.com/lakshmanpatel/tok/internal/toml"
)

var updateGolden = flag.Bool("update", false, "regenerate expected.txt from current filter output")

func TestGoldenFixtures(t *testing.T) {
	repoRoot := findRepoRoot(t)
	fixturesDir := filepath.Join(repoRoot, "tests", "filters")

	registry := toml.NewFilterRegistry()
	for _, dir := range []string{
		filepath.Join(repoRoot, "internal", "toml", "builtin"),
		filepath.Join(repoRoot, "filters"),
	} {
		loadTOMLDir(t, registry, dir)
	}

	cases, err := os.ReadDir(fixturesDir)
	if err != nil {
		t.Fatalf("read fixtures dir: %v", err)
	}

	for _, c := range cases {
		if !c.IsDir() {
			continue
		}
		caseDir := filepath.Join(fixturesDir, c.Name())
		if _, err := os.Stat(filepath.Join(caseDir, "cmd.txt")); err != nil {
			continue // not a fixture
		}
		t.Run(c.Name(), func(t *testing.T) {
			runFixture(t, registry, caseDir)
		})
	}
}

func runFixture(t *testing.T, registry *toml.FilterRegistry, dir string) {
	t.Helper()

	cmd := readTrim(t, filepath.Join(dir, "cmd.txt"))
	input := readFile(t, filepath.Join(dir, "input.txt"))

	_, _, rule := registry.FindMatchingFilter(cmd)
	if rule == nil {
		t.Fatalf("no filter matched command %q — add a filter or fix the fixture", cmd)
	}
	engine := toml.NewTOMLFilterEngine(rule)
	got, _ := engine.Apply(input, filter.Mode(""))

	expectedPath := filepath.Join(dir, "expected.txt")
	if *updateGolden {
		if err := os.WriteFile(expectedPath, []byte(got), 0644); err != nil {
			t.Fatalf("write expected.txt: %v", err)
		}
		t.Logf("updated %s", expectedPath)
		return
	}

	want := readFile(t, expectedPath)
	if got != want {
		t.Errorf("filter output mismatch for %s\n--- want ---\n%s\n--- got ---\n%s\n--- end ---",
			filepath.Base(dir), want, got)
	}
}

func loadTOMLDir(t *testing.T, registry *toml.FilterRegistry, dir string) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		return // optional dir
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".toml") {
			continue
		}
		_ = registry.LoadFile(filepath.Join(dir, e.Name()))
	}
}

func findRepoRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for dir := wd; dir != "/"; dir = filepath.Dir(dir) {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
	}
	t.Fatalf("could not locate go.mod from %s", wd)
	return ""
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(b)
}

func readTrim(t *testing.T, path string) string {
	return strings.TrimSpace(readFile(t, path))
}
