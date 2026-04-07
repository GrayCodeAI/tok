package archive

import (
	"bytes"
	"strings"
	"testing"
)

func TestHashCalculator_Calculate(t *testing.T) {
	hc := NewHashCalculator()

	tests := []struct {
		name     string
		content  []byte
		expected string // Known SHA-256 hash for "hello"
	}{
		{
			name:     "empty content",
			content:  []byte{},
			expected: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
		{
			name:     "hello world",
			content:  []byte("hello world"),
			expected: "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9",
		},
		{
			name:     "hello",
			content:  []byte("hello"),
			expected: "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := hc.Calculate(tt.content)
			if hash != tt.expected {
				t.Errorf("Calculate() = %v, want %v", hash, tt.expected)
			}

			// Verify hash length is always 64
			if len(hash) != 64 {
				t.Errorf("Hash length = %d, want 64", len(hash))
			}
		})
	}
}

func TestHashCalculator_CalculateString(t *testing.T) {
	hc := NewHashCalculator()

	content := "test content"
	hash1 := hc.CalculateString(content)
	hash2 := hc.Calculate([]byte(content))

	if hash1 != hash2 {
		t.Error("CalculateString should match Calculate")
	}
}

func TestHashCalculator_CalculateReader(t *testing.T) {
	hc := NewHashCalculator()

	content := "reader test content"
	reader := strings.NewReader(content)

	hash, err := hc.CalculateReader(reader)
	if err != nil {
		t.Fatalf("CalculateReader() error = %v", err)
	}

	expected := hc.CalculateString(content)
	if hash != expected {
		t.Errorf("CalculateReader() = %v, want %v", hash, expected)
	}
}

func TestHashCalculator_Verify(t *testing.T) {
	hc := NewHashCalculator()

	content := []byte("verify me")
	hash := hc.Calculate(content)

	// Test valid verification
	if !hc.Verify(content, hash) {
		t.Error("Verify() should return true for valid hash")
	}

	// Test invalid verification
	if hc.Verify(content, "invalidhash") {
		t.Error("Verify() should return false for invalid hash")
	}

	// Test verification with different content
	if hc.Verify([]byte("different content"), hash) {
		t.Error("Verify() should return false for different content")
	}
}

func TestHashCalculator_CalculatePartial(t *testing.T) {
	hc := NewHashCalculator()

	content := []byte("this is a longer content string")
	fullHash := hc.Calculate(content)
	partialHash := hc.CalculatePartial(content, 10)

	// Partial hash should be different from full hash
	if partialHash == fullHash {
		t.Error("Partial hash should differ from full hash")
	}

	// Partial hash of content <= n should equal full hash
	smallContent := []byte("short")
	partialOfSmall := hc.CalculatePartial(smallContent, 100)
	fullOfSmall := hc.Calculate(smallContent)
	if partialOfSmall != fullOfSmall {
		t.Error("Partial hash should equal full hash when content <= n")
	}
}

func TestIsValidHash(t *testing.T) {
	tests := []struct {
		name     string
		hash     string
		expected bool
	}{
		{
			name:     "valid hash lowercase",
			hash:     "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			expected: true,
		},
		{
			name:     "valid hash uppercase",
			hash:     "E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855",
			expected: true,
		},
		{
			name:     "valid hash mixed case",
			hash:     "E3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			expected: true,
		},
		{
			name:     "too short",
			hash:     "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b85",
			expected: false,
		},
		{
			name:     "too long",
			hash:     "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b8555",
			expected: false,
		},
		{
			name:     "invalid characters",
			hash:     "g3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			expected: false,
		},
		{
			name:     "empty string",
			hash:     "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidHash(tt.hash); got != tt.expected {
				t.Errorf("IsValidHash(%q) = %v, want %v", tt.hash, got, tt.expected)
			}
		})
	}
}

func TestHashPrefix(t *testing.T) {
	hash := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"

	tests := []struct {
		n        int
		expected string
	}{
		{2, "e3"},
		{4, "e3b0"},
		{10, "e3b0c44298"},
		{64, hash},
		{100, hash},
	}

	for _, tt := range tests {
		got := HashPrefix(hash, tt.n)
		if got != tt.expected {
			t.Errorf("HashPrefix(hash, %d) = %q, want %q", tt.n, got, tt.expected)
		}
	}
}

func TestHashSuffix(t *testing.T) {
	hash := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"

	tests := []struct {
		n        int
		expected string
	}{
		{2, "55"},
		{4, "b855"},
		{10, "1b7852b855"},
		{64, hash},
		{100, hash},
	}

	for _, tt := range tests {
		got := HashSuffix(hash, tt.n)
		if got != tt.expected {
			t.Errorf("HashSuffix(hash, %d) = %q, want %q", tt.n, got, tt.expected)
		}
	}
}

func TestHashPath(t *testing.T) {
	tests := []struct {
		hash       string
		wantPrefix string
		wantSuffix string
	}{
		{
			hash:       "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			wantPrefix: "e3",
			wantSuffix: "b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
		{
			hash:       "ab",
			wantPrefix: "ab",
			wantSuffix: "",
		},
		{
			hash:       "a",
			wantPrefix: "a",
			wantSuffix: "",
		},
		{
			hash:       "",
			wantPrefix: "",
			wantSuffix: "",
		},
	}

	for _, tt := range tests {
		gotPrefix, gotSuffix := HashPath(tt.hash)
		if gotPrefix != tt.wantPrefix || gotSuffix != tt.wantSuffix {
			t.Errorf("HashPath(%q) = (%q, %q), want (%q, %q)",
				tt.hash, gotPrefix, gotSuffix, tt.wantPrefix, tt.wantSuffix)
		}
	}
}

func TestConcurrency(t *testing.T) {
	hc := NewHashCalculator()

	// Test concurrent hashing
	done := make(chan bool, 100)
	for i := 0; i < 100; i++ {
		go func(idx int) {
			content := []byte(string(rune('a' + idx%26)))
			hash1 := hc.Calculate(content)
			hash2 := hc.Calculate(content)
			if hash1 != hash2 {
				t.Errorf("Concurrent hashes should match: %s vs %s", hash1, hash2)
			}
			done <- true
		}(i)
	}

	for i := 0; i < 100; i++ {
		<-done
	}
}

func BenchmarkHashCalculator_Calculate(b *testing.B) {
	hc := NewHashCalculator()
	content := bytes.Repeat([]byte("benchmark content "), 1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hc.Calculate(content)
	}
}

func BenchmarkHashCalculator_CalculateParallel(b *testing.B) {
	hc := NewHashCalculator()
	content := bytes.Repeat([]byte("benchmark content "), 1000)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			hc.Calculate(content)
		}
	})
}
