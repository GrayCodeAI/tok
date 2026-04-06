package analysis

import (
	"strings"
	"testing"

	"github.com/GrayCodeAI/tokman/internal/core"
	"github.com/GrayCodeAI/tokman/internal/filter"
)

func TestBenchmarkSuites(t *testing.T) {
	tests := []struct {
		name     string
		suite    string
		wantMin  int // expected minimum number of test cases
	}{
		{"git-status suite", "git-status", 1},
		{"docker-ps suite", "docker-ps", 1},
		{"test-cargo suite", "test-cargo", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cases, ok := benchmarkSuites[tt.suite]
			if !ok {
				t.Errorf("suite %q not found in benchmarkSuites", tt.suite)
				return
			}
			if len(cases) < tt.wantMin {
				t.Errorf("suite %q has %d cases, want at least %d", tt.suite, len(cases), tt.wantMin)
			}
		})
	}
}

func TestBenchmarkSuitesExist(t *testing.T) {
	requiredSuites := []string{"git-status", "docker-ps", "test-cargo", "test-go", "test-pytest", "build-cargo"}
	
	for _, suite := range requiredSuites {
		t.Run(suite, func(t *testing.T) {
			if _, ok := benchmarkSuites[suite]; !ok {
				t.Errorf("required suite %q not found", suite)
			}
		})
	}
}

func TestBenchCase(t *testing.T) {
	tc := benchCase{
		command: "git status",
		content: "On branch main\nnothing to commit",
	}

	if tc.command != "git status" {
		t.Errorf("command = %q, want %q", tc.command, "git status")
	}
	if !strings.Contains(tc.content, "On branch main") {
		t.Errorf("content missing expected text")
	}
}

func TestRunBenchmarkCase(t *testing.T) {
	tc := benchCase{
		command: "git status",
		content: gitStatusOutputClean,
	}

	cfg := filter.PipelineConfig{
		Mode: filter.ModeMinimal,
	}
	pipeline := filter.NewPipelineCoordinator(cfg)

	// Measure before tokens
	beforeTok := core.EstimateTokens(tc.content)
	if beforeTok == 0 {
		t.Fatal("beforeTok should be > 0")
	}

	// Process through pipeline
	after, _ := pipeline.Process(tc.content)

	// Measure after tokens
	afterTok := core.EstimateTokens(after)
	if afterTok == 0 {
		t.Fatal("afterTok should be > 0")
	}

	// Calculate savings
	saved := beforeTok - afterTok
	if saved < 0 {
		saved = 0
	}

	t.Logf("Before: %d tokens, After: %d tokens, Saved: %d tokens", beforeTok, afterTok, saved)
}

func TestOutputBenchmarkJSON(t *testing.T) {
	results := []benchResult{
		{
			suite:     "git-status",
			beforeTok: 100,
			afterTok:  50,
			savedTok:  50,
			pctSaved:  50.0,
			duration:  1000000,
		},
	}

	// This should not panic
	err := outputBenchmarkJSON(results, 100, 50, 50, 50.0)
	if err != nil {
		t.Errorf("outputBenchmarkJSON failed: %v", err)
	}
}

func TestOutputMarkdown(t *testing.T) {
	results := []benchResult{
		{
			suite:     "git-status",
			beforeTok: 100,
			afterTok:  50,
			savedTok:  50,
			pctSaved:  50.0,
			duration:  1000000,
		},
	}

	// This should not panic
	err := outputMarkdown(results, 100, 50, 50, 50.0)
	if err != nil {
		t.Errorf("outputMarkdown failed: %v", err)
	}
}

// TestOutputTable removed - outputTable is for report command, not benchmark

func TestBenchResultCalculations(t *testing.T) {
	br := benchResult{
		suite:     "test",
		beforeTok: 1000,
		afterTok:  200,
		savedTok:  800,
		pctSaved:  80.0,
		duration:  5000000,
	}

	if br.savedTok != 800 {
		t.Errorf("savedTok = %d, want 800", br.savedTok)
	}

	if br.pctSaved != 80.0 {
		t.Errorf("pctSaved = %.1f, want 80.0", br.pctSaved)
	}

	// Verify duration is reasonable (5ms)
	durationMs := float64(br.duration) / 1_000_000
	if durationMs != 5.0 {
		t.Errorf("Duration = %.2fms, want 5.00ms", durationMs)
	}
}

