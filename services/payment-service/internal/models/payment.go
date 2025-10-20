package models

import "time"

// PaymentRequest represents an incoming payment request
// @Description Payment processing request
type PaymentRequest struct {
	OrderID string  `json:"order_id" binding:"required" example:"order-123"`
	Amount  float64 `json:"amount" binding:"required,gt=0" example:"99.99"`
	Method  string  `json:"method" binding:"required,oneof=credit_card debit_card paypal" example:"credit_card"`
} // @name PaymentRequest

// PaymentResponse represents a payment processing result
// @Description Payment processing response
type PaymentResponse struct {
	PaymentID     string    `json:"payment_id" example:"pay-abc123"`
	OrderID       string    `json:"order_id" example:"order-123"`
	Amount        float64   `json:"amount" example:"99.99"`
	Status        string    `json:"status" example:"completed"`
	TransactionID string    `json:"transaction_id" example:"txn-xyz789"`
	ProcessedAt   time.Time `json:"processed_at" example:"2025-01-15T10:30:00Z"`
} // @name PaymentResponse
