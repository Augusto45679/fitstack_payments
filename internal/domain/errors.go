// Package domain contains the core business entities and interfaces for the payment service.
package domain

import "errors"

// Domain errors represent business rule violations.
// These are used to communicate specific error conditions from the domain layer.
var (
	// ErrGymNotFound is returned when a gym configuration cannot be found.
	ErrGymNotFound = errors.New("gym not found")

	// ErrPaymentNotEnabled is returned when attempting to process a payment
	// for a gym that doesn't have the payment feature enabled.
	ErrPaymentNotEnabled = errors.New("payment integration is not enabled for this gym")

	// ErrInvalidAccessToken is returned when the gym's Mercado Pago token is invalid or missing.
	ErrInvalidAccessToken = errors.New("invalid or missing Mercado Pago access token")

	// ErrInvalidPaymentOrder is returned when the payment order data is invalid.
	ErrInvalidPaymentOrder = errors.New("invalid payment order data")

	// ErrPaymentGatewayError is returned when there's an error communicating with Mercado Pago.
	ErrPaymentGatewayError = errors.New("payment gateway error")

	// ErrWebhookValidationFailed is returned when webhook signature validation fails.
	ErrWebhookValidationFailed = errors.New("webhook validation failed")

	// ErrCoreAPIError is returned when there's an error communicating with FitStack Core.
	ErrCoreAPIError = errors.New("error communicating with FitStack Core")
)

// PaymentError wraps a domain error with additional context.
type PaymentError struct {
	Err     error
	Message string
	Code    string
}

// Error implements the error interface.
func (e *PaymentError) Error() string {
	if e.Message != "" {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Err.Error()
}

// Unwrap allows errors.Is and errors.As to work with PaymentError.
func (e *PaymentError) Unwrap() error {
	return e.Err
}

// NewPaymentError creates a new PaymentError with the given error and message.
func NewPaymentError(err error, message, code string) *PaymentError {
	return &PaymentError{
		Err:     err,
		Message: message,
		Code:    code,
	}
}
