package router

import (
	"context"
	"net/http"
	"time"

	"github.com/Faithful001/aegis/internal/domain/admission"
	"github.com/Faithful001/aegis/internal/domain/auth"
	"github.com/Faithful001/aegis/internal/domain/inference"
	"github.com/Faithful001/aegis/internal/domain/provider"
	"github.com/Faithful001/aegis/internal/domain/usage"
	"github.com/Faithful001/aegis/internal/domain/user"
	"github.com/Faithful001/aegis/internal/infra/db"
	"github.com/Faithful001/aegis/internal/infra/middleware"
	"github.com/Faithful001/aegis/internal/infra/ratelimiter"
	"github.com/Faithful001/aegis/internal/infra/redis"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type RouterConfig struct {
	AuthService         *auth.AuthService
	AuthController      *auth.AuthController
	UserController      *user.UserController
	InferenceController *inference.InferenceController
	UsageController     *usage.UsageController
	ProviderController  *provider.ProviderController
	RateLimiter         ratelimiter.RateLimiter
	AdmissionService    *admission.AdmissionService
}

func SetupRouter(cfg RouterConfig) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestLoggingMiddleware())
	r.Use(middleware.PrometheusMiddleware())

	// Prometheus Metrics Endpoint
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

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
		kafkaStatus := "operational"
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
				"kafka":	kafkaStatus,
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

		// Protected Management & Inference Routes
		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware(cfg.AuthService))
		{
			// User Profile
			protected.GET("/user/me", cfg.UserController.GetMe)

			// User BYOK Provider Credentials Management
			if cfg.ProviderController != nil {
				protected.POST("/credentials", cfg.ProviderController.SaveCredential)
				protected.GET("/credentials", cfg.ProviderController.ListCredentials)
				protected.DELETE("/credentials/:provider", cfg.ProviderController.DeleteCredential)
			}

			// User Token Usage Metering
			if cfg.UsageController != nil {
				protected.GET("/usage", cfg.UsageController.GetUserUsage)
				protected.GET("/user/usage", cfg.UsageController.GetUserUsage)
			}

			// Chat Completions API (JWT Authenticated)
			if cfg.InferenceController != nil {
				chatGroup := protected.Group("")
				if cfg.RateLimiter != nil {
					chatGroup.Use(middleware.RateLimitMiddleware(cfg.RateLimiter))
				} else if redisClient := redis.GetClient(); redisClient != nil {
					chatGroup.Use(middleware.RateLimitMiddleware(ratelimiter.NewRedisRateLimiter(redisClient)))
				}
				if cfg.AdmissionService != nil {
					chatGroup.Use(middleware.AdmissionMiddleware(cfg.AdmissionService))
				}
				chatGroup.POST("/chat/completions", cfg.InferenceController.HandleChatCompletion)
			}
		}
	}

	// ==================================================
	// 2. INFERENCE / DATA PLANE API (/v1) - JWT Auth
	// ==================================================
	inferenceV1 := r.Group("/v1")
	inferenceV1.Use(middleware.AuthMiddleware(cfg.AuthService))

	if cfg.RateLimiter != nil {
		inferenceV1.Use(middleware.RateLimitMiddleware(cfg.RateLimiter))
	} else if redisClient := redis.GetClient(); redisClient != nil {
		inferenceV1.Use(middleware.RateLimitMiddleware(ratelimiter.NewRedisRateLimiter(redisClient)))
	}

	if cfg.AdmissionService != nil {
		inferenceV1.Use(middleware.AdmissionMiddleware(cfg.AdmissionService))
	}
	{
		// Verification / ping endpoint for User Principal
		inferenceV1.GET("/auth/verify", func(c *gin.Context) {
			userID, _ := middleware.GetUserID(c)
			email, _ := middleware.GetUserEmail(c)
			c.JSON(http.StatusOK, gin.H{
				"authenticated": true,
				"principal": gin.H{
					"user_id": userID,
					"email":   email,
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
