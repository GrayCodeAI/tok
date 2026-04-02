// Package reversible provides compression implementations.
package reversible

import (
	"bytes"
	"fmt"

	"github.com/klauspost/compress/zstd"
)

// ZstdCompressor implements zstd compression.
type ZstdCompressor struct {
	encoder *zstd.Encoder
	decoder *zstd.Decoder
}

// Compress implements Compressor.
func (z *ZstdCompressor) Compress(data []byte) ([]byte, error) {
	if z.encoder == nil {
		enc, err := zstd.NewWriter(nil, zstd.WithEncoderLevel(zstd.SpeedDefault))
		if err != nil {
			return nil, err
		}
		z.encoder = enc
	}

	return z.encoder.EncodeAll(data, nil), nil
}

// Decompress implements Compressor.
func (z *ZstdCompressor) Decompress(data []byte) ([]byte, error) {
	if z.decoder == nil {
		dec, err := zstd.NewReader(nil)
		if err != nil {
			return nil, err
		}
		z.decoder = dec
	}

	return z.decoder.DecodeAll(data, nil)
}

// Name returns the compressor name.
func (z *ZstdCompressor) Name() string {
	return "zstd"
}

// LZ4Compressor implements lz4 compression.
type LZ4Compressor struct {
	// Could use github.com/pierrec/lz4
}

// Compress implements Compressor.
func (l *LZ4Compressor) Compress(data []byte) ([]byte, error) {
	// Placeholder - would use actual LZ4 implementation
	// For now, return uncompressed with marker
	result := make([]byte, len(data)+4)
	copy(result, []byte("LZ4\x00"))
	copy(result[4:], data)
	return result, nil
}

// Decompress implements Compressor.
func (l *LZ4Compressor) Decompress(data []byte) ([]byte, error) {
	// Placeholder
	if len(data) < 4 || !bytes.HasPrefix(data, []byte("LZ4\x00")) {
		return nil, fmt.Errorf("invalid LZ4 data")
	}
	return data[4:], nil
}

// Name returns the compressor name.
func (l *LZ4Compressor) Name() string {
	return "lz4"
}

// NoOpCompressor implements no compression (pass-through).
type NoOpCompressor struct{}

// Compress implements Compressor.
func (n *NoOpCompressor) Compress(data []byte) ([]byte, error) {
	return data, nil
}

// Decompress implements Compressor.
func (n *NoOpCompressor) Decompress(data []byte) ([]byte, error) {
	return data, nil
}

// Name returns the compressor name.
func (n *NoOpCompressor) Name() string {
	return "none"
}
