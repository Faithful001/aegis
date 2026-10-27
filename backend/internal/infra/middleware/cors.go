package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

func CORSMiddleware() gin.HandlerFunc {
	allowedOrigin := os.Getenv("CLIENT_URL")
	if allowedOrigin == "" {
		allowedOrigin = "http://localhost:5173"
	}

	return func(ctx *gin.Context) {
		origin := ctx.GetHeader("Origin")

		if origin != "" && (origin == allowedOrigin || strings.HasPrefix(origin, "http://localhost:")) {
			ctx.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		} else if allowedOrigin != "" {
			ctx.Writer.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		}

		ctx.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		ctx.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		ctx.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")

		if ctx.Request.Method == http.MethodOptions {
			ctx.AbortWithStatus(http.StatusNoContent)
			return
		}

		ctx.Next()
	}
}