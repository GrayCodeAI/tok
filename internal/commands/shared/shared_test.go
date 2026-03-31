package shared

import (
	"os"
	"sync"
	"testing"
)

func TestIsVerbose(t *testing.T) {
	Verbose = 0
	if IsVerbose() {
		t.Error("expected false when Verbose=0")
	}
	Verbose = 1
	if !IsVerbose() {
		t.Error("expected true when Verbose=1")
	}
	Verbose = 2
	if !IsVerbose() {
		t.Error("expected true when Verbose=2")
	}
	Verbose = 0
}

func TestIsUltraCompact(t *testing.T) {
	UltraCompact = false
	if IsUltraCompact() {
		t.Error("expected false when UltraCompact=false")
	}
	UltraCompact = true
	if !IsUltraCompact() {
		t.Error("expected true when UltraCompact=true")
	}
	UltraCompact = false
}

func TestIsQuietMode(t *testing.T) {
	QuietMode = false
	if IsQuietMode() {
		t.Error("expected false when QuietMode=false")
	}
	QuietMode = true
	if !IsQuietMode() {
		t.Error("expected true when QuietMode=true")
	}
	QuietMode = false
}

func TestIsStreamMode(t *testing.T) {
	StreamMode = false
	if IsStreamMode() {
		t.Error("expected false when StreamMode=false")
	}
	StreamMode = true
	if !IsStreamMode() {
		t.Error("expected true when StreamMode=true")
	}
	StreamMode = false
}

