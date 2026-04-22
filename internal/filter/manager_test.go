package filter

import (
	"strings"
	"sync"
	"testing"

	"github.com/GrayCodeAI/tok/internal/config"
)

func TestCheckStructure_ValidJSON(t *testing.T) {
	m := &PipelineManager{}

	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "empty",
			input: "",
			want:  true,
		},
		{
			name:  "simple JSON object",
			input: `{"key": "value"}`,
			want:  true,
		},
		{
			name:  "nested JSON",
			input: `{"a": {"b": [1, 2, 3]}}`,
			want:  true,
		},
		{
			name:  "JSON array",
			input: `[{"name": "test"}, {"name": "test2"}]`,
			want:  true,
		},
		{
			name:  "unbalanced braces in JSON",
			input: `{"key": "value"`,
			want:  false,
		},
		{
			name:  "unbalanced brackets in JSON",
			input: `[1, 2, 3`,
			want:  false,
		},
		{
			name:  "extra closing brace in JSON",
			input: `{"key": "value"}}`,
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := m.checkStructure(tt.input)
			if got != tt.want {
				t.Errorf("checkStructure(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestCheckStructure_GeneralText(t *testing.T) {
	m := &PipelineManager{}

	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "plain text",
			input: "Hello world this is some text",
			want:  true,
		},
		{
			name:  "code with moderate imbalance",
			input: "if (x > 0) { doSomething();",
			want:  true, // general text allows moderate imbalance
		},
		{
			name:  "severe negative imbalance",
			input: strings.Repeat(")", 15),
			want:  false,
		},
		{
			name:  "excessive positive imbalance",
			input: strings.Repeat("{", 60),
			want:  false,
		},
		{
			name:  "balanced code",
			input: "func main() { if (true) { return } }",
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := m.checkStructure(tt.input)
			if got != tt.want {
				t.Errorf("checkStructure(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestCheckStructure_JSONDetection(t *testing.T) {
	m := &PipelineManager{}

	// These look like JSON due to quotes/colons/commas
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "valid JSON requires exact balance",
			input: `{"a": 1, "b": 2}`,
			want:  true,
		},
		{
			name:  "JSON with one missing brace",
			input: `{"a": 1, "b": 2`,
			want:  false,
		},
		{
			name:  "JSON-like but not JSON - no quotes",
			input: `{a: 1, b: 2}`,
			want:  true, // no quotes detected, treated as general text
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := m.checkStructure(tt.input)
			if got != tt.want {
				t.Errorf("checkStructure(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestCacheKey_Deterministic(t *testing.T) {
	m := NewPipelineManager(ManagerConfig{PipelineCfg: PipelineConfig{Mode: ModeMinimal}})
	ctx := config.CommandContext{Command: "test", Intent: "debug"}

	key1 := m.cacheKey("input", ModeMinimal, ctx)
	key2 := m.cacheKey("input", ModeMinimal, ctx)

	if key1 != key2 {
		t.Errorf("cacheKey not deterministic: %q != %q", key1, key2)
	}
}

func TestCacheKey_DifferentiatesInputs(t *testing.T) {
	m := NewPipelineManager(ManagerConfig{PipelineCfg: PipelineConfig{Mode: ModeMinimal}})
	ctx := config.CommandContext{Command: "test", Intent: "debug"}

	key1 := m.cacheKey("input1", ModeMinimal, ctx)
	key2 := m.cacheKey("input2", ModeMinimal, ctx)

	if key1 == key2 {
		t.Error("cacheKey should differentiate different inputs")
	}
}

func TestCacheKey_IncludesModeAndContext(t *testing.T) {
	m := NewPipelineManager(ManagerConfig{PipelineCfg: PipelineConfig{Mode: ModeMinimal}})
	ctx := config.CommandContext{Command: "test", Intent: "debug"}

	key1 := m.cacheKey("input", ModeMinimal, ctx)
	key2 := m.cacheKey("input", ModeAggressive, ctx)

	if key1 == key2 {
		t.Error("cacheKey should differentiate modes")
	}

	ctx2 := config.CommandContext{Command: "other", Intent: "debug"}
	key3 := m.cacheKey("input", ModeMinimal, ctx2)

	if key1 == key3 {
		t.Error("cacheKey should differentiate commands")
	}
}

func TestCoordinatorForRequest_Independent(t *testing.T) {
	cfg := ManagerConfig{
		PipelineCfg: PipelineConfig{Mode: ModeMinimal},
	}
	m := NewPipelineManager(cfg)

	// Create two coordinators for different requests
	c1 := m.coordinatorForRequest(ModeMinimal, "query1")
	c2 := m.coordinatorForRequest(ModeAggressive, "query2")

	// They should be independent (different pointers)
	if c1 == c2 {
		t.Fatal("coordinatorForRequest returned same pointer for different requests")
	}

	// Config should reflect request parameters
	if c1.config.Mode != ModeMinimal {
		t.Errorf("c1 mode = %v, want ModeMinimal", c1.config.Mode)
	}
	if c2.config.Mode != ModeAggressive {
		t.Errorf("c2 mode = %v, want ModeAggressive", c2.config.Mode)
	}
	if c1.config.QueryIntent != "query1" {
		t.Errorf("c1 intent = %q, want 'query1'", c1.config.QueryIntent)
	}
	if c2.config.QueryIntent != "query2" {
		t.Errorf("c2 intent = %q, want 'query2'", c2.config.QueryIntent)
	}
}

func TestCoordinatorForRequest_NoLockContention(t *testing.T) {
	cfg := ManagerConfig{
		PipelineCfg: PipelineConfig{Mode: ModeMinimal},
	}
	m := NewPipelineManager(cfg)

	// Run many concurrent coordinator creations
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(n int) {
			mode := ModeMinimal
			if n%2 == 0 {
				mode = ModeAggressive
			}
			_ = m.coordinatorForRequest(mode, "concurrent")
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestProcessWithBudget(t *testing.T) {
	cfg := ManagerConfig{
		MaxContextTokens: 1000000,
		PipelineCfg:      PipelineConfig{Mode: ModeMinimal},
		CacheEnabled:     false,
	}
	m := NewPipelineManager(cfg)
	ctx := config.CommandContext{Command: "test"}

	result, err := m.ProcessWithBudget("hello world", ModeMinimal, 100, ctx)
	if err != nil {
		t.Fatalf("ProcessWithBudget error = %v", err)
	}
	if result == nil {
		t.Fatal("ProcessWithBudget returned nil result")
	}
	if result.OriginalTokens == 0 {
		t.Error("OriginalTokens should be > 0")
	}
}

func TestProcess_ConcurrentMixedModes(t *testing.T) {
	cfg := ManagerConfig{
		MaxContextTokens: 1000000,
		PipelineCfg:      PipelineConfig{Mode: ModeMinimal},
		CacheEnabled:     true,
	}
	m := NewPipelineManager(cfg)

	const workers = 32
	const perWorker = 20

	var wg sync.WaitGroup
	errs := make(chan error, workers*perWorker)

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			mode := ModeMinimal
			switch n % 3 {
			case 1:
				mode = ModeAggressive
			case 2:
				mode = ModeNone
			}
			ctx := config.CommandContext{
				Command: "cmd",
				Intent:  "concurrent",
			}
			for j := 0; j < perWorker; j++ {
				input := strings.Repeat("hello world ", (n+j)%5+1)
				if _, err := m.Process(input, mode, ctx); err != nil {
					errs <- err
					return
				}
			}
		}(i)
	}

	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatalf("concurrent Process error: %v", err)
	}
}

func TestProcess_InputTooLarge(t *testing.T) {
	cfg := ManagerConfig{
		MaxContextTokens: 10,
		PipelineCfg:      PipelineConfig{Mode: ModeMinimal},
	}
	m := NewPipelineManager(cfg)
	ctx := config.CommandContext{Command: "test"}

	_, err := m.Process("this input is definitely more than ten tokens long", ModeMinimal, ctx)
	if err == nil {
		t.Error("expected error for input exceeding max tokens")
	}
}
