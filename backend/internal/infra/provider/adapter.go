package provider

import (
	"context"
	"errors"

	"github.com/Faithful001/aegis/internal/domain/inference"
	"github.com/Faithful001/aegis/internal/domain/inference/dto"
	providerDTO "github.com/Faithful001/aegis/internal/domain/provider/dto"
)

var (
	ErrProviderAPIError = errors.New("external provider API error")
	ErrUnsupportedModel = errors.New("unsupported model for provider")
)

type ProviderAdapter interface {
	Generate(ctx context.Context, apiKey, baseURL string, req dto.ChatCompletionRequest) (*dto.ChatCompletionResponse, error)
	StreamGenerate(ctx context.Context, apiKey, baseURL string, req dto.ChatCompletionRequest) (<-chan inference.StreamChunk, error)
	ListModels(ctx context.Context, apiKey, baseURL string) ([]providerDTO.ModelDescriptor, error)
}
