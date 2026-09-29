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

func (h *UsageController) GetOrganizationUsage(c *gin.Context) {
	orgID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid organization id"})
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
		inTokens, outTokens, totTokens, err := h.service.GetTotalUsage(c.Request.Context(), orgID, startTime, endTime)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
			return
		}

		res := dto.UsageSummaryResponse{
			OrganizationID: orgID,
			InputTokens:    inTokens,
			OutputTokens:   outTokens,
			TotalTokens:    totTokens,
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
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
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
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid project id"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	records, err := h.service.ListByProject(c.Request.Context(), projectID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
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
