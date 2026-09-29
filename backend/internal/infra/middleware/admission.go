package middleware

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/Faithful001/aegis/internal/domain/admission"
	"github.com/gin-gonic/gin"
)

type chatCompletionBody struct {
	Model     string `json:"model"`
	Messages  []struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"messages"`
	MaxTokens int `json:"max_tokens"`
}

func AdmissionMiddleware(admissionService *admission.AdmissionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if admissionService == nil {
			c.Next()
			return
		}

		var promptText string
		var maxOutput int
		model := "mistral-small"

		// Peek at request body for token estimation if present
		if c.Request.Body != nil {
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err == nil {
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
				var reqBody chatCompletionBody
				if json.Unmarshal(bodyBytes, &reqBody) == nil {
					if reqBody.Model != "" {
						model = reqBody.Model
					}
					maxOutput = reqBody.MaxTokens
					for _, m := range reqBody.Messages {
						promptText += m.Content + " "
					}
				}
			}
		}

		inputToks, totalToks := admission.EstimateTokens(promptText, maxOutput)
		reqID := c.GetString("requestID")
		if reqID == "" {
			reqID = c.GetHeader("X-Request-ID")
		}

		admReq := admission.AdmissionRequest{
			RequestID:            reqID,
			Model:                model,
			EstimatedInputTokens: inputToks,
			MaxOutputTokens:      totalToks - inputToks,
		}

		resp, err := admissionService.AcquireSlot(c.Request.Context(), admReq)
		if err != nil {
			if errors.Is(err, admission.ErrMaxTokensExceeded) {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": gin.H{
						"message": fmt.Sprintf("Token estimation exceeds maximum allowed limit: %v", err),
						"type":    "invalid_request_error",
						"code":    "max_tokens_exceeded",
					},
				})
				c.Abort()
				return
			}

			c.Header("Retry-After", "5")
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": gin.H{
					"message": "System overloaded. Request rejected by admission control backpressure.",
					"type":    "server_error",
					"code":    "system_overloaded",
				},
			})
			c.Abort()
			return
		}

		// Ensure slot is released upon HTTP request completion
		defer admissionService.ReleaseSlot()

		if resp != nil && resp.WaitTime > 0 {
			c.Header("X-Admission-Wait-Ms", fmt.Sprintf("%d", resp.WaitTime.Milliseconds()))
		}

		c.Next()
	}
}
