package provider

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type IProviderCredentialRepository interface {
	Save(ctx context.Context, cred *ProviderCredential) error
	GetByUserAndProvider(ctx context.Context, userID uuid.UUID, provider string) (*ProviderCredential, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]*ProviderCredential, error)
	Delete(ctx context.Context, userID uuid.UUID, provider string) error

	// UserModel methods
	UpsertUserModel(ctx context.Context, m *UserModel) error
	UpsertUserModelsBatch(ctx context.Context, models []*UserModel) error
	GetUserModel(ctx context.Context, userID uuid.UUID, provider, modelID string) (*UserModel, error)
	ListUserModels(ctx context.Context, userID uuid.UUID, provider string, activeOnly bool) ([]*UserModel, error)
	ListAllUserModels(ctx context.Context, userID uuid.UUID, activeOnly bool) ([]*UserModel, error)
	ToggleUserModel(ctx context.Context, userID uuid.UUID, provider, modelID string, isActive *bool) (*UserModel, error)
}

type ProviderCredentialRepository struct {
	db *gorm.DB
}

func NewProviderCredentialRepository(db *gorm.DB) IProviderCredentialRepository {
	return &ProviderCredentialRepository{db: db}
}

func (r *ProviderCredentialRepository) Save(ctx context.Context, cred *ProviderCredential) error {
	if r.db == nil {
		return errors.New("database connection is nil")
	}

	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "provider"}},
		DoUpdates: clause.AssignmentColumns([]string{"encrypted_api_key", "base_url", "updated_at"}),
	}).Create(cred).Error
}

func (r *ProviderCredentialRepository) GetByUserAndProvider(ctx context.Context, userID uuid.UUID, provider string) (*ProviderCredential, error) {
	if r.db == nil {
		return nil, errors.New("database connection is nil")
	}

	var cred ProviderCredential
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND provider = ?", userID, strings.ToLower(provider)).
		First(&cred).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCredentialNotFound
		}
		return nil, err
	}
	return &cred, nil
}

func (r *ProviderCredentialRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]*ProviderCredential, error) {
	if r.db == nil {
		return nil, errors.New("database connection is nil")
	}

	var creds []*ProviderCredential
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Find(&creds).Error
	return creds, err
}

func (r *ProviderCredentialRepository) Delete(ctx context.Context, userID uuid.UUID, provider string) error {
	if r.db == nil {
		return errors.New("database connection is nil")
	}

	res := r.db.WithContext(ctx).
		Where("user_id = ? AND provider = ?", userID, strings.ToLower(provider)).
		Delete(&ProviderCredential{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrCredentialNotFound
	}
	return nil
}

func (r *ProviderCredentialRepository) UpsertUserModel(ctx context.Context, m *UserModel) error {
	if r.db == nil {
		return errors.New("database connection is nil")
	}

	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "provider"}, {Name: "model_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"name", "description", "fast", "updated_at"}),
	}).Create(m).Error
}

func (r *ProviderCredentialRepository) UpsertUserModelsBatch(ctx context.Context, models []*UserModel) error {
	if r.db == nil {
		return errors.New("database connection is nil")
	}
	if len(models) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "provider"}, {Name: "model_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"name", "description", "fast", "updated_at"}),
	}).Create(&models).Error
}

func (r *ProviderCredentialRepository) GetUserModel(ctx context.Context, userID uuid.UUID, provider, modelID string) (*UserModel, error) {
	if r.db == nil {
		return nil, errors.New("database connection is nil")
	}

	var m UserModel
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND provider = ? AND model_id = ?", userID, strings.ToLower(provider), modelID).
		First(&m).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *ProviderCredentialRepository) ListUserModels(ctx context.Context, userID uuid.UUID, provider string, activeOnly bool) ([]*UserModel, error) {
	if r.db == nil {
		return nil, errors.New("database connection is nil")
	}

	var list []*UserModel
	query := r.db.WithContext(ctx).Where("user_id = ?", userID)
	if provider != "" {
		query = query.Where("provider = ?", strings.ToLower(provider))
	}
	if activeOnly {
		query = query.Where("is_active = ?", true)
	}
	err := query.Order("name ASC").Find(&list).Error
	return list, err
}

func (r *ProviderCredentialRepository) ListAllUserModels(ctx context.Context, userID uuid.UUID, activeOnly bool) ([]*UserModel, error) {
	return r.ListUserModels(ctx, userID, "", activeOnly)
}

func (r *ProviderCredentialRepository) ToggleUserModel(ctx context.Context, userID uuid.UUID, provider, modelID string, isActive *bool) (*UserModel, error) {
	if r.db == nil {
		return nil, errors.New("database connection is nil")
	}

	p := strings.ToLower(strings.TrimSpace(provider))
	mID := strings.TrimSpace(modelID)

	var existing UserModel
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND provider = ? AND model_id = ?", userID, p, mID).
		First(&existing).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			newActive := true
			if isActive != nil {
				newActive = *isActive
			}
			newModel := &UserModel{
				ID:        uuid.New(),
				UserID:    userID,
				Provider:  p,
				ModelID:   mID,
				Name:      mID,
				IsActive:  newActive,
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			}
			if err := r.db.WithContext(ctx).Create(newModel).Error; err != nil {
				return nil, err
			}
			return newModel, nil
		}
		return nil, err
	}

	if isActive != nil {
		existing.IsActive = *isActive
	} else {
		existing.IsActive = !existing.IsActive
	}
	existing.UpdatedAt = time.Now().UTC()

	if err := r.db.WithContext(ctx).Save(&existing).Error; err != nil {
		return nil, err
	}
	return &existing, nil
}
