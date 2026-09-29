package user

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserController struct {
	userService	*Service
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