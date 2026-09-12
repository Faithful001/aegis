package auth

import (
	"time"

	"github.com/google/uuid"
)

type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int64     `json:"expires_in"`
	ExpiresAt    time.Time `json:"expires_at"`
}

type TokenService interface {
	GenerateTokenPair(userID uuid.UUID, email string) (*TokenPair, *TokenClaims, *TokenClaims, error)
	ValidateToken(tokenStr string, expectedType TokenType) (*TokenClaims, error)
}
