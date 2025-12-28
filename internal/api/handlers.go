// Package api contains the HTTP handlers and routing for the payment service.
package api

import (
	"errors"
	"log"
	"net/http"

	"github.com/fitstack/fitstack-payments/internal/domain"
	"github.com/fitstack/fitstack-payments/internal/payment"
	"github.com/gin-gonic/gin"
)

// Handler contains the HTTP handlers for the payment API.
type Handler struct {
	paymentService *payment.Service
}

// NewHandler creates a new API handler with the payment service.
func NewHandler(paymentService *payment.Service) *Handler {
	return &Handler{
		paymentService: paymentService,
	}
}

// CheckoutRequest represents the JSON body for the checkout endpoint.
type CheckoutRequest struct {
	GymID      string  `json:"gym_id" binding:"required"`
	Amount     float64 `json:"amount" binding:"required,gt=0"`
	Title      string  `json:"title" binding:"required"`
	PayerEmail string  `json:"payer_email" binding:"required,email"`
}

// CheckoutResponse represents the response from the checkout endpoint.
type CheckoutResponse struct {
	Success   bool   `json:"success"`
	InitPoint string `json:"init_point,omitempty"`
}

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
}

// CreateCheckout handles POST /checkout
// Creates a Mercado Pago preference and returns the init_point URL.
func (h *Handler) CreateCheckout(c *gin.Context) {
	var req CheckoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Error:   "Invalid request body: " + err.Error(),
			Code:    "VALIDATION_ERROR",
		})
		return
	}

	// Create payment order from request
	order := domain.PaymentOrder{
		GymSlug:    req.GymID,
		Amount:     req.Amount,
		Title:      req.Title,
		PayerEmail: req.PayerEmail,
	}

	// Call service to create checkout
	preference, err := h.paymentService.CreateCheckout(c.Request.Context(), order)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, CheckoutResponse{
		Success:   true,
		InitPoint: preference.InitPoint,
	})
}

// WebhookRequest represents the JSON body from Mercado Pago webhooks.
type WebhookRequest struct {
	ID     string `json:"id"`
	Type   string `json:"type"`
	Action string `json:"action"`
	Data   struct {
		ID string `json:"id"`
	} `json:"data"`
	LiveMode    bool   `json:"live_mode"`
	DateCreated string `json:"date_created"`
}

// HandleWebhook handles POST /webhook/:gym_slug
// Receives notifications from Mercado Pago and processes them.
func (h *Handler) HandleWebhook(c *gin.Context) {
	gymSlug := c.Param("gym_slug")
	if gymSlug == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Error:   "gym_slug is required in URL",
			Code:    "MISSING_GYM_SLUG",
		})
		return
	}

	var req WebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Mercado Pago might send different formats, log and accept
		log.Printf("Webhook parsing error for gym %s: %v", gymSlug, err)
		c.JSON(http.StatusOK, gin.H{"status": "received"})
		return
	}

	// Create notification from request
	notification := domain.WebhookNotification{
		ID:          req.ID,
		Type:        req.Type,
		Action:      req.Action,
		DataID:      req.Data.ID,
		LiveMode:    req.LiveMode,
		DateCreated: req.DateCreated,
	}

	// Process webhook
	if err := h.paymentService.ProcessWebhook(c.Request.Context(), gymSlug, notification); err != nil {
		log.Printf("Webhook processing error for gym %s: %v", gymSlug, err)
		// Still return 200 to prevent MP from retrying
		c.JSON(http.StatusOK, gin.H{"status": "processed_with_error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "processed"})
}

// Health handles GET /health
func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"service": "fitstack-payments",
	})
}

// handleServiceError maps domain errors to HTTP responses.
func handleServiceError(c *gin.Context, err error) {
	var paymentErr *domain.PaymentError
	if errors.As(err, &paymentErr) {
		statusCode := http.StatusInternalServerError

		switch {
		case errors.Is(paymentErr.Err, domain.ErrGymNotFound):
			statusCode = http.StatusNotFound
		case errors.Is(paymentErr.Err, domain.ErrPaymentNotEnabled):
			statusCode = http.StatusForbidden
		case errors.Is(paymentErr.Err, domain.ErrInvalidPaymentOrder):
			statusCode = http.StatusBadRequest
		case errors.Is(paymentErr.Err, domain.ErrInvalidAccessToken):
			statusCode = http.StatusInternalServerError
		case errors.Is(paymentErr.Err, domain.ErrPaymentGatewayError):
			statusCode = http.StatusBadGateway
		}

		c.JSON(statusCode, ErrorResponse{
			Success: false,
			Error:   paymentErr.Message,
			Code:    paymentErr.Code,
		})
		return
	}

	// Generic error
	c.JSON(http.StatusInternalServerError, ErrorResponse{
		Success: false,
		Error:   "Internal server error",
		Code:    "INTERNAL_ERROR",
	})
}
