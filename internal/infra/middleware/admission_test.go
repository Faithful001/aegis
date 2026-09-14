package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Faithful001/aegis/internal/domain/admission"
	"github.com/gin-gonic/gin"
)

func setupAdmissionTestRouter(admSvc *admission.Service) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(AdmissionMiddleware(admSvc))
	r.POST("/v1/chat/completions", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	return r
}

func TestAdmissionMiddleware_Accept(t *testing.T) {
	cfg := admission.AdmissionConfig{
		MaxActiveConcurrency: 5,
		MaxQueueDepth:        5,
		MaxQueueWaitTime:     100 * time.Millisecond,
		MaxTokensPerRequest:  1000,
	}
	admSvc := admission.NewService(cfg)
	router := setupAdmissionTestRouter(admSvc)

	body, _ := json.Marshal(gin.H{
		"model": "mistral-small",
		"messages": []gin.H{
			{"role": "user", "content": "Hello admission test"},
		},
		"max_tokens": 50,
	})

	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAdmissionMiddleware_Reject503Overloaded(t *testing.T) {
	cfg := admission.AdmissionConfig{
		MaxActiveConcurrency: 1,
		MaxQueueDepth:        0, // Zero queue depth so extra request REJECTs immediately
		MaxQueueWaitTime:     10 * time.Millisecond,
		MaxTokensPerRequest:  1000,
	}
	admSvc := admission.NewService(cfg)

	// Occupy sole concurrency slot
	admReq := admission.AdmissionRequest{RequestID: "occupy_1", EstimatedInputTokens: 10, MaxOutputTokens: 10}
	_, err := admSvc.AcquireSlot(t.Context(), admReq)
	if err != nil {
		t.Fatalf("failed to occupy initial slot: %v", err)
	}

	router := setupAdmissionTestRouter(admSvc)

	body, _ := json.Marshal(gin.H{
		"model": "mistral-small",
		"messages": []gin.H{
			{"role": "user", "content": "Overload request"},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 Service Unavailable, got %d: %s", w.Code, w.Body.String())
	}

	if retryAfter := w.Header().Get("Retry-After"); retryAfter == "" {
		t.Errorf("expected Retry-After header on 503 backpressure response")
	}

	var errResp map[string]map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("failed to parse JSON error response: %v", err)
	}

	if errResp["error"]["code"] != "system_overloaded" {
		t.Errorf("expected error code system_overloaded, got %s", errResp["error"]["code"])
	}
}
