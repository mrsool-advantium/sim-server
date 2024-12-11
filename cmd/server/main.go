package main

import (
	"log"
	"sim-server/config"
	"sim-server/database"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Initialize redis
	redisDb := database.CheckRedisConnection(cfg)

	app := &Application{
		Config: &cfg,
		Redis:  redisDb,
	}

	// Create a new Gin router
	router := gin.Default()

	router.Use(corsMiddleware())

	// Initialize routes
	app.registerRoutes(router)

	// Start the server
	if err := router.Run(cfg.ServerAddress); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		//TODO:: limit the origin when we have fixed origin
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
