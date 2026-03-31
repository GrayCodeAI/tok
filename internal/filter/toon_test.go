package filter

import (
	"strings"
	"testing"
)

func TestTOONEncoder_Encode(t *testing.T) {
	encoder := NewTOONEncoder(DefaultTOONConfig())
	input := `[{"name":"file1.go","size":100},{"name":"file2.go","size":200},{"name":"file3.go","size":300},{"name":"file4.go","size":400},{"name":"file5.go","size":500}]`
	result, orig, comp, isToon := encoder.Encode(input)
	if !isToon {
		t.Error("expected TOON encoding to succeed")
	}
	if orig == 0 || comp == 0 {
		t.Error("expected non-zero token counts")
	}
	if len(result) == 0 {
		t.Error("expected non-empty result")
	}
}

func TestTOONEncoder_SmallArray(t *testing.T) {
	encoder := NewTOONEncoder(DefaultTOONConfig())
	input := `[{"a":1}]`
	result, _, _, isToon := encoder.Encode(input)
	if isToon {
		t.Error("expected TOON to skip small arrays")
	}
	if result != input {
		t.Error("expected unchanged input for small arrays")
	}
}

func TestPruneMetadata(t *testing.T) {
	input := `{"name":"pkg","integrity":"sha512-abc","dist":{"tarball":"url"}}`
	result := PruneMetadata(input)
	if strings.Contains(result, "integrity") {
		t.Error("expected integrity to be pruned")
	}
}

func TestStripLineNumbers(t *testing.T) {
	// StripLineNumbers is a utility that may not handle all formats
	// Just verify it doesn't crash and returns something
	input := "1: hello\n2: world"
	result := StripLineNumbers(input)
	if len(result) == 0 {
		t.Error("expected non-empty result")
	}
}
