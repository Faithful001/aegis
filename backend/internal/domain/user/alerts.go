package user

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AlertSettings struct {
	ID                                 uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID                             uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex" json:"user_id"`
	IsBudgetAlertEnabled               bool       `gorm:"default:true" json:"is_budget_alert_enabled"`
	MonthlyBudgetUSD                   float64    `gorm:"type:decimal(10,2);default:50.00" json:"monthly_budget_usd"`
	ThresholdPercentage                int        `gorm:"default:80" json:"threshold_percentage"`
	AlertEmail                         string     `gorm:"type:varchar(255)" json:"alert_email"`
	IsRateLimitAlertEnabled            bool       `gorm:"default:true" json:"is_rate_limit_alert_enabled"`
	IsFailedRequestAlertEnabled        bool       `gorm:"default:false" json:"is_failed_request_alert_enabled"`
	IsModifiedCredentialAlertEnabled   bool       `gorm:"default:false" json:"is_modified_credential_alert_enabled"`
	IsWeeklyUsageAlertEnabled          bool       `gorm:"default:false" json:"is_weekly_usage_alert_enabled"`
	LastAlertSentAt                    *time.Time `json:"last_alert_sent_at,omitempty"`
	CreatedAt                          time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt                          time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

func (AlertSettings) TableName() string {
	return "alert_settings"
}

func (s *AlertSettings) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

type UpdateAlertSettingsDTO struct {
	IsBudgetAlertEnabled             *bool    `json:"is_budget_alert_enabled"`
	MonthlyBudgetUSD                 *float64 `json:"monthly_budget_usd"`
	ThresholdPercentage              *int     `json:"threshold_percentage"`
	IsRateLimitAlertEnabled          *bool    `json:"is_rate_limit_alert_enabled"`
	IsFailedRequestAlertEnabled      *bool    `json:"is_failed_request_alert_enabled"`
	IsModifiedCredentialAlertEnabled *bool    `json:"is_modified_credential_alert_enabled"`
	IsWeeklyUsageAlertEnabled        *bool    `json:"is_weekly_usage_alert_enabled"`
}

type IAlertSettingsRepository interface {
	GetByUserID(ctx context.Context, userID uuid.UUID) (*AlertSettings, error)
	Upsert(ctx context.Context, settings *AlertSettings) error
}

type GormAlertSettingsRepository struct {
	db *gorm.DB
}

func NewAlertSettingsRepository(db *gorm.DB) IAlertSettingsRepository {
	return &GormAlertSettingsRepository{db: db}
}

func (r *GormAlertSettingsRepository) GetByUserID(
	ctx context.Context,
	userID uuid.UUID,
) (*AlertSettings, error) {
	if r.db == nil {
		return nil, nil
	}

	var settings AlertSettings

	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		First(&settings).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return &settings, nil
}

func (r *GormAlertSettingsRepository) Upsert(
	ctx context.Context,
	settings *AlertSettings,
) error {
	if r.db == nil {
		return nil
	}

	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "user_id"},
			},
			DoUpdates: clause.AssignmentColumns([]string{
				"is_budget_alert_enabled",
				"monthly_budget_usd",
				"threshold_percentage",
				"alert_email",
				"is_rate_limit_alert_enabled",
				"is_failed_request_alert_enabled",
				"is_modified_credential_alert_enabled",
				"is_weekly_usage_alert_enabled",
				"last_alert_sent_at",
				"updated_at",
			}),
		}).
		Create(settings).Error
}