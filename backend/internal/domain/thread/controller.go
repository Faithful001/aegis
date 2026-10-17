package thread

import (
	"errors"
	"net/http"

	"github.com/Faithful001/aegis/internal/domain/thread/dto"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ThreadController struct {
	service *ThreadService
}

func NewThreadController(service *ThreadService) *ThreadController {
	return &ThreadController{service: service}
}

func getUserIDFromContext(c *gin.Context) (uuid.UUID, bool) {
	if val, exists := c.Get("userID"); exists {
		if id, ok := val.(uuid.UUID); ok {
			return id, true
		}
	}
	return uuid.Nil, false
}

// ListThreads handles GET /api/v1/threads
func (ctl *ThreadController) ListThreads(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok || userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "unauthorized", "data": nil})
		return
	}

	threads, err := ctl.service.ListThreads(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error(), "data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    threads,
	})
}

// CreateThread handles POST /api/v1/threads
func (ctl *ThreadController) CreateThread(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok || userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "unauthorized", "data": nil})
		return
	}

	var req dto.CreateThreadRequest
	_ = c.ShouldBindJSON(&req)

	thread, err := ctl.service.CreateThread(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error(), "data": nil})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    thread,
	})
}

// GetThread handles GET /api/v1/threads/:id
func (ctl *ThreadController) GetThread(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok || userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "unauthorized", "data": nil})
		return
	}

	threadID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid thread id", "data": nil})
		return
	}

	thread, err := ctl.service.GetThread(c.Request.Context(), userID, threadID)
	if err != nil {
		if errors.Is(err, ErrThreadNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "thread not found", "data": nil})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error(), "data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    thread,
	})
}

// UpdateThread handles PATCH /api/v1/threads/:id
func (ctl *ThreadController) UpdateThread(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok || userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "unauthorized", "data": nil})
		return
	}

	threadID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid thread id", "data": nil})
		return
	}

	var req dto.UpdateThreadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error(), "data": nil})
		return
	}

	thread, err := ctl.service.UpdateThread(c.Request.Context(), userID, threadID, req)
	if err != nil {
		if errors.Is(err, ErrThreadNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "thread not found", "data": nil})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error(), "data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    thread,
	})
}

// DeleteThread handles DELETE /api/v1/threads/:id
func (ctl *ThreadController) DeleteThread(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok || userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "unauthorized", "data": nil})
		return
	}

	threadID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid thread id", "data": nil})
		return
	}

	if err := ctl.service.DeleteThread(c.Request.Context(), userID, threadID); err != nil {
		if errors.Is(err, ErrThreadNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "thread not found", "data": nil})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error(), "data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "thread deleted successfully",
	})
}

// ListMessages handles GET /api/v1/threads/:id/messages
func (ctl *ThreadController) ListMessages(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok || userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "unauthorized", "data": nil})
		return
	}

	threadID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid thread id", "data": nil})
		return
	}

	messages, err := ctl.service.ListMessages(c.Request.Context(), userID, threadID)
	if err != nil {
		if errors.Is(err, ErrThreadNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "thread not found", "data": nil})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error(), "data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    messages,
	})
}

// AddMessage handles POST /api/v1/threads/:id/messages
func (ctl *ThreadController) AddMessage(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok || userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "unauthorized", "data": nil})
		return
	}

	threadID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid thread id", "data": nil})
		return
	}

	var req dto.CreateMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error(), "data": nil})
		return
	}

	msg, err := ctl.service.AddMessage(c.Request.Context(), userID, threadID, req)
	if err != nil {
		if errors.Is(err, ErrThreadNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "thread not found", "data": nil})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error(), "data": nil})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    msg,
	})
}
