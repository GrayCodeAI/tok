package filter

import (
	"strings"
	"testing"
)

func TestNewAdaptiveLayerSelector(t *testing.T) {
	a := NewAdaptiveLayerSelector()
	if a == nil {
		t.Fatal("expected non-nil selector")
	}
	if a.codeThreshold != 0.15 {
		t.Errorf("expected codeThreshold 0.15, got %f", a.codeThreshold)
	}
}

func TestAnalyzeContent_Empty(t *testing.T) {
	a := NewAdaptiveLayerSelector()
	if got := a.AnalyzeContent(""); got != ContentTypeUnknown {
		t.Errorf("expected Unknown for empty input, got %v", got)
	}
}

func TestAnalyzeContent_Code(t *testing.T) {
	a := NewAdaptiveLayerSelector()
	input := "func main() {\n\treturn 42\n}\n"
	got := a.AnalyzeContent(input)
	if got != ContentTypeCode {
		t.Errorf("expected Code, got %v (%s)", got, got.String())
	}
}

func TestAnalyzeContent_Logs(t *testing.T) {
	a := NewAdaptiveLayerSelector()
	input := "INFO: starting service\nERROR: connection failed\nWARN: retrying\n"
	got := a.AnalyzeContent(input)
	if got != ContentTypeLogs {
		t.Errorf("expected Logs, got %v (%s)", got, got.String())
	}
}

func TestAnalyzeContent_Conversation(t *testing.T) {
	a := NewAdaptiveLayerSelector()
	input := "User: hello\nAssistant: hi there\nUser: how are you?\n"
	got := a.AnalyzeContent(input)
	if got != ContentTypeConversation {
		t.Errorf("expected Conversation, got %v (%s)", got, got.String())
	}
}

func TestAnalyzeContent_GitOutput(t *testing.T) {
	a := NewAdaptiveLayerSelector()
	input := "On branch main\nmodified: file.go\ncommit abc123\n"
	got := a.AnalyzeContent(input)
	if got != ContentTypeGitOutput {
		t.Errorf("expected GitOutput, got %v (%s)", got, got.String())
	}
}

func TestAnalyzeContent_TestOutput(t *testing.T) {
	a := NewAdaptiveLayerSelector()
	input := "=== RUN TestFoo\n--- PASS: TestFoo\nPASS\n"
	got := a.AnalyzeContent(input)
	if got != ContentTypeTestOutput {
		t.Errorf("expected TestOutput, got %v (%s)", got, got.String())
	}
}

func TestAnalyzeContent_DockerOutput(t *testing.T) {
	a := NewAdaptiveLayerSelector()
	input := "CONTAINER ID   IMAGE\nabc123         nginx\n"
	got := a.AnalyzeContent(input)
	if got != ContentTypeDockerOutput {
		t.Errorf("expected DockerOutput, got %v (%s)", got, got.String())
	}
}

func TestAnalyzeContent_Mixed(t *testing.T) {
	a := NewAdaptiveLayerSelector()
	// Mix code and logs indicators heavily
	lines := []string{
		"func main() {",
		"INFO: starting",
		"func helper() {",
		"ERROR: failed",
		"return 42",
		"WARN: retry",
	}
	input := strings.Join(lines, "\n")
	got := a.AnalyzeContent(input)
	if got != ContentTypeMixed {
		t.Errorf("expected Mixed for balanced code+logs, got %v (%s)", got, got.String())
	}
}

func TestContentTypeString(t *testing.T) {
	tests := []struct {
		ct   ContentType
		want string
	}{
		{ContentTypeUnknown, "unknown"},
		{ContentTypeCode, "code"},
		{ContentTypeLogs, "logs"},
		{ContentTypeConversation, "conversation"},
		{ContentTypeGitOutput, "git"},
		{ContentTypeTestOutput, "test"},
		{ContentTypeDockerOutput, "docker/infra"},
		{ContentTypeMixed, "mixed"},
		{ContentType(99), "unknown"},
	}
	for _, tt := range tests {
		if got := tt.ct.String(); got != tt.want {
			t.Errorf("ContentType(%d).String() = %q, want %q", tt.ct, got, tt.want)
		}
	}
}

func TestRecommendedConfig(t *testing.T) {
	a := NewAdaptiveLayerSelector()

	tests := []struct {
		ct             ContentType
		wantCompaction bool
		wantH2O        bool
	}{
		{ContentTypeCode, false, true},
		{ContentTypeLogs, false, true},
		{ContentTypeConversation, true, true},
		{ContentTypeGitOutput, false, false},
		{ContentTypeTestOutput, false, true},
		{ContentTypeDockerOutput, false, true},
		{ContentTypeMixed, true, true},
		{ContentTypeUnknown, false, true},
	}

	for _, tt := range tests {
		cfg := a.RecommendedConfig(tt.ct, ModeMinimal)
		if cfg.EnableCompaction != tt.wantCompaction {
			t.Errorf("RecommendedConfig(%v).EnableCompaction = %v, want %v",
				tt.ct, cfg.EnableCompaction, tt.wantCompaction)
		}
		if cfg.EnableH2O != tt.wantH2O {
			t.Errorf("RecommendedConfig(%v).EnableH2O = %v, want %v",
				tt.ct, cfg.EnableH2O, tt.wantH2O)
		}
	}
}

func TestOptimizePipeline(t *testing.T) {
	a := NewAdaptiveLayerSelector()
	input := "func main() {}\n"
	coord := a.OptimizePipeline(input, ModeMinimal)
	if coord == nil {
		t.Fatal("expected non-nil coordinator")
	}
}

func TestQuickTierEnable(t *testing.T) {
	tests := []struct {
		useCase string
		minLen  int
		maxLen  int
	}{
		{"minimal", 1, 3},
		{"standard", 2, 4},
		{"aggressive", 3, 5},
		{"maximum", 4, 6},
		{"unknown", 2, 4},
	}
	for _, tt := range tests {
		tiers := QuickTierEnable(tt.useCase)
		if len(tiers) < tt.minLen || len(tiers) > tt.maxLen {
			t.Errorf("QuickTierEnable(%q) returned %d tiers, expected %d-%d",
				tt.useCase, len(tiers), tt.minLen, tt.maxLen)
		}
	}
}

func TestRecommendedConfigWithTiers(t *testing.T) {
	a := NewAdaptiveLayerSelector()
	cfg := a.RecommendedConfigWithTiers(ContentTypeCode, ModeMinimal, 1000, "debug")
	if cfg.Mode != ModeMinimal {
		t.Errorf("expected ModeMinimal, got %v", cfg.Mode)
	}
}
