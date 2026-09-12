package auth

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/Faithful001/aegis/internal/domain/auth/dto"
	"github.com/Faithful001/aegis/internal/domain/user"
	"github.com/Faithful001/aegis/internal/infra/db"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"

	DefaultAccessTokenExpiry  = 15 * time.Minute
	DefaultRefreshTokenExpiry = 7 * 24 * time.Hour
)

type JWTClaims struct {
	UserID    uuid.UUID `json:"user_id"`
	Email     string    `json:"email"`
	TokenType string    `json:"token_type"`
	jwt.RegisteredClaims
}

type AuthService struct{}

func NewAuthService() *AuthService {
	return &AuthService{}
}

func (s *AuthService) Register(payload dto.RegisterRequest) (*dto.AuthResponse, error) {
	var existingUser user.User
	if err := db.DB.Where("email = ?", payload.Email).First(&existingUser).Error; err == nil {
		return nil, errors.New("user with this email already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	newUser := user.User{
		FirstName:  payload.FirstName,
		LastName:   payload.LastName,
		Email:      payload.Email,
		Password:   string(hashedPassword),
		IsVerified: false,
	}

	if err := db.DB.Create(&newUser).Error; err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return s.GenerateTokenPair(&newUser)
}

func (s *AuthService) Login(payload dto.LoginRequest) (*dto.AuthResponse, error) {
	var foundUser user.User
	if err := db.DB.Where("email = ?", payload.Email).First(&foundUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid credentials")
		}
		return nil, fmt.Errorf("database query error: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(foundUser.Password), []byte(payload.Password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	return s.GenerateTokenPair(&foundUser)
}

func (s *AuthService) RefreshToken(payload dto.RefreshRequest) (*dto.AuthResponse, error) {
	claims, err := s.ValidateToken(payload.RefreshToken, TokenTypeRefresh)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// Blacklist current refresh token (rotation mechanism)
	if err := s.BlacklistToken(claims, "refreshed"); err != nil {
		return nil, fmt.Errorf("failed to rotate refresh token: %w", err)
	}

	var foundUser user.User
	if err := db.DB.Where("id = ?", claims.UserID).First(&foundUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("database query error: %w", err)
	}

	return s.GenerateTokenPair(&foundUser)
}

func (s *AuthService) Logout(accessTokenStr string, refreshTokenStr string) error {
	if accessTokenStr != "" {
		claims, err := s.parseTokenClaims(accessTokenStr)
		if err == nil && claims != nil {
			_ = s.BlacklistToken(claims, "logout")
		}
	}

	if refreshTokenStr != "" {
		claims, err := s.parseTokenClaims(refreshTokenStr)
		if err == nil && claims != nil {
			_ = s.BlacklistToken(claims, "logout")
		}
	}

	return nil
}

func (s *AuthService) GenerateTokenPair(u *user.User) (*dto.AuthResponse, error) {
	jwtSecret := s.getJWTSecret()
	accessExpiry := s.getAccessTokenExpiry()
	refreshExpiry := s.getRefreshTokenExpiry()

	now := time.Now()

	// 1. Access Token
	accessJTI := uuid.New().String()
	accessClaims := JWTClaims{
		UserID:    u.ID,
		Email:     u.Email,
		TokenType: TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        accessJTI,
			Subject:   u.ID.String(),
			Issuer:    "aegis",
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(accessExpiry)),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString(jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	// 2. Refresh Token
	refreshJTI := uuid.New().String()
	refreshClaims := JWTClaims{
		UserID:    u.ID,
		Email:     u.Email,
		TokenType: TokenTypeRefresh,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        refreshJTI,
			Subject:   u.ID.String(),
			Issuer:    "aegis",
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(refreshExpiry)),
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString(jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return &dto.AuthResponse{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		TokenType:    "Bearer",
		ExpiresIn:    int64(accessExpiry.Seconds()),
		User:         u,
	}, nil
}

func (s *AuthService) ValidateToken(tokenStr string, expectedType string) (*JWTClaims, error) {
	claims, err := s.parseTokenClaims(tokenStr)
	if err != nil {
		return nil, err
	}

	if expectedType != "" && claims.TokenType != expectedType {
		return nil, fmt.Errorf("invalid token type: expected %s, got %s", expectedType, claims.TokenType)
	}

	// Check if token JTI is blacklisted
	isBlacklisted, err := s.IsTokenBlacklisted(claims.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to check token blacklist status: %w", err)
	}
	if isBlacklisted {
		return nil, errors.New("token has been revoked")
	}

	return claims, nil
}

func (s *AuthService) IsTokenBlacklisted(jti string) (bool, error) {
	if jti == "" {
		return false, nil
	}

	var blacklisted BlacklistedToken
	err := db.DB.Where("jti = ? AND expires_at > ?", jti, time.Now()).First(&blacklisted).Error
	if err == nil {
		return true, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	return false, err
}

func (s *AuthService) BlacklistToken(claims *JWTClaims, reason string) error {
	if claims == nil || claims.ID == "" {
		return nil
	}

	var expiresAt time.Time
	if claims.ExpiresAt != nil {
		expiresAt = claims.ExpiresAt.Time
	} else {
		expiresAt = time.Now().Add(DefaultRefreshTokenExpiry)
	}

	// Only store if the token hasn't already expired
	if expiresAt.Before(time.Now()) {
		return nil
	}

	entry := BlacklistedToken{
		JTI:       claims.ID,
		UserID:    claims.UserID,
		TokenType: claims.TokenType,
		ExpiresAt: expiresAt,
		Reason:    reason,
	}

	if err := db.DB.Create(&entry).Error; err != nil {
		return fmt.Errorf("failed to blacklist token: %w", err)
	}

	return nil
}

func (s *AuthService) parseTokenClaims(tokenStr string) (*JWTClaims, error) {
	jwtSecret := s.getJWTSecret()

	token, err := jwt.ParseWithClaims(tokenStr, &JWTClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}

func (s *AuthService) getJWTSecret() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "aegis_default_jwt_secret_key_change_in_production"
	}
	return []byte(secret)
}

func (s *AuthService) getAccessTokenExpiry() time.Duration {
	if val := os.Getenv("ACCESS_TOKEN_EXPIRY"); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			return d
		}
	}
	return DefaultAccessTokenExpiry
}

func (s *AuthService) getRefreshTokenExpiry() time.Duration {
	if val := os.Getenv("REFRESH_TOKEN_EXPIRY"); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			return d
		}
	}
	return DefaultRefreshTokenExpiry
}