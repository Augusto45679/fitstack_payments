// Package payment contains payment-related use cases.
package payment

import (
	"context"

	"github.com/fitstack/fitstack-payments/internal/core/domain/entities"
)

// PaymentGateway defines the interface for interacting with payment gateways.
// This interface is implemented by the Mercado Pago adapter in the infrastructure layer.
type PaymentGateway interface {
	// CreatePreference creates a payment preference (checkout session).
	CreatePreference(ctx context.Context, accessToken string, req CreatePaymentCommand) (*PaymentPreferenceResult, error)

	// GetPaymentInfo retrieves payment details by ID.
	GetPaymentInfo(ctx context.Context, accessToken string, paymentID string) (*entities.PaymentInfo, error)
}

// NotificationService defines the interface for sending payment notifications.
type NotificationService interface {
	// NotifyPaymentEvent sends a payment event notification to external systems.
	NotifyPaymentEvent(ctx context.Context, event PaymentEventNotification) error
}
