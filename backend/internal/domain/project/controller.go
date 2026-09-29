package project

import (
	"errors"
	"net/http"

	"github.com/Faithful001/aegis/internal/domain/organization"
	"github.com/Faithful001/aegis/internal/domain/project/dto"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ProjectController struct {
	projectService *ProjectService
}

func NewProjectController(projectService *ProjectService) *ProjectController {
	return &ProjectController{projectService: projectService}
}

func getUserID(c *gin.Context) (uuid.UUID, bool) {
	val, exists := c.Get("userID")
	if !exists {
		return uuid.Nil, false
	}
	id, ok := val.(uuid.UUID)
	return id, ok
}

func (h *ProjectController) Create(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "unauthorized", "data": nil})
		return
	}

	orgID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid organization id", "data": nil})
		return
	}

	var req dto.CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error(), "data": nil})
		return
	}

	proj, err := h.projectService.CreateProject(c.Request.Context(), orgID, userID, req)
	if err != nil {
		if errors.Is(err, organization.ErrUnauthorizedTenant) {
			c.JSON(http.StatusForbidden, gin.H{"success": false, "message": err.Error(), "data": nil})
			return
		}
		if errors.Is(err, ErrProjectSlugExists) {
			c.JSON(http.StatusConflict, gin.H{"success": false, "message": err.Error(), "data": nil})
			return
		}
		if errors.Is(err, ErrInvalidProjectData) {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error(), "data": nil})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error(), "data": nil})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "data": proj})
}

func (h *ProjectController) List(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "unauthorized", "data": nil})
		return
	}

	orgID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid organization id", "data": nil})
		return
	}

	projects, err := h.projectService.ListOrganizationProjects(c.Request.Context(), orgID, userID)
	if err != nil {
		if errors.Is(err, organization.ErrUnauthorizedTenant) {
			c.JSON(http.StatusForbidden, gin.H{"success": false, "message": err.Error(), "data": nil})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error(), "data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": projects})
}

func (h *ProjectController) Get(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "unauthorized", "data": nil})
		return
	}

	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid project id", "data": nil})
		return
	}

	proj, err := h.projectService.GetProject(c.Request.Context(), projectID, userID)
	if err != nil {
		if errors.Is(err, organization.ErrUnauthorizedTenant) {
			c.JSON(http.StatusForbidden, gin.H{"success": false, "message": err.Error(), "data": nil})
			return
		}
		if errors.Is(err, ErrProjectNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "message": err.Error(), "data": nil})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error(), "data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": proj})
}
