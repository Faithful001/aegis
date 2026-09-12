package dto

import "github.com/Faithful001/aegis/internal/domain/user"

type AuthResponse struct {
	AccessToken  string     `json:"access_token"`
	RefreshToken string     `json:"refresh_token"`
	TokenType    string     `json:"token_type"`
	ExpiresIn    int64      `json:"expires_in"` // in seconds
	User         *user.User `json:"user"`
}