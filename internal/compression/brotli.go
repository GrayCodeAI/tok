package compression

import (
	"bytes"
	"fmt"
	"io"
	"sync"

	"github.com/andybalholm/brotli"
)

// BrotliCompressor provides Brotli compression/decompression
// Brotli offers 2-4x better compression than gzip for text content
// and up to 82x for repetitive content (logs, structured data)
type BrotliCompressor struct {
	quality int
	lgwin   int
}

// BrotliConfig holds configuration for Brotli compression
type BrotliConfig struct {
	// Quality is the compression level (0-11)
	// 0 = no compression, 11 = best compression (slow)
	// Default: 4 (good balance)
	Quality int

	// LGWin is the LZ77 window size
	// 10-24, where 24 = 16MB window
	// Default: 22 (4MB window)
	LGWin int

	// MinSize is the minimum content size to compress
	// Content smaller than this won't be compressed
	// Default: 100 bytes
	MinSize int

	// MaxSize is the maximum content size for single-pass compression
	// Larger content will use streaming
	// Default: 100MB
	MaxSize int
}

// DefaultBrotliConfig returns default Brotli configuration
func DefaultBrotliConfig() BrotliConfig {
	return BrotliConfig{
		Quality: 4,
		LGWin:   22,
		MinSize: 100,
		MaxSize: 100 * 1024 * 1024, // 100MB
	}
}

// NewBrotliCompressor creates a new Brotli compressor with default settings
func NewBrotliCompressor() *BrotliCompressor {
	return NewBrotliCompressorWithConfig(DefaultBrotliConfig())
}

// NewBrotliCompressorWithConfig creates a compressor with custom config
func NewBrotliCompressorWithConfig(cfg BrotliConfig) *BrotliCompressor {
	// Validate quality
	quality := cfg.Quality
	if quality < 0 {
		quality = 0
	} else if quality > 11 {
		quality = 11
	}

	// Validate window size
	lgwin := cfg.LGWin
	if lgwin < 10 {
		lgwin = 10
	} else if lgwin > 24 {
		lgwin = 24
	}

	return &BrotliCompressor{
		quality: quality,
		lgwin:   lgwin,
	}
}

// Compress compresses data using Brotli
// Returns compressed data and nil error on success
func (bc *BrotliCompressor) Compress(data []byte) ([]byte, error) {
	// Check minimum size
	if len(data) < 100 {
		return data, nil // Too small to compress
	}

	var buf bytes.Buffer

	// Create Brotli writer
	params := brotli.WriterOptions{
		Quality: bc.quality,
		LGWin:   bc.lgwin,
	}

	writer := brotli.NewWriterOptions(&buf, params)

	// Write and close
	if _, err := writer.Write(data); err != nil {
		return nil, fmt.Errorf("brotli write failed: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("brotli close failed: %w", err)
	}

	// Check if compression actually helped
	compressed := buf.Bytes()
	if len(compressed) >= len(data) {
		// Compression didn't help, return original
		return data, nil
	}

	return compressed, nil
}

// CompressWithMetadata compresses and returns metadata
func (bc *BrotliCompressor) CompressWithMetadata(data []byte) (*CompressionResult, error) {
	compressed, err := bc.Compress(data)
	if err != nil {
		return nil, err
	}

	result := &CompressionResult{
		Algorithm:        "brotli",
		OriginalSize:     len(data),
		CompressedSize:   len(compressed),
		CompressionRatio: float64(len(compressed)) / float64(len(data)),
		SpaceSaved:       len(data) - len(compressed),
		WasCompressed:    len(compressed) < len(data),
		Data:             compressed,
	}

	return result, nil
}

// Decompress decompresses Brotli-compressed data
func (bc *BrotliCompressor) Decompress(data []byte) ([]byte, error) {
	// Check magic bytes for Brotli
	if !IsBrotliCompressed(data) {
		// Not compressed with Brotli, return as-is
		return data, nil
	}

	reader := brotli.NewReader(bytes.NewReader(data))

	result, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("brotli decompression failed: %w", err)
	}

	return result, nil
}

// CompressStream compresses data using streaming (for large content)
func (bc *BrotliCompressor) CompressStream(reader io.Reader, writer io.Writer) error {
	params := brotli.WriterOptions{
		Quality: bc.quality,
		LGWin:   bc.lgwin,
	}

	bw := brotli.NewWriterOptions(writer, params)
	defer bw.Close()

	_, err := io.Copy(bw, reader)
	if err != nil {
		return fmt.Errorf("brotli stream compression failed: %w", err)
	}

	return nil
}

