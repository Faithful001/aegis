package inference

import (
	"context"
	"testing"

	"github.com/Faithful001/aegis/internal/domain/inference/dto"
	"github.com/google/uuid"
)

func TestInferenceService_ExecuteChatCompletion(t *testing.T) {
	mockClient := NewMockWorkerClient()
	service := NewService(mockClient)

	orgID := uuid.New()
	projectID := uuid.New()

	req := dto.ChatCompletionRequest{
		Model: "mistral-small",
		Messages: []dto.ChatMessageDTO{
			{Role: "user", Content: "Hello Aegis!"},
		},
		MaxTokens:   100,
		Temperature: 0.7,
		Stream:      false,
	}

	resp, err := service.ExecuteChatCompletion(context.Background(), "req_test_123", orgID, projectID, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.ID != "req_test_123" {
		t.Errorf("expected ID req_test_123, got %s", resp.ID)
	}
	if resp.Model != "mistral-small" {
		t.Errorf("expected model mistral-small, got %s", resp.Model)
	}
	if len(resp.Choices) != 1 {
		t.Fatalf("expected 1 choice, got %d", len(resp.Choices))
	}
	if resp.Choices[0].Message.Content == "" {
		t.Errorf("expected non-empty content")
	}
	if resp.Usage.TotalTokens == 0 {
		t.Errorf("expected non-zero usage tokens")
	}
}

func TestInferenceService_ValidationErrors(t *testing.T) {
	mockClient := NewMockWorkerClient()
	service := NewService(mockClient)

	orgID := uuid.New()
	projectID := uuid.New()

	// Unsupported model
	reqInvalidModel := dto.ChatCompletionRequest{
		Model: "unsupported-model-xyz",
		Messages: []dto.ChatMessageDTO{
			{Role: "user", Content: "Hi"},
		},
	}
	_, err := service.ExecuteChatCompletion(context.Background(), "req_test", orgID, projectID, reqInvalidModel)
	if err == nil {
		t.Errorf("expected error for unsupported model, got nil")
	}

	// Empty messages
	reqEmptyMsgs := dto.ChatCompletionRequest{
		Model:    "mistral-small",
		Messages: []dto.ChatMessageDTO{},
	}
	_, err = service.ExecuteChatCompletion(context.Background(), "req_test", orgID, projectID, reqEmptyMsgs)
	if err == nil {
		t.Errorf("expected error for empty messages, got nil")
	}

	// Invalid temperature (> 2.0)
	reqInvalidTemp := dto.ChatCompletionRequest{
		Model: "mistral-small",
		Messages: []dto.ChatMessageDTO{
			{Role: "user", Content: "Hi"},
		},
		Temperature: 3.5,
	}
	_, err = service.ExecuteChatCompletion(context.Background(), "req_test", orgID, projectID, reqInvalidTemp)
	if err == nil {
		t.Errorf("expected error for invalid temperature, got nil")
	}
}

func TestInferenceService_ExecuteStreamChatCompletion(t *testing.T) {
	mockClient := NewMockWorkerClient()
	service := NewService(mockClient)

	orgID := uuid.New()
	projectID := uuid.New()

	req := dto.ChatCompletionRequest{
		Model: "mistral-small",
		Messages: []dto.ChatMessageDTO{
			{Role: "user", Content: "Stream this"},
		},
		Stream: true,
	}

	chunkChan, job, err := service.ExecuteStreamChatCompletion(context.Background(), "req_stream_123", orgID, projectID, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if job.JobID == "" {
		t.Errorf("expected non-empty job ID")
	}

	chunksReceived := 0
	for chunk := range chunkChan {
		chunksReceived++
		if chunk.Error != "" {
			t.Errorf("unexpected chunk error: %s", chunk.Error)
		}
	}

	if chunksReceived < 2 {
		t.Errorf("expected at least 2 chunks, got %d", chunksReceived)
	}
}
