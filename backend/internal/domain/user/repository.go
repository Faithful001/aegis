package user

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IUserRepository interface {
	Create(ctx context.Context, u *User) error
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByGoogleID(ctx context.Context, googleID string) (*User, error)
	GetByGithubID(ctx context.Context, githubID string) (*User, error)
	FindOrCreateOAuthUser(ctx context.Context, newUser *User) (*User, bool, error)
	Update(ctx context.Context, u *User) error
	UpdatePassword(ctx context.Context, id uuid.UUID, hashedPassword string) error
	MarkVerified(ctx context.Context, id uuid.UUID) error
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

func (r *UserRepository) GetByGoogleID(ctx context.Context, googleID string) (*User, error) {
	if r.db == nil {
		return nil, errors.New("database connection is nil")
	}
	if googleID == "" {
		return nil, ErrUserNotFound
	}
	var u User
	if err := r.db.WithContext(ctx).Where("google_id = ?", googleID).First(&u).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) GetByGithubID(ctx context.Context, githubID string) (*User, error) {
	if r.db == nil {
		return nil, errors.New("database connection is nil")
	}
	if githubID == "" {
		return nil, ErrUserNotFound
	}
	var u User
	if err := r.db.WithContext(ctx).Where("github_id = ?", githubID).First(&u).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) FindOrCreateOAuthUser(ctx context.Context, newUser *User) (*User, bool, error) {
	if r.db == nil {
		return nil, false, errors.New("database connection is nil")
	}

	var result *User
	created := false

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing User

		err := tx.
			Where("LOWER(email) = ?", strings.ToLower(newUser.Email)).
			First(&existing).Error

		if err == nil {
			result = &existing
			return nil
		}

		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		if err := tx.Create(newUser).Error; err != nil {
			return err
		}

		result = newUser
		created = true

		return nil
	})

	if err != nil {
		return nil, false, err
	}

	return result, created, nil
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

func (r *UserRepository) UpdatePassword(ctx context.Context, id uuid.UUID, hashedPassword string) error {
	if r.db == nil {
		return errors.New("database connection is nil")
	}
	return r.db.WithContext(ctx).Model(&User{}).Where("id = ?", id).Update("password", hashedPassword).Error
}

func (r *UserRepository) MarkVerified(ctx context.Context, id uuid.UUID) error {
	if r.db == nil {
		return errors.New("database connection is nil")
	}
	return r.db.WithContext(ctx).Model(&User{}).Where("id = ?", id).Update("is_verified", true).Error
}
