package provider

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrCredentialNotFound    = errors.New("provider credential not found")
	ErrInvalidProvider      = errors.New("invalid provider specified (supported: openai, anthropic, gemini, mistral)")
	ErrInvalidCredentialData = errors.New("invalid provider credential data")
)

const (
	ProviderOpenAI    	= "openai"
	ProviderAnthropic 	= "anthropic"
	ProviderGemini    	= "gemini"
	ProviderMistral   	= "mistral"
	ProviderOpenRouter	= "openrouter"
)

func IsSupportedProvider(p string) bool {
	switch strings.ToLower(strings.TrimSpace(p)) {
	case ProviderOpenAI, ProviderAnthropic, ProviderGemini, ProviderMistral, ProviderOpenRouter:
		return true
	default:
		return false
	}
}

type ProviderCredential struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrganizationID  uuid.UUID `gorm:"type:uuid;not null;index:idx_org_provider,unique" json:"organization_id"`
	Provider        string    `gorm:"type:varchar(50);not null;index:idx_org_provider,unique" json:"provider"`
	EncryptedAPIKey string    `gorm:"type:text;not null" json:"-"`
	BaseURL         string    `gorm:"type:varchar(255)" json:"base_url,omitempty"`
	CreatedAt       time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (ProviderCredential) TableName() string {
	return "provider_credentials"
}

func (c *ProviderCredential) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	c.Provider = strings.ToLower(strings.TrimSpace(c.Provider))
	return nil
}
