package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Faithful001/aegis/internal/domain/auth"
	"github.com/Faithful001/aegis/internal/domain/inference"
	"github.com/Faithful001/aegis/internal/domain/organization"
	"github.com/Faithful001/aegis/internal/domain/project"
	"github.com/Faithful001/aegis/internal/domain/user"
	infraAuth "github.com/Faithful001/aegis/internal/infra/auth"
	"github.com/Faithful001/aegis/internal/infra/config"
	"github.com/Faithful001/aegis/internal/infra/db"
	"github.com/Faithful001/aegis/internal/infra/observability"
	"github.com/Faithful001/aegis/internal/infra/redis"
	"github.com/Faithful001/aegis/internal/router"
)

func main() {
	// 1. Load Configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// 2. Initialize Structured Logger
	logger := observability.InitLogger(cfg.Log.Level, cfg.Log.Format)
	logger.Info("Starting Aegis AI Inference Platform Control Plane", "environment", cfg.Environment)

	// 3. Initialize Database (PostgreSQL)
	db.InitDB()
	database := db.GetDB()

	// Run auto migrations for core entities
	if database != nil {
		migrationModels := []interface{}{
			&user.User{},
			&auth.BlacklistedToken{},
			&organization.Organization{},
			&organization.OrganizationMember{},
			&project.Project{},
			&auth.APIKey{},
		}
		if err := db.AutoMigrate(migrationModels...); err != nil {
			logger.Error("Failed to run database migrations", "error", err)
		} else {
			logger.Info("Database migrations completed successfully")
		}
	} else {
		logger.Warn("Database is not connected. Operating in degraded mode.")
	}

	// 4. Initialize Distributed Cache / Store (Redis)
	redis.InitRedis()
	redisClient := redis.GetClient()
	if redisClient != nil {
		logger.Info("Redis connection initialized successfully")
	} else {
		logger.Warn("Redis is not connected. Ephemeral state operating with fallback.")
	}

	// 5. Initialize Domain Repositories
	userRepo := user.NewGormRepository(database)
	blacklistRepo := auth.NewTokenBlacklistRepository(redisClient, database)
	orgRepo := organization.NewGormRepository(database)
	projectRepo := project.NewGormRepository(database)
	apiKeyRepo := auth.NewAPIKeyRepository(database)

	// 6. Initialize Domain Hasher, Generators & Services
	hasher := infraAuth.NewBcryptHasher(0)
	apiKeyGenerator := auth.NewAPIKeyGenerator()
	tokenService := infraAuth.NewJWTService(
		cfg.JWT.Secret,
		cfg.JWT.Issuer,
		cfg.JWT.AccessTokenExpiry,
		cfg.JWT.RefreshTokenExpiry,
	)

	authService := auth.NewAuthService(userRepo, blacklistRepo, tokenService, hasher)
	apiKeyService := auth.NewAPIKeyService(apiKeyRepo, apiKeyGenerator)
	orgService := organization.NewService(orgRepo, userRepo)
	projectService := project.NewService(projectRepo, orgRepo)

	// 7. Initialize Inference Worker Client & Service
	var workerClient inference.WorkerClient
	grpcWorker, err := inference.NewGRPCWorkerClient(cfg.Worker.InferenceWorkerAddr)
	if err == nil {
		workerClient = grpcWorker
		logger.Info("Inference worker gRPC client initialized", "target", cfg.Worker.InferenceWorkerAddr)
	} else {
		logger.Warn("Could not connect to gRPC inference worker, using fallback mock client", "error", err)
		workerClient = inference.NewMockWorkerClient()
	}
	defer workerClient.Close()

	inferenceService := inference.NewService(workerClient)

	// 8. Initialize Domain Controllers
	authController := auth.NewAuthController(authService)
	apiKeyController := auth.NewAPIKeyController(apiKeyService)
	orgController := organization.NewController(orgService)
	projectController := project.NewController(projectService)
	inferenceController := inference.NewController(inferenceService)

	// 9. Setup HTTP Engine
	engine := router.SetupRouter(router.RouterConfig{
		AuthService:         authService,
		APIKeyService:       apiKeyService,
		AuthController:      authController,
		APIKeyController:    apiKeyController,
		OrgController:       orgController,
		ProjectController:   projectController,
		InferenceController: inferenceController,
	})

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      engine,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// 9. Start Server in Background Goroutine
	go func() {
		logger.Info("Aegis HTTP server listening", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("HTTP server failed unexpectedly", "error", err)
			os.Exit(1)
		}
	}()

	// 10. Graceful Shutdown Listener (SIGINT, SIGTERM)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Graceful shutdown initiated...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
	}

	if redisClient != nil {
		if err := redisClient.Close(); err != nil {
			logger.Warn("Error closing Redis client", "error", err)
		}
	}

	if database != nil {
		if sqlDB, err := database.DB(); err == nil {
			if err := sqlDB.Close(); err != nil {
				logger.Warn("Error closing database connection", "error", err)
			}
		}
	}

	logger.Info("Aegis control plane stopped gracefully")
}