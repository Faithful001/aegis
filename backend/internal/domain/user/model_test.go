package user_test

import (
	"testing"

	"github.com/Faithful001/aegis/internal/domain/user"
)

func TestNewUser_Success(t *testing.T) {
	u, err := user.NewUser("Alice", "Smith", "Alice@example.com", "hashed_secret_pw")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if u.Email != "alice@example.com" {
		t.Errorf("expected lowercase email, got %s", u.Email)
	}

	if u.FullName() != "Alice Smith" {
		t.Errorf("expected full name 'Alice Smith', got %s", u.FullName())
	}

	if !u.IsActive() {
		t.Errorf("expected user to be active by default")
	}
}

func TestNewUser_ValidationFailure(t *testing.T) {
	_, err := user.NewUser("", "Smith", "alice@example.com", "pw")
	if err == nil {
		t.Errorf("expected error for empty first name, got nil")
	}

	_, err = user.NewUser("Alice", "Smith", "", "pw")
	if err == nil {
		t.Errorf("expected error for empty email, got nil")
	}
}
