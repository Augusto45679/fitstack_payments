// Package handlers contains HTTP request handlers.
package handlers

import (
	"log"
	"net/http"

	"github.com/fitstack/fitstack-payments/internal/core/usecases/payment"
	"github.com/fitstack/fitstack-payments/internal/interfaces/http/dtos"
	"github.com/gin-gonic/gin"
)

// PaymentHandler handles payment-related HTTP requests.
type PaymentHandler struct {
	createPaymentUC *payment.CreatePaymentUseCase
}

// NewPaymentHandler creates a new payment handler.
func NewPaymentHandler(createPaymentUC *payment.CreatePaymentUseCase) *PaymentHandler {
	return &PaymentHandler{
		createPaymentUC: createPaymentUC,
	}
}

// CreateCheckout handles POST /api/v1/payments/checkout
// Creates a Mercado Pago preference with the provided access token.
func (h *PaymentHandler) CreateCheckout(c *gin.Context) {
	var req dtos.CreatePaymentRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponseDTO{
			Success: false,
			Error:   "Invalid request: " + err.Error(),
			Code:    "VALIDATION_ERROR",
		})
		return
	}

	// Map DTO to use case command
	cmd := payment.CreatePaymentCommand{
		GymSlug:           req.GymSlug,
		Amount:            req.Amount,
		Title:             req.Title,
		Description:       req.Description,
		PayerEmail:        req.PayerEmail,
		ExternalReference: req.ExternalReference,
		MPAccessToken:     req.MPAccessToken,
		SuccessURL:        req.SuccessURL,
		FailureURL:        req.FailureURL,
		PendingURL:        req.PendingURL,
	}

	result, err := h.createPaymentUC.Execute(c.Request.Context(), cmd)
	if err != nil {
		log.Printf("CreateCheckout error: %v", err)
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponseDTO{
			Success: false,
			Error:   "Internal server error",
			Code:    "INTERNAL_ERROR",
		})
		return
	}

	// Map result to response DTO
	response := dtos.PaymentResponseDTO{
		Success:          result.Success,
		PreferenceID:     result.PreferenceID,
		InitPoint:        result.InitPoint,
		SandboxInitPoint: result.SandboxInitPoint,
		Error:            result.Error,
		ErrorCode:        result.ErrorCode,
	}

	if !response.Success {
		c.JSON(http.StatusBadRequest, response)
		return
	}

	c.JSON(http.StatusOK, response)
}
