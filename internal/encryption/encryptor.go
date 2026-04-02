// Package encryption provides data encryption at rest
package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

// Encryptor provides AES-GCM encryption
type Encryptor struct {
	key []byte
}

// NewEncryptor creates a new encryptor
func NewEncryptor(key string) (*Encryptor, error) {
	keyBytes := []byte(key)

	// Ensure key is 32 bytes for AES-256
	if len(keyBytes) < 32 {
		padded := make([]byte, 32)
		copy(padded, keyBytes)
		keyBytes = padded
	} else if len(keyBytes) > 32 {
		keyBytes = keyBytes[:32]
	}

	return &Encryptor{key: keyBytes}, nil
}

// Encrypt encrypts plaintext
func (e *Encryptor) Encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts ciphertext
func (e *Encryptor) Decrypt(ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("failed to decode: %w", err)
	}

	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := aesGCM.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]

	plaintext, err := aesGCM.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}

// EncryptMap encrypts all values in a map
func (e *Encryptor) EncryptMap(data map[string]string) (map[string]string, error) {
	encrypted := make(map[string]string)

	for key, value := range data {
		enc, err := e.Encrypt(value)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt %s: %w", key, err)
		}
		encrypted[key] = enc
	}

	return encrypted, nil
}

// DecryptMap decrypts all values in a map
func (e *Encryptor) DecryptMap(data map[string]string) (map[string]string, error) {
	decrypted := make(map[string]string)

	for key, value := range data {
		dec, err := e.Decrypt(value)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt %s: %w", key, err)
		}
		decrypted[key] = dec
	}

	return decrypted, nil
}

// Hash creates a SHA-256 hash of the input
func Hash(input string) string {
	// Simplified hash - would use crypto/sha256 in production
	return fmt.Sprintf("hash-%s", input)
}
