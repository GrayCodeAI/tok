package tui

import (
	"strings"
	"testing"

	"github.com/GrayCodeAI/tok/internal/config"
)

func TestLoadKeyMapDefaults(t *testing.T) {
	// Empty config should return defaults
	cfg := config.KeybindingsConfig{}
	km, err := LoadKeyMap(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify some default bindings are present
	if len(km.Quit.Keys()) == 0 {
		t.Error("expected Quit to have default keys")
	}
	if len(km.NextSection.Keys()) == 0 {
		t.Error("expected NextSection to have default keys")
	}
	if len(km.HistoryBack.Keys()) == 0 {
		t.Error("expected HistoryBack to have default keys")
	}
}

func TestLoadKeyMapOverrides(t *testing.T) {
	cfg := config.KeybindingsConfig{
		Quit:        "F10",
		Refresh:     "ctrl+r",
		HistoryBack: "left",
	}

	km, err := LoadKeyMap(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check overridden keys
	if !contains(km.Quit.Keys(), "F10") {
		t.Errorf("expected Quit to have F10, got %v", km.Quit.Keys())
	}
	if !contains(km.Refresh.Keys(), "ctrl+r") {
		t.Errorf("expected Refresh to have ctrl+r, got %v", km.Refresh.Keys())
	}
	if !contains(km.HistoryBack.Keys(), "left") {
		t.Errorf("expected HistoryBack to have left, got %v", km.HistoryBack.Keys())
	}
}

func TestLoadKeyMapValidation(t *testing.T) {
	tests := []struct {
		name    string
		cfg     config.KeybindingsConfig
		wantErr string
	}{
		{
			name:    "modifier as key",
			cfg:     config.KeybindingsConfig{Quit: "ctrl"},
			wantErr: "is a modifier",
		},
		{
			name:    "empty key",
			cfg:     config.KeybindingsConfig{Refresh: ""},
			wantErr: "", // empty means "use default", not error
		},
		{
			name:    "double comma",
			cfg:     config.KeybindingsConfig{Quit: "q,,ctrl+c"},
			wantErr: "empty key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := LoadKeyMap(tt.cfg)
			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				return
			}
			if err == nil {
				t.Errorf("expected error containing %q, got nil", tt.wantErr)
				return
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("expected error containing %q, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestLoadKeyMapMultipleOverrides(t *testing.T) {
	cfg := config.KeybindingsConfig{
		Quit:           "F10",
		NextSection:    "tab",
		PrevSection:    "shift+tab",
		Refresh:        "ctrl+r",
		Up:             "up",
		Down:           "down",
		HistoryBack:    "alt+left",
		HistoryForward: "alt+right",
	}

	km, err := LoadKeyMap(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify all overridden
	if !contains(km.Quit.Keys(), "F10") {
		t.Error("Quit not overridden")
	}
	if !contains(km.NextSection.Keys(), "tab") {
		t.Error("NextSection not overridden")
	}
	if !contains(km.HistoryBack.Keys(), "alt+left") {
		t.Error("HistoryBack not overridden")
	}

	// Verify non-overridden still have defaults
	if len(km.Yank.Keys()) == 0 {
		t.Error("Yank should still have defaults")
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
