package auth_test

import (
	"testing"
	"time"

	domainAuth "github.com/Faithful001/aegis/internal/domain/auth"
	infraAuth "github.com/Faithful001/aegis/internal/infra/auth"
	"github.com/google/uuid"
)

func TestJWTService_Generate_And_Validate(t *testing.T) {
	jwtService := infraAuth.NewJWTService("secret_key_for_testing_purposes_only!", "aegis_test", 1*time.Minute, 5*time.Minute)

	userID := uuid.New()
	email := "dev@example.com"

	pair, accessClaims, refreshClaims, err := jwtService.GenerateTokenPair(userID, email)
	if err != nil {
		t.Fatalf("failed to generate token pair: %v", err)
	}

	if pair.AccessToken == "" || pair.RefreshToken == "" {
		t.Fatalf("expected non-empty tokens in pair")
	}

	if accessClaims.TokenType != domainAuth.TokenTypeAccess || refreshClaims.TokenType != domainAuth.TokenTypeRefresh {
		t.Fatalf("incorrect token types in claims")
	}

	// Validate Access Token
	validatedAccess, err := jwtService.ValidateToken(pair.AccessToken, domainAuth.TokenTypeAccess)
	if err != nil {
		t.Fatalf("failed to validate access token: %v", err)
	}
	if validatedAccess.UserID != userID || validatedAccess.Email != email {
		t.Fatalf("mismatched user claims: got %v, expected %v", validatedAccess.UserID, userID)
	}

	// Validate with wrong expected type should fail
	_, err = jwtService.ValidateToken(pair.AccessToken, domainAuth.TokenTypeRefresh)
	if err == nil {
		t.Fatalf("expected error validating access token as refresh token, got nil")
	}
}

func TestBcryptHasher(t *testing.T) {
	hasher := infraAuth.NewBcryptHasher(4)
	password := "SecurePassword123"

	hash, err := hasher.Hash(password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	if err := hasher.Compare(hash, password); err != nil {
		t.Fatalf("expected password to match hash: %v", err)
	}

	if err := hasher.Compare(hash, "wrong_password"); err == nil {
		t.Fatalf("expected wrong password to fail comparison")
	}
}
