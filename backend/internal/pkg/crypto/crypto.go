package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"
)

var (
	ErrInvalidCiphertext = errors.New("invalid ciphertext or corrupted payload")
	ErrEmptyKey          = errors.New("encryption key cannot be empty")
)

// derive32ByteKey generates a 32-byte key using SHA-256 if the provided key is not exactly 32 bytes.
func derive32ByteKey(secretKey []byte) []byte {
	if len(secretKey) == 32 {
		return secretKey
	}
	hash := sha256.Sum256(secretKey)
	return hash[:]
}

// AESEncrypt encrypts plaintext string into a hex-encoded AES-256-GCM string using secretKey.
func AESEncrypt(plainText string, secretKey []byte) (string, error) {
	if len(secretKey) == 0 {
		return "", ErrEmptyKey
	}

	key := derive32ByteKey(secretKey)
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM cipher: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	cipherText := gcm.Seal(nonce, nonce, []byte(plainText), nil)
	return hex.EncodeToString(cipherText), nil
}

// AESDecrypt decrypts hex-encoded AES-256-GCM ciphertext using secretKey.
func AESDecrypt(cipherTextHex string, secretKey []byte) (string, error) {
	if len(secretKey) == 0 {
		return "", ErrEmptyKey
	}

	cipherText, err := hex.DecodeString(cipherTextHex)
	if err != nil {
		return "", fmt.Errorf("%w: invalid hex encoding", ErrInvalidCiphertext)
	}

	key := derive32ByteKey(secretKey)
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM cipher: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(cipherText) < nonceSize {
		return "", ErrInvalidCiphertext
	}

	nonce, encryptedBytes := cipherText[:nonceSize], cipherText[nonceSize:]
	plainTextBytes, err := gcm.Open(nil, nonce, encryptedBytes, nil)
	if err != nil {
		return "", fmt.Errorf("%w: decryption failed", ErrInvalidCiphertext)
	}

	return string(plainTextBytes), nil
}

// MaskAPIKey returns a safe display string for an API key (e.g., "sk-proj-1234567890abcdef" -> "sk-proj-****cdef").
func MaskAPIKey(apiKey string) string {
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return ""
	}
	if len(apiKey) <= 8 {
		return "****"
	}
	prefix := apiKey
	if idx := strings.Index(apiKey, "-"); idx != -1 && idx < len(apiKey)-4 {
		prefix = apiKey[:idx+1]
	} else if len(apiKey) > 12 {
		prefix = apiKey[:7]
	} else {
		prefix = apiKey[:3]
	}
	suffix := apiKey[len(apiKey)-4:]
	return prefix + "****" + suffix
}
