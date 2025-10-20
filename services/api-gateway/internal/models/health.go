package models

// HealthResponse represents health check result
// @Description Service health status
type HealthResponse struct {
	Status string `json:"status" example:"healthy"`
} // @name HealthResponse
