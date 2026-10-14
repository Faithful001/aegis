package otp

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// OTPRepository handles creation, lookup, and invalidation of OTP records.
type IOTPRepository interface {
	Create(ctx context.Context, record *OTPRecord) error
	FindValid(ctx context.Context, email string, code string, purpose OTPPurpose) (*OTPRecord, error)
	MarkUsed(ctx context.Context, id uuid.UUID) error
	DeleteExpiredForUser(ctx context.Context, userID uuid.UUID, purpose OTPPurpose) error
}

type GormOTPRepository struct {
	db *gorm.DB
}

func NewOTPRepository(db *gorm.DB) IOTPRepository {
	return &GormOTPRepository{db: db}
}

func (r *GormOTPRepository) Create(ctx context.Context, record *OTPRecord) error {
	if r.db == nil {
		return errors.New("database connection is nil")
	}
	return r.db.WithContext(ctx).Create(record).Error
}

func (r *GormOTPRepository) FindValid(ctx context.Context, email, code string, purpose OTPPurpose) (*OTPRecord, error) {
	if r.db == nil {
		return nil, errors.New("database connection is nil")
	}
	var rec OTPRecord
	err := r.db.WithContext(ctx).
		Where("email = ? AND code = ? AND purpose = ? AND used_at IS NULL AND expires_at > ?",
			email, code, string(purpose), time.Now().UTC()).
		Order("created_at DESC").
		First(&rec).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidOTP
		}
		return nil, err
	}
	return &rec, nil
}

func (r *GormOTPRepository) MarkUsed(ctx context.Context, id uuid.UUID) error {
	if r.db == nil {
		return errors.New("database connection is nil")
	}
	now := time.Now().UTC()
	return r.db.WithContext(ctx).Model(&OTPRecord{}).Where("id = ?", id).Update("used_at", &now).Error
}

func (r *GormOTPRepository) DeleteExpiredForUser(ctx context.Context, userID uuid.UUID, purpose OTPPurpose) error {
	if r.db == nil {
		return errors.New("database connection is nil")
	}
	return r.db.WithContext(ctx).
		Where("user_id = ? AND purpose = ?", userID, string(purpose)).
		Delete(&OTPRecord{}).Error
}
