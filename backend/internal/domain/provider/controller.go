package provider

import (
	"errors"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/Faithful001/aegis/internal/domain/provider/dto"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ProviderController struct {
	service *ProviderCredentialService
}

func NewProviderController(service *ProviderCredentialService) *ProviderController {
	return &ProviderController{service: service}
}

func getUserIDFromContext(c *gin.Context) (uuid.UUID, bool) {
	if val, exists := c.Get("userID"); exists {
		if id, ok := val.(uuid.UUID); ok {
			return id, true
		}
	}
	return uuid.Nil, false
}

func (ctl *ProviderController) SaveCredential(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok || userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "unauthorized", "data": nil})
		return
	}

	var req dto.SaveCredentialRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error(), "data": nil})
		return
	}

	resp, err := ctl.service.SaveCredential(c.Request.Context(), userID, req)
	if err != nil {
		var valErr *ValidationError
		if errors.As(err, &valErr) || errors.Is(err, ErrInvalidProvider) || errors.Is(err, ErrInvalidCredentialData) ||
			errors.Is(err, ErrInvalidAPIKey) || errors.Is(err, ErrInvalidBaseURL) {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error(), "data": nil})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error(), "data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    resp,
	})
}

func (ctl *ProviderController) ListCredentials(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok || userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "unauthorized", "data": nil})
		return
	}

	resp, err := ctl.service.ListCredentials(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error(), "data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    resp.Data,
	})
}

func (ctl *ProviderController) DeleteCredential(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok || userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "unauthorized", "data": nil})
		return
	}

	provider := c.Param("provider")
	if err := ctl.service.DeleteCredential(c.Request.Context(), userID, provider); err != nil {
		if errors.Is(err, ErrCredentialNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "message": err.Error(), "data": nil})
			return
		}
		if errors.Is(err, ErrInvalidProvider) {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error(), "data": nil})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error(), "data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "provider credential deleted successfully",
	})
}

func (ctl *ProviderController) ListProviderModels(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok || userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "unauthorized", "data": nil})
		return
	}

	provider := c.Param("provider")
	models, err := ctl.service.ListProviderModels(c.Request.Context(), userID, provider)
	if err != nil {
		if errors.Is(err, ErrCredentialNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "API key for provider not found. Please configure your API key first.",
				"data":    nil,
			})
			return
		}
		if errors.Is(err, ErrInvalidProvider) {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": err.Error(),
				"data":    nil,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	// Filter by search query if provided
	search := strings.ToLower(strings.TrimSpace(c.Query("search")))
	if search == "" {
		search = strings.ToLower(strings.TrimSpace(c.Query("q")))
	}

	filtered := models
	if search != "" {
		filtered = make([]dto.ModelDescriptor, 0)
		for _, m := range models {
			if strings.Contains(strings.ToLower(m.ID), search) || strings.Contains(strings.ToLower(m.Name), search) {
				filtered = append(filtered, m)
			}
		}
	}

	total := len(filtered)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limit <= 0 {
		limit = 10
	} else if limit > 100 {
		limit = 100
	}

	totalPages := 0
	if total > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(limit)))
	}

	start := (page - 1) * limit
	if start > total {
		start = total
	}
	end := start + limit
	if end > total {
		end = total
	}

	pagedModels := filtered[start:end]
	if pagedModels == nil {
		pagedModels = make([]dto.ModelDescriptor, 0)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": dto.PaginatedModelListResponse{
			Items:      pagedModels,
			Total:      total,
			Page:       page,
			Limit:      limit,
			TotalPages: totalPages,
		},
	})
}

func (ctl *ProviderController) ListAllModels(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok || userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "unauthorized", "data": nil})
		return
	}

	var isActive *bool
	if activeStr := c.Query("is_active"); activeStr != "" {
		val := activeStr == "true" || activeStr == "1"
		isActive = &val
	}

	models, err := ctl.service.ListAllModels(c.Request.Context(), userID, isActive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error(), "data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    models,
	})
}

func (ctl *ProviderController) ToggleModel(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok || userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "unauthorized", "data": nil})
		return
	}

	var req dto.ToggleModelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error(), "data": nil})
		return
	}

	model, err := ctl.service.ToggleModelActive(c.Request.Context(), userID, req.Provider, req.ModelID, req.IsActive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error(), "data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    model,
	})
}

func (ctl *ProviderController) ListSupportedProviders(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok || userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "unauthorized", "data": nil})
		return
	}

	providers := ctl.service.GetSupportedProviders(c.Request.Context(), userID)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    providers,
	})
}
