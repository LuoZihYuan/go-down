package handlers

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/LuoZihYuan/go-down/services/order-service/internal/client"
	"github.com/LuoZihYuan/go-down/services/order-service/internal/models"
)

// OrderHandler handles order-related requests
type OrderHandler struct {
	paymentClient *client.PaymentClient
	orders        map[string]*models.OrderResponse
	mu            sync.RWMutex
}

// NewOrderHandler creates a new order handler
func NewOrderHandler(paymentClient *client.PaymentClient) *OrderHandler {
	return &OrderHandler{
		paymentClient: paymentClient,
		orders:        make(map[string]*models.OrderResponse),
	}
}

// CreateOrder processes a new order
// @Summary Create order
// @Description Creates a new order and processes payment
// @Tags Orders
// @Accept json
// @Produce json
// @Param order body models.OrderRequest true "Order request"
// @Success 200 {object} models.OrderResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Failure 503 {object} models.ErrorResponse
// @Router /api/orders [post]
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req models.OrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Title:  "Bad Request",
			Status: http.StatusBadRequest,
			Detail: fmt.Sprintf("Invalid order request: %v", err),
		})
		return
	}

	// Generate order ID
	orderID := fmt.Sprintf("order-%s", uuid.New().String()[:8])

	// Process payment
	paymentReq := &models.PaymentRequest{
		OrderID: orderID,
		Amount:  req.Amount,
		Method:  "credit_card",
	}

	paymentResp, err := h.paymentClient.ProcessPayment(c.Request.Context(), paymentReq)
	if err != nil {
		// Handle different error types
		if err == client.ErrCircuitOpen {
			c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{
				Title:  "Service Unavailable",
				Status: http.StatusServiceUnavailable,
				Detail: "Payment service is temporarily unavailable (circuit breaker open)",
			})
			return
		}
		if err == client.ErrBulkheadFull {
			c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{
				Title:  "Service Unavailable",
				Status: http.StatusServiceUnavailable,
				Detail: "Too many concurrent payment requests (bulkhead full)",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Title:  "Internal Server Error",
			Status: http.StatusInternalServerError,
			Detail: fmt.Sprintf("Failed to process payment: %v", err),
		})
		return
	}

	// Create order response
	order := &models.OrderResponse{
		OrderID:    orderID,
		CustomerID: req.CustomerID,
		Amount:     req.Amount,
		Status:     "completed",
		PaymentID:  paymentResp.PaymentID,
		Items:      req.Items,
		CreatedAt:  time.Now(),
	}

	// Store order (in-memory for demo)
	h.mu.Lock()
	h.orders[orderID] = order
	h.mu.Unlock()

	c.JSON(http.StatusOK, order)
}

// GetOrder retrieves an order by ID
// @Summary Get order
// @Description Retrieves an order by ID
// @Tags Orders
// @Produce json
// @Param id path string true "Order ID"
// @Success 200 {object} models.OrderResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/orders/{id} [get]
func (h *OrderHandler) GetOrder(c *gin.Context) {
	orderID := c.Param("id")

	h.mu.RLock()
	order, exists := h.orders[orderID]
	h.mu.RUnlock()

	if !exists {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Title:  "Not Found",
			Status: http.StatusNotFound,
			Detail: fmt.Sprintf("Order %s not found", orderID),
		})
		return
	}

	c.JSON(http.StatusOK, order)
}
