package compression

import (
	"bytes"
	"strings"
	"testing"
)

func TestCompareAlgorithms(t *testing.T) {
	data := bytes.Repeat([]byte("The quick brown fox jumps over the lazy dog. "), 200)

	result, err := CompareAlgorithms(data)
	if err != nil {
		t.Fatalf("CompareAlgorithms failed: %v", err)
	}

	if len(result.Algorithms) == 0 {
		t.Fatal("expected at least one algorithm result")
	}

	if result.Winner == "" {
		t.Error("expected a winner to be selected")
	}

	if result.BestRatio <= 0 || result.BestRatio > 1 {
		t.Errorf("invalid best ratio: %f", result.BestRatio)
	}

	for _, alg := range result.Algorithms {
		if alg.Algorithm == "" {
			t.Error("algorithm name should not be empty")
		}
		if alg.CompressedSize <= 0 {
			t.Errorf("compressed size should be positive for %s", alg.Algorithm)
		}
		if alg.Percentage < 0 || alg.Percentage > 100 {
			t.Errorf("invalid percentage for %s: %f", alg.Algorithm, alg.Percentage)
		}
	}
}

func TestCompareAlgorithms_Empty(t *testing.T) {
	_, err := CompareAlgorithms([]byte{})
	if err == nil {
		t.Error("expected error for empty data")
	}
}

func TestCompareResult_PrintComparison(t *testing.T) {
	result := &CompareResult{
		OriginalData: []byte("test data"),
		Algorithms: []CompressionComparison{
			{
				Algorithm:        "brotli-4",
				OriginalSize:     1000,
				CompressedSize:   500,
				CompressionRatio: 0.5,
				SpaceSaved:       500,
				Percentage:       50.0,
				Speed:            10.0,
			},
		},
		Winner:    "brotli-4",
		BestRatio: 0.5,
	}

	output := result.PrintComparison()
	if output == "" {
		t.Error("expected non-empty output")
	}
	if !strings.Contains(output, "brotli-4") {
		t.Error("output should contain algorithm name")
	}
	if !strings.Contains(output, "Winner") {
		t.Error("output should contain winner")
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes int
		want  string
	}{
		{0, "0 B"},
		{500, "500 B"},
		{1024, "1.0 KB"},
		{1024 * 1024, "1.0 MB"},
		{1024 * 1024 * 1024, "1.0 GB"},
	}
	for _, tt := range tests {
		got := formatBytes(tt.bytes)
		if !strings.Contains(got, strings.Fields(tt.want)[1]) {
			t.Errorf("formatBytes(%d) = %q, expected to contain %q", tt.bytes, got, tt.want)
		}
	}
}

func TestBrotliFilter(t *testing.T) {
	bf := NewBrotliFilter()

	if bf.Name() != "brotli" {
		t.Errorf("expected name 'brotli', got %q", bf.Name())
	}

	// Apply on small input (should not compress)
	input := "small"
	output, saved := bf.Apply(input, 0)
	if output != input {
		t.Errorf("expected unchanged output, got %q", output)
	}
	if saved != 0 {
		t.Errorf("expected 0 saved, got %d", saved)
	}
}

func TestBrotliFilterWithConfig(t *testing.T) {
	cfg := BrotliConfig{Quality: 6, LGWin: 20}
	bf := NewBrotliFilterWithConfig(cfg)

	if bf.Name() != "brotli" {
		t.Errorf("expected name 'brotli', got %q", bf.Name())
	}
}

func TestGetBuffer_PutBuffer(t *testing.T) {
	buf := GetBuffer()
	if buf == nil {
		t.Fatal("GetBuffer returned nil")
	}
	buf.WriteString("test")
	PutBuffer(buf)

	// After PutBuffer, the buffer should be reset
	buf2 := GetBuffer()
	if buf2.Len() != 0 {
		t.Error("buffer should be reset after PutBuffer")
	}
}

func TestCompressionResult_Percentage_Zero(t *testing.T) {
	cr := &CompressionResult{OriginalSize: 0, CompressionRatio: 0}
	if cr.Percentage() != 0 {
		t.Errorf("expected 0%% for zero original size, got %f", cr.Percentage())
	}
}
