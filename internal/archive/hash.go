package archive

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"sync"
)

// HashCalculator provides SHA-256 hashing functionality for content identification.
// Uses a pool of hash.Hash objects for efficient concurrent usage.
type HashCalculator struct {
	pool sync.Pool
}

// NewHashCalculator creates a new SHA-256 hash calculator with object pooling.
func NewHashCalculator() *HashCalculator {
	return &HashCalculator{
		pool: sync.Pool{
			New: func() interface{} {
				return sha256.New()
			},
		},
	}
}

// Calculate computes the SHA-256 hash of the provided content.
// Returns the hex-encoded hash string (64 characters).
func (hc *HashCalculator) Calculate(content []byte) string {
	h := hc.pool.Get().(hash.Hash)
	defer hc.pool.Put(h)
	h.Reset()

	h.Write(content)
	sum := h.Sum(nil)
	return hex.EncodeToString(sum)
}

// CalculateString computes the SHA-256 hash of a string.
// Convenience wrapper for Calculate.
func (hc *HashCalculator) CalculateString(content string) string {
	return hc.Calculate([]byte(content))
}

// CalculateReader computes the SHA-256 hash from an io.Reader.
// Useful for hashing large content without loading into memory.
func (hc *HashCalculator) CalculateReader(r io.Reader) (string, error) {
	h := hc.pool.Get().(hash.Hash)
	defer hc.pool.Put(h)
	h.Reset()

	if _, err := io.Copy(h, r); err != nil {
		return "", fmt.Errorf("failed to hash content: %w", err)
	}

	sum := h.Sum(nil)
	return hex.EncodeToString(sum), nil
}

// CalculatePartial computes the hash of the first n bytes.
// Useful for quick comparison without full content hashing.
func (hc *HashCalculator) CalculatePartial(content []byte, n int) string {
	if len(content) <= n {
		return hc.Calculate(content)
	}
	return hc.Calculate(content[:n])
}

// Verify checks if the provided hash matches the content.
// Returns true if hash is valid, false otherwise.
func (hc *HashCalculator) Verify(content []byte, expectedHash string) bool {
	actualHash := hc.Calculate(content)
	return actualHash == expectedHash
}

// VerifyString checks if the hash matches a string.
func (hc *HashCalculator) VerifyString(content string, expectedHash string) bool {
	return hc.Verify([]byte(content), expectedHash)
}

// IsValidHash checks if a string is a valid SHA-256 hex hash.
// SHA-256 hex hashes are exactly 64 hexadecimal characters.
func IsValidHash(hash string) bool {
	if len(hash) != 64 {
		return false
	}

	// Check if all characters are valid hex
	for _, c := range hash {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}

	return true
}

// HashPrefix returns the first n characters of a hash.
// Useful for partitioning or prefix-based lookup.
func HashPrefix(hash string, n int) string {
	if n >= len(hash) {
		return hash
	}
	return hash[:n]
}

// HashSuffix returns the last n characters of a hash.
func HashSuffix(hash string, n int) string {
	if n >= len(hash) {
		return hash
	}
	return hash[len(hash)-n:]
}

// HashPath generates a path-safe version of the hash.
// Splits hash into prefix/suffix for filesystem storage.
func HashPath(hash string) (prefix, suffix string) {
	if len(hash) < 4 {
		return hash, ""
	}
	return hash[:2], hash[2:]
}

// Global instance for convenience
var defaultCalculator = NewHashCalculator()

// Calculate is a convenience function using the default calculator.
func Calculate(content []byte) string {
	return defaultCalculator.Calculate(content)
}

// CalculateString is a convenience function using the default calculator.
func CalculateString(content string) string {
	return defaultCalculator.CalculateString(content)
}

// CalculateReader is a convenience function using the default calculator.
func CalculateReader(r io.Reader) (string, error) {
	return defaultCalculator.CalculateReader(r)
}

// Verify is a convenience function using the default calculator.
func Verify(content []byte, expectedHash string) bool {
	return defaultCalculator.Verify(content, expectedHash)
}

// VerifyString is a convenience function using the default calculator.
func VerifyString(content string, expectedHash string) bool {
	return defaultCalculator.VerifyString(content, expectedHash)
}
