package project

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, p *Project) error
	GetByID(ctx context.Context, id uuid.UUID) (*Project, error)
	GetByOrgAndSlug(ctx context.Context, orgID uuid.UUID, slug string) (*Project, error)
	ListByOrganization(ctx context.Context, orgID uuid.UUID) ([]*Project, error)
	Update(ctx context.Context, p *Project) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type GormRepository struct {
	db *gorm.DB
}

func NewGormRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

func (r *GormRepository) Create(ctx context.Context, p *Project) error {
	if r.db == nil {
		return errors.New("database connection is nil")
	}
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *GormRepository) GetByID(ctx context.Context, id uuid.UUID) (*Project, error) {
	if r.db == nil {
		return nil, errors.New("database connection is nil")
	}
	var p Project
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&p).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, err
	}
	return &p, nil
}

func (r *GormRepository) GetByOrgAndSlug(ctx context.Context, orgID uuid.UUID, slug string) (*Project, error) {
	if r.db == nil {
		return nil, errors.New("database connection is nil")
	}
	var p Project
	if err := r.db.WithContext(ctx).Where("organization_id = ? AND slug = ?", orgID, slug).First(&p).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, err
	}
	return &p, nil
}

func (r *GormRepository) ListByOrganization(ctx context.Context, orgID uuid.UUID) ([]*Project, error) {
	if r.db == nil {
		return nil, errors.New("database connection is nil")
	}
	var projects []*Project
	err := r.db.WithContext(ctx).Where("organization_id = ?", orgID).Find(&projects).Error
	return projects, err
}

func (r *GormRepository) Update(ctx context.Context, p *Project) error {
	if r.db == nil {
		return errors.New("database connection is nil")
	}
	return r.db.WithContext(ctx).Save(p).Error
}

func (r *GormRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if r.db == nil {
		return errors.New("database connection is nil")
	}
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&Project{}).Error
}
