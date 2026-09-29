package dto

import (
	"time"

	"github.com/google/uuid"
)

type CreateProjectRequest struct {
	Name           string `json:"name" binding:"required"`
	Slug           string `json:"slug"`
	MaxConcurrency int    `json:"max_concurrency"`
	RateLimitRPM   int    `json:"rate_limit_rpm"`
	RateLimitTPM   int    `json:"rate_limit_tpm"`
}

type ProjectResponse struct {
	ID             uuid.UUID `json:"id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	Name           string    `json:"name"`
	Slug           string    `json:"slug"`
	Status         string    `json:"status"`
	MaxConcurrency int       `json:"max_concurrency"`
	RateLimitRPM   int       `json:"rate_limit_rpm"`
	RateLimitTPM   int       `json:"rate_limit_tpm"`
	CreatedAt      time.Time `json:"created_at"`
}
