package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/statuspage/api/internal/auth"
	"github.com/statuspage/api/internal/config"
	"github.com/statuspage/api/internal/db"
	"github.com/statuspage/api/internal/handlers"
	"github.com/statuspage/api/internal/k8s"
	"github.com/statuspage/api/internal/metrics"
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
	alertHandler := handlers.NewAlertHandler(database)

	// Gin Engine
	router := gin.Default()

	// CORS & Prometheus Metrics Middleware
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(24)
			return
		}

		c.Next()

		// Record Prometheus metric
		statusStr := fmt.Sprintf("%d", c.Writer.Status())
		metrics.HttpRequestsTotal.WithLabelValues(c.Request.Method, c.FullPath(), statusStr).Inc()
	})

	// Health check
	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "UP", "service": "statuspage-api"})
	})

	// Prometheus Metrics Endpoint
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Public Routes
	v1 := router.Group("/api/v1")
	{
		v1.GET("/public/status/:tenant_id", monitorHandler.GetPublicStatusPage)

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

			alerts := protected.Group("/alerts")
			{
				alerts.POST("/webhooks", alertHandler.CreateWebhook)
				alerts.GET("/webhooks", alertHandler.ListWebhooks)
			}
		}
	}

	log.Printf("StatusPage API listening on port %s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Fatal: Server failed to start: %v", err)
	}
}
