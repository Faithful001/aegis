package dto

import (
	"time"

	"github.com/google/uuid"
)

type CreateThreadRequest struct {
	Title string `json:"title"`
	Model string `json:"model"`
}

type UpdateThreadRequest struct {
	Title *string `json:"title"`
	Model *string `json:"model"`
}

type CreateMessageRequest struct {
	Role    string `json:"role" binding:"required"`
	Content string `json:"content" binding:"required"`
	Model   string `json:"model,omitempty"`
}

type MessageResponse struct {
	ID        uuid.UUID `json:"id"`
	ThreadID  uuid.UUID `json:"thread_id"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Model     string    `json:"model,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type ThreadResponse struct {
	ID        uuid.UUID         `json:"id"`
	UserID    uuid.UUID         `json:"user_id"`
	Title     string            `json:"title"`
	Model     string            `json:"model"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
	Messages  []MessageResponse `json:"messages,omitempty"`
}
