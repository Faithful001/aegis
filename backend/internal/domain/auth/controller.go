package auth

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/Faithful001/aegis/internal/domain/auth/dto"
	"github.com/Faithful001/aegis/internal/domain/user"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuthController struct {
	authService *AuthService
}

func NewAuthController(authService *AuthService) *AuthController {
	return &AuthController{authService: authService}
}

func (h *AuthController) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error(), "data": nil})
		return
	}

	_, err := h.authService.Register(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, user.ErrEmailAlreadyExists) {
			c.JSON(http.StatusConflict, gin.H{"success": false, "message": err.Error(), "data": nil})
			return
		}
		if errors.Is(err, user.ErrInvalidUserData) {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error(), "data": nil})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error(), "data": nil})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "user registered successfully",
		"data":    nil,
	})
}

func (h *AuthController) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error(), "data": nil})
		return
	}

	result, err := h.authService.Login(c.Request.Context(), req, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) || errors.Is(err, user.ErrUserNotFound) {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "invalid email or password", "data": nil})
			return
		}
		if errors.Is(err, user.ErrUserInactive) {
			c.JSON(http.StatusForbidden, gin.H{"success": false, "message": err.Error(), "data": nil})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error(), "data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "login successful",
		"data":    result,
	})
}

func (h *AuthController) SendVerificationEmail(c *gin.Context) {
	var req dto.SendVerificationEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	if err := h.authService.SendVerificationEmail(c.Request.Context(), req.Email); err != nil {
		switch {
		case errors.Is(err, ErrEmailAlreadyVerified):
			c.JSON(http.StatusConflict, gin.H{"success": false, "message": err.Error()})
		case errors.Is(err, user.ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{"success": false, "message": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "failed to send verification email"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "verification email sent"})
}

func (h *AuthController) VerifyEmail(c *gin.Context) {
	var req dto.VerifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	result, err := h.authService.VerifyEmail(c.Request.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, ErrEmailAlreadyVerified):
			c.JSON(http.StatusConflict, gin.H{"success": false, "message": err.Error(), "data": nil})
		case errors.Is(err, ErrInvalidOTP):
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid or expired verification code", "data": nil})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "failed to verify email", "data": nil})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "email verified successfully",
		"data":    result,
	})
}

func (h *AuthController) ForgotPassword(c *gin.Context) {
	var req dto.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	// Always return success to prevent email enumeration
	_ = h.authService.ForgotPassword(c.Request.Context(), req)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "if an account with that email exists, a reset code has been sent",
	})
}

func (h *AuthController) VerifyResetOTP(c *gin.Context) {
	var req dto.VerifyResetOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	resp, err := h.authService.VerifyResetOTP(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, ErrInvalidOTP) {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid or expired reset code"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *AuthController) ResetPassword(c *gin.Context) {
	var req dto.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	token := ""
	authHeader := c.GetHeader("Authorization")
	if len(authHeader) > 7 && strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
		token = strings.TrimSpace(authHeader[7:])
	}

	if err := h.authService.ResetPassword(c.Request.Context(), token, req); err != nil {
		switch {
		case errors.Is(err, ErrInvalidOTP):
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid or expired reset code"})
		case errors.Is(err, ErrTokenExpired):
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "reset password token has expired"})
		case errors.Is(err, ErrInvalidTokenType):
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "invalid token type: must be reset_password token"})
		case errors.Is(err, ErrInvalidToken):
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "invalid reset password token"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "failed to reset password"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "password reset successfully"})
}

func (h *AuthController) ChangePassword(c *gin.Context) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "unauthorized"})
		return
	}
	userID := userIDVal.(uuid.UUID)

	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	if err := h.authService.ChangePassword(c.Request.Context(), userID, req); err != nil {
		switch {
		case errors.Is(err, ErrCurrentPasswordWrong):
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "current password is incorrect"})
		case errors.Is(err, user.ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{"success": false, "message": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "failed to change password"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "password changed successfully"})
}


func (h *AuthController) RefreshToken(c *gin.Context) {
	var req dto.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error(), "data": nil})
		return
	}

	result, err := h.authService.RefreshToken(c.Request.Context(), req, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		if errors.Is(err, ErrTokenRevoked) || errors.Is(err, ErrTokenExpired) || errors.Is(err, ErrInvalidToken) {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": err.Error(), "data": nil})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error(), "data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "token refreshed successfully",
		"data":    result,
	})
}

func (h *AuthController) Logout(c *gin.Context) {
	var req dto.LogoutRequest
	_ = c.ShouldBindJSON(&req)

	_ = h.authService.Logout(c.Request.Context(), req)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "logged out successfully",
	})
}

func (h *AuthController) ListSessions(c *gin.Context) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "unauthorized"})
		return
	}
	userID := userIDVal.(uuid.UUID)

	currentJTI := ""
	if claimsVal, exists := c.Get("userClaims"); exists && claimsVal != nil {
		if claims, ok := claimsVal.(*TokenClaims); ok {
			currentJTI = claims.JTI
		}
	}

	sessions, err := h.authService.ListSessions(c.Request.Context(), userID, currentJTI, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": sessions})
}

func (h *AuthController) RevokeOtherSessions(c *gin.Context) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "unauthorized"})
		return
	}
	userID := userIDVal.(uuid.UUID)

	currentJTI := ""
	if claimsVal, exists := c.Get("userClaims"); exists && claimsVal != nil {
		if claims, ok := claimsVal.(*TokenClaims); ok {
			currentJTI = claims.JTI
		}
	}

	if err := h.authService.RevokeOtherSessions(c.Request.Context(), userID, currentJTI); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "All other sessions have been signed out."})
}

func (h *AuthController) RevokeSession(c *gin.Context) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "unauthorized"})
		return
	}
	userID := userIDVal.(uuid.UUID)

	sessionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid session id"})
		return
	}

	if err := h.authService.RevokeSession(c.Request.Context(), userID, sessionID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Session revoked successfully."})
}

type APIKeyController struct {
	apiKeyService *APIKeyService
}

func NewAPIKeyController(apiKeyService *APIKeyService) *APIKeyController {
	return &APIKeyController{apiKeyService: apiKeyService}
}

func (h *APIKeyController) Create(c *gin.Context) {
	log.Printf("Creating API key")
	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "unauthorized", "data": nil})
		return
	}
	userID := userIDVal.(uuid.UUID)

	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid project id", "data": nil})
		return
	}

	var req dto.CreateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error(), "data": nil})
		return
	}

	keyResult, err := h.apiKeyService.CreateAPIKey(c.Request.Context(), projectID, userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error(), "data": nil})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "API key created successfully. Save your secret key now as it cannot be shown again.",
		"data":    keyResult,
	})
}

func (h *APIKeyController) List(c *gin.Context) {
	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid project id", "data": nil})
		return
	}

	keys, err := h.apiKeyService.ListProjectAPIKeys(c.Request.Context(), projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error(), "data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": keys})
}

func (h *APIKeyController) Revoke(c *gin.Context) {
	keyID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid api key id", "data": nil})
		return
	}

	if err := h.apiKeyService.RevokeAPIKey(c.Request.Context(), keyID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error(), "data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "API key revoked successfully"})
}
