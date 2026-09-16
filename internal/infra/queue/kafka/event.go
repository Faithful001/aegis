package kafka

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const (
	TopicUsageEvents = "aegis.events.usage"
	TopicJobEvents   = "aegis.events.jobs"
)

// Event is the interface implemented by all infrastructure events.
type Event interface {
	GetID() string
	GetType() string
	GetTimestamp() time.Time
}

// BaseEvent provides common metadata fields for events.
type BaseEvent struct {
	EventID   string    `json:"event_id"`
	EventType string    `json:"event_type"`
	Timestamp time.Time `json:"timestamp"`
}

func (b BaseEvent) GetID() string {
	return b.EventID
}

func (b BaseEvent) GetType() string {
	return b.EventType
}

func (b BaseEvent) GetTimestamp() time.Time {
	return b.Timestamp
}

// UsageEvent represents an LLM token consumption event recorded after an inference job.
type UsageEvent struct {
	BaseEvent
	RequestID      string    `json:"request_id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	ProjectID      uuid.UUID `json:"project_id"`
	Model          string    `json:"model"`
	InputTokens    int       `json:"input_tokens"`
	OutputTokens   int       `json:"output_tokens"`
	TotalTokens    int       `json:"total_tokens"`
	DurationMS     int64     `json:"duration_ms"`
	Status         string    `json:"status"` // "COMPLETED", "FAILED"
}

// NewUsageEvent constructs a new UsageEvent.
func NewUsageEvent(
	eventID, reqID string,
	orgID, projectID uuid.UUID,
	model string,
	inputTokens, outputTokens int,
	durationMS int64,
	status string,
) UsageEvent {
	if eventID == "" {
		eventID = "evt_" + uuid.New().String()
	}
	return UsageEvent{
		BaseEvent: BaseEvent{
			EventID:   eventID,
			EventType: "usage.recorded",
			Timestamp: time.Now().UTC(),
		},
		RequestID:      reqID,
		OrganizationID: orgID,
		ProjectID:      projectID,
		Model:          model,
		InputTokens:    inputTokens,
		OutputTokens:   outputTokens,
		TotalTokens:    inputTokens + outputTokens,
		DurationMS:     durationMS,
		Status:         status,
	}
}

// JobEvent represents a lifecycle status update for an inference job.
type JobEvent struct {
	BaseEvent
	JobID          string    `json:"job_id"`
	RequestID      string    `json:"request_id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	ProjectID      uuid.UUID `json:"project_id"`
	Model          string    `json:"model"`
	WorkerID       string    `json:"worker_id,omitempty"`
	Status         string    `json:"status"`
}

// NewJobEvent constructs a new JobEvent.
func NewJobEvent(
	eventID, jobID, reqID string,
	orgID, projectID uuid.UUID,
	model, workerID, status string,
) JobEvent {
	if eventID == "" {
		eventID = "evt_" + uuid.New().String()
	}
	return JobEvent{
		BaseEvent: BaseEvent{
			EventID:   eventID,
			EventType: "job.status_changed",
			Timestamp: time.Now().UTC(),
		},
		JobID:          jobID,
		RequestID:      reqID,
		OrganizationID: orgID,
		ProjectID:      projectID,
		Model:          model,
		WorkerID:       workerID,
		Status:         status,
	}
}

// EventProducer defines the interface for publishing events to an asynchronous event bus (Kafka).
type EventProducer interface {
	Publish(ctx context.Context, topic string, key string, evt Event) error
	Close() error
}

// EventHandler is a callback invoked when a message is received from a topic.
type EventHandler func(ctx context.Context, topic string, key string, payload []byte) error

// EventConsumer defines the interface for subscribing to asynchronous event topics (Kafka).
type EventConsumer interface {
	Subscribe(ctx context.Context, topic string, handler EventHandler) error
	Close() error
}
