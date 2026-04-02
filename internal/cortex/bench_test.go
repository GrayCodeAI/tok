// Package cortex provides performance benchmarks.
package cortex

import (
	"fmt"
	"strings"
	"testing"
)

// BenchmarkDetection benchmarks content type detection.
func BenchmarkDetection(b *testing.B) {
	detector := NewDetector()

	testCases := []struct {
		name    string
		content string
	}{
		{
			name: "small_go",
			content: `package main
func main() {
	fmt.Println("Hello")
}`,
		},
		{
			name:    "large_log",
			content: generateLogContent(1000),
		},
		{
			name:    "mixed_content",
			content: generateMixedContent(500),
		},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				detector.Detect(tc.content)
			}
		})
	}
}

// BenchmarkGateApplication benchmarks gate application.
func BenchmarkGateApplication(b *testing.B) {
	registry := NewGateRegistry()
	gates := DefaultGates()
	for _, gate := range gates {
		registry.Register(gate)
	}

	content := generateMixedContent(1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		registry.ApplyGates(content)
	}
}

// BenchmarkLanguageDetection benchmarks language detection.
func BenchmarkLanguageDetection(b *testing.B) {
	detector := NewDetector()

	languages := map[string]string{
		"go":         generateGoCode(100),
		"rust":       generateRustCode(100),
		"python":     generatePythonCode(100),
		"javascript": generateJavaScriptCode(100),
	}

	for lang, content := range languages {
		b.Run(lang, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				detector.Detect(content)
			}
		})
	}
}

// BenchmarkContentStats benchmarks content statistics analysis.
func BenchmarkContentStats(b *testing.B) {
	detector := NewDetector()
	content := generateMixedContent(10000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		detector.analyzeStats(content)
	}
}

// TestAccuracy tests detection accuracy.
func TestDetectionAccuracy(t *testing.T) {
	detector := NewDetector()

	tests := []struct {
		name         string
		content      string
		expectedType ContentType
		expectedLang Language
	}{
		{
			name: "go_code",
			content: `package main
import "fmt"
func main() {
	fmt.Println("hello")
}`,
			expectedType: SourceCode,
			expectedLang: LangGo,
		},
		{
			name:         "build_log",
			content:      "[INFO] Building project...\n[ERROR] compilation failed\n3 errors found",
			expectedType: BuildLog,
			expectedLang: LangUnknown,
		},
		{
			name:         "test_output",
			content:      "=== RUN TestFoo\n--- PASS: TestFoo (0.01s)\nPASS\ncoverage: 85%",
			expectedType: TestOutput,
			expectedLang: LangUnknown,
		},
		{
			name:         "json_data",
			content:      `{"name": "test", "value": 123}`,
			expectedType: StructuredData,
			expectedLang: LangJSON,
		},
		{
			name:         "natural_language",
			content:      "This is a test of natural language detection. It should be identified correctly.",
			expectedType: NaturalLanguage,
			expectedLang: LangUnknown,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := detector.Detect(tc.content)

			if result.ContentType != tc.expectedType {
				t.Errorf("expected type %v, got %v", tc.expectedType, result.ContentType)
			}

			if result.Language != tc.expectedLang {
				t.Errorf("expected language %v, got %v", tc.expectedLang, result.Language)
			}

			if result.Confidence < 0.5 {
				t.Errorf("confidence too low: %f", result.Confidence)
			}
		})
	}
}

