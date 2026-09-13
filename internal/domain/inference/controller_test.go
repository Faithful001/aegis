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

func TestController_HandleChatCompletion_Stream_FullSSEValidation(t *testing.T) {
	mockClient := NewMockWorkerClient()
	service := NewService(mockClient)
	ctrl := NewController(service)

	orgID := uuid.New()
	projectID := uuid.New()
	router := setupTestRouter(ctrl, orgID, projectID)

	reqPayload := dto.ChatCompletionRequest{
		Model: "mistral-small",
		Messages: []dto.ChatMessageDTO{
			{Role: "user", Content: "Detailed stream test"},
		},
		Stream: true,
	}

	body, _ := json.Marshal(reqPayload)
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}

	lines := bytes.Split(w.Body.Bytes(), []byte("\n"))
	var sseChunks []dto.ChatCompletionChunk
	foundDoneSentinel := false

	for _, line := range lines {
		lineStr := string(bytes.TrimSpace(line))
		if lineStr == "" {
			continue
		}
		if lineStr == "data: [DONE]" {
			foundDoneSentinel = true
			continue
		}
		if bytes.HasPrefix(line, []byte("data: ")) {
			jsonPayload := line[6:]
			var chunk dto.ChatCompletionChunk
			if err := json.Unmarshal(jsonPayload, &chunk); err == nil {
				sseChunks = append(sseChunks, chunk)
			}
		}
	}

	if !foundDoneSentinel {
		t.Errorf("missing data: [DONE] sentinel in SSE stream")
	}

	if len(sseChunks) == 0 {
		t.Fatalf("expected SSE chunks, got 0")
	}

	// Final chunk should carry usage statistics
	finalChunk := sseChunks[len(sseChunks)-1]
	if finalChunk.Usage == nil {
		t.Errorf("expected final chunk to contain usage information")
	} else if finalChunk.Usage.TotalTokens == 0 {
		t.Errorf("expected total_tokens > 0 in final chunk usage")
	}
}

