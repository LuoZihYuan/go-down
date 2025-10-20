package handlers

import (
	"net/http"

	"github.com/LuoZihYuan/go-down/services/api-gateway/internal/models"

	"github.com/gin-gonic/gin"
)

// RootHandler handles root-level endpoints
type RootHandler struct{}

// NewRootHandler creates a new root handler
func NewRootHandler() *RootHandler {
	return &RootHandler{}
}

// Health returns service health status
// @Summary Health check
// @Description Returns service health status
// @Tags Root
// @Produce json
// @Success 200 {object} models.HealthResponse
// @Router /health [get]
func (h *RootHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, models.HealthResponse{
		Status: "healthy",
	})
}
