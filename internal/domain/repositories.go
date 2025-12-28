// Package domain contains the core business entities and interfaces for the payment service.
package domain

import "context"

// GymRepository defines the interface for fetching gym configuration.
// This is a "port" in hexagonal architecture - the domain defines what it needs,
// and infrastructure provides the implementation.
type GymRepository interface {
	// GetGymConfig retrieves the payment configuration for a gym by its slug.
	// Returns ErrGymNotFound if the gym doesn't exist.
	GetGymConfig(ctx context.Context, gymSlug string) (*GymConfig, error)
}

// PaymentGateway defines the interface for interacting with the payment provider.
// This abstracts away the details of Mercado Pago SDK usage.
type PaymentGateway interface {
	// CreatePreference creates a payment preference in Mercado Pago.
	// The accessToken is the gym-specific token.
	// Returns the preference with the init_point URL for redirecting the user.
	CreatePreference(ctx context.Context, accessToken string, order PaymentOrder) (*PaymentPreference, error)

	// GetPaymentInfo retrieves payment information from a webhook notification.
	// Used to process webhook callbacks and get payment status.
	GetPaymentInfo(ctx context.Context, accessToken string, paymentID string) (*PaymentStatus, error)
}

// CoreNotifier defines the interface for notifying FitStack Core about payment events.
type CoreNotifier interface {
	// NotifyPaymentStatus sends payment status updates to FitStack Core.
	// The Core will update membership status, send notifications, etc.
	NotifyPaymentStatus(ctx context.Context, status *PaymentStatus) error
}
