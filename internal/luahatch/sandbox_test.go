package luahatch

import "testing"

func TestNewLuaSandbox(t *testing.T) {
	s := NewLuaSandbox()
	if s == nil {
		t.Fatal("Expected non-nil sandbox")
	}
}

func TestLuaSandboxExecute(t *testing.T) {
	s := NewLuaSandbox()

	result := s.Execute("return string.upper(input)", "hello")
	if result.Error != nil {
		t.Fatalf("Execute error: %v", result.Error)
	}
	if result.Output != "HELLO" {
		t.Errorf("Expected 'HELLO', got %s", result.Output)
	}
}

func TestLuaSandboxUnsafe(t *testing.T) {
	s := NewLuaSandbox()

	result := s.Execute("os.execute('rm -rf /')", "")
	if result.Error == nil {
		t.Error("Expected error for unsafe operation")
	}
}

func TestGetTemplate(t *testing.T) {
	tmpl := GetTemplate("uppercase")
	if tmpl == nil {
		t.Error("Expected uppercase template")
	}

	tmpl = GetTemplate("nonexistent")
	if tmpl != nil {
		t.Error("Expected nil for nonexistent template")
	}
}
