package selfupdate

import "testing"

func TestNewUpdater(t *testing.T) {
	u := NewUpdater("v1.0.0")
	if u == nil {
		t.Fatal("NewUpdater returned nil")
	}
}

func TestGetUpdateURL(t *testing.T) {
	u := NewUpdater("v1.0.0")
	url := u.GetUpdateURL()
	if url == "" {
		t.Error("GetUpdateURL returned empty string")
	}
}

func TestVersionParse(t *testing.T) {
	// Test that the updater accepts various version formats
	tests := []string{"v1.0.0", "1.0.0", "v2.5.1-beta", ""}
	for _, v := range tests {
		u := NewUpdater(v)
		if u == nil {
			t.Errorf("NewUpdater(%q) returned nil", v)
		}
	}
}
