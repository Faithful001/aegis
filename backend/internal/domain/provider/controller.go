package provider

import (
	"errors"
	"net/http"

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

// SaveCredential handles POST /api/v1/credentials
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
		if errors.Is(err, ErrInvalidProvider) || errors.Is(err, ErrInvalidCredentialData) {
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

// ListCredentials handles GET /api/v1/credentials
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

// DeleteCredential handles DELETE /api/v1/credentials/:provider
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

// ListProviderModels handles GET /api/v1/credentials/:provider/models
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

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    models,
	})
}

// ListAllModels handles GET /api/v1/models
func (ctl *ProviderController) ListAllModels(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok || userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "unauthorized", "data": nil})
		return
	}

	models, err := ctl.service.ListAllModels(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error(), "data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    models,
	})
}
