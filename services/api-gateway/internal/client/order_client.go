//go:build !stage

package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/LuoZihYuan/go-down/services/api-gateway/internal/models"
)

// OrderClient handles communication with the order service
// Resilient version: Includes timeout and circuit breaker
type OrderClient struct {
	httpClient     *http.Client
	baseURL        string
	circuitBreaker *CircuitBreaker[*models.OrderResponse]
}

// NewOrderClient creates a new resilient order client
func NewOrderClient(baseURL string) *OrderClient {
	return &OrderClient{
		httpClient: &http.Client{
			Timeout: 5 * time.Second, // 5s timeout for API Gateway
		},
		baseURL: baseURL,
		// Circuit breaker: 5 failures in 10 seconds opens circuit for 30 seconds
		circuitBreaker: NewCircuitBreaker[*models.OrderResponse]("order", 5, 30*time.Second),
	}
}

// CreateOrder sends an order creation request to the order service with resilience patterns
func (c *OrderClient) CreateOrder(ctx context.Context, req *models.OrderRequest) (*models.OrderResponse, error) {
	// Execute with circuit breaker protection
	return c.circuitBreaker.Execute(func() (*models.OrderResponse, error) {
		return c.makeCreateOrderCall(ctx, req)
	})
}

// GetOrder retrieves an order by ID from the order service
// Note: GET requests don't go through circuit breaker as they're read-only
func (c *OrderClient) GetOrder(ctx context.Context, orderID string) (*models.OrderResponse, error) {
	return c.makeGetOrderCall(ctx, orderID)
}

// makeCreateOrderCall performs the actual HTTP POST call
func (c *OrderClient) makeCreateOrderCall(ctx context.Context, req *models.OrderRequest) (*models.OrderResponse, error) {
	// Marshal request
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/orders", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	// Send request (with 5s timeout from httpClient)
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("order service returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse response
	var orderResp models.OrderResponse
	if err := json.NewDecoder(resp.Body).Decode(&orderResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &orderResp, nil
}

// makeGetOrderCall performs the actual HTTP GET call
func (c *OrderClient) makeGetOrderCall(ctx context.Context, orderID string) (*models.OrderResponse, error) {
	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/api/orders/"+orderID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Send request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("order not found")
	}
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("order service returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse response
	var orderResp models.OrderResponse
	if err := json.NewDecoder(resp.Body).Decode(&orderResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &orderResp, nil
}
