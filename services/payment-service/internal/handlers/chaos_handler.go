package handlers

import (
	"fmt"
	"net/http"

	"github.com/LuoZihYuan/go-down/services/payment-service/internal/fault"
	"github.com/LuoZihYuan/go-down/services/payment-service/internal/models"

	"github.com/gin-gonic/gin"
)

// ChaosHandler handles chaos injection endpoints
type ChaosHandler struct {
	faultInjector *fault.Injector
}

// NewChaosHandler creates a new chaos handler
func NewChaosHandler(injector *fault.Injector) *ChaosHandler {
	return &ChaosHandler{
		faultInjector: injector,
	}
}

// EnableChaos enables fault injection
// @Summary Enable chaos injection
// @Description Enables fault injection with specified delay
// @Tags Chaos
// @Accept json
// @Produce json
// @Param chaos body models.ChaosRequest true "Chaos configuration"
// @Success 200 {object} models.ChaosStatus
// @Failure 400 {object} models.ErrorResponse
// @Router /chaos/enable [post]
func (h *ChaosHandler) EnableChaos(c *gin.Context) {
	var req models.ChaosRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Title:  "Bad Request",
			Status: http.StatusBadRequest,
			Detail: fmt.Sprintf("Invalid chaos configuration: %v", err),
		})
		return
	}

	h.faultInjector.Enable(req.DelaySeconds)

	c.JSON(http.StatusOK, models.ChaosStatus{
		Enabled:      true,
		DelaySeconds: req.DelaySeconds,
	})
}

// DisableChaos disables fault injection
// @Summary Disable chaos injection
// @Description Disables fault injection
// @Tags Chaos
// @Produce json
// @Success 200 {object} models.ChaosStatus
// @Router /chaos/disable [post]
func (h *ChaosHandler) DisableChaos(c *gin.Context) {
	h.faultInjector.Disable()

	c.JSON(http.StatusOK, models.ChaosStatus{
		Enabled:      false,
		DelaySeconds: 0,
	})
}

// GetChaosStatus returns current chaos injection status
// @Summary Get chaos status
// @Description Returns current fault injection configuration
// @Tags Chaos
// @Produce json
// @Success 200 {object} models.ChaosStatus
// @Router /chaos/status [get]
func (h *ChaosHandler) GetChaosStatus(c *gin.Context) {
	enabled, delay := h.faultInjector.GetStatus()

	c.JSON(http.StatusOK, models.ChaosStatus{
		Enabled:      enabled,
		DelaySeconds: delay,
	})
}
