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

// SaveCredential handles POST /api/v1/organizations/:id/credentials
func (ctl *ProviderController) SaveCredential(c *gin.Context) {
	orgIDParam := c.Param("id")
	orgID, err := uuid.Parse(orgIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization ID"})
		return
	}

	var req dto.SaveCredentialRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := ctl.service.SaveCredential(c.Request.Context(), orgID, req)
	if err != nil {
		if errors.Is(err, ErrInvalidProvider) || errors.Is(err, ErrInvalidCredentialData) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    resp,
	})
}

// ListCredentials handles GET /api/v1/organizations/:id/credentials
func (ctl *ProviderController) ListCredentials(c *gin.Context) {
	orgIDParam := c.Param("id")
	orgID, err := uuid.Parse(orgIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization ID"})
		return
	}

	resp, err := ctl.service.ListCredentials(c.Request.Context(), orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    resp.Data,
	})
}

// DeleteCredential handles DELETE /api/v1/organizations/:id/credentials/:provider
func (ctl *ProviderController) DeleteCredential(c *gin.Context) {
	orgIDParam := c.Param("id")
	orgID, err := uuid.Parse(orgIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization ID"})
		return
	}

	provider := c.Param("provider")
	if err := ctl.service.DeleteCredential(c.Request.Context(), orgID, provider); err != nil {
		if errors.Is(err, ErrCredentialNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, ErrInvalidProvider) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "provider credential deleted successfully",
	})
}
