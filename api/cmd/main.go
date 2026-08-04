package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/statuspage/api/internal/auth"
	"github.com/statuspage/api/internal/config"
	"github.com/statuspage/api/internal/db"
	"github.com/statuspage/api/internal/handlers"
	"github.com/statuspage/api/internal/k8s"
)

func main() {
	cfg := config.LoadConfig()

	log.Println("Starting StatusPage API Service...")

	// Initialize Database
	database, err := db.InitDB(cfg)
	if err != nil {
		log.Fatalf("Fatal: Database initialization failed: %v", err)
	}

	// Initialize Kubernetes Client
	k8sClient, err := k8s.NewK8sClient(cfg.KubeConfig)
	if err != nil {
		log.Fatalf("Fatal: Kubernetes client initialization failed: %v", err)
	}

	// Handlers
	authHandler := handlers.NewAuthHandler(database, cfg.JWTSecret)
	monitorHandler := handlers.NewMonitorHandler(database, k8sClient)

	// Gin Engine
	router := gin.Default()

	// CORS / Health check
	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "UP", "service": "statuspage-api"})
	})

	// Public Routes
	v1 := router.Group("/api/v1")
	{
		authGroup := v1.Group("/auth")
		{
			authGroup.POST("/signup", authHandler.Signup)
			authGroup.POST("/login", authHandler.Login)
		}

		// Authenticated Routes
		protected := v1.Group("")
		protected.Use(auth.AuthMiddleware(cfg.JWTSecret))
		{
			monitors := protected.Group("/monitors")
			{
				monitors.POST("", monitorHandler.CreateMonitor)
				monitors.GET("", monitorHandler.ListMonitors)
				monitors.GET("/:id", monitorHandler.GetMonitor)
				monitors.DELETE("/:id", monitorHandler.DeleteMonitor)
			}
		}
	}

	log.Printf("StatusPage API listening on port %s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Fatal: Server failed to start: %v", err)
	}
}
