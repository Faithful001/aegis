package provider

import (
	"testing"
)

func TestProviderRouterGetProviderForModel(t *testing.T) {
	router := NewProviderRouter()

	tests := []struct {
		model            string
		expectedProvider string
	}{
		{"gpt-4o", "openai"},
		{"gpt-4o-mini", "openai"},
		{"gpt-3.5-turbo", "openai"},
		{"claude-3-5-sonnet-20241022", "anthropic"},
		{"claude-3-haiku-20240307", "anthropic"},
		{"gemini-1.5-pro", "gemini"},
		{"gemini-1.5-flash", "gemini"},
		{"mistral-large-latest", "mistral"},
		{"codestral-latest", "mistral"},
		{"llama-3-8b-instruct", ""},
		{"mistral-7b-v0.1-local", "mistral"}, // Note: starts with mistral- maps to mistral BYOK unless overridden
	}

	for _, tt := range tests {
		got := router.GetProviderForModel(tt.model)
		if got != tt.expectedProvider {
			t.Errorf("GetProviderForModel(%q) = %q; want %q", tt.model, got, tt.expectedProvider)
		}
	}
}
