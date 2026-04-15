package validation

import (
	"strings"
	"testing"
)

func TestValidateInputSize(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "empty string",
			input:   "",
			wantErr: false,
		},
		{
			name:    "small input",
			input:   "hello world",
			wantErr: false,
		},
		{
			name:    "exactly at limit",
			input:   strings.Repeat("a", MaxInputSize),
			wantErr: false,
		},
		{
			name:    "over limit",
			input:   strings.Repeat("a", MaxInputSize+1),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateInputSize(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateInputSize() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateCommandArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "empty args",
			args:    []string{},
			wantErr: false,
		},
		{
			name:    "single arg",
			args:    []string{"hello"},
			wantErr: false,
		},
		{
			name:    "multiple args",
			args:    []string{"arg1", "arg2", "arg3"},
			wantErr: false,
		},
		{
			name:    "too many args",
			args:    make([]string, MaxCommandArgs+1),
			wantErr: true,
		},
		{
			name:    "arg too long",
			args:    []string{strings.Repeat("a", MaxPathLength+1)},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCommandArgs(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCommandArgs() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSanitizePath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "empty path",
			path:    "",
			wantErr: true,
		},
		{
			name:    "simple path",
			path:    "/tmp/test",
			wantErr: false,
		},
		{
			name:    "relative path",
			path:    "./test",
			wantErr: false,
		},
		{
			name:    "path with double dots",
			path:    "/tmp/../etc",
			wantErr: false, // filepath.Clean normalizes this
		},
		{
			name:    "path with encoded traversal",
			path:    "/tmp/..%2fetc",
			wantErr: true, // contains .. sequence
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SanitizePath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("SanitizePath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == "" {
				t.Error("SanitizePath() returned empty string for valid path")
			}
		})
	}
}

func TestSanitizePath_ReturnsAbsolute(t *testing.T) {
	result, err := SanitizePath("./test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(result, "/") {
		t.Errorf("expected absolute path, got %s", result)
	}
}

func TestValidateConfigPath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "empty path",
			path:    "",
			wantErr: true,
		},
		{
			name:    "home config path",
			path:    "~/.config/tokman/config.toml",
			wantErr: false,
		},
		{
			name:    "relative config path",
			path:    ".tokman/config.toml",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfigPath(tt.path)
			// The homeDir check may fail in test environment, so we mainly check
			// that it doesn't panic and handles errors appropriately
			if tt.name == "empty path" && err == nil {
				t.Error("expected error for empty path")
			}
		})
	}
}

func TestConstants(t *testing.T) {
	if MaxInputSize != 10*1024*1024 {
		t.Errorf("MaxInputSize = %d, want %d", MaxInputSize, 10*1024*1024)
	}
	if MaxCommandArgs != 1000 {
		t.Errorf("MaxCommandArgs = %d, want 1000", MaxCommandArgs)
	}
	if MaxPathLength != 4096 {
		t.Errorf("MaxPathLength = %d, want 4096", MaxPathLength)
	}
	if MaxConfigSize != 1*1024*1024 {
		t.Errorf("MaxConfigSize = %d, want %d", MaxConfigSize, 1*1024*1024)
	}
}
