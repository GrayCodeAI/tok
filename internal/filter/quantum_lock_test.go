package filter

import (
	"strings"
	"testing"
)

func TestQuantumLock_ISODate(t *testing.T) {
	filter := NewQuantumLockFilter()
	input := "Current time: 2026-04-10T23:30:00Z"

	output, _ := filter.Apply(input, ModeMinimal)

	if !strings.Contains(output, "<DATE>") {
		t.Error("Expected <DATE> placeholder")
	}
	if !strings.Contains(output, "<DYNAMIC_CONTEXT>") {
		t.Error("Expected dynamic context block")
	}
	if !strings.Contains(output, "iso_date: 2026-04-10T23:30:00Z") {
		t.Error("Expected original date in context")
	}
}

func TestQuantumLock_APIKey(t *testing.T) {
	filter := NewQuantumLockFilter()
	input := "API Key: sk-abc123def456ghi789"

	output, _ := filter.Apply(input, ModeMinimal)

	if !strings.Contains(output, "<API_KEY>") {
		t.Error("Expected <API_KEY> placeholder")
	}
	if !strings.Contains(output, "api_key: sk-abc123def456ghi789") {
		t.Error("Expected original key in context")
	}
}

func TestQuantumLock_UUID(t *testing.T) {
	filter := NewQuantumLockFilter()
	input := "Request ID: 550e8400-e29b-41d4-a716-446655440000"

	output, _ := filter.Apply(input, ModeMinimal)

	if !strings.Contains(output, "<UUID>") {
		t.Error("Expected <UUID> placeholder")
	}
	if !strings.Contains(output, "uuid: 550e8400-e29b-41d4-a716-446655440000") {
		t.Error("Expected original UUID in context")
	}
}

func TestQuantumLock_JWT(t *testing.T) {
	filter := NewQuantumLockFilter()
	input := "Token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U"

	output, _ := filter.Apply(input, ModeMinimal)

	if !strings.Contains(output, "<JWT>") {
		t.Error("Expected <JWT> placeholder")
	}
}

func TestQuantumLock_MultiplePatterns(t *testing.T) {
	filter := NewQuantumLockFilter()
	input := `System prompt:
Current time: 2026-04-10T23:30:00Z
API Key: sk-test123456789abc
Request ID: 550e8400-e29b-41d4-a716-446655440000`

	output, _ := filter.Apply(input, ModeMinimal)

	if !strings.Contains(output, "<DATE>") {
		t.Error("Expected <DATE> placeholder")
	}
	if !strings.Contains(output, "<API_KEY>") {
		t.Errorf("Expected <API_KEY> placeholder, got: %s", output)
	}
	if !strings.Contains(output, "<UUID>") {
		t.Error("Expected <UUID> placeholder")
	}

	// Check context block has all values
	if !strings.Contains(output, "iso_date:") {
		t.Error("Expected iso_date in context")
	}
	if !strings.Contains(output, "api_key:") {
		t.Errorf("Expected api_key in context, got: %s", output)
	}
	if !strings.Contains(output, "uuid:") {
		t.Error("Expected uuid in context")
	}
}

func TestQuantumLock_NoDynamicContent(t *testing.T) {
	filter := NewQuantumLockFilter()
	input := "This is a static system prompt with no dynamic content"

	output, saved := filter.Apply(input, ModeMinimal)

	if output != input {
		t.Error("Expected unchanged output for static content")
	}
	if saved != 0 {
		t.Error("Expected 0 tokens saved for static content")
	}
}

func TestQuantumLock_UnixTimestamp(t *testing.T) {
	filter := NewQuantumLockFilter()
	input := "Timestamp: 1712778600"

	output, _ := filter.Apply(input, ModeMinimal)

	if !strings.Contains(output, "<TIMESTAMP>") {
		t.Error("Expected <TIMESTAMP> placeholder")
	}
}

func TestQuantumLock_HexID(t *testing.T) {
	filter := NewQuantumLockFilter()
	input := "Session: a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2"

	output, _ := filter.Apply(input, ModeMinimal)

	if !strings.Contains(output, "<HEX_ID>") {
		t.Error("Expected <HEX_ID> placeholder")
	}
}
