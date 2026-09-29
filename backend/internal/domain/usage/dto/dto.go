package dto

import (
	"time"

	"github.com/google/uuid"
)

// UsageSummaryResponse represents aggregated token usage metrics.
type UsageSummaryResponse struct {
	OrganizationID uuid.UUID  `json:"organization_id"`
	InputTokens    int        `json:"input_tokens"`
	OutputTokens   int        `json:"output_tokens"`
	TotalTokens    int        `json:"total_tokens"`
	StartTime      *time.Time `json:"start_time,omitempty"`
	EndTime        *time.Time `json:"end_time,omitempty"`
}

// UsageRecordResponse represents a formatted usage record item.
type UsageRecordResponse struct {
	ID             uuid.UUID `json:"id"`
	EventID        string    `json:"event_id"`
	RequestID      string    `json:"request_id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	ProjectID      uuid.UUID `json:"project_id"`
	Model          string    `json:"model"`
	InputTokens    int       `json:"input_tokens"`
	OutputTokens   int       `json:"output_tokens"`
	TotalTokens    int       `json:"total_tokens"`
	DurationMS     int64     `json:"duration_ms"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
}

// UsageListResponse represents a paginated list of usage records.
type UsageListResponse struct {
	Items  []UsageRecordResponse `json:"items"`
	Limit  int                   `json:"limit"`
	Offset int                   `json:"offset"`
}
