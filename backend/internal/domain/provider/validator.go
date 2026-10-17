package provider

import (
	"net/url"
	"strings"
)

// ValidationError represents a client-side validation issue with credential inputs.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// ValidateCredentialInput validates the provider, apiKey, and optional baseURL.
func ValidateCredentialInput(provider string, apiKey string, baseURL string) error {
	p := strings.ToLower(strings.TrimSpace(provider))
	if !IsSupportedProvider(p) {
		return ErrInvalidProvider
	}

	key := strings.TrimSpace(apiKey)
	if key == "" {
		return &ValidationError{Field: "api_key", Message: "API key is required and cannot be empty"}
	}

	if strings.ContainsAny(key, " \t\r\n") {
		return &ValidationError{Field: "api_key", Message: "API key cannot contain spaces or whitespace"}
	}

	if len(key) < 10 {
		return &ValidationError{Field: "api_key", Message: "API key is too short (minimum 10 characters)"}
	}

	if len(key) > 512 {
		return &ValidationError{Field: "api_key", Message: "API key is too long (maximum 512 characters)"}
	}

	// Validate BaseURL if provided
	trimmedBaseURL := strings.TrimSpace(baseURL)
	if trimmedBaseURL != "" {
		if len(trimmedBaseURL) > 2048 {
			return &ValidationError{Field: "base_url", Message: "Base URL is too long (maximum 2048 characters)"}
		}

		if strings.ContainsAny(trimmedBaseURL, " \t\r\n") {
			return &ValidationError{Field: "base_url", Message: "Base URL cannot contain spaces or whitespace"}
		}

		parsedURL, err := url.ParseRequestURI(trimmedBaseURL)
		if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") || parsedURL.Host == "" {
			return &ValidationError{
				Field:   "base_url",
				Message: "Base URL must be a valid HTTP or HTTPS URL (e.g. https://api.openai.com/v1)",
			}
		}
	}

	// Provider-specific format checks (only strictly enforced when default official endpoint is used)
	if trimmedBaseURL == "" {
		switch p {
		case ProviderOpenAI:
			if !strings.HasPrefix(key, "sk-") {
				return &ValidationError{
					Field:   "api_key",
					Message: "Invalid OpenAI API key: must start with 'sk-' (e.g. sk-proj-...)",
				}
			}
		case ProviderAnthropic:
			if !strings.HasPrefix(key, "sk-ant-") {
				return &ValidationError{
					Field:   "api_key",
					Message: "Invalid Anthropic API key: must start with 'sk-ant-'",
				}
			}
		case ProviderGemini:
			if !strings.HasPrefix(key, "AIzaSy") {
				return &ValidationError{
					Field:   "api_key",
					Message: "Invalid Google Gemini API key: must start with 'AIzaSy'",
				}
			}
		case ProviderOpenRouter:
			if !strings.HasPrefix(key, "sk-or-") {
				return &ValidationError{
					Field:   "api_key",
					Message: "Invalid OpenRouter API key: must start with 'sk-or-'",
				}
			}
		case ProviderMistral:
			if len(key) < 20 {
				return &ValidationError{
					Field:   "api_key",
					Message: "Invalid Mistral API key: must be at least 20 characters",
				}
			}
		}
	}

	return nil
}
