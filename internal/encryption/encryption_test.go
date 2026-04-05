package encryption_test

import (
	"testing"

	"github.com/GrayCodeAI/tokman/internal/encryption"
)

func TestEncryptDecrypt(t *testing.T) {
	key := "test-key-32-bytes-long!!"
	encryptor, err := encryption.NewEncryptor(key)
	if err != nil {
		t.Fatalf("NewEncryptor() error = %v", err)
	}

	plaintext := "hello world"
	encrypted, err := encryptor.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	if encrypted == plaintext {
		t.Error("encrypted output should differ from plaintext")
	}

	decrypted, err := encryptor.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("Decrypt() error = %v", err)
	}

	if decrypted != plaintext {
		t.Errorf("Decrypt(Encrypt()) = %q, want %q", decrypted, plaintext)
	}
}
