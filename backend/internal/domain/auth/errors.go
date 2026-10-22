package auth

import "errors"

var (
	ErrInvalidCredentials   	= errors.New("invalid credentials")
	ErrNotFound             	= errors.New("not found")
	ErrUserNotFound         	= errors.New("user not found")
	ErrAccountAlreadyExists 	= errors.New("an account with this email already exists")
    ErrOAuthEmailNotVerified 	= errors.New("OAuth email is not verified")
	ErrUsePasswordLogin 		= errors.New("please sign in with password")
	ErrUseGoogleLogin 			= errors.New("please sign in with Google")
    ErrUseGitHubLogin 			= errors.New("please sign in with GitHub")
    ErrPasswordLoginUnavailable = errors.New("password login is unavailable for this account")
	ErrTokenRevoked         	= errors.New("token has been revoked")
	ErrInvalidTokenType     	= errors.New("invalid token type")
	ErrTokenExpired         	= errors.New("token has expired")
	ErrInvalidToken         	= errors.New("invalid token")
	ErrMissingAuthHeader   		= errors.New("authorization header is required")
	ErrInvalidAuthFormat   		= errors.New("authorization format must be Bearer <token>")
	ErrAPIKeyNotFound      		= errors.New("API key not found")
	ErrAPIKeyInactive      		= errors.New("API key is inactive or revoked")
	ErrAPIKeyExpired       		= errors.New("API key has expired")
	ErrInvalidOTP          		= errors.New("invalid or expired OTP code")
	ErrEmailAlreadyVerified 	= errors.New("email is already verified")
	ErrCurrentPasswordWrong 	= errors.New("current password is incorrect")
)
