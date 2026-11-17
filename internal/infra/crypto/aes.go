package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"os"
	"sync"
)

var (
	globalKey     string
	globalKeyOnce sync.Once
)

// Init sets the encryption key from config (call once at startup).
func Init(key string) {
	globalKeyOnce.Do(func() { globalKey = key })
}

func encryptKey() string {
	if globalKey != "" {
		return globalKey
	}
	key := os.Getenv("ENCRYPTION_KEY")
	if key == "" {
		panic("ENCRYPTION_KEY environment variable is not set")
	}
	return key
}

// Encrypt encrypts plaintext using AES-256-GCM with the server's ENCRYPTION_KEY.
// Returns base64-encoded ciphertext (nonce || ciphertext || tag).
func Encrypt(plaintext string) (string, error) {
	return encryptWithKey(plaintext, encryptKey())
}

// Decrypt decrypts a base64-encoded ciphertext produced by Encrypt.
func Decrypt(ciphertextB64 string) (string, error) {
	return decryptWithKey(ciphertextB64, encryptKey())
}

// encryptWithKey encrypts using AES-256-GCM with an explicit key (useful for testing).
func encryptWithKey(plaintext, keyStr string) (string, error) {
	key, err := deriveKey(keyStr)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
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
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decryptWithKey decrypts using AES-256-GCM with an explicit key.
func decryptWithKey(ciphertextB64, keyStr string) (string, error) {
	key, err := deriveKey(keyStr)
	if err != nil {
		return "", err
	}
	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextB64)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("ciphertext too short")
	}
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

// deriveKey produces a 32-byte AES-256 key from the given string.
// Short strings are zero-padded; long strings are truncated.
// In production, prefer PBKDF2/bcrypt to derive from a passphrase.
func deriveKey(keyStr string) ([]byte, error) {
	if len(keyStr) < 16 {
		return nil, errors.New("encryption key must be at least 16 characters")
	}
	key := []byte(keyStr)
	if len(key) > 32 {
		key = key[:32]
	}
	if len(key) < 32 {
		padded := make([]byte, 32)
		copy(padded, key)
		key = padded
	}
	return key, nil
}
