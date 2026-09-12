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
	FirstName  string     `gorm:"type:varchar(100);not null" json:"first_name"`
	LastName   string     `gorm:"type:varchar(100);not null" json:"last_name"`
	Email      string     `gorm:"type:varchar(255);not null;uniqueIndex" json:"email"`
	Password   string     `gorm:"type:varchar(255);not null" json:"-"`
	Status     UserStatus `gorm:"type:varchar(50);not null;default:'active'" json:"status"`
	IsVerified bool       `gorm:"default:false" json:"is_verified"`
	CreatedAt  time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
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

func NewUser(firstName, lastName, email, hashedPassword string) (*User, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	firstName = strings.TrimSpace(firstName)
	lastName = strings.TrimSpace(lastName)

	if email == "" || firstName == "" || lastName == "" || hashedPassword == "" {
		return nil, ErrInvalidUserData
	}

	return &User{
		ID:         uuid.New(),
		FirstName:  firstName,
		LastName:   lastName,
		Email:      email,
		Password:   hashedPassword,
		Status:     StatusActive,
		IsVerified: false,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}, nil
}

func (u *User) FullName() string {
	return strings.TrimSpace(u.FirstName + " " + u.LastName)
}

func (u *User) IsActive() bool {
	return u.Status == StatusActive
}