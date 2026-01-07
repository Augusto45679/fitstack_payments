// Package mercadopago implements the PaymentGateway interface using Mercado Pago SDK.
package mercadopago

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/fitstack/fitstack-payments/internal/core/domain"
	"github.com/fitstack/fitstack-payments/internal/core/domain/entities"
	"github.com/fitstack/fitstack-payments/internal/core/usecases/payment"
	"github.com/mercadopago/sdk-go/pkg/config"
	mpPayment "github.com/mercadopago/sdk-go/pkg/payment"
	"github.com/mercadopago/sdk-go/pkg/preference"
)

// Adapter implements payment.PaymentGateway using Mercado Pago SDK.
type Adapter struct{}

// NewAdapter creates a new Mercado Pago adapter.
func NewAdapter() *Adapter {
	return &Adapter{}
}

// CreatePreference creates a Checkout Pro preference.
func (a *Adapter) CreatePreference(ctx context.Context, accessToken string, cmd payment.CreatePaymentCommand) (*payment.PaymentPreferenceResult, error) {
	cfg, err := config.New(accessToken)
	if err != nil {
		return nil, domain.NewServiceError(domain.ErrPaymentGatewayError,
			"failed to create MP config", "MP_CONFIG_ERROR")
	}

	client := preference.NewClient(cfg)

	// Build back URLs
	successURL := cmd.SuccessURL
	if successURL == "" {
		successURL = fmt.Sprintf("https://fitstackapp.com/gym/%s/payment/success", cmd.GymSlug)
	}
	failureURL := cmd.FailureURL
	if failureURL == "" {
		failureURL = fmt.Sprintf("https://fitstackapp.com/gym/%s/payment/failure", cmd.GymSlug)
	}
	pendingURL := cmd.PendingURL
	if pendingURL == "" {
		pendingURL = fmt.Sprintf("https://fitstackapp.com/gym/%s/payment/pending", cmd.GymSlug)
	}

	prefRequest := preference.Request{
		Items: []preference.ItemRequest{
			{
				Title:       cmd.Title,
				Description: cmd.Description,
				Quantity:    1,
				UnitPrice:   cmd.Amount,
				CurrencyID:  "ARS",
			},
		},
		Payer: &preference.PayerRequest{
			Email: cmd.PayerEmail,
		},
		ExternalReference: cmd.ExternalReference,
		AutoReturn:        "approved",
		BackURLs: &preference.BackURLsRequest{
			Success: successURL,
			Failure: failureURL,
			Pending: pendingURL,
		},
		NotificationURL: fmt.Sprintf("https://api.fitstackapp.com/webhooks/%s", cmd.GymSlug),
	}

	result, err := client.Create(ctx, prefRequest)
	if err != nil {
		return nil, domain.NewServiceError(domain.ErrPaymentGatewayError,
			"failed to create preference: "+err.Error(), "MP_PREFERENCE_ERROR")
	}

	return &payment.PaymentPreferenceResult{
		Success:          true,
		PreferenceID:     result.ID,
		InitPoint:        result.InitPoint,
		SandboxInitPoint: result.SandboxInitPoint,
	}, nil
}

// GetPaymentInfo retrieves payment details from Mercado Pago.
func (a *Adapter) GetPaymentInfo(ctx context.Context, accessToken string, paymentID string) (*entities.PaymentInfo, error) {
	cfg, err := config.New(accessToken)
	if err != nil {
		return nil, domain.NewServiceError(domain.ErrPaymentGatewayError,
			"failed to create MP config", "MP_CONFIG_ERROR")
	}

	client := mpPayment.NewClient(cfg)

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

	return &entities.PaymentInfo{
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
