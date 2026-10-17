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
	ID           uuid.UUID `json:"id"`
	UserID       uuid.UUID `json:"user_id"`
	Provider     string    `json:"provider"`
	MaskedAPIKey string    `json:"masked_api_key"`
	BaseURL      string    `json:"base_url,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type CredentialListResponse struct {
	Data []CredentialResponse `json:"data"`
}

type ModelDescriptor struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Provider    string `json:"provider"`
	Description string `json:"description,omitempty"`
	Fast        bool   `json:"fast,omitempty"`
	IsActive    bool   `json:"is_active"`
}

type ToggleModelRequest struct {
	Provider string `json:"provider" binding:"required"`
	ModelID  string `json:"model_id" binding:"required"`
	IsActive *bool  `json:"is_active,omitempty"`
}

type ModelListResponse struct {
	Data []ModelDescriptor `json:"data"`
}

type PaginatedModelListResponse struct {
	Items      []ModelDescriptor `json:"items"`
	Total      int               `json:"total"`
	Page       int               `json:"page"`
	Limit      int               `json:"limit"`
	TotalPages int               `json:"total_pages"`
}

type ProviderDescriptor struct {
	ID          	string 	`json:"id"`
	Name        	string 	`json:"name"`
	Description 	string 	`json:"desc"`
	Placeholder 	string 	`json:"placeholder"`
	KeyLink     	string 	`json:"key_link"`
	IsConfigured	bool 	`json:"is_configured"`
	MaskedAPIKey	*string `json:"masked_api_key"`
}
