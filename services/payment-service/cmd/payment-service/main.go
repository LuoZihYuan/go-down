package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/LuoZihYuan/go-down/services/payment-service/internal/fault"
	"github.com/LuoZihYuan/go-down/services/payment-service/internal/handlers"
	"github.com/LuoZihYuan/go-down/services/payment-service/internal/middleware"
)

// @title Payment Service API
// @version 1.0
// @description Payment processing service with chaos injection capabilities
// @host localhost:8082
// @BasePath /

func main() {

	// Setup router
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.MetricsMiddleware())

	// Initialize fault injector
	faultInjector := fault.NewInjector()

	// Root group
	rootHandler := handlers.NewRootHandler()
	root := router.Group("/")
	{
		root.GET("/health", rootHandler.Health)
		root.GET("/metrics", gin.WrapH(promhttp.Handler()))
	}

	// API group
	paymentHandler := handlers.NewPaymentHandler(faultInjector)
	api := router.Group("/api")
	{
		api.POST("/payments", paymentHandler.ProcessPayment)
	}

	// Chaos group
	chaosHandler := handlers.NewChaosHandler(faultInjector)
	chaos := router.Group("/chaos")
	{
		chaos.POST("/enable", chaosHandler.EnableChaos)
		chaos.POST("/disable", chaosHandler.DisableChaos)
		chaos.GET("/status", chaosHandler.GetChaosStatus)
	}

	// Swagger group (conditionally registered based on build tags)
	registerSwagger(router)

	// Start server
	log.Println("Payment Service started on :8082")
	if err := router.Run(":8082"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
