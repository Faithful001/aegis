package auth

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BlacklistedToken struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	JTI       string    `gorm:"type:varchar(255);not null;index" json:"jti"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	TokenType string    `gorm:"type:varchar(50);not null" json:"token_type"` // "access" or "refresh"
	ExpiresAt time.Time `gorm:"not null;index" json:"expires_at"`
	Reason    string    `gorm:"type:varchar(100)" json:"reason"` // "logout", "refreshed", "revoked"
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (BlacklistedToken) TableName() string {
	return "blacklisted_tokens"
}

func (b *BlacklistedToken) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}
