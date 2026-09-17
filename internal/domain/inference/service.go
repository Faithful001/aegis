package inference

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Faithful001/aegis/internal/domain/inference/dto"
	"github.com/Faithful001/aegis/internal/domain/provider"
	"github.com/Faithful001/aegis/internal/domain/scheduler"
	"github.com/Faithful001/aegis/internal/infra/queue/kafka"
	"github.com/google/uuid"
)

type ProviderGateway interface {
	GetProviderForModel(model string) string
	Generate(ctx context.Context, providerName, apiKey, baseURL string, req dto.ChatCompletionRequest) (*dto.ChatCompletionResponse, error)
	StreamGenerate(ctx context.Context, providerName, apiKey, baseURL string, req dto.ChatCompletionRequest) (<-chan StreamChunk, error)
}

type InferenceService struct {
	workerClient    WorkerClient
	scheduler       scheduler.Scheduler
	eventProducer   kafka.EventProducer
	providerSvc     *provider.ProviderCredentialService
	providerGateway ProviderGateway
}

func NewInferenceService(workerClient WorkerClient, sched scheduler.Scheduler, eventProducer ...kafka.EventProducer) *InferenceService {
	var ep kafka.EventProducer
	if len(eventProducer) > 0 {
		ep = eventProducer[0]
	}
	return &InferenceService{
		workerClient:  workerClient,
		scheduler:     sched,
		eventProducer: ep,
	}
}

func (s *InferenceService) SetProviderGateway(
	providerSvc *provider.ProviderCredentialService,
	gateway ProviderGateway,
) {
	s.providerSvc = providerSvc
	s.providerGateway = gateway
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
	startTime := time.Now()
	if reqID == "" {
		reqID = fmt.Sprintf("req_%s", uuid.New().String())
	}

	if s.providerGateway != nil && s.providerSvc != nil {
		if providerName := s.providerGateway.GetProviderForModel(req.Model); providerName != "" {
			apiKey, baseURL, err := s.providerSvc.GetDecryptedKey(ctx, orgID, providerName)
			if err != nil {
				return nil, fmt.Errorf("BYOK key for provider '%s' not configured for organization: %w", providerName, err)
			}

			resp, err := s.providerGateway.Generate(ctx, providerName, apiKey, baseURL, req)
			if err != nil {
				return nil, err
			}

			if s.eventProducer != nil {
				evt := kafka.NewUsageEvent(
					"",
					reqID,
					orgID,
					projectID,
					req.Model,
					resp.Usage.PromptTokens,
					resp.Usage.CompletionTokens,
					time.Since(startTime).Milliseconds(),
					"COMPLETED",
				)
				_ = s.eventProducer.Publish(ctx, kafka.TopicUsageEvents, reqID, evt)
			}
			return resp, nil
		}
	}

	// Local worker fallback route
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

	if s.eventProducer != nil {
		evt := kafka.NewUsageEvent(
			"",
			reqID,
			orgID,
			projectID,
			job.Model,
			result.PromptTokens,
			result.CompletionTokens,
			time.Since(startTime).Milliseconds(),
			"COMPLETED",
		)
		_ = s.eventProducer.Publish(ctx, kafka.TopicUsageEvents, reqID, evt)
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
	startTime := time.Now()
	if reqID == "" {
		reqID = fmt.Sprintf("req_%s", uuid.New().String())
	}

	// 1. Check if model routes to a BYOK frontier provider
	if s.providerGateway != nil && s.providerSvc != nil {
		if providerName := s.providerGateway.GetProviderForModel(req.Model); providerName != "" {
			apiKey, baseURL, err := s.providerSvc.GetDecryptedKey(ctx, orgID, providerName)
			if err != nil {
				return nil, nil, fmt.Errorf("BYOK key for provider '%s' not configured for organization: %w", providerName, err)
			}

			chunkChan, err := s.providerGateway.StreamGenerate(ctx, providerName, apiKey, baseURL, req)
			if err != nil {
				return nil, nil, err
			}

			job, _ := NewInferenceJob(reqID, orgID, projectID, req.Model, nil, req.MaxTokens, req.Temperature, true)

			outChan := make(chan StreamChunk)
			go func() {
				defer close(outChan)
				for chunk := range chunkChan {
					if chunk.Done && s.eventProducer != nil {
						evt := kafka.NewUsageEvent(
							"",
							reqID,
							orgID,
							projectID,
							req.Model,
							chunk.PromptTokens,
							chunk.OutputTokens,
							time.Since(startTime).Milliseconds(),
							"COMPLETED",
						)
						_ = s.eventProducer.Publish(ctx, kafka.TopicUsageEvents, reqID, evt)
					}
					outChan <- chunk
				}
			}()
			return outChan, job, nil
		}
	}

	// 2. Local worker fallback route
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

	if s.eventProducer == nil {
		return chunkChan, job, nil
	}

	outChan := make(chan StreamChunk)
	go func() {
		defer close(outChan)
		for chunk := range chunkChan {
			if chunk.Done && chunk.TotalTokens > 0 {
				evt := kafka.NewUsageEvent(
					"",
					reqID,
					orgID,
					projectID,
					job.Model,
					chunk.PromptTokens,
					chunk.OutputTokens,
					time.Since(startTime).Milliseconds(),
					"COMPLETED",
				)
				_ = s.eventProducer.Publish(ctx, kafka.TopicUsageEvents, reqID, evt)
			}
			outChan <- chunk
		}
	}()

	return outChan, job, nil
}
