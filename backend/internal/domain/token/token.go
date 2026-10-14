// Package token defines the core token types shared between the domain and
// infrastructure layers. It intentionally has no dependencies on other internal
// packages so that both domain/auth and infra/security can import it without
// creating an import cycle.
package token

import (
	"time"

	"github.com/google/uuid"
)

// Type classifies what a JWT was issued for.
type Type string

const (
	TypeAccess        Type = "access"
	TypeRefresh       Type = "refresh"
	TypeResetPassword Type = "reset_password"
)

// Claims is the decoded, validated payload extracted from a JWT.
type Claims struct {
	JTI       string    `json:"jti"`
	UserID    uuid.UUID `json:"user_id"`
	Email     string    `json:"email"`
	TokenType Type      `json:"token_type"`
	ExpiresAt time.Time `json:"expires_at"`
	IssuedAt  time.Time `json:"issued_at"`
}

// Pair holds an access/refresh token pair returned after authentication.
type Pair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int64     `json:"expires_in"`
	ExpiresAt    time.Time `json:"expires_at"`
}
