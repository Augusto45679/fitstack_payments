// Package domain contains the core business entities for the payment service.
// This is the innermost layer - no external dependencies.
package domain

import "time"

// PaymentRequest represents an incoming checkout request from Django.
// Note: mp_access_token is passed in-body (server-to-server communication).
type PaymentRequest struct {
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

// PaymentResponse represents the response after creating a payment preference.
type PaymentResponse struct {
	Success          bool   `json:"success"`
	PreferenceID     string `json:"preference_id,omitempty"`
	InitPoint        string `json:"init_point,omitempty"`
	SandboxInitPoint string `json:"sandbox_init_point,omitempty"`
	Error            string `json:"error,omitempty"`
	ErrorCode        string `json:"error_code,omitempty"`
}

// WebhookNotification represents the IPN notification from Mercado Pago.
type WebhookNotification struct {
	ID          int64  `json:"id"`
	LiveMode    bool   `json:"live_mode"`
	Type        string `json:"type"`
	DateCreated string `json:"date_created"`
	UserID      int64  `json:"user_id,omitempty"`
	APIVersion  string `json:"api_version"`
	Action      string `json:"action"`
	Data        struct {
		ID string `json:"id"`
	} `json:"data"`
}

// PaymentInfo contains the details of a confirmed payment.
type PaymentInfo struct {
	PaymentID         string    `json:"payment_id"`
	Status            string    `json:"status"`
	StatusDetail      string    `json:"status_detail"`
	ExternalReference string    `json:"external_reference"`
	Amount            float64   `json:"amount"`
	Currency          string    `json:"currency"`
	PaymentMethod     string    `json:"payment_method"`
	PaymentType       string    `json:"payment_type"`
	PayerEmail        string    `json:"payer_email"`
	DateApproved      time.Time `json:"date_approved"`
}

// DjangoWebhookPayload is sent to Django when a payment is confirmed.
type DjangoWebhookPayload struct {
	Event             string  `json:"event"`
	GymSlug           string  `json:"gym_slug"`
	ExternalReference string  `json:"external_reference"`
	PaymentID         string  `json:"payment_id"`
	PaymentStatus     string  `json:"payment_status"`
	PaymentType       string  `json:"payment_type"`
	Amount            float64 `json:"amount"`
	PayerEmail        string  `json:"payer_email"`
	Timestamp         string  `json:"timestamp"`
}

// GymCredentials holds the secrets needed for a gym.
type GymCredentials struct {
	GymSlug       string `json:"gym_slug"`
	WebhookSecret string `json:"webhook_secret"`
	AccessToken   string `json:"access_token,omitempty"`
}
