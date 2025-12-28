// Package mercadopago implements the PaymentGateway interface using the official SDK.
package mercadopago

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/fitstack/fitstack-payments/internal/core/domain"
	"github.com/mercadopago/sdk-go/pkg/config"
	"github.com/mercadopago/sdk-go/pkg/payment"
	"github.com/mercadopago/sdk-go/pkg/preference"
)

// Adapter implements ports.PaymentGateway using Mercado Pago SDK.
type Adapter struct{}

// NewAdapter creates a new Mercado Pago adapter.
func NewAdapter() *Adapter {
	return &Adapter{}
}

// CreatePreference creates a Checkout Pro preference.
func (a *Adapter) CreatePreference(ctx context.Context, accessToken string, req domain.PaymentRequest) (*domain.PaymentResponse, error) {
	cfg, err := config.New(accessToken)
	if err != nil {
		return nil, domain.NewServiceError(domain.ErrPaymentGatewayError,
			"failed to create MP config", "MP_CONFIG_ERROR")
	}

	client := preference.NewClient(cfg)

	// Build back URLs
	successURL := req.SuccessURL
	if successURL == "" {
		successURL = fmt.Sprintf("https://fitstackapp.com/gym/%s/payment/success", req.GymSlug)
	}
	failureURL := req.FailureURL
	if failureURL == "" {
		failureURL = fmt.Sprintf("https://fitstackapp.com/gym/%s/payment/failure", req.GymSlug)
	}
	pendingURL := req.PendingURL
	if pendingURL == "" {
		pendingURL = fmt.Sprintf("https://fitstackapp.com/gym/%s/payment/pending", req.GymSlug)
	}

	prefRequest := preference.Request{
		Items: []preference.ItemRequest{
			{
				Title:       req.Title,
				Description: req.Description,
				Quantity:    1,
				UnitPrice:   req.Amount,
				CurrencyID:  "ARS",
			},
		},
		Payer: &preference.PayerRequest{
			Email: req.PayerEmail,
		},
		ExternalReference: req.ExternalReference,
		AutoReturn:        "approved",
		BackURLs: &preference.BackURLsRequest{
			Success: successURL,
			Failure: failureURL,
			Pending: pendingURL,
		},
		NotificationURL: fmt.Sprintf("https://api.fitstackapp.com/webhooks/%s", req.GymSlug),
	}

	result, err := client.Create(ctx, prefRequest)
	if err != nil {
		return nil, domain.NewServiceError(domain.ErrPaymentGatewayError,
			"failed to create preference: "+err.Error(), "MP_PREFERENCE_ERROR")
	}

	return &domain.PaymentResponse{
		Success:          true,
		PreferenceID:     result.ID,
		InitPoint:        result.InitPoint,
		SandboxInitPoint: result.SandboxInitPoint,
	}, nil
}

// GetPaymentInfo retrieves payment details from Mercado Pago.
func (a *Adapter) GetPaymentInfo(ctx context.Context, accessToken string, paymentID string) (*domain.PaymentInfo, error) {
	cfg, err := config.New(accessToken)
	if err != nil {
		return nil, domain.NewServiceError(domain.ErrPaymentGatewayError,
			"failed to create MP config", "MP_CONFIG_ERROR")
	}

	client := payment.NewClient(cfg)

	id, err := strconv.Atoi(paymentID)
	if err != nil {
		return nil, domain.NewServiceError(domain.ErrInvalidRequest,
			"invalid payment ID format", "INVALID_PAYMENT_ID")
	}

	result, err := client.Get(ctx, id)
	if err != nil {
		return nil, domain.NewServiceError(domain.ErrPaymentGatewayError,
			"failed to get payment info: "+err.Error(), "MP_PAYMENT_ERROR")
	}

	dateApproved := result.DateApproved
	if dateApproved.IsZero() {
		dateApproved = time.Now()
	}

	return &domain.PaymentInfo{
		PaymentID:         paymentID,
		Status:            result.Status,
		StatusDetail:      result.StatusDetail,
		ExternalReference: result.ExternalReference,
		Amount:            result.TransactionAmount,
		Currency:          result.CurrencyID,
		PaymentMethod:     result.PaymentMethodID,
		PaymentType:       result.PaymentTypeID,
		PayerEmail:        result.Payer.Email,
		DateApproved:      dateApproved,
	}, nil
}
