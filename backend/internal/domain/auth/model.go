package auth

import (
	"time"

	"github.com/Faithful001/aegis/internal/domain/token"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TokenType and its constants are type aliases over domain/token so that existing
// code inside this package (service.go, repository.go, etc.) keeps compiling
// unchanged. External callers should prefer the canonical token.Type.
type TokenType = token.Type

const (
	TokenTypeAccess        TokenType = token.TypeAccess
	TokenTypeRefresh       TokenType = token.TypeRefresh
	TokenTypeResetPassword TokenType = token.TypeResetPassword
)

// TokenClaims is an alias for token.Claims, kept here so domain/auth callers
type TokenClaims = token.Claims

// TokenPair is an alias for token.Pair.
type TokenPair = token.Pair

// IJWTService is the domain-owned interface for token operations.
type IJWTService interface {
	GenerateTokenPair(userID uuid.UUID, email string) (*TokenPair, *TokenClaims, *TokenClaims, error)
	GenerateResetPasswordToken(userID uuid.UUID, email string) (string, error)
	ValidateToken(tokenStr string, expectedType TokenType) (*TokenClaims, error)
}

// IBcryptPasswordHasher is the domain-owned interface for password hashing.
type IBcryptPasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hashedPassword, plainPassword string) error
}

type BlacklistedToken struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	JTI       string    `gorm:"type:varchar(255);not null;index" json:"jti"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	TokenType string    `gorm:"type:varchar(50);not null" json:"token_type"` // "access" or "refresh"
	ExpiresAt time.Time `gorm:"not null;index" json:"expires_at"`
	Reason    string    `gorm:"type:varchar(100)" json:"reason"` // "logout", "refreshed", "revoked"
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (BlacklistedToken) TableName() string {
	return "blacklisted_tokens"
}

func (b *BlacklistedToken) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}
