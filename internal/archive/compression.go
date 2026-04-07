package archive

import (
	"github.com/GrayCodeAI/tokman/internal/compression"
)

// CompressionEngine handles compression for archives
type CompressionEngine struct {
	compressor *compression.BrotliCompressor
	enabled    bool
}

// NewCompressionEngine creates a new compression engine
func NewCompressionEngine() *CompressionEngine {
	return &CompressionEngine{
		compressor: compression.NewBrotliCompressor(),
		enabled:    true,
	}
}

// NewCompressionEngineWithConfig creates engine with custom config
func NewCompressionEngineWithConfig(cfg compression.BrotliConfig) *CompressionEngine {
	return &CompressionEngine{
		compressor: compression.NewBrotliCompressorWithConfig(cfg),
		enabled:    true,
	}
}

// Enable enables compression
func (ce *CompressionEngine) Enable() {
	ce.enabled = true
}

// Disable disables compression
func (ce *CompressionEngine) Disable() {
	ce.enabled = false
}

// IsEnabled returns whether compression is enabled
func (ce *CompressionEngine) IsEnabled() bool {
	return ce.enabled
}

// Compress compresses content for storage
func (ce *CompressionEngine) Compress(content []byte) ([]byte, *compression.CompressionResult, error) {
	if !ce.enabled {
		return content, &compression.CompressionResult{
			Algorithm:     "none",
			OriginalSize:  len(content),
			WasCompressed: false,
			Data:          content,
		}, nil
	}

	result, err := ce.compressor.CompressWithMetadata(content)
	if err != nil {
		return content, nil, err
	}

	return result.Data, result, nil
}

// Decompress decompresses content from storage
func (ce *CompressionEngine) Decompress(content []byte) ([]byte, error) {
	return ce.compressor.Decompress(content)
}

// GetCompressor returns the underlying compressor
func (ce *CompressionEngine) GetCompressor() *compression.BrotliCompressor {
	return ce.compressor
}

// SetQuality sets the compression quality (0-11)
func (ce *CompressionEngine) SetQuality(quality int) {
	cfg := compression.BrotliConfig{
		Quality: quality,
		LGWin:   22,
	}
	ce.compressor = compression.NewBrotliCompressorWithConfig(cfg)
}
