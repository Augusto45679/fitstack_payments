// Package service implements the core business logic.
package service

import (
	"context"
	"log"
	"time"

	"github.com/fitstack/fitstack-payments/internal/core/domain"
	"github.com/fitstack/fitstack-payments/internal/core/ports"
)

// PaymentService orchestrates payment operations.
type PaymentService struct {
	gateway         ports.PaymentGateway
	credProvider    ports.GymCredentialProvider
	djangoNotifier  ports.DjangoNotifier
	webhookValidator ports.WebhookValidator
}

// NewPaymentService creates a new payment service.
func NewPaymentService(
	gateway ports.PaymentGateway,
	credProvider ports.GymCredentialProvider,
	djangoNotifier ports.DjangoNotifier,
	webhookValidator ports.WebhookValidator,
) *PaymentService {
	return &PaymentService{
		gateway:         gateway,
		credProvider:    credProvider,
		djangoNotifier:  djangoNotifier,
		webhookValidator: webhookValidator,
	}
}

// CreateCheckout creates a payment preference in Mercado Pago.
// The access token is provided in the request (stateless).
func (s *PaymentService) CreateCheckout(ctx context.Context, req domain.PaymentRequest) (*domain.PaymentResponse, error) {
	// Validate required fields
	if req.MPAccessToken == "" {
		return &domain.PaymentResponse{
			Success:   false,
			Error:     "mp_access_token is required",
			ErrorCode: "VALIDATION_ERROR",
		}, nil
	}

	if req.GymSlug == "" || req.Amount <= 0 || req.Title == "" {
		return &domain.PaymentResponse{
			Success:   false,
			Error:     "gym_slug, amount, and title are required",
			ErrorCode: "VALIDATION_ERROR",
		}, nil
	}

	// Create preference using the provided token
	response, err := s.gateway.CreatePreference(ctx, req.MPAccessToken, req)
	if err != nil {
		log.Printf("Failed to create preference for gym %s: %v", req.GymSlug, err)
		return &domain.PaymentResponse{
			Success:   false,
			Error:     "Failed to create payment preference",
			ErrorCode: "GATEWAY_ERROR",
		}, nil
	}

	log.Printf("Created preference %s for gym %s, amount: %.2f", 
		response.PreferenceID, req.GymSlug, req.Amount)

	return response, nil
}

// ProcessWebhook handles incoming Mercado Pago webhook notifications.
func (s *PaymentService) ProcessWebhook(
	ctx context.Context,
	gymSlug string,
	notification domain.WebhookNotification,
	xSignature string,
	xRequestID string,
) error {
	// Step 1: Get webhook secret for this gym
	secret, err := s.credProvider.GetWebhookSecret(ctx, gymSlug)
	if err != nil {
		log.Printf("Failed to get webhook secret for gym %s: %v", gymSlug, err)
		return domain.NewServiceError(domain.ErrGymNotFound,
			"gym not found: "+gymSlug, "GYM_NOT_FOUND")
	}

	// Step 2: Validate webhook signature
	dataID := notification.Data.ID
	if !s.webhookValidator.ValidateSignature(xSignature, xRequestID, dataID, secret) {
		log.Printf("Webhook signature validation failed for gym %s", gymSlug)
		return domain.ErrWebhookValidationFailed
	}

	// Step 3: Only process payment notifications
	if notification.Type != "payment" {
		log.Printf("Ignoring webhook type: %s for gym %s", notification.Type, gymSlug)
		return nil
	}

	// Step 4: Get access token to fetch payment info
	accessToken, err := s.credProvider.GetAccessToken(ctx, gymSlug)
	if err != nil {
		log.Printf("Failed to get access token for gym %s: %v", gymSlug, err)
		return err
	}

	// Step 5: Get payment details from Mercado Pago
	paymentInfo, err := s.gateway.GetPaymentInfo(ctx, accessToken, dataID)
	if err != nil {
		log.Printf("Failed to get payment info %s for gym %s: %v", dataID, gymSlug, err)
		return err
	}

	// Step 6: Determine event type based on status
	event := mapStatusToEvent(paymentInfo.Status)
	
	// Step 7: Notify Django backend
	payload := domain.DjangoWebhookPayload{
		Event:             event,
		GymSlug:           gymSlug,
		ExternalReference: paymentInfo.ExternalReference,
		PaymentID:         paymentInfo.PaymentID,
		PaymentStatus:     paymentInfo.Status,
		PaymentType:       paymentInfo.PaymentType,
		Amount:            paymentInfo.Amount,
		PayerEmail:        paymentInfo.PayerEmail,
		Timestamp:         time.Now().Format(time.RFC3339),
	}

	if err := s.djangoNotifier.NotifyPaymentConfirmed(ctx, payload); err != nil {
		log.Printf("Failed to notify Django for payment %s: %v", dataID, err)
		return err
	}

	log.Printf("Webhook processed: payment %s, status %s, gym %s", 
		dataID, paymentInfo.Status, gymSlug)

	return nil
}

// mapStatusToEvent maps MP payment status to event name.
func mapStatusToEvent(status string) string {
	switch status {
	case "approved":
		return "payment.approved"
	case "pending", "in_process":
		return "payment.pending"
	case "rejected":
		return "payment.rejected"
	case "cancelled":
		return "payment.cancelled"
	case "refunded":
		return "payment.refunded"
	default:
		return "payment.updated"
	}
}
