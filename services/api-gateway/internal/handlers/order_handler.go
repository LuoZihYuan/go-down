package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/LuoZihYuan/go-down/services/api-gateway/internal/client"
	"github.com/LuoZihYuan/go-down/services/api-gateway/internal/models"
)

// OrderHandler handles order-related requests by proxying to order service
type OrderHandler struct {
	orderClient *client.OrderClient
}

// NewOrderHandler creates a new order handler
func NewOrderHandler(orderClient *client.OrderClient) *OrderHandler {
	return &OrderHandler{
		orderClient: orderClient,
	}
}

// CreateOrder proxies order creation to the order service
// @Summary Create order
// @Description Creates a new order via order service
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

	// Proxy to order service
	order, err := h.orderClient.CreateOrder(c.Request.Context(), &req)
	if err != nil {
		// Handle circuit breaker error
		if err == client.ErrCircuitOpen {
			c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{
				Title:  "Service Unavailable",
				Status: http.StatusServiceUnavailable,
				Detail: "Order service is temporarily unavailable (circuit breaker open)",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Title:  "Internal Server Error",
			Status: http.StatusInternalServerError,
			Detail: fmt.Sprintf("Failed to create order: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, order)
}

// GetOrder proxies order retrieval to the order service
// @Summary Get order
// @Description Retrieves an order by ID via order service
// @Tags Orders
// @Produce json
// @Param id path string true "Order ID"
// @Success 200 {object} models.OrderResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/orders/{id} [get]
func (h *OrderHandler) GetOrder(c *gin.Context) {
	orderID := c.Param("id")

	// Proxy to order service
	order, err := h.orderClient.GetOrder(c.Request.Context(), orderID)
	if err != nil {
		// Check if order not found
		if err.Error() == "order not found" {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Title:  "Not Found",
				Status: http.StatusNotFound,
				Detail: fmt.Sprintf("Order %s not found", orderID),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Title:  "Internal Server Error",
			Status: http.StatusInternalServerError,
			Detail: fmt.Sprintf("Failed to retrieve order: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, order)
}
