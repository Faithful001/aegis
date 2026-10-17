package thread

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Thread struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index:idx_thread_user" json:"user_id"`
	Title     string    `gorm:"type:varchar(255);not null" json:"title"`
	Model     string    `gorm:"type:varchar(100)" json:"model"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	Messages  []Message `gorm:"foreignKey:ThreadID;constraint:OnDelete:CASCADE" json:"messages,omitempty"`
}

func (Thread) TableName() string {
	return "chat_threads"
}

func (t *Thread) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	if t.Title == "" {
		t.Title = "New Exploration"
	}
	return nil
}

type Message struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ThreadID  uuid.UUID `gorm:"type:uuid;not null;index:idx_message_thread" json:"thread_id"`
	Role      string    `gorm:"type:varchar(20);not null" json:"role"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	Model     string    `gorm:"type:varchar(100)" json:"model,omitempty"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (Message) TableName() string {
	return "chat_messages"
}

func (m *Message) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}
