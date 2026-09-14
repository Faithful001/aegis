package inference

import (
	"context"
	"fmt"
	"time"

	"github.com/Faithful001/aegis/internal/domain/inference/dto"
	"github.com/google/uuid"
)

type InferenceService struct {
	workerClient WorkerClient
}

func NewInferenceService(workerClient WorkerClient) *InferenceService {
	return &InferenceService{
		workerClient: workerClient,
	}
}

func (s *InferenceService) ExecuteChatCompletion(
	ctx context.Context,
	reqID string,
	orgID, projectID uuid.UUID,
	req dto.ChatCompletionRequest,
) (*dto.ChatCompletionResponse, error) {
	if reqID == "" {
		reqID = fmt.Sprintf("req_%s", uuid.New().String())
	}

	domainMessages := make([]Message, len(req.Messages))
	for i, m := range req.Messages {
		domainMessages[i] = Message{
			Role:    m.Role,
			Content: m.Content,
		}
	}

	job, err := NewInferenceJob(
		reqID,
		orgID,
		projectID,
		req.Model,
		domainMessages,
		req.MaxTokens,
		req.Temperature,
		false,
	)
	if err != nil {
		return nil, err
	}

	result, err := s.workerClient.Generate(ctx, job)
	if err != nil {
		return nil, err
	}

	return &dto.ChatCompletionResponse{
		ID:      reqID,
		Object:  "chat.completion",
		Created: time.Now().UTC().Unix(),
		Model:   job.Model,
		Choices: []dto.ChoiceDTO{
			{
				Index: 0,
				Message: dto.ChoiceMessageDTO{
					Role:    "assistant",
					Content: result.Content,
				},
				FinishReason: result.FinishReason,
			},
		},
		Usage: dto.UsageDTO{
			PromptTokens:     result.PromptTokens,
			CompletionTokens: result.CompletionTokens,
			TotalTokens:      result.TotalTokens,
		},
	}, nil
}

func (s *InferenceService) ExecuteStreamChatCompletion(
	ctx context.Context,
	reqID string,
	orgID, projectID uuid.UUID,
	req dto.ChatCompletionRequest,
) (<-chan StreamChunk, *InferenceJob, error) {
	if reqID == "" {
		reqID = fmt.Sprintf("req_%s", uuid.New().String())
	}

	domainMessages := make([]Message, len(req.Messages))
	for i, m := range req.Messages {
		domainMessages[i] = Message{
			Role:    m.Role,
			Content: m.Content,
		}
	}

	job, err := NewInferenceJob(
		reqID,
		orgID,
		projectID,
		req.Model,
		domainMessages,
		req.MaxTokens,
		req.Temperature,
		true,
	)
	if err != nil {
		return nil, nil, err
	}

	chunkChan, err := s.workerClient.StreamGenerate(ctx, job)
	if err != nil {
		return nil, nil, err
	}

	return chunkChan, job, nil
}
