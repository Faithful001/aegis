package dto

import (
	"time"

	"github.com/google/uuid"
)

type SaveCredentialRequest struct {
	Provider string `json:"provider" binding:"required"`
	APIKey   string `json:"api_key" binding:"required"`
	BaseURL  string `json:"base_url,omitempty"`
}

type CredentialResponse struct {
	ID             uuid.UUID `json:"id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	Provider       string    `json:"provider"`
	MaskedAPIKey   string    `json:"masked_api_key"`
	BaseURL        string    `json:"base_url,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type CredentialListResponse struct {
	Data []CredentialResponse `json:"data"`
}
