// Package errors provides error handling utilities.
package errors

import (
	"log"

	"github.com/gin-gonic/gin"
)

// ErrorHandler provides error handling utilities for HTTP responses.
type ErrorHandler struct {
}

// NewErrorHandler creates a new error handler.
func NewErrorHandler() *ErrorHandler {
	return &ErrorHandler{}
}

// HandleError converts an error to an HTTP response.
func (h *ErrorHandler) HandleError(c *gin.Context, err error) {
	// Check if it's an AppError
	if appErr, ok := err.(*AppError); ok {
		log.Printf("Application error: %s - %v", appErr.Code, appErr)
		c.JSON(appErr.HTTPStatus, gin.H{
			"success": false,
			"error":   appErr.Message,
			"code":    appErr.Code,
		})
		return
	}

	// Generic error
	log.Printf("Unexpected error: %v", err)
	c.JSON(500, gin.H{
		"success": false,
		"error":   "Internal server error",
		"code":    "INTERNAL_ERROR",
	})
}

// LogError logs an error with context.
func (h *ErrorHandler) LogError(context string, err error) {
	log.Printf("[ERROR] %s: %v", context, err)
}
