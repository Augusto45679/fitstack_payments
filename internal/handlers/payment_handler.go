// Package handlers contains the HTTP handlers for the payment service.
package handlers

import (
	"log"
	"net/http"

	"github.com/fitstack/fitstack-payments/internal/core/domain"
	"github.com/fitstack/fitstack-payments/internal/core/service"
	"github.com/gin-gonic/gin"
)

// PaymentHandler handles HTTP requests for payments.
type PaymentHandler struct {
	service *service.PaymentService
}

// NewPaymentHandler creates a new payment handler.
func NewPaymentHandler(svc *service.PaymentService) *PaymentHandler {
	return &PaymentHandler{service: svc}
}

// CreateCheckout handles POST /api/v1/payments/checkout
// Creates a Mercado Pago preference with the provided access token.
func (h *PaymentHandler) CreateCheckout(c *gin.Context) {
	var req domain.PaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.PaymentResponse{
			Success:   false,
			Error:     "Invalid request: " + err.Error(),
			ErrorCode: "VALIDATION_ERROR",
		})
		return
	}

	response, err := h.service.CreateCheckout(c.Request.Context(), req)
	if err != nil {
		log.Printf("CreateCheckout error: %v", err)
		c.JSON(http.StatusInternalServerError, domain.PaymentResponse{
			Success:   false,
			Error:     "Internal server error",
			ErrorCode: "INTERNAL_ERROR",
		})
		return
	}

	if !response.Success {
		c.JSON(http.StatusBadRequest, response)
		return
	}

	c.JSON(http.StatusOK, response)
}

// HandleWebhook handles POST /webhooks/:gym_slug
// Receives Mercado Pago IPN notifications.
func (h *PaymentHandler) HandleWebhook(c *gin.Context) {
	gymSlug := c.Param("gym_slug")
	if gymSlug == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "gym_slug is required",
		})
		return
	}

	// Extract security headers
	xSignature := c.GetHeader("x-signature")
	xRequestID := c.GetHeader("x-request-id")

	// Parse notification body
	var notification domain.WebhookNotification
	if err := c.ShouldBindJSON(&notification); err != nil {
		// MP may send different formats, log and accept
		log.Printf("Webhook parse error for gym %s: %v", gymSlug, err)
		c.JSON(http.StatusOK, gin.H{"status": "received"})
		return
	}

	// Process the webhook
	err := h.service.ProcessWebhook(
		c.Request.Context(),
		gymSlug,
		notification,
		xSignature,
		xRequestID,
	)

	if err != nil {
		log.Printf("Webhook processing error for gym %s: %v", gymSlug, err)
		// Return 200 to prevent MP from retrying (we log the error)
		c.JSON(http.StatusOK, gin.H{
			"status": "processed_with_error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "processed"})
}

// Health handles GET /health
func (h *PaymentHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"service": "fitstack-payments",
		"version": "1.0.0",
	})
}
