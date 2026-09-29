package user

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IUserRepository interface {
	Create(ctx context.Context, u *User) error
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, u *User) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) IUserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, u *User) error {
	if r.db == nil {
		return errors.New("database connection is nil")
	}
	return r.db.WithContext(ctx).Create(u).Error
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	if r.db == nil {
		return nil, errors.New("database connection is nil")
	}
	var u User
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&u).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	if r.db == nil {
		return nil, errors.New("database connection is nil")
	}
	var u User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&u).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) Update(ctx context.Context, u *User) error {
	if r.db == nil {
		return errors.New("database connection is nil")
	}
	return r.db.WithContext(ctx).Save(u).Error
}

func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if r.db == nil {
		return errors.New("database connection is nil")
	}
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&User{}).Error
}
