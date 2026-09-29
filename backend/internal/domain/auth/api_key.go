package auth

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type APIKeyStatus string

const (
	APIKeyStatusActive  APIKeyStatus = "active"
	APIKeyStatusRevoked APIKeyStatus = "revoked"
	APIKeyStatusExpired APIKeyStatus = "expired"
)

type APIKey struct {
	ID             uuid.UUID    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrganizationID uuid.UUID    `gorm:"type:uuid;not null;index" json:"organization_id"`
	ProjectID      uuid.UUID    `gorm:"type:uuid;not null;index" json:"project_id"`
	UserID         uuid.UUID    `gorm:"type:uuid;not null;index" json:"user_id"`
	Name           string       `gorm:"type:varchar(100);not null" json:"name"`
	KeyPrefix      string       `gorm:"type:varchar(20);not null;index" json:"key_prefix"` // e.g. "aeg_live_a1b2..."
	KeyHash        string       `gorm:"type:varchar(64);not null;uniqueIndex" json:"-"`     // SHA-256 hash (64 hex chars)
	Status         APIKeyStatus `gorm:"type:varchar(50);not null;default:'active'" json:"status"`
	ExpiresAt      *time.Time   `gorm:"index" json:"expires_at,omitempty"`
	LastUsedAt     *time.Time   `json:"last_used_at,omitempty"`
	CreatedAt      time.Time    `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time    `gorm:"autoUpdateTime" json:"updated_at"`
}

func (APIKey) TableName() string {
	return "api_keys"
}

func (k *APIKey) BeforeCreate(tx *gorm.DB) error {
	if k.ID == uuid.Nil {
		k.ID = uuid.New()
	}
	if k.Status == "" {
		k.Status = APIKeyStatusActive
	}
	return nil
}

func (k *APIKey) IsActive() bool {
	if k.Status != APIKeyStatusActive {
		return false
	}
	if k.ExpiresAt != nil && k.ExpiresAt.Before(time.Now().UTC()) {
		return false
	}
	return true
}

func (k *APIKey) Revoke() {
	k.Status = APIKeyStatusRevoked
	k.UpdatedAt = time.Now().UTC()
}
