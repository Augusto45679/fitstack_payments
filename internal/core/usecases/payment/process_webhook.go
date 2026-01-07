// Package payment contains payment use cases.
package payment

import (
	"context"
	"log"
	"time"

	"github.com/fitstack/fitstack-payments/internal/core/domain"
	"github.com/fitstack/fitstack-payments/internal/core/domain/entities"
	"github.com/fitstack/fitstack-payments/internal/core/domain/repository_interfaces"
	"github.com/fitstack/fitstack-payments/internal/core/domain/value_objects"
	"github.com/fitstack/fitstack-payments/internal/core/services"
)

// ProcessWebhookCommand represents the input for processing a webhook.
type ProcessWebhookCommand struct {
	GymSlug      string
	Notification entities.WebhookNotification
	XSignature   string
	XRequestID   string
}

// PaymentEventNotification represents a payment event to be sent to external systems.
type PaymentEventNotification struct {
	Event             string
	GymSlug           string
	ExternalReference string
	PaymentID         string
	PaymentStatus     string
	PaymentType       string
	Amount            float64
	PayerEmail        string
	Timestamp         string
}

// ProcessWebhookUseCase handles incoming webhook notifications.
type ProcessWebhookUseCase struct {
	gateway            PaymentGateway
	tenantRepo         repository_interfaces.TenantRepository
	webhookValidator   services.WebhookValidator
	notificationSvc    NotificationService
}

// NewProcessWebhookUseCase creates a new instance of ProcessWebhookUseCase.
func NewProcessWebhookUseCase(
	gateway PaymentGateway,
	tenantRepo repository_interfaces.TenantRepository,
	webhookValidator services.WebhookValidator,
	notificationSvc NotificationService,
) *ProcessWebhookUseCase {
	return &ProcessWebhookUseCase{
		gateway:          gateway,
		tenantRepo:       tenantRepo,
		webhookValidator: webhookValidator,
		notificationSvc:  notificationSvc,
	}
}

// Execute processes a webhook notification.
func (uc *ProcessWebhookUseCase) Execute(ctx context.Context, cmd ProcessWebhookCommand) error {
	// Step 1: Validate tenant ID
	tenantID, err := value_objects.NewTenantID(cmd.GymSlug)
	if err != nil {
		log.Printf("Invalid tenant ID: %v", err)
		return domain.NewServiceError(domain.ErrGymNotFound,
			"invalid gym_slug: "+cmd.GymSlug, "INVALID_TENANT")
	}

	// Step 2: Get webhook secret for this tenant
	secret, err := uc.tenantRepo.GetWebhookSecret(ctx, tenantID)
	if err != nil {
		log.Printf("Failed to get webhook secret for gym %s: %v", cmd.GymSlug, err)
		return domain.NewServiceError(domain.ErrGymNotFound,
			"gym not found: "+cmd.GymSlug, "GYM_NOT_FOUND")
	}

	// Step 3: Validate webhook signature
	dataID := cmd.Notification.GetPaymentID()
	if !uc.webhookValidator.ValidateSignature(cmd.XSignature, cmd.XRequestID, dataID, secret) {
		log.Printf("Webhook signature validation failed for gym %s", cmd.GymSlug)
		return domain.ErrWebhookValidationFailed
	}

	// Step 4: Only process payment notifications
	if !cmd.Notification.IsPaymentNotification() {
		log.Printf("Ignoring webhook type: %s for gym %s", cmd.Notification.Type, cmd.GymSlug)
		return nil
	}

	// Step 5: Get access token to fetch payment info
	accessToken, err := uc.tenantRepo.GetAccessToken(ctx, tenantID)
	if err != nil {
		log.Printf("Failed to get access token for gym %s: %v", cmd.GymSlug, err)
		return err
	}

	// Step 6: Get payment details from payment gateway
	paymentInfo, err := uc.gateway.GetPaymentInfo(ctx, accessToken, dataID)
	if err != nil {
		log.Printf("Failed to get payment info %s for gym %s: %v", dataID, cmd.GymSlug, err)
		return err
	}

	// Step 7: Parse payment status
	paymentStatus, err := value_objects.NewPaymentStatus(paymentInfo.Status)
	if err != nil {
		log.Printf("Invalid payment status %s: %v", paymentInfo.Status, err)
		paymentStatus = value_objects.PaymentStatus(paymentInfo.Status) // Use as-is if unknown
	}

	// Step 8: Determine event type based on status
	event := paymentStatus.ToEventName()

	// Step 9: Notify external systems (Django backend)
	notification := PaymentEventNotification{
		Event:             event,
		GymSlug:           cmd.GymSlug,
		ExternalReference: paymentInfo.ExternalReference,
		PaymentID:         paymentInfo.PaymentID,
		PaymentStatus:     paymentInfo.Status,
		PaymentType:       paymentInfo.PaymentType,
		Amount:            paymentInfo.Amount,
		PayerEmail:        paymentInfo.PayerEmail,
		Timestamp:         time.Now().Format(time.RFC3339),
	}

	if err := uc.notificationSvc.NotifyPaymentEvent(ctx, notification); err != nil {
		log.Printf("Failed to notify payment event for payment %s: %v", dataID, err)
		return err
	}

	log.Printf("Webhook processed: payment %s, status %s, gym %s", 
		dataID, paymentInfo.Status, cmd.GymSlug)

	return nil
}
