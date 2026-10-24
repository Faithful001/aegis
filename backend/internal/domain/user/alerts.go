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

// ThresholdUSD calculates the exact dollar threshold before an alert is triggered.
// e.g. MonthlyBudgetUSD = $50.00, ThresholdPercentage = 80 -> $40.00
func (s *AlertSettings) ThresholdUSD() float64 {
	pct := float64(s.ThresholdPercentage)
	if pct <= 0 {
		pct = 80.0
	}
	return s.MonthlyBudgetUSD * (pct / 100.0)
}

// IsBudgetExceeded checks if the current monthly spend has breached the threshold.
func (s *AlertSettings) IsBudgetExceeded(currentSpendUSD float64) bool {
	if !s.IsBudgetAlertEnabled || s.MonthlyBudgetUSD <= 0 {
		return false
	}
	return currentSpendUSD >= s.ThresholdUSD()
}

// PercentOfBudgetUsed calculates the percentage of monthly budget consumed so far.
func (s *AlertSettings) PercentOfBudgetUsed(currentSpendUSD float64) int {
	if s.MonthlyBudgetUSD <= 0 {
		return 0
	}
	return int((currentSpendUSD / s.MonthlyBudgetUSD) * 100.0)
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
	ListBudgetAlertEnabled(ctx context.Context) ([]AlertSettings, error)
	ListWeeklyUsageAlertEnabled(ctx context.Context) ([]AlertSettings, error)
	UpdateLastAlertSentAt(ctx context.Context, userID uuid.UUID, t time.Time) error
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

func (r *GormAlertSettingsRepository) ListBudgetAlertEnabled(ctx context.Context) ([]AlertSettings, error) {
	if r.db == nil {
		return nil, nil
	}
	var results []AlertSettings
	err := r.db.WithContext(ctx).
		Where("is_budget_alert_enabled = ?", true).
		Find(&results).Error
	return results, err
}

func (r *GormAlertSettingsRepository) ListWeeklyUsageAlertEnabled(ctx context.Context) ([]AlertSettings, error) {
	if r.db == nil {
		return nil, nil
	}
	var results []AlertSettings
	err := r.db.WithContext(ctx).
		Where("is_weekly_usage_alert_enabled = ?", true).
		Find(&results).Error
	return results, err
}

func (r *GormAlertSettingsRepository) UpdateLastAlertSentAt(ctx context.Context, userID uuid.UUID, t time.Time) error {
	if r.db == nil {
		return nil
	}
	return r.db.WithContext(ctx).Model(&AlertSettings{}).
		Where("user_id = ?", userID).
		Update("last_alert_sent_at", t).Error
}