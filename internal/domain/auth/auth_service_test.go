package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Faithful001/aegis/internal/domain/auth"
	authDto "github.com/Faithful001/aegis/internal/domain/auth/dto"
	"github.com/Faithful001/aegis/internal/domain/user"
	infraAuth "github.com/Faithful001/aegis/internal/infra/auth"
	"github.com/google/uuid"
)

type mockUserRepo struct {
	users map[string]*user.User
	byID  map[uuid.UUID]*user.User
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		users: make(map[string]*user.User),
		byID:  make(map[uuid.UUID]*user.User),
	}
}

func (m *mockUserRepo) Create(ctx context.Context, u *user.User) error {
	if _, exists := m.users[u.Email]; exists {
		return user.ErrEmailAlreadyExists
	}
	m.users[u.Email] = u
	m.byID[u.ID] = u
	return nil
}

func (m *mockUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	u, exists := m.byID[id]
	if !exists {
		return nil, user.ErrUserNotFound
	}
	return u, nil
}

func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	u, exists := m.users[email]
	if !exists {
		return nil, user.ErrUserNotFound
	}
	return u, nil
}

func (m *mockUserRepo) Update(ctx context.Context, u *user.User) error {
	m.users[u.Email] = u
	m.byID[u.ID] = u
	return nil
}

func (m *mockUserRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if u, exists := m.byID[id]; exists {
		delete(m.users, u.Email)
		delete(m.byID, id)
	}
	return nil
}

type mockBlacklistRepo struct {
	blacklisted map[string]bool
}

func newMockBlacklistRepo() *mockBlacklistRepo {
	return &mockBlacklistRepo{blacklisted: make(map[string]bool)}
}

func (m *mockBlacklistRepo) BlacklistToken(ctx context.Context, jti string, userID uuid.UUID, tokenType auth.TokenType, expiresAt time.Time, reason string) error {
	m.blacklisted[jti] = true
	return nil
}

func (m *mockBlacklistRepo) IsTokenBlacklisted(ctx context.Context, jti string) (bool, error) {
	return m.blacklisted[jti], nil
}

func TestAuthService_Register_And_Login(t *testing.T) {
	userRepo := newMockUserRepo()
	blacklistRepo := newMockBlacklistRepo()
	tokenService := infraAuth.NewJWTService("test_secret_key_32_bytes_long_jwt!", "aegis_test", 15*time.Minute, 7*24*time.Hour)
	hasher := infraAuth.NewBcryptHasher(4)

	svc := auth.NewAuthService(userRepo, blacklistRepo, tokenService, hasher)
	ctx := context.Background()

	// 1. Register
	regReq := authDto.RegisterRequest{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john.doe@example.com",
		Password:  "SecurePassword123!",
	}

	result, err := svc.Register(ctx, regReq)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	if result.AccessToken == "" || result.RefreshToken == "" {
		t.Errorf("expected access and refresh tokens, got empty strings")
	}

	// 2. Duplicate Registration
	_, err = svc.Register(ctx, regReq)
	if !errors.Is(err, user.ErrEmailAlreadyExists) {
		t.Errorf("expected ErrEmailAlreadyExists, got %v", err)
	}

	// 3. Login
	loginReq := authDto.LoginRequest{
		Email:    "john.doe@example.com",
		Password: "SecurePassword123!",
	}
	loginResult, err := svc.Login(ctx, loginReq)
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	// 4. Refresh Token
	refreshReq := authDto.RefreshRequest{
		RefreshToken: loginResult.RefreshToken,
	}
	refreshed, err := svc.RefreshToken(ctx, refreshReq)
	if err != nil {
		t.Fatalf("RefreshToken failed: %v", err)
	}
	if refreshed.AccessToken == "" {
		t.Errorf("expected new access token, got empty")
	}

	// 5. Old token revoked
	_, err = svc.RefreshToken(ctx, refreshReq)
	if !errors.Is(err, auth.ErrTokenRevoked) {
		t.Errorf("expected ErrTokenRevoked on reused refresh token, got %v", err)
	}
}
