package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
)

const (
	APIKeyPrefixLive = "aeg_live_"
	base62Chars      = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
)

type GeneratedAPIKey struct {
	PlaintextKey string
	KeyPrefix    string
	KeyHash      string
}

type APIKeyGenerator struct{}

func NewAPIKeyGenerator() *APIKeyGenerator {
	return &APIKeyGenerator{}
}

func (g *APIKeyGenerator) Generate() (*GeneratedAPIKey, error) {
	randomPart, err := generateRandomBase62(32)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random API key: %w", err)
	}

	plaintext := fmt.Sprintf("%s%s", APIKeyPrefixLive, randomPart)
	keyHash := HashAPIKey(plaintext)
	prefix := plaintext[:14]

	return &GeneratedAPIKey{
		PlaintextKey: plaintext,
		KeyPrefix:    prefix,
		KeyHash:      keyHash,
	}, nil
}

func HashAPIKey(plaintext string) string {
	hash := sha256.Sum256([]byte(plaintext))
	return hex.EncodeToString(hash[:])
}

func generateRandomBase62(length int) (string, error) {
	chars := make([]byte, length)
	charsLen := big.NewInt(int64(len(base62Chars)))

	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, charsLen)
		if err != nil {
			return "", err
		}
		chars[i] = base62Chars[num.Int64()]
	}

	return string(chars), nil
}
