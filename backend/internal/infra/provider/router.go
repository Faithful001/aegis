package provider

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/Faithful001/aegis/internal/domain/inference"
	"github.com/Faithful001/aegis/internal/domain/inference/dto"
	domainProvider "github.com/Faithful001/aegis/internal/domain/provider"
	providerDTO "github.com/Faithful001/aegis/internal/domain/provider/dto"
)

type ProviderRouter struct {
	adapters    map[string]ProviderAdapter
	knownModels sync.Map
}

func NewProviderRouter() *ProviderRouter {
	return &ProviderRouter{
		adapters: map[string]ProviderAdapter{
			domainProvider.ProviderOpenAI:     NewOpenAIAdapter(),
			domainProvider.ProviderAnthropic:  NewAnthropicAdapter(),
			domainProvider.ProviderGemini:     NewGeminiAdapter(),
			domainProvider.ProviderMistral:    NewMistralAdapter(),
			domainProvider.ProviderOpenRouter: NewOpenRouterAdapter(),
		},
	}
}

func (r *ProviderRouter) RegisterAdapter(providerName string, adapter ProviderAdapter) {
	r.adapters[strings.ToLower(providerName)] = adapter
}

func (r *ProviderRouter) RegisterModel(modelID, providerName string) {
	if modelID != "" && providerName != "" {
		r.knownModels.Store(strings.ToLower(strings.TrimSpace(modelID)), strings.ToLower(strings.TrimSpace(providerName)))
	}
}

// GetProviderForModel returns the provider string ("openai", "anthropic", "gemini", "mistral", "openrouter") or empty string "".
func (r *ProviderRouter) GetProviderForModel(model string) string {
	m := strings.ToLower(strings.TrimSpace(model))

	// 1. Check dynamically registered models cache
	if p, ok := r.knownModels.Load(m); ok {
		return p.(string)
	}

	// 2. Prefix matching heuristics
	if strings.HasPrefix(m, "gpt-") || strings.HasPrefix(m, "o1") || strings.HasPrefix(m, "o3") || strings.HasPrefix(m, "chatgpt-") || strings.HasPrefix(m, "text-embedding") {
		return domainProvider.ProviderOpenAI
	}
	if strings.HasPrefix(m, "claude-") {
		return domainProvider.ProviderAnthropic
	}
	if strings.HasPrefix(m, "gemini-") {
		return domainProvider.ProviderGemini
	}
	if strings.HasPrefix(m, "mistral-") || strings.HasPrefix(m, "codestral") || strings.HasPrefix(m, "pixtral") || strings.HasPrefix(m, "ministral") || strings.HasPrefix(m, "open-mistral") {
		return domainProvider.ProviderMistral
	}

	if strings.Contains(m, "/") {
		return domainProvider.ProviderOpenRouter
	}

	return ""
}

func (r *ProviderRouter) GetAdapter(providerName string) (ProviderAdapter, error) {
	p := strings.ToLower(providerName)
	adapter, ok := r.adapters[p]
	if !ok {
		return nil, fmt.Errorf("no adapter registered for provider '%s'", providerName)
	}
	return adapter, nil
}

func (r *ProviderRouter) ListModels(
	ctx context.Context,
	providerName, apiKey, baseURL string,
) ([]providerDTO.ModelDescriptor, error) {
	adapter, err := r.GetAdapter(providerName)
	if err != nil {
		return nil, err
	}

	models, err := adapter.ListModels(ctx, apiKey, baseURL)
	if err != nil {
		return nil, err
	}

	for _, m := range models {
		r.RegisterModel(m.ID, providerName)
	}

	return models, nil
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
