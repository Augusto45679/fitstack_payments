// Package dtos contains HTTP data transfer objects.
package dtos

// PaymentResponseDTO represents the HTTP response for payment operations.
type PaymentResponseDTO struct {
	Success          bool   `json:"success"`
	PreferenceID     string `json:"preference_id,omitempty"`
	InitPoint        string `json:"init_point,omitempty"`
	SandboxInitPoint string `json:"sandbox_init_point,omitempty"`
	Error            string `json:"error,omitempty"`
	ErrorCode        string `json:"error_code,omitempty"`
}

// ErrorResponseDTO represents a generic error response.
type ErrorResponseDTO struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Code    string `json:"code"`
}
