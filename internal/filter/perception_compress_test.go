package filter

import (
	"strings"
	"testing"
)

func TestPerceptionCompressFilter_Name(t *testing.T) {
	f := NewPerceptionCompressFilter()
	if f.Name() != "25_perception_compress" {
		t.Errorf("unexpected name: %s", f.Name())
	}
}

func TestPerceptionCompressFilter_ModeNone(t *testing.T) {
	f := NewPerceptionCompressFilter()
	input := strings.Repeat("the quick brown fox jumps over the lazy dog\n", 20)
	out, saved := f.Apply(input, ModeNone)
	if out != input || saved != 0 {
		t.Error("ModeNone should return input unchanged")
	}
}

func TestPerceptionCompressFilter_ShortInput(t *testing.T) {
	f := NewPerceptionCompressFilter()
	input := "line one\nline two\nline three"
	out, _ := f.Apply(input, ModeMinimal)
	if out != input {
		t.Error("short input should pass through unchanged")
	}
}

func TestPerceptionCompressFilter_RemovesRedundantProse(t *testing.T) {
	f := NewPerceptionCompressFilter()
	// Consecutive lines with very high term overlap
	lines := []string{
		"The system processes incoming requests through the authentication layer.",
		"Requests are processed by the system via the authentication layer pipeline.",
		"Authentication layer receives and processes system requests from the pipeline.",
		"All requests to the system go through the authentication and processing layer.",
		"The pipeline authentication layer processes each incoming system request.",
		"",
		"Error: connection refused at port 5432 database unavailable",
		"",
		"The authentication layer handles system requests via processing pipeline.",
		"System authentication processes all incoming requests through the layer.",
		"Processing requests in the system authentication layer pipeline.",
	}
	input := strings.Join(lines, "\n")
	out, saved := f.Apply(input, ModeMinimal)

	if saved <= 0 {
		t.Error("expected savings on repetitive prose")
	}
	// Error line must always be preserved
	if !strings.Contains(out, "connection refused") {
		t.Error("error line must be preserved regardless of overlap")
	}
}

func TestPerceptionCompressFilter_PreservesDistinctLines(t *testing.T) {
	f := NewPerceptionCompressFilter()
	lines := []string{
		"The server starts listening on port 8080.",
		"Database migration completed successfully.",
		"Worker pool initialised with 4 threads.",
		"Cache warm-up finished loading 1200 entries.",
		"Health check endpoint registered at slash health.",
		"TLS certificate loaded from slash etc slash certs.",
		"Configuration loaded from environment variables.",
		"Application ready to serve incoming connections.",
	}
	input := strings.Join(lines, "\n")
	out, _ := f.Apply(input, ModeMinimal)

	outLines := strings.Split(strings.TrimSpace(out), "\n")
	// All lines are distinct — most should survive
	if len(outLines) < len(lines)-2 {
		t.Errorf("distinct lines should not be aggressively removed; kept %d of %d", len(outLines), len(lines))
	}
}

func TestPerceptionCompressFilter_AggressiveMoreReduction(t *testing.T) {
	f := NewPerceptionCompressFilter()
	lines := []string{
		"Processing authentication request for user with token credentials.",
		"User authentication processing with token and credential validation.",
		"Token validation for user credentials during authentication processing.",
		"Credentials processed for user token during authentication pipeline.",
		"Authentication pipeline processes user token credential validation.",
		"",
		"CRITICAL: authentication service unreachable",
		"",
		"Validating user authentication token credentials in pipeline.",
		"Processing token authentication credentials for user validation.",
	}
	input := strings.Join(lines, "\n")
	_, savedMin := f.Apply(input, ModeMinimal)
	_, savedAgg := f.Apply(input, ModeAggressive)
	if savedAgg < savedMin {
		t.Error("aggressive mode should save at least as many tokens as minimal")
	}
}

func TestPerceptionCompressFilter_AlwaysKeepsFirstAndLast(t *testing.T) {
	f := NewPerceptionCompressFilter()
	var lines []string
	lines = append(lines, "FIRST: unique opening statement with distinct terminology")
	for i := 0; i < 15; i++ {
		lines = append(lines, "middle line repeating the same words over and over again continuously")
	}
	lines = append(lines, "LAST: unique closing statement with distinct conclusion marker")
	input := strings.Join(lines, "\n")
	out, _ := f.Apply(input, ModeAggressive)

	if !strings.Contains(out, "FIRST:") {
		t.Error("first line must always be preserved")
	}
	if !strings.Contains(out, "LAST:") {
		t.Error("last line must always be preserved")
	}
}
