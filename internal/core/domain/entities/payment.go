// Package entities contains core business entities for the payment domain.
// This is the innermost layer - no external dependencies.
package entities

import (
	"time"
)

// Payment represents a payment transaction in the system.
// This is the aggregate root for payment operations.
type Payment struct {
	ID                string
	ExternalReference string
	Amount            float64
	Currency          string
	Status            string
	StatusDetail      string
	PaymentMethod     string
	PaymentType       string
	PayerEmail        string
	GymSlug           string
	Title             string
	Description       string
	PreferenceID      string
	DateCreated       time.Time
	DateApproved      time.Time
}

// PaymentInfo contains detailed information about a confirmed payment.
// This is typically retrieved from the payment gateway.
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

// IsApproved checks if the payment is approved.
func (p *PaymentInfo) IsApproved() bool {
	return p.Status == "approved"
}

// IsPending checks if the payment is pending.
func (p *PaymentInfo) IsPending() bool {
	return p.Status == "pending" || p.Status == "in_process"
}

// IsRejected checks if the payment is rejected.
func (p *PaymentInfo) IsRejected() bool {
	return p.Status == "rejected"
}
