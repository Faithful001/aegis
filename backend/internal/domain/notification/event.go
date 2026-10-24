package notification

import (
	"time"

	"github.com/Faithful001/aegis/internal/infra/queue/kafka"
	"github.com/google/uuid"
)

type FailedRequestEvent struct {
	kafka.BaseEvent
	RequestID      string    `json:"request_id"`
	UserID         uuid.UUID `json:"user_id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	ProjectID      uuid.UUID `json:"project_id"`
	Model          string    `json:"model"`
	Title          string    `json:"title"`
	Message        string    `json:"message"`
}

func NewFailedRequestEvent(
	requestID string,
	userID uuid.UUID,
	organizationID uuid.UUID,
	projectID uuid.UUID,
	model string,
	title string,
	message string,
) *FailedRequestEvent {
	return &FailedRequestEvent{
		BaseEvent: kafka.BaseEvent{
			EventID:   uuid.NewString(),
			EventType: string(TypeFailedRequest),
			Timestamp: time.Now().UTC(),
		},
		RequestID:      requestID,
		UserID:         userID,
		OrganizationID: organizationID,
		ProjectID:      projectID,
		Model:          model,
		Title:          title,
		Message:        message,
	}
}

type BudgetAlertEvent struct {
	kafka.BaseEvent
	UserID         uuid.UUID `json:"user_id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	CurrentSpend   float64   `json:"current_spend"`
	BudgetLimit    float64   `json:"budget_limit"`
	Percentage     int       `json:"percentage"`
	Title          string    `json:"title"`
	Message        string    `json:"message"`
}

func NewBudgetAlertEvent(
	userID uuid.UUID,
	organizationID uuid.UUID,
	currentSpend, budgetLimit float64,
	percentage int,
	title, message string,
) *BudgetAlertEvent {
	return &BudgetAlertEvent{
		BaseEvent: kafka.BaseEvent{
			EventID:   uuid.NewString(),
			EventType: string(TypeBudgetAlert),
			Timestamp: time.Now().UTC(),
		},
		UserID:         userID,
		OrganizationID: organizationID,
		CurrentSpend:   currentSpend,
		BudgetLimit:    budgetLimit,
		Percentage:     percentage,
		Title:          title,
		Message:        message,
	}
}

type RateLimitEvent struct {
	kafka.BaseEvent
	UserID         uuid.UUID `json:"user_id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	ProjectID      uuid.UUID `json:"project_id"`
	Endpoint       string    `json:"endpoint"`
	Title          string    `json:"title"`
	Message        string    `json:"message"`
}

func NewRateLimitEvent(
	userID uuid.UUID,
	organizationID uuid.UUID,
	projectID uuid.UUID,
	endpoint, title, message string,
) *RateLimitEvent {
	return &RateLimitEvent{
		BaseEvent: kafka.BaseEvent{
			EventID:   uuid.NewString(),
			EventType: string(TypeRateLimit),
			Timestamp: time.Now().UTC(),
		},
		UserID:         userID,
		OrganizationID: organizationID,
		ProjectID:      projectID,
		Endpoint:       endpoint,
		Title:          title,
		Message:        message,
	}
}

type ModifiedCredentialEvent struct {
	kafka.BaseEvent
	UserID    uuid.UUID `json:"user_id"`
	Action    string    `json:"action"` // e.g. "password_changed", "api_key_regenerated"
	IPAddress string    `json:"ip_address,omitempty"`
	Title     string    `json:"title"`
	Message   string    `json:"message"`
}

func NewModifiedCredentialEvent(
	userID uuid.UUID,
	action, ipAddress, title, message string,
) *ModifiedCredentialEvent {
	return &ModifiedCredentialEvent{
		BaseEvent: kafka.BaseEvent{
			EventID:   uuid.NewString(),
			EventType: string(TypeModifiedCredential),
			Timestamp: time.Now().UTC(),
		},
		UserID:    userID,
		Action:    action,
		IPAddress: ipAddress,
		Title:     title,
		Message:   message,
	}
}

type WeeklyUsageEvent struct {
	kafka.BaseEvent
	UserID         uuid.UUID `json:"user_id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	TotalRequests  int64     `json:"total_requests"`
	TotalTokens    int64     `json:"total_tokens"`
	TotalCostUSD   float64   `json:"total_cost_usd"`
	Title          string    `json:"title"`
	Message        string    `json:"message"`
}

func NewWeeklyUsageEvent(
	userID uuid.UUID,
	organizationID uuid.UUID,
	totalRequests, totalTokens int64,
	totalCostUSD float64,
	title, message string,
) *WeeklyUsageEvent {
	return &WeeklyUsageEvent{
		BaseEvent: kafka.BaseEvent{
			EventID:   uuid.NewString(),
			EventType: string(TypeWeeklyUsage),
			Timestamp: time.Now().UTC(),
		},
		UserID:         userID,
		OrganizationID: organizationID,
		TotalRequests:  totalRequests,
		TotalTokens:    totalTokens,
		TotalCostUSD:   totalCostUSD,
		Title:          title,
		Message:        message,
	}
}

type GenericNotificationEvent struct {
	kafka.BaseEvent
	UserID   uuid.UUID        `json:"user_id"`
	Type     NotificationType `json:"type"`
	Title    string           `json:"title"`
	Message  string           `json:"message"`
	Metadata map[string]any   `json:"metadata,omitempty"`
}

func NewGenericNotificationEvent(
	userID uuid.UUID,
	notifType NotificationType,
	title, message string,
	metadata map[string]any,
) *GenericNotificationEvent {
	return &GenericNotificationEvent{
		BaseEvent: kafka.BaseEvent{
			EventID:   uuid.NewString(),
			EventType: string(notifType),
			Timestamp: time.Now().UTC(),
		},
		UserID:   userID,
		Type:     notifType,
		Title:    title,
		Message:  message,
		Metadata: metadata,
	}
}