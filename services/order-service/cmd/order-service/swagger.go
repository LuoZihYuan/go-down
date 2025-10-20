//go:build !prod

package main

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/LuoZihYuan/go-down/services/order-service/docs"
)

// registerSwagger registers Swagger UI endpoints
// This function is only compiled in dev and stage builds
func registerSwagger(router *gin.Engine) {
	swagger := router.Group("/swagger")
	{
		swagger.GET("/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}
}
