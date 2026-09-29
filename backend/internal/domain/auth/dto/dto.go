package dto

import (
	"time"

	"github.com/Faithful001/aegis/internal/domain/user"
	"github.com/google/uuid"
)

type RegisterRequest struct {
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type LogoutRequest struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type AuthResponse struct {
	AccessToken  string     `json:"access_token"`
	RefreshToken string     `json:"refresh_token"`
	TokenType    string     `json:"token_type"`
	ExpiresIn    int64      `json:"expires_in"`
	User         *user.User `json:"user"`
}

type CreateAPIKeyRequest struct {
	Name      string     `json:"name" binding:"required"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

type CreatedAPIKeyResponse struct {
	ID        uuid.UUID  `json:"id"`
	Name      string     `json:"name"`
	SecretKey string     `json:"secret_key"` // Returned only ONCE
	KeyPrefix string     `json:"key_prefix"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

type APIKeyResponse struct {
	ID             uuid.UUID  `json:"id"`
	OrganizationID uuid.UUID  `json:"organization_id"`
	ProjectID      uuid.UUID  `json:"project_id"`
	UserID         uuid.UUID  `json:"user_id"`
	Name           string     `json:"name"`
	KeyPrefix      string     `json:"key_prefix"`
	Status         string     `json:"status"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
	LastUsedAt     *time.Time `json:"last_used_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

type AuthenticatedPrincipal struct {
	APIKeyID       uuid.UUID
	OrganizationID uuid.UUID
	ProjectID      uuid.UUID
	UserID         uuid.UUID
	KeyPrefix      string
}
