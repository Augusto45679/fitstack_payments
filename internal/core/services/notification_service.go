// Package services contains domain services.
package services

import (
	"context"
	"log"

	"github.com/fitstack/fitstack-payments/internal/core/usecases/payment"
)

// NotificationService handles payment event notifications.
type NotificationService struct {
	djangoNotifier DjangoNotifier
}

// DjangoNotifier sends notifications to Django backend.
type DjangoNotifier interface {
	// NotifyPaymentConfirmed sends payment confirmation to Django.
	NotifyPaymentConfirmed(ctx context.Context, event payment.PaymentEventNotification) error
}

// NewNotificationService creates a new notification service.
func NewNotificationService(djangoNotifier DjangoNotifier) *NotificationService {
	return &NotificationService{
		djangoNotifier: djangoNotifier,
	}
}

// NotifyPaymentEvent sends a payment event notification.
func (s *NotificationService) NotifyPaymentEvent(ctx context.Context, event payment.PaymentEventNotification) error {
	log.Printf("Notifying payment event: %s for gym %s, payment %s", 
		event.Event, event.GymSlug, event.PaymentID)
	
	return s.djangoNotifier.NotifyPaymentConfirmed(ctx, event)
}
