package security

import (
	"errors"
	"fmt"
	"time"

	"github.com/Faithful001/aegis/internal/domain/token"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrTokenExpired     = errors.New("token has expired")
	ErrInvalidToken     = errors.New("invalid token")
	ErrInvalidTokenType = errors.New("invalid token type")
)

type IJWTService interface {
	GenerateTokenPair(userID uuid.UUID, email string) (*token.Pair, *token.Claims, *token.Claims, error)
	GenerateResetPasswordToken(userID uuid.UUID, email string) (string, error)
	ValidateToken(tokenStr string, expectedType token.Type) (*token.Claims, error)
}


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

func (s *JWTService) GenerateTokenPair(userID uuid.UUID, email string) (*token.Pair, *token.Claims, *token.Claims, error) {
	now := time.Now().UTC()

	// 1. Access Token
	accessJTI := uuid.New().String()
	accessExpiresAt := now.Add(s.accessTokenExpiry)
	accessClaims := JWTClaims{
		UserID:    userID,
		Email:     email,
		TokenType: string(token.TypeAccess),
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
		TokenType: string(token.TypeRefresh),
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

	domainAccessClaims := &token.Claims{
		JTI:       accessJTI,
		UserID:    userID,
		Email:     email,
		TokenType: token.TypeAccess,
		ExpiresAt: accessExpiresAt,
		IssuedAt:  now,
	}

	domainRefreshClaims := &token.Claims{
		JTI:       refreshJTI,
		UserID:    userID,
		Email:     email,
		TokenType: token.TypeRefresh,
		ExpiresAt: refreshExpiresAt,
		IssuedAt:  now,
	}

	pair := &token.Pair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.accessTokenExpiry.Seconds()),
		ExpiresAt:    accessExpiresAt,
	}

	return pair, domainAccessClaims, domainRefreshClaims, nil
}

func (s *JWTService) GenerateResetPasswordToken(userID uuid.UUID, email string) (string, error) {
	now := time.Now().UTC()
	expiresAt := now.Add(15 * time.Minute) // 15 minutes validity for password reset
	jti := uuid.New().String()

	claims := JWTClaims{
		UserID:    userID,
		Email:     email,
		TokenType: string(token.TypeResetPassword),
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			Subject:   userID.String(),
			Issuer:    s.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(s.secret)
}

func (s *JWTService) ValidateToken(tokenStr string, expectedType token.Type) (*token.Claims, error) {
	t, err := jwt.ParseWithClaims(tokenStr, &JWTClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.secret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	claims, ok := t.Claims.(*JWTClaims)
	if !ok || !t.Valid {
		return nil, ErrInvalidToken
	}

	if expectedType != "" && claims.TokenType != string(expectedType) {
		return nil, ErrInvalidTokenType
	}

	var exp time.Time
	if claims.ExpiresAt != nil {
		exp = claims.ExpiresAt.Time
	}
	var iat time.Time
	if claims.IssuedAt != nil {
		iat = claims.IssuedAt.Time
	}

	return &token.Claims{
		JTI:       claims.ID,
		UserID:    claims.UserID,
		Email:     claims.Email,
		TokenType: token.Type(claims.TokenType),
		ExpiresAt: exp,
		IssuedAt:  iat,
	}, nil
}
