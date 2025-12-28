// Package payment implements the core business logic for payment processing.
// This is the service/use-case layer in Clean Architecture.
package payment

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/fitstack/fitstack-payments/internal/domain"
)

// Service implements the payment business logic.
// It orchestrates between the repository (to get gym config) and
// the payment gateway (to create preferences in Mercado Pago).
type Service struct {
	gymRepo        domain.GymRepository
	paymentGateway domain.PaymentGateway
	coreNotifier   domain.CoreNotifier
}

// NewService creates a new payment service with the required dependencies.
func NewService(
	gymRepo domain.GymRepository,
	paymentGateway domain.PaymentGateway,
	coreNotifier domain.CoreNotifier,
) *Service {
	return &Service{
		gymRepo:        gymRepo,
		paymentGateway: paymentGateway,
		coreNotifier:   coreNotifier,
	}
}

// CreateCheckout handles the checkout flow:
// 1. Fetches the gym configuration
// 2. Validates the gym has payment enabled
// 3. Creates a payment preference in Mercado Pago
// 4. Returns the init_point URL for the user to complete payment
func (s *Service) CreateCheckout(ctx context.Context, order domain.PaymentOrder) (*domain.PaymentPreference, error) {
	// Validate order data
	if err := validateOrder(order); err != nil {
		return nil, err
	}

	// Step 1: Get gym configuration
	gymConfig, err := s.gymRepo.GetGymConfig(ctx, order.GymSlug)
	if err != nil {
		if errors.Is(err, domain.ErrGymNotFound) {
			return nil, domain.NewPaymentError(err, 
				fmt.Sprintf("gym '%s' not found", order.GymSlug), 
				"GYM_NOT_FOUND")
		}
		return nil, domain.NewPaymentError(domain.ErrCoreAPIError,
			"failed to fetch gym configuration",
			"CORE_API_ERROR")
	}

	// Step 2: Validate gym has payment enabled
	if !gymConfig.IsPaymentEnabled {
		log.Printf("Payment attempt for gym %s but payment is not enabled", order.GymSlug)
		return nil, domain.NewPaymentError(domain.ErrPaymentNotEnabled,
			fmt.Sprintf("gym '%s' does not have payment integration enabled", order.GymSlug),
			"PAYMENT_NOT_ENABLED")
	}

	// Step 3: Validate access token exists
	if gymConfig.AccessToken == "" {
		log.Printf("Gym %s has payment enabled but no access token", order.GymSlug)
		return nil, domain.NewPaymentError(domain.ErrInvalidAccessToken,
			"gym payment configuration is incomplete",
			"INVALID_TOKEN")
	}

	// Step 4: Create payment preference in Mercado Pago
	preference, err := s.paymentGateway.CreatePreference(ctx, gymConfig.AccessToken, order)
	if err != nil {
		log.Printf("Failed to create MP preference for gym %s: %v", order.GymSlug, err)
		return nil, domain.NewPaymentError(domain.ErrPaymentGatewayError,
			"failed to create payment preference",
			"GATEWAY_ERROR")
	}

	log.Printf("Created payment preference %s for gym %s, amount: %.2f", 
		preference.ID, order.GymSlug, order.Amount)

	return preference, nil
}

// ProcessWebhook handles incoming webhook notifications from Mercado Pago.
// It validates the notification, fetches payment details, and notifies the Core.
func (s *Service) ProcessWebhook(ctx context.Context, gymSlug string, notification domain.WebhookNotification) error {
	// Only process payment notifications
	if notification.Type != "payment" {
		log.Printf("Ignoring webhook type: %s", notification.Type)
		return nil
	}

	// Get gym configuration to retrieve access token
	gymConfig, err := s.gymRepo.GetGymConfig(ctx, gymSlug)
	if err != nil {
		return domain.NewPaymentError(err,
			"failed to fetch gym configuration for webhook processing",
			"WEBHOOK_GYM_ERROR")
	}

	// Get payment info from Mercado Pago
	paymentStatus, err := s.paymentGateway.GetPaymentInfo(ctx, gymConfig.AccessToken, notification.DataID)
	if err != nil {
		return domain.NewPaymentError(domain.ErrPaymentGatewayError,
			"failed to get payment info",
			"WEBHOOK_GATEWAY_ERROR")
	}

	// Set the gym slug for the status
	paymentStatus.GymSlug = gymSlug

	log.Printf("Webhook processed: payment %s for gym %s, status: %s",
		paymentStatus.PaymentID, gymSlug, paymentStatus.Status)

	// Notify FitStack Core about the payment status
	if err := s.coreNotifier.NotifyPaymentStatus(ctx, paymentStatus); err != nil {
		log.Printf("Failed to notify core about payment %s: %v", paymentStatus.PaymentID, err)
		return domain.NewPaymentError(domain.ErrCoreAPIError,
			"failed to notify core about payment",
			"WEBHOOK_NOTIFY_ERROR")
	}

	return nil
}

// validateOrder performs basic validation on the payment order.
func validateOrder(order domain.PaymentOrder) error {
	if order.GymSlug == "" {
		return domain.NewPaymentError(domain.ErrInvalidPaymentOrder,
			"gym_id is required",
			"VALIDATION_ERROR")
	}
	if order.Amount <= 0 {
		return domain.NewPaymentError(domain.ErrInvalidPaymentOrder,
			"amount must be greater than 0",
			"VALIDATION_ERROR")
	}
	if order.Title == "" {
		return domain.NewPaymentError(domain.ErrInvalidPaymentOrder,
			"title is required",
			"VALIDATION_ERROR")
	}
	if order.PayerEmail == "" {
		return domain.NewPaymentError(domain.ErrInvalidPaymentOrder,
			"payer_email is required",
			"VALIDATION_ERROR")
	}
	return nil
}
