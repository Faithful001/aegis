package auth

import (
	"errors"
	"fmt"
	"time"

	domainAuth "github.com/Faithful001/aegis/internal/domain/auth"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTClaims struct {
	UserID    uuid.UUID `json:"user_id"`
	Email     string    `json:"email"`
	TokenType string    `json:"token_type"`
	jwt.RegisteredClaims
}

type JWTService struct {
	secret             []byte
	issuer             string
	accessTokenExpiry  time.Duration
	refreshTokenExpiry time.Duration
}

func NewJWTService(secret string, issuer string, accessExpiry, refreshExpiry time.Duration) *JWTService {
	return &JWTService{
		secret:             []byte(secret),
		issuer:             issuer,
		accessTokenExpiry:  accessExpiry,
		refreshTokenExpiry: refreshExpiry,
	}
}

func (s *JWTService) GenerateTokenPair(userID uuid.UUID, email string) (*domainAuth.TokenPair, *domainAuth.TokenClaims, *domainAuth.TokenClaims, error) {
	now := time.Now().UTC()

	// 1. Access Token
	accessJTI := uuid.New().String()
	accessExpiresAt := now.Add(s.accessTokenExpiry)
	accessClaims := JWTClaims{
		UserID:    userID,
		Email:     email,
		TokenType: string(domainAuth.TokenTypeAccess),
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        accessJTI,
			Subject:   userID.String(),
			Issuer:    s.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(accessExpiresAt),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString(s.secret)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	// 2. Refresh Token
	refreshJTI := uuid.New().String()
	refreshExpiresAt := now.Add(s.refreshTokenExpiry)
	refreshClaims := JWTClaims{
		UserID:    userID,
		Email:     email,
		TokenType: string(domainAuth.TokenTypeRefresh),
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        refreshJTI,
			Subject:   userID.String(),
			Issuer:    s.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(refreshExpiresAt),
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString(s.secret)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	domainAccessClaims := &domainAuth.TokenClaims{
		JTI:       accessJTI,
		UserID:    userID,
		Email:     email,
		TokenType: domainAuth.TokenTypeAccess,
		ExpiresAt: accessExpiresAt,
		IssuedAt:  now,
	}

	domainRefreshClaims := &domainAuth.TokenClaims{
		JTI:       refreshJTI,
		UserID:    userID,
		Email:     email,
		TokenType: domainAuth.TokenTypeRefresh,
		ExpiresAt: refreshExpiresAt,
		IssuedAt:  now,
	}

	pair := &domainAuth.TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.accessTokenExpiry.Seconds()),
		ExpiresAt:    accessExpiresAt,
	}

	return pair, domainAccessClaims, domainRefreshClaims, nil
}

func (s *JWTService) ValidateToken(tokenStr string, expectedType domainAuth.TokenType) (*domainAuth.TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &JWTClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.secret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, domainAuth.ErrTokenExpired
		}
		return nil, fmt.Errorf("%w: %v", domainAuth.ErrInvalidToken, err)
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, domainAuth.ErrInvalidToken
	}

	if expectedType != "" && claims.TokenType != string(expectedType) {
		return nil, domainAuth.ErrInvalidTokenType
	}

	var exp time.Time
	if claims.ExpiresAt != nil {
		exp = claims.ExpiresAt.Time
	}
	var iat time.Time
	if claims.IssuedAt != nil {
		iat = claims.IssuedAt.Time
	}

	return &domainAuth.TokenClaims{
		JTI:       claims.ID,
		UserID:    claims.UserID,
		Email:     claims.Email,
		TokenType: domainAuth.TokenType(claims.TokenType),
		ExpiresAt: exp,
		IssuedAt:  iat,
	}, nil
}
