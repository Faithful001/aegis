package user

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserStatus string

const (
	StatusActive    UserStatus = "active"
	StatusInactive  UserStatus = "inactive"
	StatusSuspended UserStatus = "suspended"
)

type User struct {
	ID         uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	FullName   string     `gorm:"type:varchar(255);not null;default:''" json:"full_name"`
	Email      string     `gorm:"type:varchar(255);not null;uniqueIndex:idx_users_email" json:"email"`
	Password   string     `gorm:"type:varchar(255);not null;default:''" json:"-"`
	GoogleID   string     `gorm:"type:varchar(255);not null;default:'';uniqueIndex:idx_users_google_id,where:google_id <> ''" json:"-"`
	GithubID   string     `gorm:"type:varchar(255);not null;default:'';uniqueIndex:idx_users_github_id,where:github_id <> ''" json:"-"`
	AvatarURL  string     `gorm:"type:text;not null;default:''" json:"avatar_url,omitempty"`
	Status     UserStatus `gorm:"type:varchar(50);not null;default:'active';check:chk_users_status,status IN ('active','inactive','suspended')" json:"status"`
	IsVerified bool       `gorm:"not null;default:false" json:"is_verified"`
	CreatedAt  time.Time  `gorm:"not null;autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time  `gorm:"not null;autoUpdateTime" json:"updated_at"`
}

func (User) TableName() string {
	return "users"
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	if u.Status == "" {
		u.Status = StatusActive
	}
	return nil
}

func NewUser(fullName, email, hashedPassword string) (*User, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	fullName = strings.TrimSpace(fullName)

	if email == "" || fullName == "" || hashedPassword == "" {
		return nil, ErrInvalidUserData
	}

	return &User{
		ID:         uuid.New(),
		FullName:   fullName,
		Email:      email,
		Password:   hashedPassword,
		Status:     StatusActive,
		IsVerified: false,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}, nil
}

func (u *User) IsActive() bool {
	return u.Status == StatusActive
}

func (u *User) HasPassword() bool {
	return strings.TrimSpace(u.Password) != ""
}

func NewOAuthUser(fullName, email, avatarURL string) *User {
	email = strings.ToLower(strings.TrimSpace(email))
	fullName = strings.TrimSpace(fullName)
	avatarURL = strings.TrimSpace(avatarURL)

	if fullName == "" {
		if parts := strings.SplitN(email, "@", 2); len(parts) == 2 && parts[0] != "" {
			fullName = parts[0]
		} else {
			fullName = "User"
		}
	}

	return &User{
		ID:         uuid.New(),
		FullName:   fullName,
		Email:      email,
		Password:   "",
		AvatarURL:  avatarURL,
		Status:     StatusActive,
		IsVerified: true,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}
}