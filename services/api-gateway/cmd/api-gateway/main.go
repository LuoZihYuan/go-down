package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/LuoZihYuan/go-down/services/api-gateway/internal/client"
	"github.com/LuoZihYuan/go-down/services/api-gateway/internal/handlers"
	"github.com/LuoZihYuan/go-down/services/api-gateway/internal/middleware"
)

// @title API Gateway
// @version 1.0
// @description API Gateway with resilience patterns
// @host localhost:8080
// @BasePath /

func main() {
	// Get order service URL from environment
	orderServiceURL := os.Getenv("ORDER_SERVICE_URL")
	if orderServiceURL == "" {
		log.Fatal("ORDER_SERVICE_URL environment variable is required")
	}

	// Setup router
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.MetricsMiddleware())

	// Initialize clients
	orderClient := client.NewOrderClient(orderServiceURL)

	// Root group
	rootHandler := handlers.NewRootHandler()
	root := router.Group("/")
	{
		root.GET("/health", rootHandler.Health)
		root.GET("/metrics", gin.WrapH(promhttp.Handler()))
	}

	// API group
	orderHandler := handlers.NewOrderHandler(orderClient)
	api := router.Group("/api")
	{
		api.POST("/orders", orderHandler.CreateOrder)
		api.GET("/orders/:id", orderHandler.GetOrder)
	}

	// Swagger group (conditionally registered based on build tags)
	registerSwagger(router)

	// Start server
	log.Println("API Gateway started")
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
