package user

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserAlertSettings struct {
	ID                  uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID              uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex" json:"user_id"`
	BudgetAlertEnabled  bool       `gorm:"default:true" json:"budget_alert_enabled"`
	MonthlyBudgetUSD    float64    `gorm:"type:decimal(10,2);default:50.00" json:"monthly_budget_usd"`
	ThresholdPercentage int        `gorm:"default:80" json:"threshold_percentage"`
	AlertEmail          string     `gorm:"type:varchar(255)" json:"alert_email"`
	LastAlertSentAt     *time.Time `json:"last_alert_sent_at,omitempty"`
	CreatedAt           time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt           time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

func (UserAlertSettings) TableName() string {
	return "user_alert_settings"
}

func (s *UserAlertSettings) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

type UpdateAlertSettingsDTO struct {
	BudgetAlertEnabled  *bool    `json:"budget_alert_enabled"`
	MonthlyBudgetUSD    *float64 `json:"monthly_budget_usd"`
	ThresholdPercentage *int     `json:"threshold_percentage"`
	AlertEmail          *string  `json:"alert_email"`
}

type IAlertSettingsRepository interface {
	GetByUserID(ctx context.Context, userID uuid.UUID) (*UserAlertSettings, error)
	Upsert(ctx context.Context, settings *UserAlertSettings) error
}

type GormAlertSettingsRepository struct {
	db *gorm.DB
}

func NewAlertSettingsRepository(db *gorm.DB) IAlertSettingsRepository {
	return &GormAlertSettingsRepository{db: db}
}

func (r *GormAlertSettingsRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*UserAlertSettings, error) {
	if r.db == nil {
		return nil, nil
	}
	var s UserAlertSettings
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&s).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *GormAlertSettingsRepository) Upsert(ctx context.Context, settings *UserAlertSettings) error {
	if r.db == nil {
		return nil
	}
	return r.db.WithContext(ctx).Save(settings).Error
}
