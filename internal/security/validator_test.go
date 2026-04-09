package security_test

import (
	"testing"

	"github.com/GrayCodeAI/tokman/internal/security"
)

func TestValidator_New(t *testing.T) {
	v := security.NewValidator()

	if v == nil {
		t.Fatal("expected validator to not be nil")
	}
}

func TestValidatePreset(t *testing.T) {
	v := security.NewValidator()

	tests := []struct {
		preset  string
		wantErr bool
	}{
		{"fast", false},
		{"balanced", false},
		{"full", false},
		{"", false},
		{"invalid", true},
	}

	for _, tt := range tests {
		err := v.ValidatePreset(tt.preset)
		if (err != nil) != tt.wantErr {
			t.Errorf("ValidatePreset(%q) error = %v, wantErr %v", tt.preset, err, tt.wantErr)
		}
	}
}

func TestValidateMode(t *testing.T) {
	v := security.NewValidator()

	tests := []struct {
		mode    string
		wantErr bool
	}{
		{"minimal", false},
		{"aggressive", false},
		{"", false},
		{"invalid", true},
	}

	for _, tt := range tests {
		err := v.ValidateMode(tt.mode)
		if (err != nil) != tt.wantErr {
			t.Errorf("ValidateMode(%q) error = %v, wantErr %v", tt.mode, err, tt.wantErr)
		}
	}
}

func TestValidateBudget(t *testing.T) {
	v := security.NewValidator()

	tests := []struct {
		budget  int
		wantErr bool
	}{
		{0, false},
		{1000, false},
		{10000000, false},
		{-1, true},
		{-100, true},
		{10000001, true}, // exceeds max
	}

	for _, tt := range tests {
		err := v.ValidateBudget(tt.budget)
		if (err != nil) != tt.wantErr {
			t.Errorf("ValidateBudget(%d) error = %v, wantErr %v", tt.budget, err, tt.wantErr)
		}
	}
}

func TestValidatePath(t *testing.T) {
	v := security.NewValidator()

	tests := []struct {
		path    string
		wantErr bool
	}{
		{"/tmp/test", false},
		{"./relative", false},
		{"../etc/passwd", true},    // traversal
		{"../../etc/passwd", true}, // traversal
		{"/tmp/test\x00", true},    // null byte
		{"test", false},
		{"", false},
	}

	for _, tt := range tests {
		err := v.ValidatePath(tt.path)
		if (err != nil) != tt.wantErr {
			t.Errorf("ValidatePath(%q) error = %v, wantErr %v", tt.path, err, tt.wantErr)
		}
	}
}

func TestValidateCommandName(t *testing.T) {
	v := security.NewValidator()

	tests := []struct {
		name    string
		wantErr bool
	}{
		{"ls", false},
		{"git", false},
		{"", true},      // empty
		{"ls;rm", true}, // shell meta
		{"ls|grep", true},
		{"`ls`", true},
		{"$(ls)", true},
	}

	for _, tt := range tests {
		err := v.ValidateCommandName(tt.name)
		if (err != nil) != tt.wantErr {
			t.Errorf("ValidateCommandName(%q) error = %v, wantErr %v", tt.name, err, tt.wantErr)
		}
	}
}

func TestSanitizeInput(t *testing.T) {
	v := security.NewValidator()

	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "hello"},
		{"hello\x00world", "helloworld"},
		{"test\x01\x02", "test"},
		{"", ""},
	}

	for _, tt := range tests {
		got := v.SanitizeInput(tt.input)
		if got != tt.expected {
			t.Errorf("SanitizeInput(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestIsSafeFilename(t *testing.T) {
	v := security.NewValidator()

	tests := []struct {
		filename string
		expected bool
	}{
		{"test.txt", true},
		{"file.go", true},
		{"", false},
		{"../etc/passwd", false},
		{"test/../passwd", false},
		{"test\x00.txt", false},
	}

	for _, tt := range tests {
		got := v.IsSafeFilename(tt.filename)
		if got != tt.expected {
			t.Errorf("IsSafeFilename(%q) = %v, want %v", tt.filename, got, tt.expected)
		}
	}
}

func TestValidateLayerName(t *testing.T) {
	v := security.NewValidator()

	tests := []struct {
		layer   string
		wantErr bool
	}{
		{"entropy", false},
		{"perplexity", false},
		{"h2o", false},
		{"compaction", false},
		{"invalid", true},
		{"", true},
	}

	for _, tt := range tests {
		err := v.ValidateLayerName(tt.layer)
		if (err != nil) != tt.wantErr {
			t.Errorf("ValidateLayerName(%q) error = %v, wantErr %v", tt.layer, err, tt.wantErr)
		}
	}
}

func TestValidateProfile(t *testing.T) {
	v := security.NewValidator()

	tests := []struct {
		profile string
		wantErr bool
	}{
		{"surface", false},
		{"trim", false},
		{"extract", false},
		{"core", false},
		{"code", false},
		{"log", false},
		{"thread", false},
		{"", false},
		{"invalid", true},
	}

	for _, tt := range tests {
		err := v.ValidateProfile(tt.profile)
		if (err != nil) != tt.wantErr {
			t.Errorf("ValidateProfile(%q) error = %v, wantErr %v", tt.profile, err, tt.wantErr)
		}
	}
}
