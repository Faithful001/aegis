package main

import "github.com/gin-gonic/gin"

func main() {
	r := gin.Default()

	r.GET("/health", func (c *gin.Context) {
		c.JSON(200, gin.H {
			"success": true, 
			"message": "server up and running",
			"data": gin.H{
				"database": "operational",
				"server":   "operational",
			},
		})
	})

	r.Run(":5000")
}