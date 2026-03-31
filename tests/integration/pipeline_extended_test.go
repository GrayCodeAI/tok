package integration

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestPipelinePresets tests different pipeline presets
func TestPipelinePresets(t *testing.T) {
	binPath := filepath.Join(t.TempDir(), "tokman")
	buildCmd := exec.Command("go", "build", "-o", binPath, "./cmd/tokman")
	buildCmd.Dir = filepath.Join("..", "..")
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build: %v\n%s", err, output)
	}

	input := strings.Repeat("This is a test line with some content.\n", 100)
	inputFile := filepath.Join(t.TempDir(), "input.txt")
	if err := os.WriteFile(inputFile, []byte(input), 0644); err != nil {
		t.Fatalf("Failed to write input: %v", err)
	}

	presets := []string{"fast", "balanced", "full"}
	for _, preset := range presets {
		t.Run(preset, func(t *testing.T) {
			cmd := exec.Command(binPath, "summary", "--preset", preset, inputFile)
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("Command failed for preset %s: %v\n%s", preset, err, output)
			}
			if len(output) == 0 {
				t.Errorf("Expected non-empty output for preset %s", preset)
			}
		})
	}
}

// TestPipelineModes tests minimal vs aggressive modes
func TestPipelineModes(t *testing.T) {
	binPath := filepath.Join(t.TempDir(), "tokman")
	buildCmd := exec.Command("go", "build", "-o", binPath, "./cmd/tokman")
	buildCmd.Dir = filepath.Join("..", "..")
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build: %v\n%s", err, output)
	}

	input := strings.Repeat("Test content line.\n", 200)
	inputFile := filepath.Join(t.TempDir(), "input.txt")
	if err := os.WriteFile(inputFile, []byte(input), 0644); err != nil {
		t.Fatalf("Failed to write input: %v", err)
	}

	modes := []string{"minimal", "aggressive"}
	for _, mode := range modes {
		t.Run(mode, func(t *testing.T) {
			cmd := exec.Command(binPath, "summary", "--mode", mode, inputFile)
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("Command failed for mode %s: %v\n%s", mode, err, output)
			}
			if len(output) == 0 {
				t.Errorf("Expected non-empty output for mode %s", mode)
			}
		})
	}
}

// TestPipelineWithBudget tests budget-constrained compression
func TestPipelineWithBudget(t *testing.T) {
	binPath := filepath.Join(t.TempDir(), "tokman")
	buildCmd := exec.Command("go", "build", "-o", binPath, "./cmd/tokman")
	buildCmd.Dir = filepath.Join("..", "..")
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build: %v\n%s", err, output)
	}

	input := strings.Repeat("Content line with words.\n", 500)
	inputFile := filepath.Join(t.TempDir(), "input.txt")
	if err := os.WriteFile(inputFile, []byte(input), 0644); err != nil {
		t.Fatalf("Failed to write input: %v", err)
	}

	cmd := exec.Command(binPath, "summary", "--budget", "500", inputFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Command failed: %v\n%s", err, output)
	}
	if len(output) == 0 {
		t.Error("Expected non-empty output")
	}
}

// TestPipelineJSONOutput tests JSON output format
func TestPipelineJSONOutput(t *testing.T) {
	binPath := filepath.Join(t.TempDir(), "tokman")
	buildCmd := exec.Command("go", "build", "-o", binPath, "./cmd/tokman")
	buildCmd.Dir = filepath.Join("..", "..")
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build: %v\n%s", err, output)
	}

	input := strings.Repeat("JSON test content.\n", 50)
	inputFile := filepath.Join(t.TempDir(), "input.txt")
	if err := os.WriteFile(inputFile, []byte(input), 0644); err != nil {
		t.Fatalf("Failed to write input: %v", err)
	}

	cmd := exec.Command(binPath, "json", inputFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Skipf("json command not available: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		t.Errorf("Output is not valid JSON: %v\nOutput: %s", err, string(output))
	}
}

// TestPipelineWithQueryIntent tests query-dependent compression
func TestPipelineWithQueryIntent(t *testing.T) {
	binPath := filepath.Join(t.TempDir(), "tokman")
	buildCmd := exec.Command("go", "build", "-o", binPath, "./cmd/tokman")
	buildCmd.Dir = filepath.Join("..", "..")
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build: %v\n%s", err, output)
	}

	input := `Error: connection refused
Stack trace:
  at main.go:42
  at handler.go:100
The server failed to start because port 8080 is already in use.
Check if another process is running on this port.`
	inputFile := filepath.Join(t.TempDir(), "input.txt")
	if err := os.WriteFile(inputFile, []byte(input), 0644); err != nil {
		t.Fatalf("Failed to write input: %v", err)
	}

	cmd := exec.Command(binPath, "summary", "--query", "port error", inputFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Command failed: %v\n%s", err, output)
	}
	if len(output) == 0 {
		t.Error("Expected non-empty output")
	}
}

