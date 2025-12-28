// Package domain contains the core business entities and interfaces for the payment service.
// This is the innermost layer of the Clean Architecture - it has no dependencies on
// external frameworks or infrastructure.
package domain

import "time"

// GymConfig represents the payment configuration for a specific gym (tenant).
// The AccessToken is the decrypted Mercado Pago token for this gym.
type GymConfig struct {
	ID               string `json:"id"`
	Slug             string `json:"slug"`
	Name             string `json:"name"`
	AccessToken      string `json:"access_token"` // Decrypted MP access token
	IsPaymentEnabled bool   `json:"is_payment_enabled"`
}

// PaymentOrder represents a request to create a payment.
type PaymentOrder struct {
	GymSlug     string  `json:"gym_id"`      // Gym slug identifier
	Amount      float64 `json:"amount"`      // Amount in the gym's currency
	Title       string  `json:"title"`       // Payment title (e.g., "Plan Mensual")
	Description string  `json:"description"` // Optional description
	PayerEmail  string  `json:"payer_email"` // Payer's email address
}

// PaymentPreference represents a created Mercado Pago preference.
type PaymentPreference struct {
	ID        string `json:"id"`         // Mercado Pago preference ID
	InitPoint string `json:"init_point"` // URL to redirect user for payment
}

// WebhookNotification represents an incoming webhook from Mercado Pago.
type WebhookNotification struct {
	ID          string `json:"id"`
	Type        string `json:"type"`         // "payment", "merchant_order", etc.
	Action      string `json:"action"`       // "payment.created", "payment.updated", etc.
	DataID      string `json:"data_id"`      // The ID of the resource (payment ID, etc.)
	LiveMode    bool   `json:"live_mode"`    // true for production
	DateCreated string `json:"date_created"` // ISO 8601 timestamp
}

// PaymentStatus represents the status of a payment after webhook processing.
type PaymentStatus struct {
	PaymentID       string    `json:"payment_id"`
	GymSlug         string    `json:"gym_slug"`
	Status          string    `json:"status"`           // "approved", "pending", "rejected", etc.
	StatusDetail    string    `json:"status_detail"`    // More detailed status
	ExternalRef     string    `json:"external_ref"`     // Our reference (gym_slug + order info)
	Amount          float64   `json:"amount"`           // Amount paid
	PayerEmail      string    `json:"payer_email"`      // Payer's email
	TransactionDate time.Time `json:"transaction_date"` // When the transaction occurred
}
