// Package testing provides test fixtures and utilities.
package testing

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// FixtureLoader loads test fixtures from disk.
type FixtureLoader struct {
	basePath string
}

// NewFixtureLoader creates a new fixture loader.
func NewFixtureLoader(basePath string) *FixtureLoader {
	return &FixtureLoader{basePath: basePath}
}

// Load loads a fixture by name.
func (f *FixtureLoader) Load(name string) (string, error) {
	path := filepath.Join(f.basePath, name)
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to load fixture %s: %w", name, err)
	}
	return string(data), nil
}

// LoadBytes loads a fixture as bytes.
func (f *FixtureLoader) LoadBytes(name string) ([]byte, error) {
	path := filepath.Join(f.basePath, name)
	return os.ReadFile(path)
}

// MustLoad loads a fixture or fails the test.
func (f *FixtureLoader) MustLoad(t testing.TB, name string) string {
	t.Helper()
	content, err := f.Load(name)
	if err != nil {
		t.Fatalf("Failed to load fixture %s: %v", name, err)
	}
	return content
}

// List returns all fixture names.
func (f *FixtureLoader) List() ([]string, error) {
	entries, err := os.ReadDir(f.basePath)
	if err != nil {
		return nil, err
	}

	var names []string
	for _, entry := range entries {
		if !entry.IsDir() {
			names = append(names, entry.Name())
		}
	}
	return names, nil
}

// TestContentGenerator generates test content of various types.
type TestContentGenerator struct {
	rng *rand.Rand
}

