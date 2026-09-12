package db

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Note: .env file not found or failed to load: %v", err)
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Printf("Warning: DATABASE_URL environment variable is not set")
		return
	}

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
		return
	}

	log.Printf("Database connection successful")
}

func GetDB() *gorm.DB {
	return DB
}

func AutoMigrate(dst ...interface{}) error {
	if DB == nil {
		return nil
	}
	return DB.AutoMigrate(dst...)
}