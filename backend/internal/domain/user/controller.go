package user

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserController struct {
	userService *Service
}

func NewUserController(userService *Service) *UserController {
	return &UserController{
		userService: userService,
	}
}

func (c *UserController) GetMe(ctx *gin.Context) {
	userID, ok := ctx.Get("userID")
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Unauthorized", "data": nil})
		return
	}

	user, err := c.userService.GetUserByID(ctx, userID.(uuid.UUID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to get user", "data": nil})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": true, "message": "User fetched successfully", "data": user})
}

func (c *UserController) GetAlertSettings(ctx *gin.Context) {
	userIDVal, ok := ctx.Get("userID")
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Unauthorized"})
		return
	}
	userID := userIDVal.(uuid.UUID)

	userEmail, _ := ctx.Get("userEmail")
	emailStr := ""
	if userEmail != nil {
		emailStr = userEmail.(string)
	}

	settings, err := c.userService.GetAlertSettings(ctx.Request.Context(), userID, emailStr)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": true, "data": settings})
}

func (c *UserController) UpdateAlertSettings(ctx *gin.Context) {
	userIDVal, ok := ctx.Get("userID")
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Unauthorized"})
		return
	}
	userID := userIDVal.(uuid.UUID)

	var req UpdateAlertSettingsDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	settings, err := c.userService.UpdateAlertSettings(ctx.Request.Context(), userID, req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": true, "message": "Alert settings updated successfully", "data": settings})
}

func (c *UserController) ExportData(ctx *gin.Context) {
	userIDVal, ok := ctx.Get("userID")
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Unauthorized"})
		return
	}
	userID := userIDVal.(uuid.UUID)

	data, err := c.userService.ExportUserData(ctx.Request.Context(), userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": true, "data": data})
}

func (c *UserController) DeleteAccount(ctx *gin.Context) {
	userIDVal, ok := ctx.Get("userID")
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Unauthorized"})
		return
	}
	userID := userIDVal.(uuid.UUID)

	if err := c.userService.DeleteUserAccount(ctx.Request.Context(), userID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": true, "message": "Account successfully deleted"})
}