// NewTestContentGenerator creates a new generator.
func NewTestContentGenerator() *TestContentGenerator {
	return &TestContentGenerator{
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// GenerateLogOutput generates realistic log output.
func (g *TestContentGenerator) GenerateLogOutput(lines int) string {
	levels := []string{"DEBUG", "INFO", "WARN", "ERROR"}
	messages := []string{
		"Connection established",
		"Request processed",
		"Cache miss",
		"Database query executed",
		"Authentication successful",
		"Rate limit hit",
		"Retrying operation",
		"Batch processing complete",
	}

	var sb strings.Builder
	baseTime := time.Now().Add(-time.Duration(lines) * time.Second)

	for i := 0; i < lines; i++ {
		timestamp := baseTime.Add(time.Duration(i) * time.Second)
		level := levels[g.rng.Intn(len(levels))]
		message := messages[g.rng.Intn(len(messages))]
		id := g.rng.Int63n(1000000)

		fmt.Fprintf(&sb, "%s [%s] request_id=%d %s\n",
			timestamp.Format(time.RFC3339),
			level,
			id,
			message,
		)
	}

	return sb.String()
}

// GenerateBuildOutput generates build tool output.
func (g *TestContentGenerator) GenerateBuildOutput(errors, warnings int) string {
	var sb strings.Builder

	// Compiling messages
	files := []string{"src/main.rs", "src/lib.rs", "src/utils.rs", "tests/integration.rs"}
	for _, file := range files {
		fmt.Fprintf(&sb, "   Compiling %s\n", file)
	}

	// Warnings
	for i := 0; i < warnings; i++ {
		file := files[g.rng.Intn(len(files))]
		line := g.rng.Intn(100) + 1
		fmt.Fprintf(&sb, "warning: unused variable: `x`\n  --> %s:%d:5\n\n", file, line)
	}

	// Errors
	for i := 0; i < errors; i++ {
		file := files[g.rng.Intn(len(files))]
		line := g.rng.Intn(100) + 1
		fmt.Fprintf(&sb, "error[E0%d]: mismatched types\n  --> %s:%d:10\n   expected `String`\n\n",
			g.rng.Intn(999), file, line)
	}

	if errors == 0 {
		sb.WriteString("    Finished dev [unoptimized + debuginfo] target(s) in 1.23s\n")
	} else {
		fmt.Fprintf(&sb, "error: could not compile due to %d previous errors\n", errors)
	}

	return sb.String()
}

// GenerateTestOutput generates test runner output.
func (g *TestContentGenerator) GenerateTestOutput(passed, failed, skipped int) string {
	var sb strings.Builder

	tests := []string{
		"test_parse_simple", "test_parse_complex", "test_serialize",
		"test_roundtrip", "test_edge_cases", "test_performance",
		"test_integration", "test_errors", "test_concurrent",
	}

	// Run tests
	for i := 0; i < passed+failed+skipped; i++ {
		testName := tests[i%len(tests)]
		if i > 0 {
			testName = fmt.Sprintf("%s_%d", testName, i)
		}

		switch {
		case i < passed:
			fmt.Fprintf(&sb, "test %s ... ok\n", testName)
		case i < passed+failed:
			fmt.Fprintf(&sb, "test %s ... FAILED\n", testName)
			file := "src/lib.rs"
			line := g.rng.Intn(100) + 1
			fmt.Fprintf(&sb, "    --> %s:%d\n", file, line)
			fmt.Fprintf(&sb, "    expected: 42\n    actual: 43\n\n")
		default:
			fmt.Fprintf(&sb, "test %s ... ignored\n", testName)
		}
	}

	// Summary
	fmt.Fprintf(&sb, "\ntest result: %s. %d passed; %d failed; %d ignored\n",
		map[bool]string{true: "FAILED", false: "ok"}[failed > 0],
		passed, failed, skipped)
	fmt.Fprintf(&sb, "test %s::tests ... bench: %d ns/iter (+/- %d)\n",
		tests[0], g.rng.Intn(10000)+1000, g.rng.Intn(100))

	return sb.String()
}

// GenerateCode generates source code of specified language.
func (g *TestContentGenerator) GenerateCode(language string, lines int) string {
	switch language {
	case "go":
		return g.generateGoCode(lines)
	case "rust":
		return g.generateRustCode(lines)
	case "typescript":
		return g.generateTypeScriptCode(lines)
	default:
		return g.generateGenericCode(lines)
	}
}

func (g *TestContentGenerator) generateGoCode(lines int) string {
	var sb strings.Builder
	sb.WriteString("package main\n\n")
	sb.WriteString("import \"fmt\"\n\n")
	sb.WriteString("func main() {\n")

	for i := 0; i < lines-5; i++ {
		fmt.Fprintf(&sb, "\t// Line %d: process item\n", i+1)
		fmt.Fprintf(&sb, "\tfmt.Printf(\"Processing %%d\\n\", %d)\n", i)
	}

	sb.WriteString("}\n")
	return sb.String()
}

func (g *TestContentGenerator) generateRustCode(lines int) string {
	var sb strings.Builder
	sb.WriteString("fn main() {\n")

	for i := 0; i < lines-2; i++ {
		fmt.Fprintf(&sb, "    // Line %d\n", i+1)
		fmt.Fprintf(&sb, "    let x%d = %d;\n", i, g.rng.Intn(100))
	}

	sb.WriteString("}\n")
	return sb.String()
}

func (g *TestContentGenerator) generateTypeScriptCode(lines int) string {
	var sb strings.Builder
	sb.WriteString("function main(): void {\n")

	for i := 0; i < lines-2; i++ {
		fmt.Fprintf(&sb, "  // Line %d\n", i+1)
		fmt.Fprintf(&sb, "  const x%d: number = %d;\n", i, g.rng.Intn(100))
	}

	sb.WriteString("}\n")
	return sb.String()
}

func (g *TestContentGenerator) generateGenericCode(lines int) string {
	var sb strings.Builder
	for i := 0; i < lines; i++ {
		fmt.Fprintf(&sb, "// Line %d: some operation here\n", i+1)
	}
	return sb.String()
}

// GeneratePII generates content with PII for testing detection.
func (g *TestContentGenerator) GeneratePII() string {
	emails := []string{
		"user@example.com",
		"john.doe@company.org",
		"admin@site.net",
	}
	phones := []string{
		"(555) 123-4567",
		"555-987-6543",
		"+1 555 555 5555",
	}
	ssns := []string{
		"123-45-6789",
		"987-65-4321",
	}
	apiKeys := []string{
		"sk-abc123def456ghi789",
		"api-key-x1y2z3a4b5c6d7",
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Contact: %s\n", emails[g.rng.Intn(len(emails))])
	fmt.Fprintf(&sb, "Phone: %s\n", phones[g.rng.Intn(len(phones))])
	fmt.Fprintf(&sb, "SSN: %s\n", ssns[g.rng.Intn(len(ssns))])
	fmt.Fprintf(&sb, "API Key: %s\n", apiKeys[g.rng.Intn(len(apiKeys))])

	return sb.String()
}

// GoldenFile manages golden files for snapshot testing.
type GoldenFile struct {
	dir string
}

// NewGoldenFile creates a new golden file manager.
func NewGoldenFile(dir string) *GoldenFile {
	return &GoldenFile{dir: dir}
}

// Path returns the path to a golden file.
func (g *GoldenFile) Path(name string) string {
	return filepath.Join(g.dir, name+".golden")
}

// Update updates a golden file.
func (g *GoldenFile) Update(name string, content []byte) error {
	if err := os.MkdirAll(g.dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(g.Path(name), content, 0644)
}

// Read reads a golden file.
func (g *GoldenFile) Read(name string) ([]byte, error) {
	return os.ReadFile(g.Path(name))
}

// Assert compares content against golden file.
func (g *GoldenFile) Assert(t testing.TB, name string, got []byte) {
	t.Helper()

	path := g.Path(name)

	// Check if update flag is set
	if os.Getenv("UPDATE_GOLDEN") != "" {
		if err := g.Update(name, got); err != nil {
			t.Fatalf("Failed to update golden file: %v", err)
		}
		return
	}

	want, err := g.Read(name)
	if err != nil {
		t.Fatalf("Failed to read golden file %s: %v", path, err)
	}

	if string(got) != string(want) {
		t.Errorf("Golden file mismatch:\nwant:\n%s\ngot:\n%s", want, got)
	}
}

// BenchmarkRunner runs compression benchmarks.
type BenchmarkRunner struct {
	sizes []int
}

// NewBenchmarkRunner creates a new benchmark runner.
func NewBenchmarkRunner() *BenchmarkRunner {
	return &BenchmarkRunner{
		sizes: []int{100, 1000, 10000, 100000},
	}
}

// Run runs benchmarks for different content sizes.
func (b *BenchmarkRunner) Run(name string, fn func(content string) (string, int), gen *TestContentGenerator) map[string]BenchmarkResult {
	results := make(map[string]BenchmarkResult)

	for _, size := range b.sizes {
		content := gen.GenerateLogOutput(size)
		compressed, saved := fn(content)

		key := fmt.Sprintf("%s_%d", name, size)
		results[key] = BenchmarkResult{
			OriginalSize:   len(content),
			CompressedSize: len(compressed),
			TokensSaved:    saved,
			ReductionPct:   float64(len(content)-len(compressed)) / float64(len(content)) * 100,
		}
	}

	return results
}

// BenchmarkResult contains benchmark results.
type BenchmarkResult struct {
	OriginalSize   int
	CompressedSize int
	TokensSaved    int
	ReductionPct   float64
}

// TempDir creates a temporary directory for tests.
func TempDir(t testing.TB) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "tokman-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	return dir
}

// WriteTempFile writes a temporary file for tests.
func WriteTempFile(t testing.TB, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	return path
}
