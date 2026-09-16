package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Environment string
	Port        string
	Database    DatabaseConfig
	Redis       RedisConfig
	JWT         JWTConfig
	Log         LogConfig
	Worker      WorkerConfig
	Kafka       KafkaConfig
}

type KafkaConfig struct {
	Brokers []string
	GroupID string
}

type WorkerConfig struct {
	InferenceWorkerAddr string
}

type DatabaseConfig struct {
	URL          string
	MaxOpenConns int
	MaxIdleConns int
	MaxIdleTime  time.Duration
}

type RedisConfig struct {
	URL      string
	Addr     string
	Password string
	DB       int
}

type JWTConfig struct {
	Secret             string
	AccessTokenExpiry  time.Duration
	RefreshTokenExpiry time.Duration
	Issuer             string
}

type LogConfig struct {
	Level  string
	Format string // "json" or "text"
}

func Load() (*Config, error) {
	// Attempt to load .env file if it exists, ignore if not present
	_ = godotenv.Load()

	cfg := &Config{
		Environment: getEnv("ENVIRONMENT", "development"),
		Port:        getEnv("PORT", "5000"),
		Database: DatabaseConfig{
			URL:          getEnv("DATABASE_URL", ""),
			MaxOpenConns: getEnvInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns: getEnvInt("DB_MAX_IDLE_CONNS", 10),
			MaxIdleTime:  getEnvDuration("DB_MAX_IDLE_TIME", 15*time.Minute),
		},
		Redis: RedisConfig{
			URL:      getEnv("REDIS_URL", ""),
			Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvInt("REDIS_DB", 0),
		},
		JWT: JWTConfig{
			Secret:             getEnv("JWT_SECRET", "aegis_default_jwt_secret_key_change_in_production"),
			AccessTokenExpiry:  getEnvDuration("ACCESS_TOKEN_EXPIRY", 15*time.Minute),
			RefreshTokenExpiry: getEnvDuration("REFRESH_TOKEN_EXPIRY", 7*24*time.Hour),
			Issuer:             getEnv("JWT_ISSUER", "aegis"),
		},
		Log: LogConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
		Worker: WorkerConfig{
			InferenceWorkerAddr: getEnv("INFERENCE_WORKER_ADDR", "localhost:50051"),
		},
		Kafka: KafkaConfig{
			Brokers: []string{getEnv("KAFKA_BROKERS", "localhost:9092")},
			GroupID: getEnv("KAFKA_GROUP_ID", "aegis-consumer-group"),
		},
	}

	return cfg, nil
}

func getEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return defaultVal
}

func getEnvDuration(key string, defaultVal time.Duration) time.Duration {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		if dur, err := time.ParseDuration(val); err == nil {
			return dur
		}
		if intVal, err := strconv.Atoi(val); err == nil {
			return time.Duration(intVal) * time.Second
		}
		fmt.Printf("Warning: invalid duration for %s=%s, using default %v\n", key, val, defaultVal)
	}
	return defaultVal
}
