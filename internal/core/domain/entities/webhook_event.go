// Package entities contains core business entities.
package entities

import "time"

// WebhookEvent represents a webhook notification from the payment gateway.
// This provides an event sourcing foundation for webhook processing.
type WebhookEvent struct {
	ID          string
	GymSlug     string
	EventType   string
	PaymentID   string
	Status      string
	ProcessedAt time.Time
	RawPayload  string
}

// WebhookNotification represents the IPN notification from Mercado Pago.
// This is the raw structure received from the payment gateway.
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

// IsPaymentNotification checks if this is a payment-related webhook.
func (w *WebhookNotification) IsPaymentNotification() bool {
	return w.Type == "payment"
}

// GetPaymentID extracts the payment ID from the notification.
func (w *WebhookNotification) GetPaymentID() string {
	return w.Data.ID
}
