package main

import (
	"log"

	"github.com/Faithful001/aegis/internal/domain/auth"
	"github.com/Faithful001/aegis/internal/domain/user"
	"github.com/Faithful001/aegis/internal/infra/db"
	"github.com/Faithful001/aegis/internal/infra/redis"
	"github.com/Faithful001/aegis/internal/router"
	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize PostgreSQL database
	db.InitDB()

	// Initialize Redis cache
	redis.InitRedis()

	// Run auto migrations
	if err := db.AutoMigrate(&user.User{}, &auth.BlacklistedToken{}); err != nil {
		log.Printf("Failed to run database migrations: %v", err)
	}

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"success": true,
			"message": "server up and running",
			"data": gin.H{
				"database": "operational",
				"server":   "operational",
			},
		})
	})

	router.SetupRoutes(r)

	if err := r.Run(":5000"); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}