func TestAvailableSuitesStr(t *testing.T) {
	str := availableSuitesStr()
	
	// Should contain some expected suites
	expectedSuites := []string{"git-status", "docker", "cargo"}
	for _, suite := range expectedSuites {
		if !strings.Contains(str, suite) {
			t.Errorf("availableSuitesStr() missing suite %q", suite)
		}
	}
}

func TestSampleOutputConstants(t *testing.T) {
	// Test that sample outputs are non-empty
	samples := map[string]string{
		"gitStatusOutputClean": gitStatusOutputClean,
		"gitLogOutput":         gitLogOutput,
		"gitDiffOutput":        gitDiffOutput,
		"cargoTestOutput":      cargoTestOutput,
		"cargoTestFailures":    cargoTestFailures,
		"pytestOutput":         pytestOutput,
		"goTestOutput":         goTestOutput,
		"cargoBuildOutput":     cargoBuildOutput,
		"eslintOutput":         eslintOutput,
		"ruffOutput":           ruffOutput,
		"dockerPsOutput":       dockerPsOutput,
		"lsOutputLarge":        lsOutputLarge,
	}

	for name, content := range samples {
		if len(content) == 0 {
			t.Errorf("sample output %q is empty", name)
		}
		if len(content) < 10 {
			t.Errorf("sample output %q is too short: %d chars", name, len(content))
		}
	}
}

func TestBenchmarkCaseContentQuality(t *testing.T) {
	// Verify test cases have realistic content
	for suiteName, cases := range benchmarkSuites {
		t.Run(suiteName, func(t *testing.T) {
			for i, tc := range cases {
				if len(tc.content) < 10 {
					t.Errorf("case %d in suite %q has content too short: %d chars", 
						i, suiteName, len(tc.content))
				}
				if len(tc.command) == 0 {
					t.Errorf("case %d in suite %q has empty command", i, suiteName)
				}
			}
		})
	}
}

func TestTokenEstimationConsistency(t *testing.T) {
	testCases := []struct {
		name    string
		content string
		minTok  int
		maxTok  int
	}{
		{"short", "hello world", 2, 10},
		{"medium", strings.Repeat("test ", 50), 40, 60},
		{"long", strings.Repeat("content ", 500), 400, 600},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tokens := core.EstimateTokens(tc.content)
			if tokens < tc.minTok || tokens > tc.maxTok {
				t.Errorf("EstimateTokens(%q) = %d, want between %d and %d",
					tc.name, tokens, tc.minTok, tc.maxTok)
			}
		})
	}
}

func TestPipelineProcessing(t *testing.T) {
	cfg := filter.PipelineConfig{
		Mode: filter.ModeMinimal,
	}
	pipeline := filter.NewPipelineCoordinator(cfg)

	input := gitStatusOutputClean
	output, _ := pipeline.Process(input)

	// Output should be shorter or equal
	if len(output) > len(input) {
		t.Errorf("processed output is longer than input: %d > %d", len(output), len(input))
	}

	// Output should not be empty
	if len(output) == 0 {
		t.Error("processed output is empty")
	}
}

func TestAggressiveMode(t *testing.T) {
	minimal := filter.PipelineConfig{Mode: filter.ModeMinimal}
	aggressive := filter.PipelineConfig{Mode: filter.ModeAggressive}

	pipelineMin := filter.NewPipelineCoordinator(minimal)
	pipelineAgg := filter.NewPipelineCoordinator(aggressive)

	input := cargoTestOutput
	outputMin, _ := pipelineMin.Process(input)
	outputAgg, _ := pipelineAgg.Process(input)

	// Aggressive should produce shorter or equal output
	if len(outputAgg) > len(outputMin) {
		t.Errorf("aggressive mode produced longer output: %d > %d", len(outputAgg), len(outputMin))
	}
}
