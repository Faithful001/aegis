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
	"github.com/Faithful001/aegis/internal/domain/otp"
	"github.com/Faithful001/aegis/internal/domain/provider"
	"github.com/Faithful001/aegis/internal/domain/scheduler"
	"github.com/Faithful001/aegis/internal/domain/usage"
	"github.com/Faithful001/aegis/internal/domain/user"
	"github.com/Faithful001/aegis/internal/domain/worker"
	"github.com/Faithful001/aegis/internal/infra/config"
	"github.com/Faithful001/aegis/internal/infra/db"
	infraEmail "github.com/Faithful001/aegis/internal/infra/email"
	"github.com/Faithful001/aegis/internal/infra/observability"
	infraProvider "github.com/Faithful001/aegis/internal/infra/provider"
	events "github.com/Faithful001/aegis/internal/infra/queue/kafka"
	"github.com/Faithful001/aegis/internal/infra/redis"
	infraAuth "github.com/Faithful001/aegis/internal/infra/security"
	"github.com/Faithful001/aegis/internal/infra/workerregistry"
	"github.com/Faithful001/aegis/internal/router"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	logger := observability.InitLogger(cfg.Log.Level, cfg.Log.Format)
	logger.Info("Starting Aegis AI Inference Platform Control Plane", "environment", cfg.Environment)

	// initialize db (postgres)
	db.InitDB()
	database := db.GetDB()

	// run auto migrations for the core entities
	if database != nil {
		migrationModels := []interface{}{
			&user.User{},
			&auth.BlacklistedToken{},
			&otp.OTPRecord{},
			&usage.UsageRecord{},
			&provider.ProviderCredential{},
			&provider.UserModel{},
		}
		if err := db.AutoMigrate(migrationModels...); err != nil {
			logger.Error("Failed to run database migrations", "error", err)
		} else {
			logger.Info("Database migrations completed successfully")
		}
	} else {
		logger.Warn("Database is not connected. Operating in degraded mode.")
	}

	// initialize redis
	redis.InitRedis()
	redisClient := redis.GetClient()
	if redisClient != nil {
		logger.Info("Redis connection initialized successfully")
	} else {
		logger.Warn("Redis is not connected. Ephemeral state operating with fallback.")
	}

	// initialize kafka
	var eventProducer events.EventProducer = events.NewKafkaProducer(cfg.Kafka.Brokers, logger)
	defer eventProducer.Close()
	events.SetBrokers(cfg.Kafka.Brokers)
	logger.Info("Kafka event infrastructure initialized", "brokers", cfg.Kafka.Brokers)

	workerReg := workerregistry.NewRedisWorkerRegistry(redisClient)
	logger.Info("Worker registry backed by Redis")
	
	workerSvc := worker.NewWorkerService(workerReg, worker.DefaultHeartbeatMonitorConfig(), logger)
	workerSvc.StartHeartbeatMonitor(context.Background())
	defer workerSvc.StopHeartbeatMonitor()
	logger.Info("Worker heartbeat monitor started")

	// initialize domain repositories
	userRepo := user.NewUserRepository(database)
	blacklistRepo := auth.NewTokenBlacklistRepository(redisClient, database)
	usageRepo := usage.NewUsageRepository(database)
	providerRepo := provider.NewProviderCredentialRepository(database)
	otpRepo := otp.NewOTPRepository(database)

	// initialize domain hasher, generators and services
	hasher := infraAuth.NewBcryptHasher(0)
	tokenService := infraAuth.NewJWTService(
		cfg.JWT.Secret,
		cfg.JWT.Issuer,
		cfg.JWT.AccessTokenExpiry,
		cfg.JWT.RefreshTokenExpiry,
	)

	emailService := infraEmail.NewEmailService(cfg.Resend.APIKey, cfg.Resend.FromAddr)
	
	otpService := otp.NewOTPService(userRepo, otpRepo, hasher)
	
	authService := auth.NewAuthService(
		userRepo,
		otpRepo,
		otpService,
		emailService,
		tokenService,
		hasher,
		blacklistRepo,
		eventProducer,
	)

	userService := user.NewUserService(userRepo)
	usageService := usage.NewUsageService(usageRepo, logger)
	providerService := provider.NewProviderCredentialService(providerRepo, cfg.JWT.Secret)

	// initialize email + OTP service
	if cfg.Resend.APIKey != "" {
		logger.Info("Resend email service initialized", "from", cfg.Resend.FromAddr)
	} else {
		logger.Warn("RESEND_API_KEY is not set — email features will be unavailable")
	}
	
	// authService.SetEmailService(emailSvc)
	// authService.SetOTP(otpRepo, otpService)

	// initialize event consumer (kafka)
	var eventConsumer events.EventConsumer = events.NewKafkaConsumer(cfg.Kafka.Brokers, cfg.Kafka.GroupID, logger)
	defer eventConsumer.Close()
	meteringConsumer := usage.NewMeteringConsumer(eventConsumer, usageService, logger)
	if err := meteringConsumer.Start(context.Background()); err != nil {
		logger.Warn("Could not start metering consumer background listener", "error", err)
	}

	emailConsumer := auth.NewEmailConsumer(eventConsumer, emailService, logger)
	if err := emailConsumer.Start(context.Background()); err != nil {
		logger.Warn("Could not start email consumer background listener", "error", err)
	}

	// initialize worker client, capacity scheduler & byok provider gateway
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

	// setup frontier provider router
	providerRouter := infraProvider.NewProviderRouter()
	providerService.SetGateway(providerRouter)
	inferenceService.SetProviderGateway(providerService, providerRouter)
	logger.Info("BYOK Frontier Provider Gateway initialized (OpenAI, Anthropic, Gemini, Mistral AI, OpenRouter)")

	// initialize domain controllers
	authController := auth.NewAuthController(authService)
	inferenceController := inference.NewInferenceController(inferenceService)
	usageController := usage.NewUsageController(usageService)
	providerController := provider.NewProviderController(providerService)
	userController := user.NewUserController(userService)

	// setup http engine
	engine := router.SetupRouter(router.RouterConfig{
		AuthService:         authService,
		AuthController:      authController,
		InferenceController: inferenceController,
		UsageController:     usageController,
		ProviderController:  providerController,
		UserController:      userController,
	})

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      engine,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// start server in background goroutine
	go func() {
		logger.Info("Aegis HTTP server listening", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("HTTP server failed unexpectedly", "error", err)
			os.Exit(1)
		}
	}()

	// wait for interrupt signal
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