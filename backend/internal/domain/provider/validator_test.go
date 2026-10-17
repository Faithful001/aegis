package provider

import (
	"testing"
)

func TestValidateCredentialInput(t *testing.T) {
	tests := []struct {
		name      string
		provider  string
		apiKey    string
		baseURL   string
		expectErr bool
	}{
		{
			name:      "Valid OpenAI key",
			provider:  "openai",
			apiKey:    "sk-proj-1234567890abcdef",
			baseURL:   "",
			expectErr: false,
		},
		{
			name:      "Invalid OpenAI key without sk- prefix",
			provider:  "openai",
			apiKey:    "bad-key-1234567890abcdef",
			baseURL:   "",
			expectErr: true,
		},
		{
			name:      "Custom baseURL relaxes prefix check",
			provider:  "openai",
			apiKey:    "custom-proxy-key-123456",
			baseURL:   "https://my-proxy.company.com/v1",
			expectErr: false,
		},
		{
			name:      "Valid Anthropic key",
			provider:  "anthropic",
			apiKey:    "sk-ant-1234567890abcdef",
			baseURL:   "",
			expectErr: false,
		},
		{
			name:      "Invalid Anthropic key",
			provider:  "anthropic",
			apiKey:    "sk-1234567890abcdef",
			baseURL:   "",
			expectErr: true,
		},
		{
			name:      "Valid Gemini key",
			provider:  "gemini",
			apiKey:    "AIzaSyD1234567890abcdefghijklmnop",
			baseURL:   "",
			expectErr: false,
		},
		{
			name:      "Invalid Gemini key",
			provider:  "gemini",
			apiKey:    "wrong-prefix-key-12345",
			baseURL:   "",
			expectErr: true,
		},
		{
			name:      "Valid OpenRouter key",
			provider:  "openrouter",
			apiKey:    "sk-or-v1-abcdef1234567890",
			baseURL:   "",
			expectErr: false,
		},
		{
			name:      "Valid Mistral key",
			provider:  "mistral",
			apiKey:    "abcdefghijklmnopqrst12345",
			baseURL:   "",
			expectErr: false,
		},
		{
			name:      "Mistral key too short",
			provider:  "mistral",
			apiKey:    "short-key",
			baseURL:   "",
			expectErr: true,
		},
		{
			name:      "Unsupported provider",
			provider:  "unknown-ai",
			apiKey:    "sk-1234567890abcdef",
			baseURL:   "",
			expectErr: true,
		},
		{
			name:      "Key with space",
			provider:  "openai",
			apiKey:    "sk-proj- 1234567890abcdef",
			baseURL:   "",
			expectErr: true,
		},
		{
			name:      "Invalid Base URL protocol",
			provider:  "openai",
			apiKey:    "sk-proj-1234567890abcdef",
			baseURL:   "ftp://invalid-url.com",
			expectErr: true,
		},
		{
			name:      "Valid Base URL HTTPS",
			provider:  "openai",
			apiKey:    "sk-proj-1234567890abcdef",
			baseURL:   "https://api.openai.com/v1",
			expectErr: false,
		},
		{
			name:      "Valid Base URL HTTP (local proxy)",
			provider:  "openai",
			apiKey:    "sk-proj-1234567890abcdef",
			baseURL:   "http://localhost:8000/v1",
			expectErr: false,
		},
		{
			name:      "Malformed Base URL",
			provider:  "openai",
			apiKey:    "sk-proj-1234567890abcdef",
			baseURL:   "not a url at all",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCredentialInput(tt.provider, tt.apiKey, tt.baseURL)
			if (err != nil) != tt.expectErr {
				t.Errorf("ValidateCredentialInput(%q, %q, %q) error = %v, expectErr %v",
					tt.provider, tt.apiKey, tt.baseURL, err, tt.expectErr)
			}
		})
	}
}
