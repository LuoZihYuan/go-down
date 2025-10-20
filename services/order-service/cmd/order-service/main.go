package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/LuoZihYuan/go-down/services/order-service/internal/client"
	"github.com/LuoZihYuan/go-down/services/order-service/internal/handlers"
	"github.com/LuoZihYuan/go-down/services/order-service/internal/middleware"
)

// @title Order Service API
// @version 1.0
// @description Order processing service with resilience patterns
// @host localhost:8081
// @BasePath /

func main() {
	// Get payment service URL from environment
	paymentServiceURL := os.Getenv("PAYMENT_SERVICE_URL")
	if paymentServiceURL == "" {
		log.Fatal("PAYMENT_SERVICE_URL environment variable is required")
	}

	// Setup router
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.MetricsMiddleware())

	// Initialize clients
	paymentClient := client.NewPaymentClient(paymentServiceURL)

	// Root group
	rootHandler := handlers.NewRootHandler()
	root := router.Group("/")
	{
		root.GET("/health", rootHandler.Health)
		root.GET("/metrics", gin.WrapH(promhttp.Handler()))
	}

	// Swagger group (conditionally registered based on build tags)
	registerSwagger(router)

	// API group
	orderHandler := handlers.NewOrderHandler(paymentClient)
	api := router.Group("/api")
	{
		api.POST("/orders", orderHandler.CreateOrder)
		api.GET("/orders/:id", orderHandler.GetOrder)
	}

	// Start server
	log.Println("Order Service started")
	if err := router.Run(":8081"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
