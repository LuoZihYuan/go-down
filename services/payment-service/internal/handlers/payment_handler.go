package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/LuoZihYuan/go-down/services/payment-service/internal/fault"
	"github.com/LuoZihYuan/go-down/services/payment-service/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// PaymentHandler handles payment-related requests
type PaymentHandler struct {
	faultInjector *fault.Injector
}

// NewPaymentHandler creates a new payment handler
func NewPaymentHandler(injector *fault.Injector) *PaymentHandler {
	return &PaymentHandler{
		faultInjector: injector,
	}
}

// ProcessPayment processes a payment request
// @Summary Process payment
// @Description Processes a payment transaction (returns mock data)
// @Tags Payments
// @Accept json
// @Produce json
// @Param payment body models.PaymentRequest true "Payment request"
// @Success 200 {object} models.PaymentResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/payments [post]
func (h *PaymentHandler) ProcessPayment(c *gin.Context) {
	var req models.PaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Title:  "Bad Request",
			Status: http.StatusBadRequest,
			Detail: fmt.Sprintf("Invalid payment request: %v", err),
		})
		return
	}

	// Inject fault if chaos is enabled
	h.faultInjector.Inject()

	// Generate mock payment response
	response := models.PaymentResponse{
		PaymentID:     fmt.Sprintf("pay-%s", uuid.New().String()[:8]),
		OrderID:       req.OrderID,
		Amount:        req.Amount,
		Status:        "completed",
		TransactionID: fmt.Sprintf("txn-%s", uuid.New().String()[:8]),
		ProcessedAt:   time.Now(),
	}

	c.JSON(http.StatusOK, response)
}
