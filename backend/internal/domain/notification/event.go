package notification

import (
	"github.com/Faithful001/aegis/internal/infra/queue/kafka"
	"github.com/google/uuid"
)

type FailedRequestEvent struct {
	kafka.BaseEvent
	RequestID      	string    		`json:"request_id"`
	UserID 			uuid.UUID 		`json:"user_id"`
	OrganizationID 	uuid.UUID 		`json:"organization_id"`
	ProjectID      	uuid.UUID 		`json:"project_id"`
	Model          	string    		`json:"model"`
	Title			string 			`json:"title"`
	Message			string 			`json:"message"`
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
		RequestID:      requestID,
		UserID: 		userID,
		OrganizationID: organizationID,
		ProjectID:      projectID,
		Model:          model,
		Title:          title,
		Message:        message,
	}
}