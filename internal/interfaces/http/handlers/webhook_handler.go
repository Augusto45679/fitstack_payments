// Package handlers contains HTTP request handlers.
package handlers

import (
	"log"
	"net/http"

	"github.com/fitstack/fitstack-payments/internal/core/domain/entities"
	"github.com/fitstack/fitstack-payments/internal/core/usecases/payment"
	"github.com/fitstack/fitstack-payments/internal/interfaces/http/dtos"
	"github.com/gin-gonic/gin"
)

// WebhookHandler handles webhook-related HTTP requests.
type WebhookHandler struct {
	processWebhookUC *payment.ProcessWebhookUseCase
}

// NewWebhookHandler creates a new webhook handler.
func NewWebhookHandler(processWebhookUC *payment.ProcessWebhookUseCase) *WebhookHandler {
	return &WebhookHandler{
		processWebhookUC: processWebhookUC,
	}
}

// HandleWebhook handles POST /webhooks/:gym_slug
// Receives Mercado Pago IPN notifications.
func (h *WebhookHandler) HandleWebhook(c *gin.Context) {
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
	var notificationDTO dtos.WebhookNotificationDTO
	if err := c.ShouldBindJSON(&notificationDTO); err != nil {
		// MP may send different formats, log and accept
		log.Printf("Webhook parse error for gym %s: %v", gymSlug, err)
		c.JSON(http.StatusOK, gin.H{"status": "received"})
		return
	}

	// Map DTO to domain entity
	notification := entities.WebhookNotification{
		ID:          notificationDTO.ID,
		LiveMode:    notificationDTO.LiveMode,
		Type:        notificationDTO.Type,
		DateCreated: notificationDTO.DateCreated,
		UserID:      notificationDTO.UserID,
		APIVersion:  notificationDTO.APIVersion,
		Action:      notificationDTO.Action,
		Data: struct {
			ID string `json:"id"`
		}{
			ID: notificationDTO.Data.ID,
		},
	}

	// Create use case command
	cmd := payment.ProcessWebhookCommand{
		GymSlug:      gymSlug,
		Notification: notification,
		XSignature:   xSignature,
		XRequestID:   xRequestID,
	}

	// Process the webhook
	err := h.processWebhookUC.Execute(c.Request.Context(), cmd)
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
