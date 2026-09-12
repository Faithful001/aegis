package inference

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Faithful001/aegis/internal/domain/inference/dto"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func setupTestRouter(ctrl *Controller, orgID, projectID uuid.UUID) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("tenantID", orgID)
		c.Set("projectID", projectID)
		c.Next()
	})
	r.POST("/v1/chat/completions", ctrl.HandleChatCompletion)
	return r
}

func TestController_HandleChatCompletion_Unary(t *testing.T) {
	mockClient := NewMockWorkerClient()
	service := NewService(mockClient)
	ctrl := NewController(service)

	orgID := uuid.New()
	projectID := uuid.New()
	router := setupTestRouter(ctrl, orgID, projectID)

	reqPayload := dto.ChatCompletionRequest{
		Model: "mistral-small",
		Messages: []dto.ChatMessageDTO{
			{Role: "user", Content: "Hello world!"},
		},
		MaxTokens: 50,
		Stream:    false,
	}

	body, _ := json.Marshal(reqPayload)
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", w.Code, w.Body.String())
	}

	var resp dto.ChatCompletionResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode JSON response: %v", err)
	}

	if resp.Object != "chat.completion" {
		t.Errorf("expected object chat.completion, got %s", resp.Object)
	}
	if len(resp.Choices) != 1 {
		t.Errorf("expected 1 choice, got %d", len(resp.Choices))
	}
}

func TestController_HandleChatCompletion_Stream(t *testing.T) {
	mockClient := NewMockWorkerClient()
	service := NewService(mockClient)
	ctrl := NewController(service)

	orgID := uuid.New()
	projectID := uuid.New()
	router := setupTestRouter(ctrl, orgID, projectID)

	reqPayload := dto.ChatCompletionRequest{
		Model: "mistral-small",
		Messages: []dto.ChatMessageDTO{
			{Role: "user", Content: "Stream tokens!"},
		},
		Stream: true,
	}

	body, _ := json.Marshal(reqPayload)
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", w.Code, w.Body.String())
	}

	if w.Header().Get("Content-Type") != "text/event-stream" {
		t.Errorf("expected text/event-stream content type, got %s", w.Header().Get("Content-Type"))
	}

	bodyStr := w.Body.String()
	if !bytes.Contains([]byte(bodyStr), []byte("data: [DONE]")) {
		t.Errorf("expected stream to contain [DONE] sentinel")
	}
}
