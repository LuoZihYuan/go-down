package models

// ErrorResponse represents an error response
// @Description Standard error response
type ErrorResponse struct {
	Title  string `json:"title" example:"Bad Request"`
	Status int    `json:"status" example:"400"`
	Detail string `json:"detail" example:"Invalid payment request"`
} // @name ErrorResponse
