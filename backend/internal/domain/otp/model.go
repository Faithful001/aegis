package otp

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OTPPurpose string

const (
	OTPPurposeEmailVerification OTPPurpose = "email_verification"
	OTPPurposePasswordReset     OTPPurpose = "password_reset"
)

// OTPRecord stores one-time passwords for email verification and password resets.
type OTPRecord struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	UserID    uuid.UUID      `gorm:"type:uuid;index;not null" json:"user_id"`
	Email     string         `gorm:"type:varchar(255);index;not null" json:"email"`
	Code      string         `gorm:"type:varchar(10);not null" json:"code"`
	Purpose   OTPPurpose     `gorm:"type:varchar(50);not null" json:"purpose"`
	ExpiresAt time.Time      `gorm:"not null" json:"expires_at"`
	UsedAt    *time.Time     `gorm:"default:null" json:"used_at"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (OTPRecord) TableName() string {
	return "otp_records"
}

func (r *OTPRecord) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}
