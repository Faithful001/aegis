package upload

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UploadedFile struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index:idx_upload_user" json:"user_id"`
	Filename  string    `gorm:"type:varchar(255);not null" json:"filename"`
	MimeType  string    `gorm:"type:varchar(100);not null" json:"mime_type"`
	Size      int64     `gorm:"type:bigint;not null" json:"size"`
	FilePath  string    `gorm:"type:varchar(500);not null" json:"file_path"`
	FileType  string    `gorm:"type:varchar(20);not null" json:"file_type"` // "image" or "document"
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (UploadedFile) TableName() string {
	return "uploaded_files"
}

func (u *UploadedFile) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

