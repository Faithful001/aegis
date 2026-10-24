package notification

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type NotificationType string

const (
	TypeBudgetAlert        NotificationType = "budget_alert"
	TypeRateLimit          NotificationType = "rate_limit"
	TypeFailedRequest      NotificationType = "failed_request"
	TypeModifiedCredential NotificationType = "modified_credential"
	TypeWeeklyUsage        NotificationType = "weekly_usage"
)

type Notification struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;index:idx_notifications_user" json:"user_id"`
	Type      string    `gorm:"type:varchar(50);not null" json:"type"`
	Title     string    `gorm:"type:text;not null" json:"title"`
	Message   string    `gorm:"type:text;not null" json:"message"`
	IsRead    bool      `gorm:"not null;default:false" json:"is_read"`
	CreatedAt time.Time `gorm:"not null;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null;autoUpdateTime" json:"updated_at"`
}

func (n *Notification) TableName() string {
	return "notifications"
}

func (n *Notification) BeforeCreate(tx *gorm.DB) error {
	if n.ID == uuid.Nil {
		n.ID = uuid.New()
	}
	return nil
}