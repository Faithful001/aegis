package upload

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrFileNotFound = errors.New("file not found")
)

type UploadRepository interface {
	Create(ctx context.Context, file *UploadedFile) error
	GetByID(ctx context.Context, id uuid.UUID) (*UploadedFile, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]UploadedFile, error)
	Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
}

type uploadRepository struct {
	db *gorm.DB
}

func NewUploadRepository(db *gorm.DB) UploadRepository {
	return &uploadRepository{db: db}
}

func (r *uploadRepository) Create(ctx context.Context, file *UploadedFile) error {
	if r.db == nil {
		return errors.New("database not initialized")
	}
	return r.db.WithContext(ctx).Create(file).Error
}

func (r *uploadRepository) GetByID(ctx context.Context, id uuid.UUID) (*UploadedFile, error) {
	if r.db == nil {
		return nil, errors.New("database not initialized")
	}
	var file UploadedFile
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&file).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrFileNotFound
		}
		return nil, err
	}
	return &file, nil
}

func (r *uploadRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]UploadedFile, error) {
	if r.db == nil {
		return nil, errors.New("database not initialized")
	}
	var files []UploadedFile
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at desc").Find(&files).Error
	return files, err
}

func (r *uploadRepository) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	if r.db == nil {
		return errors.New("database not initialized")
	}
	result := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).Delete(&UploadedFile{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrFileNotFound
	}
	return nil
}

