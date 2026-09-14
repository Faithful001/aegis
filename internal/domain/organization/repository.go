package organization

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IOrganizationRepository interface {
	Create(ctx context.Context, org *Organization, ownerID uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*Organization, error)
	GetBySlug(ctx context.Context, slug string) (*Organization, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]*Organization, error)
	Update(ctx context.Context, org *Organization) error

	AddMember(ctx context.Context, member *OrganizationMember) error
	GetMember(ctx context.Context, orgID, userID uuid.UUID) (*OrganizationMember, error)
	ListMembers(ctx context.Context, orgID uuid.UUID) ([]*OrganizationMember, error)
	RemoveMember(ctx context.Context, orgID, userID uuid.UUID) error
}

type OrganizationRepository struct {
	db *gorm.DB
}

func NewOrganizationRepository(db *gorm.DB) IOrganizationRepository {
	return &OrganizationRepository{db: db}
}

func (r *OrganizationRepository) Create(ctx context.Context, org *Organization, ownerID uuid.UUID) error {
	if r.db == nil {
		return errors.New("database connection is nil")
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(org).Error; err != nil {
			return err
		}

		member := &OrganizationMember{
			ID:             uuid.New(),
			OrganizationID: org.ID,
			UserID:         ownerID,
			Role:           RoleOwner,
		}

		return tx.Create(member).Error
	})
}

func (r *OrganizationRepository) GetByID(ctx context.Context, id uuid.UUID) (*Organization, error) {
	if r.db == nil {
		return nil, errors.New("database connection is nil")
	}
	var org Organization
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&org).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrgNotFound
		}
		return nil, err
	}
	return &org, nil
}

func (r *OrganizationRepository) GetBySlug(ctx context.Context, slug string) (*Organization, error) {
	if r.db == nil {
		return nil, errors.New("database connection is nil")
	}
	var org Organization
	if err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&org).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrgNotFound
		}
		return nil, err
	}
	return &org, nil
}

func (r *OrganizationRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]*Organization, error) {
	if r.db == nil {
		return nil, errors.New("database connection is nil")
	}
	var orgs []*Organization
	err := r.db.WithContext(ctx).
		Joins("JOIN organization_members ON organization_members.organization_id = organizations.id").
		Where("organization_members.user_id = ?", userID).
		Find(&orgs).Error
	return orgs, err
}

func (r *OrganizationRepository) Update(ctx context.Context, org *Organization) error {
	if r.db == nil {
		return errors.New("database connection is nil")
	}
	return r.db.WithContext(ctx).Save(org).Error
}

func (r *OrganizationRepository) AddMember(ctx context.Context, member *OrganizationMember) error {
	if r.db == nil {
		return errors.New("database connection is nil")
	}
	return r.db.WithContext(ctx).Create(member).Error
}

func (r *OrganizationRepository) GetMember(ctx context.Context, orgID, userID uuid.UUID) (*OrganizationMember, error) {
	if r.db == nil {
		return nil, errors.New("database connection is nil")
	}
	var member OrganizationMember
	if err := r.db.WithContext(ctx).Where("organization_id = ? AND user_id = ?", orgID, userID).First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrMemberNotFound
		}
		return nil, err
	}
	return &member, nil
}

func (r *OrganizationRepository) ListMembers(ctx context.Context, orgID uuid.UUID) ([]*OrganizationMember, error) {
	if r.db == nil {
		return nil, errors.New("database connection is nil")
	}
	var members []*OrganizationMember
	err := r.db.WithContext(ctx).Where("organization_id = ?", orgID).Find(&members).Error
	return members, err
}

func (r *OrganizationRepository) RemoveMember(ctx context.Context, orgID, userID uuid.UUID) error {
	if r.db == nil {
		return errors.New("database connection is nil")
	}
	return r.db.WithContext(ctx).Where("organization_id = ? AND user_id = ?", orgID, userID).Delete(&OrganizationMember{}).Error
}
