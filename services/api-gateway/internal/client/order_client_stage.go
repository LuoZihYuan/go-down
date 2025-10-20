//go:build stage

package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/LuoZihYuan/go-down/services/api-gateway/internal/models"
)

// OrderClient handles communication with the order service
// Stage version: No resilience patterns, no timeout
type OrderClient struct {
	httpClient *http.Client
	baseURL    string
}

// NewOrderClient creates a new order client
func NewOrderClient(baseURL string) *OrderClient {
	return &OrderClient{
		httpClient: &http.Client{
			// No timeout in stage - allows full cascade failure
		},
		baseURL: baseURL,
	}
}

// CreateOrder sends an order creation request to the order service
func (c *OrderClient) CreateOrder(ctx context.Context, req *models.OrderRequest) (*models.OrderResponse, error) {
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

	// Send request (no timeout, no circuit breaker)
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

// GetOrder retrieves an order by ID from the order service
func (c *OrderClient) GetOrder(ctx context.Context, orderID string) (*models.OrderResponse, error) {
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