// TestPipelineEdgeCases tests edge cases
func TestPipelineEdgeCases(t *testing.T) {
	binPath := filepath.Join(t.TempDir(), "tokman")
	buildCmd := exec.Command("go", "build", "-o", binPath, "./cmd/tokman")
	buildCmd.Dir = filepath.Join("..", "..")
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build: %v\n%s", err, output)
	}

	tests := []struct {
		name  string
		input string
	}{
		{"empty_input", ""},
		{"single_char", "a"},
		{"unicode", "Hello 世界 🌍"},
		{"special_chars", "!@#$%^&*()_+-=[]{}|;':\",./<>?"},
		{"whitespace_only", "   \n\n\t\t   "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputFile := filepath.Join(t.TempDir(), "input.txt")
			if err := os.WriteFile(inputFile, []byte(tt.input), 0644); err != nil {
				t.Fatalf("Failed to write input: %v", err)
			}

			cmd := exec.Command(binPath, "summary", inputFile)
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("Command failed: %v\n%s", err, output)
			}
			// Should not crash even with edge case inputs
			t.Logf("Output length: %d bytes", len(output))
		})
	}
}

// TestPipelineLargeInput tests handling of large inputs
func TestPipelineLargeInput(t *testing.T) {
	binPath := filepath.Join(t.TempDir(), "tokman")
	buildCmd := exec.Command("go", "build", "-o", binPath, "./cmd/tokman")
	buildCmd.Dir = filepath.Join("..", "..")
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build: %v\n%s", err, output)
	}

	// Generate 100KB of content
	var sb strings.Builder
	for i := 0; i < 2000; i++ {
		sb.WriteString("Line ")
		sb.WriteString(strings.Repeat("content ", 10))
		sb.WriteString("\n")
	}
	input := sb.String()

	inputFile := filepath.Join(t.TempDir(), "large.txt")
	if err := os.WriteFile(inputFile, []byte(input), 0644); err != nil {
		t.Fatalf("Failed to write input: %v", err)
	}

	cmd := exec.Command(binPath, "summary", inputFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Command failed: %v\n%s", err, output)
	}

	// Verify significant compression
	originalLen := len(input)
	compressedLen := len(output)
	if compressedLen >= originalLen {
		t.Errorf("Expected compression, but output (%d bytes) >= input (%d bytes)", compressedLen, originalLen)
	}

	reduction := float64(originalLen-compressedLen) / float64(originalLen) * 100
	t.Logf("Compression: %.1f%% (%d -> %d bytes)", reduction, originalLen, compressedLen)
}

// TestPipelineCodeInput tests compression of code content
func TestPipelineCodeInput(t *testing.T) {
	binPath := filepath.Join(t.TempDir(), "tokman")
	buildCmd := exec.Command("go", "build", "-o", binPath, "./cmd/tokman")
	buildCmd.Dir = filepath.Join("..", "..")
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build: %v\n%s", err, output)
	}

	input := `package main

import (
	"fmt"
	"net/http"
)

type Server struct {
	addr string
	mux  *http.ServeMux
}

func NewServer(addr string) *Server {
	return &Server{
		addr: addr,
		mux:  http.NewServeMux(),
	}
}

func (s *Server) Start() error {
	s.mux.HandleFunc("/", s.handleRoot)
	s.mux.HandleFunc("/health", s.handleHealth)
	return http.ListenAndServe(s.addr, s.mux)
}

func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, World!")
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "OK")
}

func main() {
	server := NewServer(":8080")
	if err := server.Start(); err != nil {
		fmt.Printf("Server failed: %v\n", err)
	}
}`

	inputFile := filepath.Join(t.TempDir(), "server.go")
	if err := os.WriteFile(inputFile, []byte(input), 0644); err != nil {
		t.Fatalf("Failed to write input: %v", err)
	}

	cmd := exec.Command(binPath, "summary", inputFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Command failed: %v\n%s", err, output)
	}

	t.Logf("Input: %d bytes, Output: %d bytes", len(input), len(output))
}
