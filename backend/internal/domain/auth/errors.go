package auth

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrTokenRevoked        = errors.New("token has been revoked")
	ErrInvalidTokenType    = errors.New("invalid token type")
	ErrTokenExpired        = errors.New("token has expired")
	ErrInvalidToken        = errors.New("invalid token")
	ErrMissingAuthHeader   = errors.New("authorization header is required")
	ErrInvalidAuthFormat   = errors.New("authorization format must be Bearer <token>")
	ErrAPIKeyNotFound      = errors.New("API key not found")
	ErrAPIKeyInactive      = errors.New("API key is inactive or revoked")
	ErrAPIKeyExpired       = errors.New("API key has expired")
)
