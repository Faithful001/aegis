package otp

import (
	"crypto/rand"
	"fmt"
	"time"

	"github.com/Faithful001/aegis/internal/domain/user"
)

const (
	OTPLength = 6
	OTPExpiry = 10 * time.Minute
)

type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hashedPassword, plainPassword string) error
}
type IOTPService interface {
	GenerateOTPCode() (string, error)
}

type OTPService struct {
	userRepo     user.IUserRepository
	otpRepo      IOTPRepository
	hasher       PasswordHasher
}

func NewOTPService(
	userRepo user.IUserRepository,
	otpRepo IOTPRepository,
	hasher PasswordHasher,
) *OTPService {
	return &OTPService{
		userRepo:     userRepo,
		otpRepo:      otpRepo,
		hasher:       hasher,
	}
}

// generateOTPCode generates a cryptographically secure 6-digit numeric OTP.
func (s *OTPService) GenerateOTPCode() (string, error) {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	num := (uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])) % 1_000_000
	return fmt.Sprintf("%06d", num), nil
}