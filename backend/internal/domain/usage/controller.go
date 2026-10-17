package usage

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Faithful001/aegis/internal/domain/usage/dto"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UsageController struct {
	service *UsageService
}

func NewUsageController(service *UsageService) *UsageController {
	return &UsageController{service: service}
}

func (h *UsageController) GetUserUsage(c *gin.Context) {
	val, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "unauthorized", "data": nil})
		return
	}
	userID := val.(uuid.UUID)

	summaryMode := c.Query("summary") == "true"

	var startTime, endTime time.Time
	if stStr := c.Query("start_time"); stStr != "" {
		if t, err := time.Parse(time.RFC3339, stStr); err == nil {
			startTime = t
		}
	}
	if etStr := c.Query("end_time"); etStr != "" {
		if t, err := time.Parse(time.RFC3339, etStr); err == nil {
			endTime = t
		}
	}

	if summaryMode {
		summary, err := h.service.GetTotalUsage(c.Request.Context(), userID, startTime, endTime)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error(), "data": nil})
			return
		}

		res := dto.UsageSummaryResponse{
			OrganizationID: userID,
			InputTokens:    summary.InputTokens,
			OutputTokens:   summary.OutputTokens,
			TotalTokens:    summary.TotalTokens,
			TotalRequests:  summary.TotalRequests,
			TotalCostUSD:   summary.TotalCostUSD,
		}
		if !startTime.IsZero() {
			res.StartTime = &startTime
		}
		if !endTime.IsZero() {
			res.EndTime = &endTime
		}

		c.JSON(http.StatusOK, gin.H{"success": true, "data": res})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	records, err := h.service.ListByOrganization(c.Request.Context(), userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error(), "data": nil})
		return
	}

	items := make([]dto.UsageRecordResponse, len(records))
	for i, r := range records {
		items[i] = mapRecordToResponse(r)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": dto.UsageListResponse{
			Items:  items,
			Limit:  limit,
			Offset: offset,
		},
	})
}

func (h *UsageController) GetOrganizationUsage(c *gin.Context) {
	orgID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid organization id", "data": nil})
		return
	}

	summaryMode := c.Query("summary") == "true"

	var startTime, endTime time.Time
	if stStr := c.Query("start_time"); stStr != "" {
		if t, err := time.Parse(time.RFC3339, stStr); err == nil {
			startTime = t
		}
	}
	if etStr := c.Query("end_time"); etStr != "" {
		if t, err := time.Parse(time.RFC3339, etStr); err == nil {
			endTime = t
		}
	}

	if summaryMode {
		summary, err := h.service.GetTotalUsage(c.Request.Context(), orgID, startTime, endTime)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error(), "data": nil})
			return
		}

		res := dto.UsageSummaryResponse{
			OrganizationID: orgID,
			InputTokens:    summary.InputTokens,
			OutputTokens:   summary.OutputTokens,
			TotalTokens:    summary.TotalTokens,
			TotalRequests:  summary.TotalRequests,
			TotalCostUSD:   summary.TotalCostUSD,
		}
		if !startTime.IsZero() {
			res.StartTime = &startTime
		}
		if !endTime.IsZero() {
			res.EndTime = &endTime
		}

		c.JSON(http.StatusOK, gin.H{"success": true, "data": res})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	records, err := h.service.ListByOrganization(c.Request.Context(), orgID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error(), "data": nil})
		return
	}

	items := make([]dto.UsageRecordResponse, len(records))
	for i, r := range records {
		items[i] = mapRecordToResponse(r)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": dto.UsageListResponse{
			Items:  items,
			Limit:  limit,
			Offset: offset,
		},
	})
}

func (h *UsageController) GetProjectUsage(c *gin.Context) {
	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid project id", "data": nil})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	records, err := h.service.ListByProject(c.Request.Context(), projectID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error(), "data": nil})
		return
	}

	items := make([]dto.UsageRecordResponse, len(records))
	for i, r := range records {
		items[i] = mapRecordToResponse(r)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": dto.UsageListResponse{
			Items:  items,
			Limit:  limit,
			Offset: offset,
		},
	})
}

func mapRecordToResponse(r *UsageRecord) dto.UsageRecordResponse {
	return dto.UsageRecordResponse{
		ID:             r.ID,
		EventID:        r.EventID,
		RequestID:      r.RequestID,
		OrganizationID: r.OrganizationID,
		ProjectID:      r.ProjectID,
		Model:          r.Model,
		InputTokens:    r.InputTokens,
		OutputTokens:   r.OutputTokens,
		TotalTokens:    r.TotalTokens,
		DurationMS:     r.DurationMS,
		Status:         r.Status,
		CreatedAt:      r.CreatedAt,
	}
}

// GetTokenActivity returns a GitHub-style heatmap payload.
// Query params:
//   - granularity: "daily" (default) | "weekly" | "cumulative"
//   - since: RFC3339 timestamp (default: 12 months ago)
func (h *UsageController) GetTokenActivity(c *gin.Context) {
	val, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "unauthorized", "data": nil})
		return
	}
	userID := val.(uuid.UUID)

	granularity := c.DefaultQuery("granularity", "daily")

	var since time.Time
	if sinceStr := c.Query("since"); sinceStr != "" {
		if t, err := time.Parse(time.RFC3339, sinceStr); err == nil {
			since = t
		}
	}

	result, err := h.service.GetTokenActivity(c.Request.Context(), userID, granularity, since)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error(), "data": nil})
		return
	}

	// Map service result → DTO
	buckets := make([]dto.TokenActivityBucket, len(result.Buckets))
	for i, b := range result.Buckets {
		buckets[i] = dto.TokenActivityBucket{Date: b.Date, Tokens: b.Tokens, Requests: b.Requests}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": dto.TokenActivityResponse{
			Granularity: result.Granularity,
			Buckets:     buckets,
			MaxTokens:   result.MaxTokens,
		},
	})
}
