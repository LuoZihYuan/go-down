//go:build stage

package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/LuoZihYuan/go-down/services/order-service/internal/models"
)

// PaymentClient handles communication with the payment service
// Stage version: No resilience patterns, no timeout
type PaymentClient struct {
	httpClient *http.Client
	baseURL    string
}

// NewPaymentClient creates a new payment client
func NewPaymentClient(baseURL string) *PaymentClient {
	return &PaymentClient{
		httpClient: &http.Client{
			// No timeout in stage - allows full cascade failure
		},
		baseURL: baseURL,
	}
}

// ProcessPayment sends a payment request to the payment service
func (c *PaymentClient) ProcessPayment(ctx context.Context, req *models.PaymentRequest) (*models.PaymentResponse, error) {
	// Marshal request
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/payments", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	// Send request (no timeout, no circuit breaker, no bulkhead)
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("payment service returned status %d", resp.StatusCode)
	}

	// Parse response
	var paymentResp models.PaymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&paymentResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &paymentResp, nil
}
