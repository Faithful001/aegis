package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/Faithful001/aegis/internal/infra/ratelimiter"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	HeaderXRateLimitLimitRPM     = "X-RateLimit-Limit-RPM"
	HeaderXRateLimitRemainingRPM = "X-RateLimit-Remaining-RPM"
	HeaderXRateLimitResetRPM     = "X-RateLimit-Reset-RPM"
	HeaderRetryAfter             = "Retry-After"

	DefaultRPM = 600
	DefaultTPM = 100000
)

func RateLimitMiddleware(limiter ratelimiter.RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		if limiter == nil {
			c.Next()
			return
		}

		var userID uuid.UUID
		if val, exists := c.Get("userID"); exists {
			if id, ok := val.(uuid.UUID); ok {
				userID = id
			}
		}

		if userID == uuid.Nil {
			c.Next()
			return
		}

		ctx := c.Request.Context()
		rpmLimit := DefaultRPM

		key := fmt.Sprintf("ratelimit:user:%s:rpm", userID.String())
		res, err := limiter.CheckRateLimit(ctx, ratelimiter.RateLimitRequest{
			Tier:   ratelimiter.TierUser,
			Key:    key,
			Limit:  rpmLimit,
			Window: time.Minute,
			Cost:   1,
		})
		if err == nil && res != nil {
			c.Header(HeaderXRateLimitLimitRPM, strconv.Itoa(res.Limit))
			c.Header(HeaderXRateLimitRemainingRPM, strconv.Itoa(res.Remaining))
			c.Header(HeaderXRateLimitResetRPM, strconv.FormatInt(res.ResetTime.Unix(), 10))

			if !res.Allowed {
				retrySecs := int(res.RetryAfter.Seconds())
				if retrySecs < 1 {
					retrySecs = 1
				}
				c.Header(HeaderRetryAfter, strconv.Itoa(retrySecs))
				c.JSON(http.StatusTooManyRequests, gin.H{
					"error": gin.H{
						"message": fmt.Sprintf("Rate limit exceeded for User. Limit: %d RPM.", res.Limit),
						"type":    "rate_limit_error",
						"code":    "rate_limit_exceeded",
					},
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}
