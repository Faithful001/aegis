package user

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID			uuid.UUID	`gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	FirstName	string		`gorm:"type:varchar(100);not null" json:"first_name"`
	LastName	string		`gorm:"type:varchar(100);not null" json:"last_name"`
	Email		string		`gorm:"type:varchar(255);not null;uniqueIndex" json:"email"`
	Password    string     `gorm:"type:varchar(255);not null" json:"-"`
	IsVerified  bool       `gorm:"default:false" json:"is_verified"`
	CreatedAt  	time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt	time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

func (User) TableName() string {
	return "users"
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}

	return nil
}