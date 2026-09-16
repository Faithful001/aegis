package inference

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Faithful001/aegis/internal/domain/inference/dto"
	"github.com/Faithful001/aegis/internal/domain/scheduler"
	"github.com/google/uuid"
)

type InferenceService struct {
	workerClient WorkerClient
	scheduler    scheduler.Scheduler
}

func NewInferenceService(workerClient WorkerClient, sched ...scheduler.Scheduler) *InferenceService {
	var s scheduler.Scheduler
	if len(sched) > 0 {
		s = sched[0]
	}
	return &InferenceService{
		workerClient: workerClient,
		scheduler:    s,
	}
}

func (s *InferenceService) selectWorker(ctx context.Context, req dto.ChatCompletionRequest, job *InferenceJob) error {
	if s.scheduler == nil {
		return nil
	}

	estTokens := req.MaxTokens
	for _, m := range req.Messages {
		estTokens += len(m.Content) / 4
	}

	schedRes, err := s.scheduler.SelectWorker(ctx, scheduler.ScheduleRequest{
		Model:           req.Model,
		EstimatedTokens: estTokens,
	})
	if err != nil {
		if errors.Is(err, scheduler.ErrNoWorkersForModel) {
			return fmt.Errorf("%w: %s", ErrModelNotSupported, req.Model)
		}
		if errors.Is(err, scheduler.ErrNoWorkersAvailable) {
			return ErrWorkerUnavailable
		}
		return err
	}

	if job != nil {
		job.WorkerID = schedRes.WorkerID
		job.WorkerAddress = schedRes.Address
	}
	return nil
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

	if err := s.selectWorker(ctx, req, job); err != nil {
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

	if err := s.selectWorker(ctx, req, job); err != nil {
		return nil, nil, err
	}

	chunkChan, err := s.workerClient.StreamGenerate(ctx, job)
	if err != nil {
		return nil, nil, err
	}

	return chunkChan, job, nil
}

