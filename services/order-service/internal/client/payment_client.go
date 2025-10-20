//go:build !stage

package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/LuoZihYuan/go-down/services/order-service/internal/models"
)

// PaymentClient handles communication with the payment service
// Resilient version: Includes timeout, circuit breaker, and bulkhead
type PaymentClient struct {
	httpClient     *http.Client
	baseURL        string
	circuitBreaker *CircuitBreaker[*models.PaymentResponse]
	bulkhead       *Bulkhead
}

// NewPaymentClient creates a new resilient payment client
func NewPaymentClient(baseURL string) *PaymentClient {
	return &PaymentClient{
		httpClient: &http.Client{
			Timeout: 3 * time.Second, // Fail fast timeout
		},
		baseURL: baseURL,
		// Circuit breaker: 5 failures in 10 seconds opens circuit for 30 seconds
		circuitBreaker: NewCircuitBreaker[*models.PaymentResponse]("payment", 5, 30*time.Second),
		// Bulkhead: Max 10 concurrent payment requests
		bulkhead: NewBulkhead("payment", 10),
	}
}

// ProcessPayment sends a payment request to the payment service with resilience patterns
func (c *PaymentClient) ProcessPayment(ctx context.Context, req *models.PaymentRequest) (*models.PaymentResponse, error) {
	var result *models.PaymentResponse

	// Execute with circuit breaker protection
	result, err := c.circuitBreaker.Execute(func() (*models.PaymentResponse, error) {
		// Execute with bulkhead protection
		var callErr error
		bulkheadErr := c.bulkhead.TryExecute(func() error {
			result, callErr = c.makePaymentCall(ctx, req)
			return callErr
		})

		// If bulkhead rejected, return that error
		if bulkheadErr != nil {
			return nil, bulkheadErr
		}

		// Return the actual call result
		return result, callErr
	})

	return result, err
}

// makePaymentCall performs the actual HTTP call
func (c *PaymentClient) makePaymentCall(ctx context.Context, req *models.PaymentRequest) (*models.PaymentResponse, error) {
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

	// Send request (with 3s timeout from httpClient)
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
