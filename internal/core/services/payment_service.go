// Package services contains domain services.
package services

import (
	"github.com/fitstack/fitstack-payments/internal/core/domain/value_objects"
)

// PaymentService provides payment domain logic.
type PaymentService struct {
}

// NewPaymentService creates a new payment service.
func NewPaymentService() *PaymentService {
	return &PaymentService{}
}

// MapStatusToEvent maps a payment status to an event name.
// This was moved from the old service layer to a domain service.
func (s *PaymentService) MapStatusToEvent(status value_objects.PaymentStatus) string {
	return status.ToEventName()
}

// ValidatePaymentAmount validates a payment amount.
func (s *PaymentService) ValidatePaymentAmount(money *value_objects.Money) error {
	if !money.IsPositive() {
		return value_objects.NewMoneyError("amount must be positive")
	}
	return nil
}
