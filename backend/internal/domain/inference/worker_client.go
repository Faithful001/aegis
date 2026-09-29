package inference

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	inferencepb "github.com/Faithful001/aegis/proto/inference"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type WorkerClient interface {
	Generate(ctx context.Context, job *InferenceJob) (*InferenceResult, error)
	StreamGenerate(ctx context.Context, job *InferenceJob) (<-chan StreamChunk, error)
	HealthCheck(ctx context.Context) (string, []string, error)
	Close() error
}

type GRPCWorkerClient struct {
	targetAddress string
	conn          *grpc.ClientConn
	client        inferencepb.InferenceServiceClient
	mu            sync.RWMutex
}

func NewGRPCWorkerClient(targetAddress string) (*GRPCWorkerClient, error) {
	conn, err := grpc.NewClient(
		targetAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to dial inference worker target %s: %w", targetAddress, err)
	}

	client := inferencepb.NewInferenceServiceClient(conn)
	return &GRPCWorkerClient{
		targetAddress: targetAddress,
		conn:          conn,
		client:        client,
	}, nil
}

func (c *GRPCWorkerClient) buildProtoRequest(job *InferenceJob, stream bool) *inferencepb.GenerateRequest {
	protoMsgs := make([]*inferencepb.ChatMessage, len(job.Messages))
	for i, m := range job.Messages {
		protoMsgs[i] = &inferencepb.ChatMessage{
			Role:    m.Role,
			Content: m.Content,
		}
	}

	return &inferencepb.GenerateRequest{
		RequestId:      job.RequestID,
		JobId:          job.JobID,
		Model:          job.Model,
		Messages:       protoMsgs,
		MaxTokens:      int32(job.MaxTokens),
		Temperature:    job.Temperature,
		Stream:         stream,
		OrganizationId: job.OrganizationID.String(),
		ProjectId:      job.ProjectID.String(),
	}
}

func (c *GRPCWorkerClient) Generate(ctx context.Context, job *InferenceJob) (*InferenceResult, error) {
	req := c.buildProtoRequest(job, false)
	stream, err := c.client.Generate(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInferenceFailed, err)
	}

	var finalResult InferenceResult
	var fullContent string

	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrInferenceFailed, err)
		}

		if resp.GetError() != "" {
			return nil, fmt.Errorf("%w: %s", ErrInferenceFailed, resp.GetError())
		}

		if resp.GetContent() != "" {
			fullContent += resp.GetContent()
		}

		if resp.GetDone() {
			finalResult.FinishReason = resp.GetFinishReason()
			if finalResult.FinishReason == "" {
				finalResult.FinishReason = "stop"
			}
			if usage := resp.GetUsage(); usage != nil {
				finalResult.PromptTokens = int(usage.GetPromptTokens())
				finalResult.CompletionTokens = int(usage.GetCompletionTokens())
				finalResult.TotalTokens = int(usage.GetTotalTokens())
			}
		}
	}

	finalResult.Content = fullContent
	if finalResult.TotalTokens == 0 {
		finalResult.PromptTokens = len(job.Messages) * 10
		finalResult.CompletionTokens = len(fullContent) / 4
		finalResult.TotalTokens = finalResult.PromptTokens + finalResult.CompletionTokens
	}
	if finalResult.FinishReason == "" {
		finalResult.FinishReason = "stop"
	}

	return &finalResult, nil
}

func (c *GRPCWorkerClient) StreamGenerate(ctx context.Context, job *InferenceJob) (<-chan StreamChunk, error) {
	req := c.buildProtoRequest(job, true)
	stream, err := c.client.Generate(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrStreamFailed, err)
	}

	chunkChan := make(chan StreamChunk, 32)

	go func() {
		defer close(chunkChan)
		for {
			select {
			case <-ctx.Done():
				chunkChan <- StreamChunk{Error: ctx.Err().Error(), Done: true}
				return
			default:
				resp, err := stream.Recv()
				if err == io.EOF {
					return
				}
				if err != nil {
					chunkChan <- StreamChunk{Error: err.Error(), Done: true}
					return
				}

				var promptTokens, outputTokens, totalTokens int
				if usage := resp.GetUsage(); usage != nil {
					promptTokens = int(usage.GetPromptTokens())
					outputTokens = int(usage.GetCompletionTokens())
					totalTokens = int(usage.GetTotalTokens())
				}

				chunk := StreamChunk{
					Content:      resp.GetContent(),
					Done:         resp.GetDone(),
					FinishReason: resp.GetFinishReason(),
					PromptTokens: promptTokens,
					OutputTokens: outputTokens,
					TotalTokens:  totalTokens,
					Error:        resp.GetError(),
				}

				select {
				case chunkChan <- chunk:
				case <-ctx.Done():
					return
				}

				if resp.GetDone() {
					return
				}
			}
		}
	}()

	return chunkChan, nil
}

func (c *GRPCWorkerClient) HealthCheck(ctx context.Context) (string, []string, error) {
	resp, err := c.client.HealthCheck(ctx, &inferencepb.HealthCheckRequest{
		WorkerId: "health_checker",
	})
	if err != nil {
		return "", nil, err
	}
	return resp.GetStatus(), resp.GetSupportedModels(), nil
}

func (c *GRPCWorkerClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// MockWorkerClient for standalone testing and fallback
type MockWorkerClient struct {
	CustomResponse string
	CustomError    error
}

func NewMockWorkerClient() *MockWorkerClient {
	return &MockWorkerClient{}
}

func (m *MockWorkerClient) Generate(ctx context.Context, job *InferenceJob) (*InferenceResult, error) {
	if m.CustomError != nil {
		return nil, m.CustomError
	}
	content := m.CustomResponse
	if content == "" {
		content = fmt.Sprintf("Mock inference response for model %s on request %s", job.Model, job.RequestID)
	}

	return &InferenceResult{
		Content:          content,
		PromptTokens:     15,
		CompletionTokens: 20,
		TotalTokens:      35,
		FinishReason:     "stop",
	}, nil
}

func (m *MockWorkerClient) StreamGenerate(ctx context.Context, job *InferenceJob) (<-chan StreamChunk, error) {
	if m.CustomError != nil {
		return nil, m.CustomError
	}

	ch := make(chan StreamChunk, 4)
	go func() {
		defer close(ch)
		tokens := []string{"Mock ", "streaming ", "inference ", "response."}
		for _, tok := range tokens {
			select {
			case <-ctx.Done():
				return
			case <-time.After(5 * time.Millisecond):
				ch <- StreamChunk{Content: tok, Done: false}
			}
		}
		ch <- StreamChunk{
			Content:      "",
			Done:         true,
			FinishReason: "stop",
			PromptTokens: 15,
			OutputTokens: 20,
			TotalTokens:  35,
		}
	}()

	return ch, nil
}

func (m *MockWorkerClient) HealthCheck(ctx context.Context) (string, []string, error) {
	return "SERVING", []string{"mistral-small", "mistral-small-latest"}, nil
}

func (m *MockWorkerClient) Close() error {
	return nil
}
