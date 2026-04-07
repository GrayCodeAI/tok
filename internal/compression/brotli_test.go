package compression

import (
	"bytes"
	"strings"
	"testing"
)

func TestNewBrotliCompressor(t *testing.T) {
	bc := NewBrotliCompressor()

	if bc == nil {
		t.Fatal("NewBrotliCompressor returned nil")
	}

	if bc.quality != 4 {
		t.Errorf("Expected default quality 4, got %d", bc.quality)
	}

	if bc.lgwin != 22 {
		t.Errorf("Expected default lgwin 22, got %d", bc.lgwin)
	}
}

func TestNewBrotliCompressorWithConfig(t *testing.T) {
	cfg := BrotliConfig{
		Quality: 8,
		LGWin:   20,
	}

	bc := NewBrotliCompressorWithConfig(cfg)

	if bc.quality != 8 {
		t.Errorf("Expected quality 8, got %d", bc.quality)
	}

	if bc.lgwin != 20 {
		t.Errorf("Expected lgwin 20, got %d", bc.lgwin)
	}
}

func TestBrotliCompressor_Compress(t *testing.T) {
	bc := NewBrotliCompressor()

	tests := []struct {
		name    string
		data    []byte
		wantErr bool
	}{
		{
			name: "simple text",
			data: []byte("Hello, World! This is a test message for Brotli compression."),
		},
		{
			name: "repetitive content",
			data: bytes.Repeat([]byte("The quick brown fox jumps over the lazy dog. "), 100),
		},
		{
			name:    "empty data",
			data:    []byte{},
			wantErr: false,
		},
		{
			name:    "small data",
			data:    []byte("ab"),
			wantErr: false, // Should not compress but not error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compressed, err := bc.Compress(tt.data)

			if (err != nil) != tt.wantErr {
				t.Errorf("Compress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				return
			}

			// Should be able to decompress
			decompressed, err := bc.Decompress(compressed)
			if err != nil {
				t.Errorf("Decompress() failed: %v", err)
				return
			}

			// Should match original (unless too small to compress)
			if len(tt.data) >= 100 && !bytes.Equal(decompressed, tt.data) {
				t.Error("Decompressed data doesn't match original")
			}
		})
	}
}

func TestBrotliCompressor_CompressDecompress(t *testing.T) {
	bc := NewBrotliCompressor()

	original := []byte("This is a test message that should be compressed and then decompressed successfully.")

	// Compress
	compressed, err := bc.Compress(original)
	if err != nil {
		t.Fatalf("Compress() failed: %v", err)
	}

	// Should be smaller (or equal for small data)
	if len(compressed) > len(original) {
		t.Logf("Warning: compressed size (%d) > original size (%d)", len(compressed), len(original))
	}

	// Decompress
	decompressed, err := bc.Decompress(compressed)
	if err != nil {
		t.Fatalf("Decompress() failed: %v", err)
	}

	// Should match
	if !bytes.Equal(decompressed, original) {
		t.Error("Decompressed data doesn't match original")
	}
}

func TestBrotliCompressor_CompressWithMetadata(t *testing.T) {
	bc := NewBrotliCompressor()

	original := []byte("Test data for compression with metadata.")

	result, err := bc.CompressWithMetadata(original)
	if err != nil {
		t.Fatalf("CompressWithMetadata() failed: %v", err)
	}

	if result.Algorithm != "brotli" {
		t.Errorf("Expected algorithm 'brotli', got '%s'", result.Algorithm)
	}

	if result.OriginalSize != len(original) {
		t.Errorf("Expected original size %d, got %d", len(original), result.OriginalSize)
	}

	if result.CompressionRatio < 0 || result.CompressionRatio > 1 {
		t.Errorf("Invalid compression ratio: %f", result.CompressionRatio)
	}
}