// DecompressStream decompresses using streaming
func (bc *BrotliCompressor) DecompressStream(reader io.Reader, writer io.Writer) error {
	br := brotli.NewReader(reader)

	_, err := io.Copy(writer, br)
	if err != nil {
		return fmt.Errorf("brotli stream decompression failed: %w", err)
	}

	return nil
}

// IsBrotliCompressed checks if data appears to be Brotli compressed
// Brotli doesn't have a standard magic number, but we can check some heuristics
func IsBrotliCompressed(data []byte) bool {
	if len(data) < 10 {
		return false
	}

	// Try to create a reader and read a small amount
	// If it succeeds, it's likely valid Brotli
	reader := brotli.NewReader(bytes.NewReader(data))
	buf := make([]byte, 1)
	_, err := reader.Read(buf)

	// If we can read even 1 byte, it's probably valid
	return err == nil
}

// CompressionResult holds compression results
type CompressionResult struct {
	Algorithm        string  `json:"algorithm"`
	OriginalSize     int     `json:"original_size"`
	CompressedSize   int     `json:"compressed_size"`
	CompressionRatio float64 `json:"compression_ratio"`
	SpaceSaved       int     `json:"space_saved"`
	WasCompressed    bool    `json:"was_compressed"`
	Data             []byte  `json:"-"`
}

// Percentage returns space saved as percentage
func (cr *CompressionResult) Percentage() float64 {
	if cr.OriginalSize == 0 {
		return 0
	}
	return (1.0 - cr.CompressionRatio) * 100.0
}

// Summary returns a human-readable summary
func (cr *CompressionResult) Summary() string {
	if !cr.WasCompressed {
		return "Not compressed (too small or incompressible)"
	}
	return fmt.Sprintf("Compressed %.1f%% (%d → %d bytes)",
		cr.Percentage(), cr.OriginalSize, cr.CompressedSize)
}

// BrotliFilter implements the filter.Layer interface for Brotli compression
type BrotliFilter struct {
	compressor *BrotliCompressor
	config     BrotliConfig
}

// NewBrotliFilter creates a new Brotli filter
func NewBrotliFilter() *BrotliFilter {
	return &BrotliFilter{
		compressor: NewBrotliCompressor(),
		config:     DefaultBrotliConfig(),
	}
}

// NewBrotliFilterWithConfig creates a filter with custom config
func NewBrotliFilterWithConfig(cfg BrotliConfig) *BrotliFilter {
	return &BrotliFilter{
		compressor: NewBrotliCompressorWithConfig(cfg),
		config:     cfg,
	}
}

// Name returns the filter name
func (bf *BrotliFilter) Name() string {
	return "brotli"
}

// Apply compresses input using Brotli
// Note: This is used for storage compression, not pipeline filtering
func (bf *BrotliFilter) Apply(input string, mode int) (string, int) {
	result, err := bf.compressor.CompressWithMetadata([]byte(input))
	if err != nil {
		// Compression failed, return original
		return input, 0
	}

	if !result.WasCompressed {
		return input, 0
	}

	// Return base64-encoded compressed data with prefix
	// Note: In practice, you'd store binary data, not return as string
	return string(result.Data), result.SpaceSaved
}

// Pool for reusing buffers
var bufferPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

// GetBuffer gets a buffer from the pool
func GetBuffer() *bytes.Buffer {
	return bufferPool.Get().(*bytes.Buffer)
}

// PutBuffer returns a buffer to the pool
func PutBuffer(buf *bytes.Buffer) {
	buf.Reset()
	bufferPool.Put(buf)
}

// Quality names for display
var QualityNames = map[int]string{
	0:  "None (fastest)",
	1:  "Fast",
	2:  "Fast",
	3:  "Balanced",
	4:  "Balanced",
	5:  "Balanced",
	6:  "Good",
	7:  "Good",
	8:  "Best",
	9:  "Best",
	10: "Maximum",
	11: "Maximum (slowest)",
}

// GetQualityName returns human-readable quality name
func GetQualityName(quality int) string {
	if name, ok := QualityNames[quality]; ok {
		return name
	}
	return "Unknown"
}

// EstimateCompressedSize estimates the compressed size without actually compressing
// This is a rough estimate based on content type
func EstimateCompressedSize(data []byte, quality int) int {
	if len(data) < 100 {
		return len(data)
	}

	// Estimate based on quality level
	ratios := []float64{
		1.0,  // 0 - no compression
		0.9,  // 1
		0.85, // 2
		0.8,  // 3
		0.75, // 4
		0.7,  // 5
		0.65, // 6
		0.6,  // 7
		0.55, // 8
		0.5,  // 9
		0.45, // 10
		0.4,  // 11
	}

	if quality < 0 || quality > 11 {
		quality = 4
	}

	return int(float64(len(data)) * ratios[quality])
}
