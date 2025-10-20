//go:build prod

package main

import (
	"github.com/gin-gonic/gin"
)

// registerSwagger is a no-op in production builds
// Swagger is completely excluded from the production binary
func registerSwagger(router *gin.Engine) {
	// No-op - Swagger not included in production
}
