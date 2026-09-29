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

		principal, hasPrincipal := GetPrincipal(c)

		var apiKeyID, projectID, orgID uuid.UUID
		rpmLimit := DefaultRPM

		if hasPrincipal && principal != nil {
			apiKeyID = principal.APIKeyID
			projectID = principal.ProjectID
			orgID = principal.OrganizationID
		} else {
			if id, ok := GetOrgID(c); ok {
				orgID = id
			}
			if id, ok := GetProjectID(c); ok {
				projectID = id
			}
		}

		ctx := c.Request.Context()

		// 1. Check API Key Rate Limit (if present)
		if apiKeyID != uuid.Nil {
			key := fmt.Sprintf("ratelimit:key:%s:rpm", apiKeyID.String())
			res, err := limiter.CheckRateLimit(ctx, ratelimiter.RateLimitRequest{
				Tier:   ratelimiter.TierAPIKey,
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
							"message": fmt.Sprintf("Rate limit exceeded for API Key. Limit: %d RPM.", res.Limit),
							"type":    "rate_limit_error",
							"code":    "rate_limit_exceeded",
						},
					})
					c.Abort()
					return
				}
			}
		}

		// 2. Check Project Rate Limit (RPM)
		if projectID != uuid.Nil {
			key := fmt.Sprintf("ratelimit:project:%s:rpm", projectID.String())
			res, err := limiter.CheckRateLimit(ctx, ratelimiter.RateLimitRequest{
				Tier:   ratelimiter.TierProject,
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
							"message": fmt.Sprintf("Rate limit exceeded for Project. Limit: %d RPM.", res.Limit),
							"type":    "rate_limit_error",
							"code":    "rate_limit_exceeded",
						},
					})
					c.Abort()
					return
				}
			}
		}

		// 3. Check Organization Rate Limit (RPM fallback)
		if orgID != uuid.Nil && projectID == uuid.Nil && apiKeyID == uuid.Nil {
			key := fmt.Sprintf("ratelimit:org:%s:rpm", orgID.String())
			res, err := limiter.CheckRateLimit(ctx, ratelimiter.RateLimitRequest{
				Tier:   ratelimiter.TierOrganization,
				Key:    key,
				Limit:  rpmLimit,
				Window: time.Minute,
				Cost:   1,
			})
			if err == nil && res != nil && !res.Allowed {
				retrySecs := int(res.RetryAfter.Seconds())
				if retrySecs < 1 {
					retrySecs = 1
				}
				c.Header(HeaderRetryAfter, strconv.Itoa(retrySecs))
				c.JSON(http.StatusTooManyRequests, gin.H{
					"error": gin.H{
						"message": fmt.Sprintf("Rate limit exceeded for Organization. Limit: %d RPM.", res.Limit),
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
