package middleware

import (
	"net/http"
	"strings"

	"github.com/Faithful001/aegis/internal/domain/auth"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	ContextKeyUserID    = "userID"
	ContextKeyUserEmail = "userEmail"
	ContextKeyClaims    = "userClaims"
)

func AuthMiddleware(authService *auth.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Authorization header is required",
			})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && strings.EqualFold(parts[0], "Bearer")) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Authorization format must be Bearer <token>",
			})
			c.Abort()
			return
		}

		tokenStr := parts[1]
		claims, err := authService.ValidateToken(tokenStr, auth.TokenTypeAccess)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": err.Error(),
			})
			c.Abort()
			return
		}

		c.Set(ContextKeyUserID, claims.UserID)
		c.Set(ContextKeyUserEmail, claims.Email)
		c.Set(ContextKeyClaims, claims)

		c.Next()
	}
}

func GetUserID(c *gin.Context) (uuid.UUID, bool) {
	val, exists := c.Get(ContextKeyUserID)
	if !exists {
		return uuid.Nil, false
	}
	id, ok := val.(uuid.UUID)
	return id, ok
}

func GetUserEmail(c *gin.Context) (string, bool) {
	val, exists := c.Get(ContextKeyUserEmail)
	if !exists {
		return "", false
	}
	email, ok := val.(string)
	return email, ok
}

func GetClaims(c *gin.Context) (*auth.JWTClaims, bool) {
	val, exists := c.Get(ContextKeyClaims)
	if !exists {
		return nil, false
	}
	claims, ok := val.(*auth.JWTClaims)
	return claims, ok
}
