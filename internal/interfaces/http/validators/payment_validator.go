// Package validators contains custom HTTP validators.
package validators

import (
	"regexp"
)

// PaymentValidator validates payment-related data.
type PaymentValidator struct {
}

// NewPaymentValidator creates a new payment validator.
func NewPaymentValidator() *PaymentValidator {
	return &PaymentValidator{}
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// ValidateEmail validates email format.
func (v *PaymentValidator) ValidateEmail(email string) bool {
	return emailRegex.MatchString(email)
}

// ValidateAmount ensures amount is positive and reasonable.
func (v *PaymentValidator) ValidateAmount(amount float64) bool {
	return amount > 0 && amount < 10000000 // Max 10 million
}

// ValidateExternalReference validates external reference format.
func (v *PaymentValidator) ValidateExternalReference(ref string) bool {
	return len(ref) >= 3 && len(ref) <= 256
}