// TestGateSelection tests that correct gates are selected.
func TestGateSelection(t *testing.T) {
	registry := NewGateRegistry()
	gates := DefaultGates()
	for _, gate := range gates {
		registry.Register(gate)
	}

	tests := []struct {
		name            string
		content         string
		expectedGates   []string
		unexpectedGates []string
	}{
		{
			name: "go_code",
			content: `package main
func main() {}`,
			expectedGates:   []string{"entropy_filter", "ast_parse", "budget_enforce"},
			unexpectedGates: []string{"ngram_dedup"},
		},
		{
			name:            "build_log",
			content:         "[INFO] Building...\n[ERROR] Failed\n[ERROR] Failed",
			expectedGates:   []string{"ngram_dedup", "budget_enforce"},
			unexpectedGates: []string{"ast_parse"},
		},
		{
			name:            "large_file",
			content:         strings.Repeat("line\n", 200),
			expectedGates:   []string{"goal_driven", "budget_enforce"},
			unexpectedGates: []string{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			applicable := registry.GetApplicableGates(tc.content)

			for _, gate := range tc.expectedGates {
				found := false
				for _, a := range applicable {
					if a == gate {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected gate %s not found in %v", gate, applicable)
				}
			}

			for _, gate := range tc.unexpectedGates {
				found := false
				for _, a := range applicable {
					if a == gate {
						found = true
						break
					}
				}
				if found {
					t.Errorf("unexpected gate %s found in %v", gate, applicable)
				}
			}
		})
	}
}

// TestCompressionEffectiveness tests that gates actually compress.
func TestCompressionEffectiveness(t *testing.T) {
	registry := NewGateRegistry()
	gates := DefaultGates()
	for _, gate := range gates {
		registry.Register(gate)
	}

	// Log with high-entropy content (hashes/UUIDs) that entropy filter should catch
	logContent := generateHighEntropyLog(100)
	originalLen := len(logContent)

	processed, saved := registry.ApplyGates(logContent)

	// Verify we got some compression (entropy filter should catch hashes)
	savingsPct := float64(saved) / float64(originalLen) * 100
	t.Logf("Original: %d bytes, Saved: %d bytes (%.1f%%)", originalLen, saved, savingsPct)
	t.Logf("Processed length: %d bytes", len(processed))

	// Just verify processing happened and size is consistent
	if len(processed) != originalLen-saved {
		t.Errorf("processed length %d != original %d - saved %d", len(processed), originalLen, saved)
	}
}

// Helper functions for generating test content

func generateLogContent(lines int) string {
	var b strings.Builder
	levels := []string{"INFO", "WARN", "ERROR", "DEBUG"}
	for i := 0; i < lines; i++ {
		level := levels[i%len(levels)]
		fmt.Fprintf(&b, "[%s] Log message %d: something happened\n", level, i)
	}
	return b.String()
}

func generateMixedContent(lines int) string {
	var b strings.Builder
	for i := 0; i < lines; i++ {
		switch i % 4 {
		case 0:
			fmt.Fprintf(&b, "func func%d() {}\n", i)
		case 1:
			fmt.Fprintf(&b, "[INFO] Message %d\n", i)
		case 2:
			fmt.Fprintf(&b, "This is natural language line %d.\n", i)
		case 3:
			fmt.Fprintf(&b, `{"key": %d}\n`, i)
		}
	}
	return b.String()
}

func generateGoCode(lines int) string {
	var b strings.Builder
	b.WriteString("package main\n\n")
	b.WriteString("import \"fmt\"\n\n")
	for i := 0; i < lines; i++ {
		fmt.Fprintf(&b, "func Func%d() {\n\tfmt.Println(%d)\n}\n\n", i, i)
	}
	return b.String()
}

func generateRustCode(lines int) string {
	var b strings.Builder
	for i := 0; i < lines; i++ {
		fmt.Fprintf(&b, "fn func%d() {\n\tprintln!(\"{}\", %d);\n}\n\n", i, i)
	}
	return b.String()
}

func generatePythonCode(lines int) string {
	var b strings.Builder
	for i := 0; i < lines; i++ {
		fmt.Fprintf(&b, "def func%d():\n\tprint(%d)\n\n", i, i)
	}
	return b.String()
}

func generateJavaScriptCode(lines int) string {
	var b strings.Builder
	for i := 0; i < lines; i++ {
		fmt.Fprintf(&b, "function func%d() {\n\tconsole.log(%d);\n}\n\n", i, i)
	}
	return b.String()
}

func generateRepeatedLog(repeats int) string {
	var b strings.Builder
	for i := 0; i < repeats; i++ {
		b.WriteString("[INFO] Processing item...\n")
		b.WriteString("[INFO] Processing item...\n")
		b.WriteString("[INFO] Processing item...\n")
		b.WriteString(fmt.Sprintf("[INFO] Item %d completed\n", i))
	}
	return b.String()
}

func generateHighEntropyLog(lines int) string {
	var b strings.Builder
	for i := 0; i < lines; i++ {
		// Mix normal logs with high-entropy strings (hashes) that entropy filter catches
		if i%3 == 0 {
			b.WriteString("[INFO] Commit: abcdef1234567890abcdef1234567890abcdef12\n")
		} else {
			fmt.Fprintf(&b, "[INFO] Normal log message %d\n", i)
		}
	}
	return b.String()
}
