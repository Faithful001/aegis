package inference

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type JobStatus string

const (
	JobStatusPending   JobStatus = "PENDING"
	JobStatusRunning   JobStatus = "RUNNING"
	JobStatusCompleted JobStatus = "COMPLETED"
	JobStatusFailed    JobStatus = "FAILED"
	JobStatusCanceled  JobStatus = "CANCELED"
)

var DefaultSupportedModels = map[string]bool{
	"mistral-small":         true,
	"mistral-small-latest":  true,
	"mistral-medium":        true,
	"mistral-medium-latest": true,
	"mistral-large":         true,
	"mistral-large-latest":  true,
	"open-mistral-7b":       true,
	"open-mixtral-8x7b":     true,
	"open-mixtral-8x22b":    true,
	"codestral-latest":      true,
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type InferenceJob struct {
	JobID          string     `json:"job_id"`
	RequestID      string     `json:"request_id"`
	OrganizationID uuid.UUID  `json:"organization_id"`
	ProjectID      uuid.UUID  `json:"project_id"`
	Model          string     `json:"model"`
	Messages       []Message  `json:"messages"`
	MaxTokens      int        `json:"max_tokens"`
	Temperature    float32    `json:"temperature"`
	Stream         bool       `json:"stream"`
	Status         JobStatus  `json:"status"`
	WorkerID       string     `json:"worker_id,omitempty"`
	WorkerAddress  string     `json:"worker_address,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
}

type InferenceResult struct {
	Content          string `json:"content"`
	PromptTokens     int    `json:"prompt_tokens"`
	CompletionTokens int    `json:"completion_tokens"`
	TotalTokens      int    `json:"total_tokens"`
	FinishReason     string `json:"finish_reason"`
}

type StreamChunk struct {
	Content      string
	Done         bool
	FinishReason string
	PromptTokens int
	OutputTokens int
	TotalTokens  int
	Error        string
}

func NewInferenceJob(
	reqID string,
	orgID, projectID uuid.UUID,
	model string,
	messages []Message,
	maxTokens int,
	temperature float32,
	stream bool,
) (*InferenceJob, error) {
	if len(messages) == 0 {
		return nil, ErrEmptyMessages
	}

	if temperature < 0.0 || temperature > 2.0 {
		return nil, ErrInvalidTemperature
	}

	if maxTokens <= 0 {
		maxTokens = 500 // default max tokens
	}

	if model == "" {
		model = "mistral-small-latest"
	}

	if !DefaultSupportedModels[model] {
		return nil, fmt.Errorf("%w: %s", ErrModelNotSupported, model)
	}

	jobID := fmt.Sprintf("job_%s", uuid.New().String())

	return &InferenceJob{
		JobID:          jobID,
		RequestID:      reqID,
		OrganizationID: orgID,
		ProjectID:      projectID,
		Model:          model,
		Messages:       messages,
		MaxTokens:      maxTokens,
		Temperature:    temperature,
		Stream:         stream,
		Status:         JobStatusPending,
		CreatedAt:      time.Now().UTC(),
	}, nil
}
