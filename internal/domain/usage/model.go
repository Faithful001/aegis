package usage

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	// ErrDuplicateUsageEvent is returned when a usage event with an existing EventID is processed.
	ErrDuplicateUsageEvent = errors.New("duplicate usage event: record already exists")

	// ErrUsageRecordNotFound is returned when the requested usage record does not exist.
	ErrUsageRecordNotFound = errors.New("usage record not found")

	// ErrInvalidUsagePayload is returned when required fields in the usage payload are missing.
	ErrInvalidUsagePayload = errors.New("invalid usage payload")
)

// UsageRecord represents an immutable database entity recording model token consumption.
type UsageRecord struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	EventID        string    `gorm:"type:varchar(255);uniqueIndex:idx_usage_event_id;not null" json:"event_id"`
	RequestID      string    `gorm:"type:varchar(255);index:idx_usage_request_id;not null" json:"request_id"`
	OrganizationID uuid.UUID `gorm:"type:uuid;index:idx_usage_org_id;not null" json:"organization_id"`
	ProjectID      uuid.UUID `gorm:"type:uuid;index:idx_usage_project_id;not null" json:"project_id"`
	Model          string    `gorm:"type:varchar(255);not null" json:"model"`
	InputTokens    int       `gorm:"type:integer;not null;default:0" json:"input_tokens"`
	OutputTokens   int       `gorm:"type:integer;not null;default:0" json:"output_tokens"`
	TotalTokens    int       `gorm:"type:integer;not null;default:0" json:"total_tokens"`
	DurationMS     int64     `gorm:"type:bigint;not null;default:0" json:"duration_ms"`
	Status         string    `gorm:"type:varchar(50);not null" json:"status"`
	CreatedAt      time.Time `gorm:"autoCreateTime;index" json:"created_at"`
}

// TableName explicitly configures the PostgreSQL table name.
func (UsageRecord) TableName() string {
	return "usage_records"
}

// NewUsageRecord constructs a validated UsageRecord.
func NewUsageRecord(
	eventID, reqID string,
	orgID, projectID uuid.UUID,
	model string,
	inputTokens, outputTokens int,
	durationMS int64,
	status string,
) (*UsageRecord, error) {
	if eventID == "" || reqID == "" {
		return nil, ErrInvalidUsagePayload
	}

	if status == "" {
		status = "COMPLETED"
	}

	totalTokens := inputTokens + outputTokens
	if totalTokens < 0 {
		totalTokens = 0
	}

	return &UsageRecord{
		ID:             uuid.New(),
		EventID:        eventID,
		RequestID:      reqID,
		OrganizationID: orgID,
		ProjectID:      projectID,
		Model:          model,
		InputTokens:    inputTokens,
		OutputTokens:   outputTokens,
		TotalTokens:    totalTokens,
		DurationMS:     durationMS,
		Status:         status,
		CreatedAt:      time.Now().UTC(),
	}, nil
}
