// Package errors provides application-level error handling.
package errors

import (
	"fmt"
	"net/http"
)

// AppError represents an application error with HTTP status code.
type AppError struct {
	Err        error
	Message    string
	Code       string
	HTTPStatus int
}

// Error implements the error interface.
func (e *AppError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("%s: %s", e.Message, e.Err.Error())
	}
	return e.Err.Error()
}

// Unwrap returns the underlying error.
func (e *AppError) Unwrap() error {
	return e.Err
}

// NewAppError creates a new application error.
func NewAppError(err error, message, code string, httpStatus int) *AppError {
	return &AppError{
		Err:        err,
		Message:    message,
		Code:       code,
		HTTPStatus: httpStatus,
	}
}

// NewBadRequestError creates a 400 error.
func NewBadRequestError(err error, message string) *AppError {
	return NewAppError(err, message, "BAD_REQUEST", http.StatusBadRequest)
}

// NewUnauthorizedError creates a 401 error.
func NewUnauthorizedError(err error, message string) *AppError {
	return NewAppError(err, message, "UNAUTHORIZED", http.StatusUnauthorized)
}

// NewNotFoundError creates a 404 error.
func NewNotFoundError(err error, message string) *AppError {
	return NewAppError(err, message, "NOT_FOUND", http.StatusNotFound)
}

// NewInternalError creates a 500 error.
func NewInternalError(err error, message string) *AppError {
	return NewAppError(err, message, "INTERNAL_ERROR", http.StatusInternalServerError)
}
