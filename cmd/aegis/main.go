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
	"github.com/Faithful001/aegis/internal/domain/provider"
	"github.com/Faithful001/aegis/internal/domain/scheduler"
	"github.com/Faithful001/aegis/internal/domain/usage"
	"github.com/Faithful001/aegis/internal/domain/user"
	"github.com/Faithful001/aegis/internal/domain/worker"
	infraAuth "github.com/Faithful001/aegis/internal/infra/auth"
	"github.com/Faithful001/aegis/internal/infra/config"
	"github.com/Faithful001/aegis/internal/infra/db"
	"github.com/Faithful001/aegis/internal/infra/observability"
	infraProvider "github.com/Faithful001/aegis/internal/infra/provider"
	events "github.com/Faithful001/aegis/internal/infra/queue/kafka"
	"github.com/Faithful001/aegis/internal/infra/redis"
	"github.com/Faithful001/aegis/internal/infra/workerregistry"
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
			&usage.UsageRecord{},
			&provider.ProviderCredential{},
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

	// 4c. Initialize Kafka Event Producer (Phase 10)
	var eventProducer events.EventProducer = events.NewKafkaProducer(cfg.Kafka.Brokers, logger)
	defer eventProducer.Close()
	events.SetBrokers(cfg.Kafka.Brokers)
	logger.Info("Kafka event infrastructure initialized", "brokers", cfg.Kafka.Brokers)

	var workerReg worker.WorkerRegistry
	if redisClient != nil {
		workerReg = workerregistry.NewRedisWorkerRegistry(redisClient)
		logger.Info("Worker registry backed by Redis")
	} else {
		workerReg = workerregistry.NewMemoryWorkerRegistry()
		logger.Warn("Worker registry using in-memory fallback (not suitable for production)")
	}
	workerSvc := worker.NewWorkerService(workerReg, worker.DefaultHeartbeatMonitorConfig(), logger)
	workerSvc.StartHeartbeatMonitor(context.Background())
	defer workerSvc.StopHeartbeatMonitor()
	logger.Info("Worker heartbeat monitor started")

	// 5. Initialize Domain Repositories
	userRepo := user.NewUserRepository(database)
	blacklistRepo := auth.NewTokenBlacklistRepository(redisClient, database)
	orgRepo := organization.NewOrganizationRepository(database)
	projectRepo := project.NewProjectRepository(database)
	apiKeyRepo := auth.NewAPIKeyRepository(database)
	usageRepo := usage.NewUsageRepository(database)
	providerRepo := provider.NewProviderCredentialRepository(database)

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
	orgService := organization.NewOrganizationService(orgRepo, userRepo)
	projectService := project.NewProjectService(projectRepo, orgRepo)
	usageService := usage.NewUsageService(usageRepo, logger)
	providerService := provider.NewProviderCredentialService(providerRepo, cfg.JWT.Secret)

	// Start Metering Consumer (Phase 11)
	var eventConsumer events.EventConsumer = events.NewKafkaConsumer(cfg.Kafka.Brokers, cfg.Kafka.GroupID, logger)
	defer eventConsumer.Close()
	meteringConsumer := usage.NewMeteringConsumer(eventConsumer, usageService, logger)
	if err := meteringConsumer.Start(context.Background()); err != nil {
		logger.Warn("Could not start metering consumer background listener", "error", err)
	}

	// 7. Initialize Inference Worker Client, Capacity Scheduler & BYOK Provider Gateway
	sched := scheduler.NewWeightedScoreScheduler(workerReg, logger)

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

	inferenceService := inference.NewInferenceService(workerClient, sched, eventProducer)

	// Setup Frontier Provider Router
	providerRouter := infraProvider.NewProviderRouter()
	inferenceService.SetProviderGateway(providerService, providerRouter)
	logger.Info("BYOK Frontier Provider Gateway initialized (OpenAI, Anthropic, Gemini, Mistral AI)")

	// 8. Initialize Domain Controllers
	authController := auth.NewAuthController(authService)
	apiKeyController := auth.NewAPIKeyController(apiKeyService)
	orgController := organization.NewController(orgService)
	projectController := project.NewProjectController(projectService)
	inferenceController := inference.NewInferenceController(inferenceService)
	usageController := usage.NewUsageController(usageService)
	providerController := provider.NewProviderController(providerService)

	// 9. Setup HTTP Engine
	engine := router.SetupRouter(router.RouterConfig{
		AuthService:         authService,
		APIKeyService:       apiKeyService,
		AuthController:      authController,
		APIKeyController:    apiKeyController,
		OrgController:       orgController,
		ProjectController:   projectController,
		InferenceController: inferenceController,
		UsageController:     usageController,
		ProviderController:  providerController,
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