package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Faithful001/aegis/internal/domain/auth/dto"
	"github.com/Faithful001/aegis/internal/infra/ratelimiter"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func setupRateLimitTestRouter(limiter ratelimiter.RateLimiter, principal *dto.AuthenticatedPrincipal) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		if principal != nil {
			c.Set(ContextKeyPrincipal, principal)
			c.Set(ContextKeyOrgID, principal.OrganizationID)
			c.Set(ContextKeyProjectID, principal.ProjectID)
			c.Set(ContextKeyAPIKeyID, principal.APIKeyID)
		}
		c.Next()
	})
	r.Use(RateLimitMiddleware(limiter))
	r.POST("/v1/chat/completions", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	return r
}

func TestMemoryRateLimiter_SlidingWindow(t *testing.T) {
	limiter := ratelimiter.NewMemoryRateLimiter()
	ctx := t.Context()

	req := ratelimiter.RateLimitRequest{
		Tier:   ratelimiter.TierProject,
		Key:    "test:project:123:rpm",
		Limit:  3,
		Window: time.Minute,
		Cost:   1,
	}

	// Request 1: Allowed (Remaining: 2)
	res1, err := limiter.CheckRateLimit(ctx, req)
	if err != nil || !res1.Allowed {
		t.Fatalf("expected request 1 to be allowed, got err=%v, allowed=%v", err, res1.Allowed)
	}
	if res1.Remaining != 2 {
		t.Errorf("expected remaining 2, got %d", res1.Remaining)
	}

	// Request 2: Allowed (Remaining: 1)
	res2, err := limiter.CheckRateLimit(ctx, req)
	if err != nil || !res2.Allowed {
		t.Fatalf("expected request 2 to be allowed, got err=%v, allowed=%v", err, res2.Allowed)
	}
	if res2.Remaining != 1 {
		t.Errorf("expected remaining 1, got %d", res2.Remaining)
	}

	// Request 3: Allowed (Remaining: 0)
	res3, err := limiter.CheckRateLimit(ctx, req)
	if err != nil || !res3.Allowed {
		t.Fatalf("expected request 3 to be allowed, got err=%v, allowed=%v", err, res3.Allowed)
	}
	if res3.Remaining != 0 {
		t.Errorf("expected remaining 0, got %d", res3.Remaining)
	}

	// Request 4: Exceeded! (Allowed: false, Remaining: 0)
	res4, err := limiter.CheckRateLimit(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res4.Allowed {
		t.Errorf("expected request 4 to be blocked by rate limit")
	}
	if res4.Remaining != 0 {
		t.Errorf("expected remaining 0 on blocked request, got %d", res4.Remaining)
	}
	if res4.RetryAfter <= 0 {
		t.Errorf("expected positive RetryAfter duration on rate limit error")
	}
}

func TestMemoryRateLimiter_CostBatching(t *testing.T) {
	limiter := ratelimiter.NewMemoryRateLimiter()
	ctx := t.Context()

	req := ratelimiter.RateLimitRequest{
		Tier:   ratelimiter.TierProject,
		Key:    "test:project:tpm:456",
		Limit:  100,
		Window: time.Minute,
		Cost:   60,
	}

	// Batch 1 (60 tokens): Allowed
	res1, err := limiter.CheckRateLimit(ctx, req)
	if err != nil || !res1.Allowed {
		t.Fatalf("expected batch 1 to be allowed, got allowed=%v", res1.Allowed)
	}
	if res1.Remaining != 40 {
		t.Errorf("expected remaining 40, got %d", res1.Remaining)
	}

	// Batch 2 (60 tokens): Exceeded (60 + 60 = 120 > 100)
	res2, err := limiter.CheckRateLimit(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res2.Allowed {
		t.Errorf("expected batch 2 to be blocked")
	}
}

func TestRateLimitMiddleware_AllowedAndHeaders(t *testing.T) {
	memLimiter := ratelimiter.NewMemoryRateLimiter()
	principal := &dto.AuthenticatedPrincipal{
		APIKeyID:       uuid.New(),
		ProjectID:      uuid.New(),
		OrganizationID: uuid.New(),
	}

	router := setupRateLimitTestRouter(memLimiter, principal)

	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", w.Code, w.Body.String())
	}

	if limitHeader := w.Header().Get(HeaderXRateLimitLimitRPM); limitHeader == "" {
		t.Errorf("expected header %s to be set", HeaderXRateLimitLimitRPM)
	}
	if remHeader := w.Header().Get(HeaderXRateLimitRemainingRPM); remHeader == "" {
		t.Errorf("expected header %s to be set", HeaderXRateLimitRemainingRPM)
	}
}

func TestRateLimitMiddleware_Exceeded429(t *testing.T) {
	memLimiter := ratelimiter.NewMemoryRateLimiter()
	apiKeyID := uuid.New()
	projectID := uuid.New()
	principal := &dto.AuthenticatedPrincipal{
		APIKeyID:       apiKeyID,
		ProjectID:      projectID,
		OrganizationID: uuid.New(),
	}

	// Pre-fill memory limiter to hit project rate limit
	ctx := t.Context()
	projectKey := "ratelimit:project:" + projectID.String() + ":rpm"
	for i := 0; i < DefaultRPM; i++ {
		_, _ = memLimiter.CheckRateLimit(ctx, ratelimiter.RateLimitRequest{
			Tier:   ratelimiter.TierProject,
			Key:    projectKey,
			Limit:  DefaultRPM,
			Window: time.Minute,
			Cost:   1,
		})
	}

	router := setupRateLimitTestRouter(memLimiter, principal)

	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 Too Many Requests, got %d: %s", w.Code, w.Body.String())
	}

	if retryAfter := w.Header().Get(HeaderRetryAfter); retryAfter == "" {
		t.Errorf("expected Retry-After header on 429 error response")
	}

	var errResp map[string]map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("failed to decode JSON error response: %v", err)
	}

	if errResp["error"]["code"] != "rate_limit_exceeded" {
		t.Errorf("expected code rate_limit_exceeded, got %s", errResp["error"]["code"])
	}
}
