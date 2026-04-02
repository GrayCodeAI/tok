package toon

import "testing"

func TestTOONEncoder(t *testing.T) {
	e := NewTOONEncoder()

	json := `[
		{"name": "Alice", "age": 30},
		{"name": "Bob", "age": 25},
		{"name": "Charlie", "age": 35}
	]`

	encoded, err := e.EncodeJSON(json)
	if err != nil {
		t.Fatalf("EncodeJSON error: %v", err)
	}
	if encoded == "" {
		t.Error("Expected non-empty encoded output")
	}

	decoded, err := e.Decode(encoded)
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}
	if decoded == "" {
		t.Error("Expected non-empty decoded output")
	}
}

func TestIsHomogeneous(t *testing.T) {
	e := NewTOONEncoder()

	// Check with homogeneous slice
	if !e.IsHomogeneous([]int{1, 2, 3, 4, 5}) {
		t.Error("Expected homogeneous int slice to return true")
	}

	if e.IsHomogeneous([]int{1, 2}) {
		t.Error("Expected short slice to return false")
	}
}

func TestCompressionRatio(t *testing.T) {
	e := NewTOONEncoder()
	ratio := e.CompressionRatio("long original text", "short")
	if ratio <= 0 {
		t.Error("Expected positive compression ratio")
	}
}
