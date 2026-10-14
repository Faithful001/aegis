package otp

import "errors"

var (
	ErrEmailAlreadyVerified = errors.New("email is already verified")
	ErrInvalidOTP           = errors.New("invalid or expired verification code")
	ErrCurrentPasswordWrong = errors.New("current password is incorrect")
)
