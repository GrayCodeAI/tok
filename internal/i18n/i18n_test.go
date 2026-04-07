package i18n

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestSetLanguage(t *testing.T) {
	SetupTestLocales(t)
	
	// Test valid language
	result := SetLanguage("en")
	if result != "en" {
		t.Errorf("SetLanguage(en) = %s, want en", result)
	}
	
	result = SetLanguage("fr")
	if result != "fr" {
		t.Errorf("SetLanguage(fr) = %s, want fr", result)
	}
	
	// Test invalid language (should fallback to en)
	result = SetLanguage("xx")
	if result != "en" {
		t.Errorf("SetLanguage(xx) fallback = %s, want en", result)
	}
}

func TestT(t *testing.T) {
	SetupTestLocales(t)
	
	SetLanguage("en")
	
	// Test with no args
	msg := T("common.success")
	if msg == "" || msg == "common.success" {
		t.Errorf("T returned untranslated key: %s", msg)
	}
	
	// Test with substitution args
	msg = T("filter.tokens_in")
	if msg == "" {
		t.Error("expected non-empty translation")
	}
}

func TestFallsBackToEnglish(t *testing.T) {
	SetupTestLocales(t)
	
	SetLanguage("en")
	en := T("common.success")
	
	SetLanguage("nonexistent")
	fallback := T("common.success")
	
	if fallback != en {
		t.Errorf("fallback = %q, want %q", fallback, en)
	}
}

func TestAvailableLanguages(t *testing.T) {
	SetupTestLocales(t)
	
	langs := GetAvailableLanguages()
	if len(langs) == 0 {
		t.Error("expected at least one language")
	}
	
	// Check English is always available
	found := false
	for _, l := range langs {
		if l == "en" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected English to be available")
	}
}

func TestGetCurrentLanguage(t *testing.T) {
	SetupTestLocales(t)
	
	SetLanguage("fr")
	if GetCurrentLanguage() != "fr" {
		t.Errorf("GetCurrentLanguage() = %s, want fr", GetCurrentLanguage())
	}
}

func TestGetLanguageName(t *testing.T) {
	tests := []struct {
		code string
		want string
	}{
		{"en", "English"},
		{"fr", "Français"},
		{"zh", "中文"},
		{"ja", "日本語"},
		{"es", "Español"},
		{"de", "Deutsch"},
		{"ko", "한국어"},
		{"xx", "xx"},
	}
	
	for _, tt := range tests {
		got := GetLanguageName(tt.code)
		if got != tt.want {
			t.Errorf("GetLanguageName(%s) = %s, want %s", tt.code, got, tt.want)
		}
	}
}

// Helper to setup test locales
func SetupTestLocales(t *testing.T) {
	// Create minimal test locale files in temp dir
	tmpDir, err := os.MkdirTemp("", "i18n-test-*")
	if err != nil {
		t.Fatalf("create temp: %v", err)
	}
	
	en := `
[common]
success = "completed successfully"
error = "Error: {error}"
warning = "Warning: {message}"

[filter]
tokens_in = "Input tokens"
tokens_out = "Output tokens"
`
	
	fr := `
[common]
success = "terminé avec succès"
error = "Erreur: {error}"

[filter]
tokens_in = "Tokens d'entrée"
`
	
	os.WriteFile(filepath.Join(tmpDir, "en.toml"), []byte(en), 0644)
	os.WriteFile(filepath.Join(tmpDir, "fr.toml"), []byte(fr), 0644)
	
	// Reset once for clean test
	once = sync.Once{}
	
	// Load the test locales
	translations = make(map[string]map[string]string)
	for _, file := range []string{"en", "fr"} {
		flat, _ := flattenTOML(filepath.Join(tmpDir, file+".toml"))
		translations[file] = flat
	}
	SetLanguage("en")
}
