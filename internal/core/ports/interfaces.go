// Package ports defines the interfaces (ports) for the payment service.
// These are contracts that adapters must implement.
package ports

import (
	"context"

	"github.com/fitstack/fitstack-payments/internal/core/domain"
)

// PaymentGateway defines the interface for interacting with Mercado Pago.
type PaymentGateway interface {
	// CreatePreference creates a Checkout Pro preference.
	// Returns the preference ID and init_point URLs.
	CreatePreference(ctx context.Context, accessToken string, req domain.PaymentRequest) (*domain.PaymentResponse, error)

	// GetPaymentInfo retrieves payment details by ID.
	GetPaymentInfo(ctx context.Context, accessToken string, paymentID string) (*domain.PaymentInfo, error)
}

// GymCredentialProvider retrieves gym credentials for webhook validation.
// In production, this will call Django. For now, it's an interface.
type GymCredentialProvider interface {
	// GetWebhookSecret retrieves the webhook secret for a gym.
	GetWebhookSecret(ctx context.Context, gymSlug string) (string, error)

	// GetAccessToken retrieves the access token for a gym (optional, for webhook processing).
	GetAccessToken(ctx context.Context, gymSlug string) (string, error)
}

// DjangoNotifier sends payment confirmations to Django backend.
type DjangoNotifier interface {
	// NotifyPaymentConfirmed sends payment confirmation to Django.
	NotifyPaymentConfirmed(ctx context.Context, payload domain.DjangoWebhookPayload) error
}

// WebhookValidator validates Mercado Pago webhook signatures.
type WebhookValidator interface {
	// ValidateSignature validates the x-signature header from Mercado Pago.
	ValidateSignature(xSignature, xRequestID, dataID, secret string) bool
}
