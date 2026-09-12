package router

import (
	"github.com/Faithful001/aegis/internal/domain/auth"
	"github.com/Faithful001/aegis/internal/infra/middleware"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	authService := auth.NewAuthService()
	authController := auth.NewAuthController(authService)

	api := r.Group("/api/v1")
	{
		authRoutes := api.Group("/auth")
		{
			authRoutes.POST("/register", authController.Register)
			authRoutes.POST("/login", authController.Login)
			authRoutes.POST("/refresh", authController.RefreshToken)
			authRoutes.POST("/logout", authController.Logout)
		}

		protected := api.Group("/user")
		protected.Use(middleware.AuthMiddleware(authService))
		{
			protected.GET("/profile", func(c *gin.Context) {
				userID, _ := middleware.GetUserID(c)
				userEmail, _ := middleware.GetUserEmail(c)
				c.JSON(200, gin.H{
					"success": true,
					"data": gin.H{
						"user_id": userID,
						"email":   userEmail,
					},
				})
			})
		}
	}
}
