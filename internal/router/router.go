package router

import (
	"context"
	"net/http"
	"time"

	"github.com/Faithful001/aegis/internal/domain/auth"
	"github.com/Faithful001/aegis/internal/domain/inference"
	"github.com/Faithful001/aegis/internal/domain/organization"
	"github.com/Faithful001/aegis/internal/domain/project"
	"github.com/Faithful001/aegis/internal/infra/db"
	"github.com/Faithful001/aegis/internal/infra/middleware"
	"github.com/Faithful001/aegis/internal/infra/ratelimiter"
	"github.com/Faithful001/aegis/internal/infra/redis"
	"github.com/gin-gonic/gin"
)

type RouterConfig struct {
	AuthService         *auth.AuthService
	APIKeyService       *auth.APIKeyService
	AuthController      *auth.AuthController
	APIKeyController    *auth.APIKeyController
	OrgController       *organization.Controller
	ProjectController   *project.Controller
	InferenceController *inference.Controller
	RateLimiter         ratelimiter.RateLimiter
}

func SetupRouter(cfg RouterConfig) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestLoggingMiddleware())

	// Liveness Probe
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	})

	// Readiness Probe
	r.GET("/ready", func(c *gin.Context) {
		dbStatus := "operational"
		redisStatus := "operational"
		isReady := true

		if database := db.GetDB(); database != nil {
			sqlDB, err := database.DB()
			if err != nil || sqlDB.Ping() != nil {
				dbStatus = "unavailable"
				isReady = false
			}
		} else {
			dbStatus = "not_configured"
		}

		if redisClient := redis.GetClient(); redisClient != nil {
			ctx, cancel := context.WithTimeout(c.Request.Context(), 1*time.Second)
			defer cancel()
			if err := redisClient.Ping(ctx).Err(); err != nil {
				redisStatus = "unavailable"
				isReady = false
			}
		} else {
			redisStatus = "not_configured"
		}

		statusCode := http.StatusOK
		if !isReady {
			statusCode = http.StatusServiceUnavailable
		}

		c.JSON(statusCode, gin.H{
			"status": map[string]string{
				"database": dbStatus,
				"redis":    redisStatus,
			},
			"ready": isReady,
		})
	})

	// ==========================================
	// 1. CONTROL PLANE API (/api/v1) - JWT Auth
	// ==========================================
	api := r.Group("/api/v1")
	{
		// Public Auth Endpoints
		authRoutes := api.Group("/auth")
		{
			authRoutes.POST("/register", cfg.AuthController.Register)
			authRoutes.POST("/login", cfg.AuthController.Login)
			authRoutes.POST("/refresh", cfg.AuthController.RefreshToken)
			authRoutes.POST("/logout", cfg.AuthController.Logout)
		}

		// Protected Management Routes
		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware(cfg.AuthService))
		{
			// User Profile
			protected.GET("/user/profile", func(c *gin.Context) {
				userID, _ := middleware.GetUserID(c)
				userEmail, _ := middleware.GetUserEmail(c)
				c.JSON(http.StatusOK, gin.H{
					"success": true,
					"data": gin.H{
						"user_id": userID,
						"email":   userEmail,
					},
				})
			})

			// Organizations
			orgs := protected.Group("/organizations")
			{
				orgs.POST("", cfg.OrgController.Create)
				orgs.GET("", cfg.OrgController.List)
				orgs.GET("/:id", cfg.OrgController.Get)
				orgs.POST("/:id/members", cfg.OrgController.AddMember)
				orgs.GET("/:id/members", cfg.OrgController.ListMembers)

				// Organization Projects
				orgs.POST("/:org_id/projects", cfg.ProjectController.Create)
				orgs.GET("/:org_id/projects", cfg.ProjectController.List)
			}

			// Projects & API Keys
			projects := protected.Group("/projects")
			{
				projects.GET("/:id", cfg.ProjectController.Get)
				projects.POST("/:id/api-keys", cfg.APIKeyController.Create)
				projects.GET("/:id/api-keys", cfg.APIKeyController.List)
			}

			// Direct API Key Management
			protected.DELETE("/api-keys/:id", cfg.APIKeyController.Revoke)
		}
	}

	// ==================================================
	// 2. INFERENCE / DATA PLANE API (/v1) - API Key Auth
	// ==================================================
	inferenceV1 := r.Group("/v1")
	inferenceV1.Use(middleware.APIKeyAuthMiddleware(cfg.APIKeyService))

	if cfg.RateLimiter != nil {
		inferenceV1.Use(middleware.RateLimitMiddleware(cfg.RateLimiter))
	} else if redisClient := redis.GetClient(); redisClient != nil {
		inferenceV1.Use(middleware.RateLimitMiddleware(ratelimiter.NewRedisRateLimiter(redisClient)))
	}
	{
		// Verification / ping endpoint for API Key Principal
		inferenceV1.GET("/auth/verify", func(c *gin.Context) {
			principal, _ := middleware.GetPrincipal(c)
			c.JSON(http.StatusOK, gin.H{
				"authenticated": true,
				"principal": gin.H{
					"organization_id": principal.OrganizationID,
					"project_id":      principal.ProjectID,
					"api_key_id":      principal.APIKeyID,
					"key_prefix":      principal.KeyPrefix,
				},
			})
		})

		// OpenAI-compatible Chat Completions API
		if cfg.InferenceController != nil {
			inferenceV1.POST("/chat/completions", cfg.InferenceController.HandleChatCompletion)
		}
	}

	return r
}
