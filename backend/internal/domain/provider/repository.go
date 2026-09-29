package provider

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type IProviderCredentialRepository interface {
	Save(ctx context.Context, cred *ProviderCredential) error
	GetByUserAndProvider(ctx context.Context, userID uuid.UUID, provider string) (*ProviderCredential, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]*ProviderCredential, error)
	Delete(ctx context.Context, userID uuid.UUID, provider string) error
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
