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
	ErrInvalidProvider      = errors.New("invalid provider specified (supported: openai, anthropic, gemini, mistral, openrouter)")
	ErrInvalidCredentialData = errors.New("invalid provider credential data")
	ErrInvalidAPIKey         = errors.New("invalid API key")
	ErrInvalidBaseURL        = errors.New("invalid base URL")
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
	UserID          uuid.UUID `gorm:"type:uuid;not null;index:idx_user_provider,unique" json:"user_id"`
	Provider        string    `gorm:"type:varchar(50);not null;index:idx_user_provider,unique" json:"provider"`
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

type UserModel struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID      uuid.UUID `gorm:"type:uuid;not null;index:idx_user_provider_model,unique" json:"user_id"`
	Provider    string    `gorm:"type:varchar(50);not null;index:idx_user_provider_model,unique" json:"provider"`
	ModelID     string    `gorm:"type:varchar(255);not null;index:idx_user_provider_model,unique" json:"model_id"`
	Name        string    `gorm:"type:varchar(255);not null" json:"name"`
	Description string    `gorm:"type:text" json:"description,omitempty"`
	Fast        bool      `gorm:"default:false" json:"fast"`
	IsActive    bool      `gorm:"default:false;index" json:"is_active"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (UserModel) TableName() string {
	return "user_models"
}

func (m *UserModel) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	m.Provider = strings.ToLower(strings.TrimSpace(m.Provider))
	return nil
}

