package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/Faithful001/aegis/internal/domain/auth"
	authDto "github.com/Faithful001/aegis/internal/domain/auth/dto"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	ContextKeyPrincipal = "authPrincipal"
	ContextKeyOrgID     = "orgID"
	ContextKeyProjectID = "projectID"
	ContextKeyAPIKeyID  = "apiKeyID"
	HeaderXAPIKey       = "X-API-Key"
)

type contextKey string

const PrincipalContextKey contextKey = "principal"

func APIKeyAuthMiddleware(apiKeyService *auth.APIKeyService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var apiKey string

		// 1. Check Authorization header (e.g. "Bearer aeg_live_...")
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
				apiKey = strings.TrimSpace(parts[1])
			} else {
				apiKey = strings.TrimSpace(authHeader)
			}
		}

		// 2. Fallback to X-API-Key header if Authorization header was not provided
		if apiKey == "" {
			apiKey = strings.TrimSpace(c.GetHeader(HeaderXAPIKey))
		}

		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"message": "Missing API key. Pass via 'Authorization: Bearer <key>' or 'X-API-Key: <key>' header.",
					"type":    "invalid_request_error",
					"code":    "api_key_missing",
				},
			})
			c.Abort()
			return
		}

		// 3. Authenticate API key against database
		principal, err := apiKeyService.AuthenticateAPIKey(c.Request.Context(), apiKey)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"message": "Invalid or revoked API key.",
					"type":    "invalid_request_error",
					"code":    "invalid_api_key",
				},
			})
			c.Abort()
			return
		}

		// 4. Inject Principal into Gin Context and standard Request Context
		c.Set(ContextKeyPrincipal, principal)
		c.Set(ContextKeyOrgID, principal.OrganizationID)
		c.Set(ContextKeyProjectID, principal.ProjectID)
		c.Set(ContextKeyAPIKeyID, principal.APIKeyID)
		c.Set(ContextKeyUserID, principal.UserID)

		reqCtx := context.WithValue(c.Request.Context(), PrincipalContextKey, principal)
		c.Request = c.Request.WithContext(reqCtx)

		c.Next()
	}
}

func GetPrincipal(c *gin.Context) (*authDto.AuthenticatedPrincipal, bool) {
	val, exists := c.Get(ContextKeyPrincipal)
	if !exists {
		return nil, false
	}
	p, ok := val.(*authDto.AuthenticatedPrincipal)
	return p, ok
}

func GetOrgID(c *gin.Context) (uuid.UUID, bool) {
	val, exists := c.Get(ContextKeyOrgID)
	if !exists {
		return uuid.Nil, false
	}
	id, ok := val.(uuid.UUID)
	return id, ok
}

func GetProjectID(c *gin.Context) (uuid.UUID, bool) {
	val, exists := c.Get(ContextKeyProjectID)
	if !exists {
		return uuid.Nil, false
	}
	id, ok := val.(uuid.UUID)
	return id, ok
}
