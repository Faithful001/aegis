package crypto

import (
	"testing"
)

func TestAESEncryptionDecryption(t *testing.T) {
	secretKey := []byte("super-secret-aegis-byok-key-32b!")
	plainText := "sk-proj-test1234567890abcdefghijklmnopqrstuvwxyz"

	encryptedHex, err := AESEncrypt(plainText, secretKey)
	if err != nil {
		t.Fatalf("AESEncrypt failed: %v", err)
	}

	if encryptedHex == plainText {
		t.Fatalf("Ciphertext should not equal plaintext")
	}

	decrypted, err := AESDecrypt(encryptedHex, secretKey)
	if err != nil {
		t.Fatalf("AESDecrypt failed: %v", err)
	}

	if decrypted != plainText {
		t.Errorf("Expected decrypted text '%s', got '%s'", plainText, decrypted)
	}
}

func TestAESEncryptionInvalidKey(t *testing.T) {
	_, err := AESEncrypt("test", []byte{})
	if err == nil {
		t.Errorf("Expected error for empty key, got nil")
	}
}

func TestMaskAPIKey(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"sk-1234567890abcdef", "sk-****cdef"},
		{"sk-proj-abcdefghijklmnopqrstuvwxyz", "sk-****wxyz"},
		{"short", "****"},
		{"", ""},
	}

	for _, tt := range tests {
		result := MaskAPIKey(tt.input)
		if result != tt.expected {
			t.Errorf("MaskAPIKey(%q) = %q; want %q", tt.input, result, tt.expected)
		}
	}
}
