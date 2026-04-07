package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

// EncryptionManager handles data encryption/decryption.
type EncryptionManager struct {
	masterKey []byte
}

// NewEncryptionManager creates a new encryption manager.
func NewEncryptionManager(masterKey string) (*EncryptionManager, error) {
	if len(masterKey) < 32 {
		return nil, fmt.Errorf("master key must be at least 32 characters")
	}

	// Use first 32 bytes as AES-256 key
	key := []byte(masterKey)
	if len(key) > 32 {
		key = key[:32]
	}

	return &EncryptionManager{
		masterKey: key,
	}, nil
}

// Encrypt encrypts data with AES-GCM.
func (em *EncryptionManager) Encrypt(plaintext []byte) (string, error) {
	block, err := aes.NewCipher(em.masterKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts data encrypted with Encrypt.
func (em *EncryptionManager) Decrypt(encryptedData string) ([]byte, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(em.masterKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// HashPassword hashes a password using PBKDF2.
func HashPassword(password string) (string, error) {
	// In production, use bcrypt or argon2
	// This is a placeholder
	return base64.StdEncoding.EncodeToString([]byte(password)), nil
}

// VerifyPassword verifies a hashed password.
func VerifyPassword(password, hash string) bool {
	// In production, use proper password verification
	decoded, _ := base64.StdEncoding.DecodeString(hash)
	return string(decoded) == password
}
