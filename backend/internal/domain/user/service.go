package user

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Service struct {
	repo      IUserRepository
	alertRepo IAlertSettingsRepository
	db        *gorm.DB
}

func NewUserService(repo IUserRepository, alertRepo ...IAlertSettingsRepository) *Service {
	var ar IAlertSettingsRepository
	if len(alertRepo) > 0 {
		ar = alertRepo[0]
	}
	return &Service{
		repo:      repo,
		alertRepo: ar,
	}
}

func (s *Service) SetDB(db *gorm.DB) {
	s.db = db
}

func (s *Service) SetAlertRepo(alertRepo IAlertSettingsRepository) {
	s.alertRepo = alertRepo
}

func (s *Service) GetUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) UpdateProfile(ctx context.Context, userID uuid.UUID, fullName string) (*User, error) {
	u, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	fullName = strings.TrimSpace(fullName)
	if fullName == "" {
		return nil, ErrInvalidUserData
	}
	u.FullName = fullName
	u.UpdatedAt = time.Now().UTC()
	if err := s.repo.Update(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *Service) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	return s.repo.GetByEmail(ctx, email)
}

func (s *Service) GetAlertSettings(ctx context.Context, userID uuid.UUID, defaultEmail string) (*UserAlertSettings, error) {
	if s.alertRepo == nil {
		return &UserAlertSettings{
			UserID:              userID,
			BudgetAlertEnabled:  true,
			MonthlyBudgetUSD:    50.00,
			ThresholdPercentage: 80,
			AlertEmail:          defaultEmail,
		}, nil
	}

	settings, err := s.alertRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if settings == nil {
		settings = &UserAlertSettings{
			UserID:              userID,
			BudgetAlertEnabled:  true,
			MonthlyBudgetUSD:    50.00,
			ThresholdPercentage: 80,
			AlertEmail:          defaultEmail,
		}
		_ = s.alertRepo.Upsert(ctx, settings)
	}
	return settings, nil
}

func (s *Service) UpdateAlertSettings(ctx context.Context, userID uuid.UUID, dto UpdateAlertSettingsDTO) (*UserAlertSettings, error) {
	if s.alertRepo == nil {
		return nil, nil
	}
	settings, err := s.alertRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if settings == nil {
		settings = &UserAlertSettings{
			UserID:              userID,
			BudgetAlertEnabled:  true,
			MonthlyBudgetUSD:    50.00,
			ThresholdPercentage: 80,
		}
	}

	if dto.BudgetAlertEnabled != nil {
		settings.BudgetAlertEnabled = *dto.BudgetAlertEnabled
	}
	if dto.MonthlyBudgetUSD != nil {
		settings.MonthlyBudgetUSD = *dto.MonthlyBudgetUSD
	}
	if dto.ThresholdPercentage != nil {
		settings.ThresholdPercentage = *dto.ThresholdPercentage
	}
	if dto.AlertEmail != nil {
		settings.AlertEmail = *dto.AlertEmail
	}

	if err := s.alertRepo.Upsert(ctx, settings); err != nil {
		return nil, err
	}
	return settings, nil
}

func (s *Service) ExportUserData(ctx context.Context, userID uuid.UUID) (map[string]interface{}, error) {
	u, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"exported_at": time.Now().UTC().Format(time.RFC3339),
		"profile": map[string]interface{}{
			"id":          u.ID,
			"full_name":   u.FullName,
			"email":       u.Email,
			"status":      u.Status,
			"is_verified": u.IsVerified,
			"created_at":  u.CreatedAt,
			"updated_at":  u.UpdatedAt,
		},
	}

	if s.db != nil {
		var threads []struct {
			ID        uuid.UUID `json:"id"`
			Title     string    `json:"title"`
			CreatedAt time.Time `json:"created_at"`
		}
		_ = s.db.WithContext(ctx).Table("threads").Where("user_id = ?", userID).Find(&threads).Error
		result["threads"] = threads

		var usageStats struct {
			TotalTokens   int64   `json:"total_tokens"`
			TotalRequests int64   `json:"total_requests"`
			TotalCostUSD  float64 `json:"total_cost_usd"`
		}
		_ = s.db.WithContext(ctx).Table("usage_records").
			Select("COALESCE(SUM(total_tokens), 0) as total_tokens, COUNT(id) as total_requests, COALESCE(SUM(cost_usd), 0) as total_cost_usd").
			Where("user_id = ?", userID).Scan(&usageStats).Error
		result["usage_summary"] = usageStats

		var providers []string
		_ = s.db.WithContext(ctx).Table("provider_credentials").
			Where("user_id = ?", userID).Pluck("provider", &providers).Error
		result["configured_providers"] = providers
	}

	return result, nil
}

func (s *Service) DeleteUserAccount(ctx context.Context, userID uuid.UUID) error {
	if s.db != nil {
		_ = s.db.WithContext(ctx).Exec("DELETE FROM provider_credentials WHERE user_id = ?", userID).Error
		_ = s.db.WithContext(ctx).Exec("DELETE FROM user_sessions WHERE user_id = ?", userID).Error
		_ = s.db.WithContext(ctx).Exec("DELETE FROM user_alert_settings WHERE user_id = ?", userID).Error
	}
	return s.repo.Delete(ctx, userID)
}
