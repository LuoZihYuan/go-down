package models

// ChaosRequest represents chaos injection configuration
// @Description Chaos injection settings
type ChaosRequest struct {
	DelaySeconds int `json:"delay_seconds" binding:"required,min=1,max=300" example:"30"`
} // @name ChaosRequest

// ChaosStatus represents current chaos state
// @Description Current chaos injection status
type ChaosStatus struct {
	Enabled      bool `json:"enabled" example:"true"`
	DelaySeconds int  `json:"delay_seconds" example:"30"`
} // @name ChaosStatus
