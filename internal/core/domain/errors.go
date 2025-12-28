// Package domain contains the core business entities for the payment service.
package domain

import "errors"

// Domain errors - represent business rule violations.
var (
	// ErrInvalidRequest is returned for malformed requests.
	ErrInvalidRequest = errors.New("invalid request")

	// ErrPaymentGatewayError is returned when Mercado Pago fails.
	ErrPaymentGatewayError = errors.New("payment gateway error")

	// ErrWebhookValidationFailed is returned when x-signature is invalid.
	ErrWebhookValidationFailed = errors.New("webhook signature validation failed")

	// ErrGymNotFound is returned when gym credentials are not found.
	ErrGymNotFound = errors.New("gym not found")

	// ErrDjangoCallbackFailed is returned when Django notification fails.
	ErrDjangoCallbackFailed = errors.New("failed to notify Django backend")
)

// ServiceError wraps errors with additional context.
type ServiceError struct {
	Err     error
	Message string
	Code    string
}

func (e *ServiceError) Error() string {
	if e.Message != "" {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Err.Error()
}

func (e *ServiceError) Unwrap() error {
	return e.Err
}

// NewServiceError creates a new ServiceError.
func NewServiceError(err error, message, code string) *ServiceError {
	return &ServiceError{Err: err, Message: message, Code: code}
}
