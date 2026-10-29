package upload

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UploadController struct {
	service *UploadService
}

func NewUploadController(service *UploadService) *UploadController {
	return &UploadController{service: service}
}

func getUserID(c *gin.Context) (uuid.UUID, bool) {
	if val, exists := c.Get("userID"); exists {
		if id, ok := val.(uuid.UUID); ok {
			return id, true
		}
	}
	return uuid.Nil, false
}

// UploadFile handles POST /api/v1/upload
func (ctl *UploadController) UploadFile(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok || userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Unauthorized",
			"data":    nil,
		})
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "File field 'file' is required in form-data",
			"data":    nil,
		})
		return
	}

	requestedType := c.DefaultPostForm("type", "")

	result, err := ctl.service.SaveUploadedFile(c.Request.Context(), userID, fileHeader, requestedType)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, ErrInvalidFileType) || errors.Is(err, ErrFileTooLarge) || errors.Is(err, ErrEmptyFile) {
			status = http.StatusBadRequest
		}
		c.JSON(status, gin.H{
			"success": false,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "File uploaded successfully",
		"data": gin.H{
			"id":        result.File.ID,
			"filename":  result.File.Filename,
			"mime_type": result.File.MimeType,
			"size":      result.File.Size,
			"file_type": result.File.FileType,
			"url":       result.URL,
			"content":   result.Content,
		},
	})
}

// UploadMultipleFiles handles POST /api/v1/upload/multiple
func (ctl *UploadController) UploadMultipleFiles(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok || userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Unauthorized",
			"data":    nil,
		})
		return
	}

	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Failed to parse multipart form: " + err.Error(),
			"data":    nil,
		})
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		files = form.File["file"]
	}

	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "No files provided in 'files' or 'file' form field",
			"data":    nil,
		})
		return
	}

	var requestedType string
	if types := form.Value["type"]; len(types) > 0 {
		requestedType = types[0]
	}

	results := make([]gin.H, 0, len(files))
	var lastErr error

	for _, fileHeader := range files {
		res, err := ctl.service.SaveUploadedFile(c.Request.Context(), userID, fileHeader, requestedType)
		if err != nil {
			lastErr = err
			continue
		}

		results = append(results, gin.H{
			"id":        res.File.ID,
			"filename":  res.File.Filename,
			"mime_type": res.File.MimeType,
			"size":      res.File.Size,
			"file_type": res.File.FileType,
			"url":       res.URL,
			"content":   res.Content,
		})
	}

	if len(results) == 0 && lastErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": lastErr.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": fmt.Sprintf("%d file(s) uploaded successfully", len(results)),
		"data":    results,
	})
}

// GetFile handles GET /api/v1/upload/:id
func (ctl *UploadController) GetFile(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok || userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Unauthorized", "data": nil})
		return
	}

	idParam := c.Param("id")
	fileID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid file ID", "data": nil})
		return
	}

	file, err := ctl.service.GetFile(c.Request.Context(), fileID, userID)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, ErrFileNotFound) {
			status = http.StatusNotFound
		} else if errors.Is(err, ErrUnauthorizedAccess) {
			status = http.StatusForbidden
		}
		c.JSON(status, gin.H{"success": false, "message": err.Error(), "data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"id":        file.ID,
			"filename":  file.Filename,
			"mime_type": file.MimeType,
			"size":      file.Size,
			"file_type": file.FileType,
			"url":       fmt.Sprintf("/api/v1/upload/%s/file", file.ID.String()),
		},
	})
}

// ServeFile handles GET /api/v1/upload/:id/file
func (ctl *UploadController) ServeFile(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok || userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Unauthorized", "data": nil})
		return
	}

	idParam := c.Param("id")
	fileID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid file ID", "data": nil})
		return
	}

	file, err := ctl.service.GetFile(c.Request.Context(), fileID, userID)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, ErrFileNotFound) {
			status = http.StatusNotFound
		} else if errors.Is(err, ErrUnauthorizedAccess) {
			status = http.StatusForbidden
		}
		c.JSON(status, gin.H{"success": false, "message": err.Error(), "data": nil})
		return
	}

	if _, err := os.Stat(file.FilePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "File not found on storage", "data": nil})
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", file.Filename))
	c.Header("Content-Type", file.MimeType)
	c.File(file.FilePath)
}

