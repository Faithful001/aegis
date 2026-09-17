package inference

import (
	"context"
	"testing"

	"github.com/Faithful001/aegis/internal/domain/inference/dto"
	"github.com/google/uuid"
)

type mockGateway struct {
	providerForModel map[string]string
}

func (m *mockGateway) GetProviderForModel(model string) string {
	return m.providerForModel[model]
}

func (m *mockGateway) Generate(
	ctx context.Context,
	providerName, apiKey, baseURL string,
	req dto.ChatCompletionRequest,
) (*dto.ChatCompletionResponse, error) {
	return &dto.ChatCompletionResponse{
		ID:    "mock-byok-id",
		Model: req.Model,
		Choices: []dto.ChoiceDTO{
			{
				Message: dto.ChoiceMessageDTO{
					Role:    "assistant",
					Content: "Hello from " + providerName,
				},
				FinishReason: "stop",
			},
		},
		Usage: dto.UsageDTO{
			PromptTokens:     10,
			CompletionTokens: 5,
			TotalTokens:      15,
		},
	}, nil
}

func (m *mockGateway) StreamGenerate(
	ctx context.Context,
	providerName, apiKey, baseURL string,
	req dto.ChatCompletionRequest,
) (<-chan StreamChunk, error) {
	ch := make(chan StreamChunk, 2)
	ch <- StreamChunk{Content: "Hello ", Done: false}
	ch <- StreamChunk{Content: "world", Done: true, PromptTokens: 10, OutputTokens: 5, TotalTokens: 15}
	close(ch)
	return ch, nil
}

func TestInferenceServiceBYOKRouting(t *testing.T) {
	mockWorker := NewMockWorkerClient()
	svc := NewInferenceService(mockWorker, nil)

	// Attach mock gateway without providerSvc configured yet -> should fail if key not found
	gw := &mockGateway{
		providerForModel: map[string]string{
			"gpt-4o": "openai",
		},
	}
	svc.SetProviderGateway(nil, gw)

	orgID := uuid.New()
	projectID := uuid.New()

	req := dto.ChatCompletionRequest{
		Model: "gpt-4o",
		Messages: []dto.ChatMessageDTO{
			{Role: "user", Content: "Hello"},
		},
	}

	// Should fail because providerSvc is nil
	_, err := svc.ExecuteChatCompletion(context.Background(), "req-1", orgID, projectID, req)
	if err == nil {
		t.Errorf("Expected error when providerSvc is nil, got nil")
	}
}
