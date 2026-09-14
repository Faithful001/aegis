package inference

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Faithful001/aegis/internal/domain/inference/dto"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type InferenceController struct {
	service *InferenceService
}

func NewInferenceController(service *InferenceService) *InferenceController {
	return &InferenceController{service: service}
}

func (ctrl *InferenceController) HandleChatCompletion(c *gin.Context) {
	var req dto.ChatCompletionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"message": fmt.Sprintf("invalid request body: %v", err),
				"type":    "invalid_request_error",
			},
		})
		return
	}

	reqID := c.GetString("requestID")
	if reqID == "" {
		reqID = c.GetHeader("X-Request-ID")
	}
	if reqID == "" {
		reqID = fmt.Sprintf("req_%s", uuid.New().String())
		c.Set("requestID", reqID)
	}

	orgID, _ := c.Get("tenantID")
	projectID, _ := c.Get("projectID")

	var orgUUID, projectUUID uuid.UUID
	if val, ok := orgID.(uuid.UUID); ok {
		orgUUID = val
	}
	if val, ok := projectID.(uuid.UUID); ok {
		projectUUID = val
	}

	if req.Stream {
		ctrl.handleStreamChatCompletion(c, reqID, orgUUID, projectUUID, req)
		return
	}

	resp, err := ctrl.service.ExecuteChatCompletion(c.Request.Context(), reqID, orgUUID, projectUUID, req)
	if err != nil {
		ctrl.handleInferenceError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (ctrl *InferenceController) handleStreamChatCompletion(
	c *gin.Context,
	reqID string,
	orgUUID, projectUUID uuid.UUID,
	req dto.ChatCompletionRequest,
) {
	chunkChan, job, err := ctrl.service.ExecuteStreamChatCompletion(c.Request.Context(), reqID, orgUUID, projectUUID, req)
	if err != nil {
		ctrl.handleInferenceError(c, err)
		return
	}

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")
	c.Writer.Header().Set("X-Accel-Buffering", "no")
	c.Writer.WriteHeader(http.StatusOK)
	c.Writer.Flush()

	created := time.Now().UTC().Unix()

	for chunk := range chunkChan {
		if chunk.Error != "" {
			errPayload, _ := json.Marshal(gin.H{
				"error": gin.H{
					"message": chunk.Error,
					"type":    "stream_error",
				},
			})
			_, _ = fmt.Fprintf(c.Writer, "data: %s\n\n", errPayload)
			c.Writer.Flush()
			return
		}

		var finishReason *string
		if chunk.FinishReason != "" {
			fr := chunk.FinishReason
			finishReason = &fr
		}

		streamResp := dto.ChatCompletionChunk{
			ID:      reqID,
			Object:  "chat.completion.chunk",
			Created: created,
			Model:   job.Model,
			Choices: []dto.StreamChoiceDTO{
				{
					Index: 0,
					Delta: dto.DeltaDTO{
						Content: chunk.Content,
					},
					FinishReason: finishReason,
				},
			},
		}

		if chunk.Done && chunk.TotalTokens > 0 {
			streamResp.Usage = &dto.UsageDTO{
				PromptTokens:     chunk.PromptTokens,
				CompletionTokens: chunk.OutputTokens,
				TotalTokens:      chunk.TotalTokens,
			}
		}

		payload, err := json.Marshal(streamResp)
		if err == nil {
			_, _ = fmt.Fprintf(c.Writer, "data: %s\n\n", payload)
			c.Writer.Flush()
		}

		if chunk.Done {
			break
		}
	}

	_, _ = io.WriteString(c.Writer, "data: [DONE]\n\n")
	c.Writer.Flush()
}

func (ctrl *InferenceController) handleInferenceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrModelNotSupported):
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"message": err.Error(),
				"type":    "invalid_request_error",
				"code":    "model_not_found",
			},
		})
	case errors.Is(err, ErrEmptyMessages), errors.Is(err, ErrInvalidTemperature), errors.Is(err, ErrInvalidMaxTokens):
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"message": err.Error(),
				"type":    "invalid_request_error",
			},
		})
	case errors.Is(err, ErrWorkerUnavailable):
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": gin.H{
				"message": "No inference worker available to process request",
				"type":    "server_error",
				"code":    "worker_unavailable",
			},
		})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"message": fmt.Sprintf("internal server error: %v", err),
				"type":    "api_error",
			},
		})
	}
}
