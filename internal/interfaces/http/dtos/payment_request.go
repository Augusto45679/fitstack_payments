// Package dtos contains HTTP data transfer objects.
package dtos

// CreatePaymentRequestDTO represents the HTTP request for creating a payment.
type CreatePaymentRequestDTO struct {
	GymSlug           string  `json:"gym_slug" binding:"required"`
	Amount            float64 `json:"amount" binding:"required,gt=0"`
	Title             string  `json:"title" binding:"required"`
	Description       string  `json:"description"`
	PayerEmail        string  `json:"payer_email" binding:"required,email"`
	ExternalReference string  `json:"external_reference" binding:"required"`
	MPAccessToken     string  `json:"mp_access_token" binding:"required"`
	// Optional: Redirect URLs
	SuccessURL string `json:"success_url"`
	FailureURL string `json:"failure_url"`
	PendingURL string `json:"pending_url"`
}
