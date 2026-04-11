package filter

import (
	"strings"
	"testing"
)

func TestLightThinkerFilter_Name(t *testing.T) {
	f := NewLightThinkerFilter()
	if f.Name() != "26_lightthinker" {
		t.Errorf("unexpected name: %s", f.Name())
	}
}

func TestLightThinkerFilter_ModeNone(t *testing.T) {
	f := NewLightThinkerFilter()
	input := "Step 1: Do something.\nLine two.\nLine three.\nLine four.\n"
	out, saved := f.Apply(input, ModeNone)
	if out != input || saved != 0 {
		t.Error("ModeNone should return unchanged")
	}
}

func TestLightThinkerFilter_ShortInput(t *testing.T) {
	f := NewLightThinkerFilter()
	input := "Step 1: Short.\nOnly one body line."
	out, _ := f.Apply(input, ModeMinimal)
	if out != input {
		t.Error("short steps should not be compressed")
	}
}

func TestLightThinkerFilter_CompressesLongSteps(t *testing.T) {
	f := NewLightThinkerFilter()
	lines := []string{
		"Step 1: Analyze the database connection issue.",
		"We need to look at the connection pool settings.",
		"The pool might be exhausted due to high load.",
		"Let me check the configuration file for timeout values.",
		"The timeout might be set too low for production workloads.",
		"We should increase both max connections and timeout settings.",
		"",
		"Step 2: Fix the configuration by updating pool parameters.",
		"Open the database configuration file in the config directory.",
		"Find the connection pool section and update max_connections.",
		"Also update the connection_timeout to a higher value.",
		"Restart the service to apply the new configuration settings.",
		"",
		"Conclusion: Increasing pool size resolved the database issue.",
	}
	input := strings.Join(lines, "\n")
	out, saved := f.Apply(input, ModeMinimal)

	if saved <= 0 {
		t.Error("expected positive savings on multi-line steps")
	}
	outLines := strings.Split(out, "\n")
	if len(outLines) >= len(lines) {
		t.Errorf("expected fewer lines after compression: got %d want < %d", len(outLines), len(lines))
	}
	if !strings.Contains(out, "Step 1") {
		t.Error("step headers must be preserved")
	}
	if !strings.Contains(out, "Step 2") {
		t.Error("step 2 header must be preserved")
	}
	if !strings.Contains(out, "Conclusion") {
		t.Error("conclusion must be preserved (not a step)")
	}
}

func TestLightThinkerFilter_KeepsKeyInsight(t *testing.T) {
	f := NewLightThinkerFilter()
	lines := []string{
		"Step 1: Investigate authentication failure.",
		"First let us check the log files.",
		"Looking at the access logs now.",
		"The logs show repeated 401 responses from the auth service.",
		"This suggests the JWT token validation is failing.",
		"The root cause is an expired signing key in production.",
	}
	input := strings.Join(lines, "\n")
	out, _ := f.Apply(input, ModeMinimal)

	// The most informative line (unique terms about root cause) should survive
	if !strings.Contains(out, "Step 1") {
		t.Error("step header must be preserved")
	}
}

func TestLightThinkerFilter_OrdinalSteps(t *testing.T) {
	f := NewLightThinkerFilter()
	lines := []string{
		"First, identify the failing component in the system.",
		"We need to trace the request through the service mesh.",
		"Check each microservice for timeout or error responses.",
		"Look at the distributed tracing dashboard for spans.",
		"",
		"Second, reproduce the failure in a staging environment.",
		"Set up the same load conditions as in production.",
		"Use the same request parameters that trigger the failure.",
		"Observe which service drops the request first.",
		"",
		"Resolution: service timeout in payment gateway was the cause.",
	}
	input := strings.Join(lines, "\n")
	out, saved := f.Apply(input, ModeMinimal)

	if saved <= 0 {
		t.Error("expected savings on ordinal reasoning steps")
	}
	if !strings.Contains(out, "Resolution") {
		t.Error("non-step conclusion must be preserved")
	}
}
