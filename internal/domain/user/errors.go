package user

import "errors"

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailAlreadyExists = errors.New("user with this email already exists")
	ErrInvalidUserData    = errors.New("invalid user data")
	ErrUserInactive       = errors.New("user account is inactive or disabled")
)