func TestGetQueryIntent(t *testing.T) {
	os.Unsetenv("TOKMAN_QUERY")

	// Flag takes precedence
	QueryIntent = "debug"
	if got := GetQueryIntent(); got != "debug" {
		t.Errorf("expected 'debug', got %q", got)
	}

	// Falls back to env var
	QueryIntent = ""
	os.Setenv("TOKMAN_QUERY", "fix bug")
	if got := GetQueryIntent(); got != "fix bug" {
		t.Errorf("expected 'fix bug', got %q", got)
	}

	// Empty when neither set
	os.Unsetenv("TOKMAN_QUERY")
	if got := GetQueryIntent(); got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

func TestIsLLMEnabled(t *testing.T) {
	os.Unsetenv("TOKMAN_LLM")

	LLMEnabled = true
	if !IsLLMEnabled() {
		t.Error("expected true when LLMEnabled=true")
	}

	LLMEnabled = false
	os.Setenv("TOKMAN_LLM", "true")
	if !IsLLMEnabled() {
		t.Error("expected true when TOKMAN_LLM=true")
	}

	os.Unsetenv("TOKMAN_LLM")
	if IsLLMEnabled() {
		t.Error("expected false when neither flag nor env set")
	}
}

func TestGetTokenBudget(t *testing.T) {
	os.Unsetenv("TOKMAN_BUDGET")

	// Flag takes precedence
	TokenBudget = 1000
	if got := GetTokenBudget(); got != 1000 {
		t.Errorf("expected 1000, got %d", got)
	}

	// Falls back to env var
	TokenBudget = 0
	os.Setenv("TOKMAN_BUDGET", "2000")
	if got := GetTokenBudget(); got != 2000 {
		t.Errorf("expected 2000, got %d", got)
	}

	// Zero when neither set
	os.Unsetenv("TOKMAN_BUDGET")
	if got := GetTokenBudget(); got != 0 {
		t.Errorf("expected 0, got %d", got)
	}
}

func TestGetLayerPreset(t *testing.T) {
	os.Unsetenv("TOKMAN_PRESET")

	LayerPreset = "fast"
	if got := GetLayerPreset(); got != "fast" {
		t.Errorf("expected 'fast', got %q", got)
	}

	LayerPreset = ""
	os.Setenv("TOKMAN_PRESET", "balanced")
	if got := GetLayerPreset(); got != "balanced" {
		t.Errorf("expected 'balanced', got %q", got)
	}

	os.Unsetenv("TOKMAN_PRESET")
}

func TestIsReversibleEnabled(t *testing.T) {
	os.Unsetenv("TOKMAN_REVERSIBLE")

	ReversibleEnabled = true
	if !IsReversibleEnabled() {
		t.Error("expected true when ReversibleEnabled=true")
	}

	ReversibleEnabled = false
	os.Setenv("TOKMAN_REVERSIBLE", "true")
	if !IsReversibleEnabled() {
		t.Error("expected true when TOKMAN_REVERSIBLE=true")
	}

	os.Unsetenv("TOKMAN_REVERSIBLE")
	if IsReversibleEnabled() {
		t.Error("expected false when neither set")
	}
}

func TestIsRemoteMode(t *testing.T) {
	os.Unsetenv("TOKMAN_REMOTE")

	RemoteMode = true
	if !IsRemoteMode() {
		t.Error("expected true when RemoteMode=true")
	}

	RemoteMode = false
	os.Setenv("TOKMAN_REMOTE", "true")
	if !IsRemoteMode() {
		t.Error("expected true when TOKMAN_REMOTE=true")
	}

	os.Unsetenv("TOKMAN_REMOTE")
	if IsRemoteMode() {
		t.Error("expected false when neither set")
	}
}

func TestGetCompressionAddr(t *testing.T) {
	os.Unsetenv("TOKMAN_COMPRESSION_ADDR")

	CompressionAddr = "localhost:9090"
	if got := GetCompressionAddr(); got != "localhost:9090" {
		t.Errorf("expected 'localhost:9090', got %q", got)
	}

	CompressionAddr = ""
	os.Setenv("TOKMAN_COMPRESSION_ADDR", "remote:9090")
	if got := GetCompressionAddr(); got != "remote:9090" {
		t.Errorf("expected 'remote:9090', got %q", got)
	}

	os.Unsetenv("TOKMAN_COMPRESSION_ADDR")
}

func TestGetAnalyticsAddr(t *testing.T) {
	os.Unsetenv("TOKMAN_ANALYTICS_ADDR")

	AnalyticsAddr = "localhost:9091"
	if got := GetAnalyticsAddr(); got != "localhost:9091" {
		t.Errorf("expected 'localhost:9091', got %q", got)
	}

	AnalyticsAddr = ""
	os.Setenv("TOKMAN_ANALYTICS_ADDR", "remote:9091")
	if got := GetAnalyticsAddr(); got != "remote:9091" {
		t.Errorf("expected 'remote:9091', got %q", got)
	}

	os.Unsetenv("TOKMAN_ANALYTICS_ADDR")
}

func TestGetRemoteTimeout(t *testing.T) {
	RemoteTimeout = 60
	if got := GetRemoteTimeout(); got != 60 {
		t.Errorf("expected 60, got %d", got)
	}

	RemoteTimeout = 0
	if got := GetRemoteTimeout(); got != 30 {
		t.Errorf("expected default 30, got %d", got)
	}
}

func TestGetEnableLayers(t *testing.T) {
	EnableLayers = []string{"entropy", "compaction"}
	layers := GetEnableLayers()
	if len(layers) != 2 {
		t.Errorf("expected 2 layers, got %d", len(layers))
	}
	if layers[0] != "entropy" {
		t.Errorf("expected 'entropy', got %q", layers[0])
	}
	EnableLayers = nil
}

func TestGetDisableLayers(t *testing.T) {
	DisableLayers = []string{"h2o", "attention_sink"}
	layers := GetDisableLayers()
	if len(layers) != 2 {
		t.Errorf("expected 2 layers, got %d", len(layers))
	}
	DisableLayers = nil
}

func TestSetFlags(t *testing.T) {
	cfg := FlagConfig{
		Verbose:             2,
		DryRun:              true,
		UltraCompact:        true,
		QueryIntent:         "test",
		LLMEnabled:          true,
		TokenBudget:         500,
		LayerPreset:         "fast",
		QuietMode:           true,
		JSONOutput:          true,
		RemoteMode:          true,
		CompressionAddr:     "localhost:9090",
		AnalyticsAddr:       "localhost:9091",
		RemoteTimeout:       45,
		CompactionEnabled:   true,
		CompactionThreshold: 1000,
		StreamMode:          true,
		EnableLayers:        []string{"entropy"},
		DisableLayers:       []string{"h2o"},
	}

	SetFlags(cfg)

	if Verbose != 2 {
		t.Errorf("Verbose: expected 2, got %d", Verbose)
	}
	if !DryRun {
		t.Error("DryRun: expected true")
	}
	if !UltraCompact {
		t.Error("UltraCompact: expected true")
	}
	if QueryIntent != "test" {
		t.Errorf("QueryIntent: expected 'test', got %q", QueryIntent)
	}
	if !LLMEnabled {
		t.Error("LLMEnabled: expected true")
	}
	if TokenBudget != 500 {
		t.Errorf("TokenBudget: expected 500, got %d", TokenBudget)
	}
	if LayerPreset != "fast" {
		t.Errorf("LayerPreset: expected 'fast', got %q", LayerPreset)
	}
	if !QuietMode {
		t.Error("QuietMode: expected true")
	}
	if !JSONOutput {
		t.Error("JSONOutput: expected true")
	}
	if !RemoteMode {
		t.Error("RemoteMode: expected true")
	}
	if CompressionAddr != "localhost:9090" {
		t.Errorf("CompressionAddr: expected 'localhost:9090', got %q", CompressionAddr)
	}
	if AnalyticsAddr != "localhost:9091" {
		t.Errorf("AnalyticsAddr: expected 'localhost:9091', got %q", AnalyticsAddr)
	}
	if RemoteTimeout != 45 {
		t.Errorf("RemoteTimeout: expected 45, got %d", RemoteTimeout)
	}
	if !CompactionEnabled {
		t.Error("CompactionEnabled: expected true")
	}
	if CompactionThreshold != 1000 {
		t.Errorf("CompactionThreshold: expected 1000, got %d", CompactionThreshold)
	}
	if !StreamMode {
		t.Error("StreamMode: expected true")
	}
	if len(EnableLayers) != 1 || EnableLayers[0] != "entropy" {
		t.Errorf("EnableLayers: expected ['entropy'], got %v", EnableLayers)
	}
	if len(DisableLayers) != 1 || DisableLayers[0] != "h2o" {
		t.Errorf("DisableLayers: expected ['h2o'], got %v", DisableLayers)
	}
}

func TestSetFlags_Concurrent(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(v int) {
			defer wg.Done()
			cfg := FlagConfig{
				Verbose:     v,
				TokenBudget: v * 100,
			}
			SetFlags(cfg)
			_ = IsVerbose()
			_ = GetTokenBudget()
		}(i)
	}
	wg.Wait()
}

func TestTeeOnFailure_NoError(t *testing.T) {
	result := TeeOnFailure("output", "test-cmd", nil)
	if result != "" {
		t.Errorf("expected empty string on no error, got %q", result)
	}
}

func TestGetModelFamily(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"claude-3-opus", "claude"},
		{"claude-3.5-sonnet", "claude"},
		{"gpt-4", "gpt"},
		{"gpt-4o", "gpt"},
		{"o1-preview", "gpt"},
		{"o3-mini", "gpt"},
		{"gemini-pro", "gemini"},
		{"llama-3", "llama"},
		{"meta-llama", "llama"},
		{"qwen-2.5", "qwen"},
		{"deepseek-v3", "deepseek"},
		{"mistral-large", "mistral"},
		{"mixtral", "mistral"},
		{"unknown-model", "other"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := GetModelFamily(tt.input)
			if got != tt.expected {
				t.Errorf("GetModelFamily(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
