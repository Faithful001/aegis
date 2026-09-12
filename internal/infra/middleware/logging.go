package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/Faithful001/aegis/internal/infra/observability"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const HeaderXRequestID = "X-Request-ID"

func RequestLoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		reqID := c.GetHeader(HeaderXRequestID)
		if reqID == "" {
			reqID = fmt.Sprintf("req_%s", uuid.New().String())
		}
		c.Header(HeaderXRequestID, reqID)

		ctx := context.WithValue(c.Request.Context(), observability.RequestIDKey, reqID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()

		duration := time.Since(start)
		status := c.Writer.Status()

		logger := observability.WithContext(c.Request.Context())
		logLevel := slog.LevelInfo
		if status >= http.StatusInternalServerError {
			logLevel = slog.LevelError
		} else if status >= http.StatusBadRequest {
			logLevel = slog.LevelWarn
		}

		logger.Log(
			c.Request.Context(),
			logLevel,
			"HTTP request processed",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", status,
			"duration_ms", duration.Milliseconds(),
			"client_ip", c.ClientIP(),
		)
	}
}
