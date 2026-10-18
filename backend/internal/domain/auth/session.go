package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserSession struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID       uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	AccessJTI    string    `gorm:"type:varchar(255);not null;index" json:"access_jti"`
	RefreshJTI   string    `gorm:"type:varchar(255);index" json:"refresh_jti"`
	IPAddress    string    `gorm:"type:varchar(100);index" json:"ip_address"`
	UserAgent    string    `gorm:"type:text" json:"user_agent"`
	DeviceName   string    `gorm:"type:varchar(100)" json:"device_name"`
	LastActiveAt time.Time `gorm:"not null" json:"last_active_at"`
	ExpiresAt    time.Time `gorm:"not null" json:"expires_at"`
	IsRevoked    bool      `gorm:"default:false;index" json:"is_revoked"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (UserSession) TableName() string {
	return "user_sessions"
}

func (s *UserSession) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

type SessionResponse struct {
	ID           uuid.UUID `json:"id"`
	DeviceName   string    `json:"device_name"`
	IPAddress    string    `json:"ip_address"`
	LastActiveAt time.Time `json:"last_active_at"`
	CreatedAt    time.Time `json:"created_at"`
	IsCurrent    bool      `json:"is_current"`
}

func ParseDeviceName(ua string) string {
	if strings.TrimSpace(ua) == "" {
		return "Web Client"
	}
	lower := strings.ToLower(ua)

	os := "Unknown Device"
	switch {
	case strings.Contains(lower, "windows"):
		os = "Windows"
	case strings.Contains(lower, "macintosh") || strings.Contains(lower, "mac os"):
		os = "macOS"
	case strings.Contains(lower, "iphone"):
		os = "iPhone"
	case strings.Contains(lower, "ipad"):
		os = "iPad"
	case strings.Contains(lower, "android"):
		os = "Android"
	case strings.Contains(lower, "linux"):
		os = "Linux"
	}

	browser := "Web Client"
	switch {
	case strings.Contains(lower, "edg/"):
		browser = "Edge"
	case strings.Contains(lower, "chrome/") && !strings.Contains(lower, "edg/"):
		browser = "Chrome"
	case strings.Contains(lower, "firefox/"):
		browser = "Firefox"
	case strings.Contains(lower, "safari/") && !strings.Contains(lower, "chrome/"):
		browser = "Safari"
	case strings.Contains(lower, "postman"):
		browser = "Postman"
	}

	return fmt.Sprintf("%s on %s", browser, os)
}

// SessionRepository interface
type ISessionRepository interface {
	Create(ctx context.Context, s *UserSession) error
	Update(ctx context.Context, s *UserSession) error
	GetByAccessJTI(ctx context.Context, jti string) (*UserSession, error)
	GetByUserIDAndIPAddress(
		ctx context.Context,
		userID uuid.UUID,
		ipAddress string,
	) (*UserSession, error)
	ListActiveByUserID(ctx context.Context, userID uuid.UUID) ([]*UserSession, error)
	RevokeDeviceSessions(ctx context.Context, userID uuid.UUID, ipAddress, userAgent string) error
	RevokeOthers(ctx context.Context, userID uuid.UUID, currentJTI string) ([]*UserSession, error)
	RevokeByID(ctx context.Context, userID uuid.UUID, sessionID uuid.UUID) (*UserSession, error)
	UpdateLastActive(ctx context.Context, jti string, lastActive time.Time) error
}

type GormSessionRepository struct {
	db *gorm.DB
}

func NewSessionRepository(db *gorm.DB) ISessionRepository {
	return &GormSessionRepository{db: db}
}

func (r *GormSessionRepository) Create(ctx context.Context, s *UserSession) error {
	if r.db == nil {
		return nil
	}
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *GormSessionRepository) Update(ctx context.Context, s *UserSession) error {
	if r.db == nil {
		return nil
	}
	return r.db.WithContext(ctx).
		Model(&UserSession{}).
		Where("id = ?", s.ID).
		Select("access_jti", "refresh_jti", "user_agent", "device_name", "last_active_at", "expires_at", "is_revoked").
		Updates(s).Error
}


func (r *GormSessionRepository) GetByUserIDAndIPAddress(
	ctx context.Context,
	userID uuid.UUID,
	ipAddress string,
) (*UserSession, error) {
	if r.db == nil {
		return nil, gorm.ErrRecordNotFound
	}
	var s UserSession
	err := r.db.WithContext(ctx).Where(
		"user_id = ? AND ip_address = ?",
		userID,
		ipAddress,
	).First(&s).Error

	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *GormSessionRepository) GetByAccessJTI(ctx context.Context, jti string) (*UserSession, error) {
	if r.db == nil {
		return nil, gorm.ErrRecordNotFound
	}
	var s UserSession
	err := r.db.WithContext(ctx).Where("access_jti = ? OR refresh_jti = ?", jti, jti).First(&s).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *GormSessionRepository) ListActiveByUserID(ctx context.Context, userID uuid.UUID) ([]*UserSession, error) {
	if r.db == nil {
		return nil, nil
	}
	var sessions []*UserSession
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND is_revoked = false AND expires_at > ?", userID, time.Now().UTC()).
		Order("last_active_at desc").
		Find(&sessions).Error
	return sessions, err
}

func (r *GormSessionRepository) RevokeDeviceSessions(ctx context.Context, userID uuid.UUID, ipAddress, userAgent string) error {
	if r.db == nil {
		return nil
	}
	return r.db.WithContext(ctx).
		Model(&UserSession{}).
		Where("user_id = ? AND ip_address = ? AND user_agent = ? AND is_revoked = false", userID, ipAddress, userAgent).
		Update("is_revoked", true).Error
}

func (r *GormSessionRepository) RevokeOthers(ctx context.Context, userID uuid.UUID, currentJTI string) ([]*UserSession, error) {
	if r.db == nil {
		return nil, nil
	}
	var sessionsToRevoke []*UserSession
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND is_revoked = false AND access_jti != ? AND refresh_jti != ?", userID, currentJTI, currentJTI).
		Find(&sessionsToRevoke).Error
	if err != nil {
		return nil, err
	}

	if len(sessionsToRevoke) > 0 {
		var ids []uuid.UUID
		for _, s := range sessionsToRevoke {
			ids = append(ids, s.ID)
		}
		_ = r.db.WithContext(ctx).Model(&UserSession{}).Where("id IN ?", ids).Update("is_revoked", true).Error
	}

	return sessionsToRevoke, nil
}

func (r *GormSessionRepository) RevokeByID(ctx context.Context, userID uuid.UUID, sessionID uuid.UUID) (*UserSession, error) {
	if r.db == nil {
		return nil, nil
	}
	var s UserSession
	err := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", sessionID, userID).First(&s).Error
	if err != nil {
		return nil, err
	}
	s.IsRevoked = true
	_ = r.db.WithContext(ctx).Save(&s).Error
	return &s, nil
}

func (r *GormSessionRepository) UpdateLastActive(ctx context.Context, jti string, lastActive time.Time) error {
	if r.db == nil {
		return nil
	}
	return r.db.WithContext(ctx).Model(&UserSession{}).
		Where("access_jti = ? OR refresh_jti = ?", jti, jti).
		Update("last_active_at", lastActive).Error
}
