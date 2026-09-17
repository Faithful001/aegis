package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/Faithful001/aegis/internal/domain/inference"
	"github.com/Faithful001/aegis/internal/domain/inference/dto"
	domainProvider "github.com/Faithful001/aegis/internal/domain/provider"
)

type ProviderRouter struct {
	adapters map[string]ProviderAdapter
}

func NewProviderRouter() *ProviderRouter {
	return &ProviderRouter{
		adapters: map[string]ProviderAdapter{
			domainProvider.ProviderOpenAI:    NewOpenAIAdapter(),
			domainProvider.ProviderAnthropic: NewAnthropicAdapter(),
			domainProvider.ProviderGemini:    NewGeminiAdapter(),
			domainProvider.ProviderMistral:   NewMistralAdapter(),
		},
	}
}

func (r *ProviderRouter) RegisterAdapter(providerName string, adapter ProviderAdapter) {
	r.adapters[strings.ToLower(providerName)] = adapter
}

// GetProviderForModel returns the provider string ("openai", "anthropic", "gemini", "mistral") or empty string "" if local worker.
func (r *ProviderRouter) GetProviderForModel(model string) string {
	m := strings.ToLower(strings.TrimSpace(model))

	if strings.HasPrefix(m, "gpt-") || strings.HasPrefix(m, "o1") || strings.HasPrefix(m, "o3") || strings.HasPrefix(m, "text-embedding") {
		return domainProvider.ProviderOpenAI
	}
	if strings.HasPrefix(m, "claude-") {
		return domainProvider.ProviderAnthropic
	}
	if strings.HasPrefix(m, "gemini-") {
		return domainProvider.ProviderGemini
	}
	if strings.HasPrefix(m, "mistral-") || strings.HasPrefix(m, "codestral") || strings.HasPrefix(m, "pixtral") {
		return domainProvider.ProviderMistral
	}

	return "" // Local worker model fallback
}

func (r *ProviderRouter) GetAdapter(providerName string) (ProviderAdapter, error) {
	p := strings.ToLower(providerName)
	adapter, ok := r.adapters[p]
	if !ok {
		return nil, fmt.Errorf("no adapter registered for provider '%s'", providerName)
	}
	return adapter, nil
}

func (r *ProviderRouter) Generate(
	ctx context.Context,
	providerName, apiKey, baseURL string,
	req dto.ChatCompletionRequest,
) (*dto.ChatCompletionResponse, error) {
	adapter, err := r.GetAdapter(providerName)
	if err != nil {
		return nil, err
	}
	return adapter.Generate(ctx, apiKey, baseURL, req)
}

func (r *ProviderRouter) StreamGenerate(
	ctx context.Context,
	providerName, apiKey, baseURL string,
	req dto.ChatCompletionRequest,
) (<-chan inference.StreamChunk, error) {
	adapter, err := r.GetAdapter(providerName)
	if err != nil {
		return nil, err
	}
	return adapter.StreamGenerate(ctx, apiKey, baseURL, req)
}