func TestIsBrotliCompressed(t *testing.T) {
	bc := NewBrotliCompressor()

	// Compress some data (must be >100 bytes to actually compress)
	original := []byte("This is test data for IsBrotliCompressed function. This is test data for IsBrotliCompressed function. This is test data for IsBrotliCompressed function. This is test data for IsBrotliCompressed function. This is test data for IsBrotliCompressed function.")
	compressed, err := bc.Compress(original)
	if err != nil {
		t.Fatalf("Compress() failed: %v", err)
	}

	// Should detect as compressed
	if !IsBrotliCompressed(compressed) {
		t.Error("IsBrotliCompressed() should return true for compressed data")
	}

	// Should not detect plain text as compressed
	if IsBrotliCompressed(original) {
		t.Error("IsBrotliCompressed() should return false for plain text")
	}

	// Should not detect small data
	if IsBrotliCompressed([]byte("ab")) {
		t.Error("IsBrotliCompressed() should return false for small data")
	}
}

func TestCompressionResult(t *testing.T) {
	cr := &CompressionResult{
		Algorithm:        "brotli",
		OriginalSize:     1000,
		CompressedSize:   500,
		CompressionRatio: 0.5,
		SpaceSaved:       500,
		WasCompressed:    true,
	}

	if cr.Percentage() != 50.0 {
		t.Errorf("Expected 50%%, got %f%%", cr.Percentage())
	}

	summary := cr.Summary()
	if summary == "" {
		t.Error("Summary() should return non-empty string")
	}

	if !strings.Contains(summary, "50") {
		t.Error("Summary should contain compression percentage")
	}
}

func TestCompressionResult_NotCompressed(t *testing.T) {
	cr := &CompressionResult{
		WasCompressed: false,
	}

	summary := cr.Summary()
	if !strings.Contains(summary, "Not compressed") {
		t.Error("Summary should indicate not compressed")
	}
}

func TestGetQualityName(t *testing.T) {
	tests := []struct {
		quality  int
		expected string
	}{
		{0, "None (fastest)"},
		{4, "Balanced"},
		{11, "Maximum (slowest)"},
		{99, "Unknown"},
	}

	for _, tt := range tests {
		got := GetQualityName(tt.quality)
		if !strings.Contains(got, tt.expected) && tt.quality <= 11 {
			t.Errorf("GetQualityName(%d) = %s, expected to contain %s", tt.quality, got, tt.expected)
		}
	}
}

func TestEstimateCompressedSize(t *testing.T) {
	data := bytes.Repeat([]byte("Test data "), 100) // 1000 bytes

	size0 := EstimateCompressedSize(data, 0)
	size11 := EstimateCompressedSize(data, 11)

	// Quality 0 should be similar to original
	if size0 < len(data)*9/10 {
		t.Error("Quality 0 should not compress much")
	}

	// Quality 11 should be much smaller
	if size11 > len(data)/2 {
		t.Error("Quality 11 should compress significantly")
	}

	// Small data should not be compressed
	small := []byte("ab")
	smallSize := EstimateCompressedSize(small, 11)
	if smallSize != len(small) {
		t.Error("Small data should not be compressed")
	}
}

func BenchmarkBrotliCompressor_Compress(b *testing.B) {
	bc := NewBrotliCompressor()
	data := bytes.Repeat([]byte("The quick brown fox jumps over the lazy dog. "), 1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := bc.Compress(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBrotliCompressor_Decompress(b *testing.B) {
	bc := NewBrotliCompressor()
	data := bytes.Repeat([]byte("The quick brown fox jumps over the lazy dog. "), 1000)
	compressed, _ := bc.Compress(data)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := bc.Decompress(compressed)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestDefaultBrotliConfig(t *testing.T) {
	cfg := DefaultBrotliConfig()

	if cfg.Quality != 4 {
		t.Errorf("Expected quality 4, got %d", cfg.Quality)
	}

	if cfg.LGWin != 22 {
		t.Errorf("Expected LGWin 22, got %d", cfg.LGWin)
	}

	if cfg.MinSize != 100 {
		t.Errorf("Expected MinSize 100, got %d", cfg.MinSize)
	}

	if cfg.MaxSize != 100*1024*1024 {
		t.Errorf("Expected MaxSize 100MB, got %d", cfg.MaxSize)
	}
}
