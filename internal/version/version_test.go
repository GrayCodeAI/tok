package version

import (
	"testing"
)

func TestVersionNotEmpty(t *testing.T) {
	// Version is set to "dev" by default; build-time ldflags may override it.
	if Version == "" {
		t.Error("Version should not be empty")
	}
}
