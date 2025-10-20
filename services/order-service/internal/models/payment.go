package models

import "time"

// PaymentRequest represents a payment request to payment service
type PaymentRequest struct {
	OrderID string  `json:"order_id"`
	Amount  float64 `json:"amount"`
	Method  string  `json:"method"`
}

// PaymentResponse represents a payment response from payment service
type PaymentResponse struct {
	PaymentID     string    `json:"payment_id"`
	OrderID       string    `json:"order_id"`
	Amount        float64   `json:"amount"`
	Status        string    `json:"status"`
	TransactionID string    `json:"transaction_id"`
	ProcessedAt   time.Time `json:"processed_at"`
}
