package models

import "time"

// OrderRequest represents an incoming order request
// @Description Order creation request
type OrderRequest struct {
	CustomerID string  `json:"customer_id" binding:"required" example:"cust-123"`
	Amount     float64 `json:"amount" binding:"required,gt=0" example:"99.99"`
	Items      []Item  `json:"items" binding:"required,min=1"`
} // @name OrderRequest

// Item represents an order item
// @Description Order line item
type Item struct {
	ProductID string  `json:"product_id" binding:"required" example:"prod-456"`
	Quantity  int     `json:"quantity" binding:"required,gt=0" example:"2"`
	Price     float64 `json:"price" binding:"required,gt=0" example:"49.99"`
} // @name Item

// OrderResponse represents an order processing result
// @Description Order processing response
type OrderResponse struct {
	OrderID    string    `json:"order_id" example:"order-abc123"`
	CustomerID string    `json:"customer_id" example:"cust-123"`
	Amount     float64   `json:"amount" example:"99.99"`
	Status     string    `json:"status" example:"completed"`
	PaymentID  string    `json:"payment_id" example:"pay-xyz789"`
	Items      []Item    `json:"items"`
	CreatedAt  time.Time `json:"created_at" example:"2025-01-15T10:30:00Z"`
} // @name OrderResponse
