// Package mercadopago implements the PaymentGateway interface using the Mercado Pago SDK.
package mercadopago

import (
	"context"
	"fmt"
	"time"

	"github.com/fitstack/fitstack-payments/internal/domain"
	"github.com/mercadopago/sdk-go/pkg/config"
	"github.com/mercadopago/sdk-go/pkg/payment"
	"github.com/mercadopago/sdk-go/pkg/preference"
)

// Adapter implements the domain.PaymentGateway interface using Mercado Pago SDK.
type Adapter struct {
	// No shared state - each request uses the gym's specific access token
}

// NewAdapter creates a new Mercado Pago adapter.
func NewAdapter() *Adapter {
	return &Adapter{}
}

// CreatePreference creates a payment preference in Mercado Pago.
// Each gym has its own access token, so we create a new client per request.
func (a *Adapter) CreatePreference(ctx context.Context, accessToken string, order domain.PaymentOrder) (*domain.PaymentPreference, error) {
	// Create config with the gym's specific access token
	cfg, err := config.New(accessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create MP config: %w", err)
	}

	// Create preference client
	client := preference.NewClient(cfg)

	// Build the preference request
	// external_reference is used to identify the gym when we receive the webhook
	externalRef := fmt.Sprintf("%s|%d", order.GymSlug, time.Now().UnixNano())

	// Build back URLs
	successURL := fmt.Sprintf("https://fitstackapp.com/gym/%s/payment/success", order.GymSlug)
	failureURL := fmt.Sprintf("https://fitstackapp.com/gym/%s/payment/failure", order.GymSlug)
	pendingURL := fmt.Sprintf("https://fitstackapp.com/gym/%s/payment/pending", order.GymSlug)
	notificationURL := fmt.Sprintf("https://api.fitstackapp.com/payments/webhook/%s", order.GymSlug)

	request := preference.Request{
		Items: []preference.ItemRequest{
			{
				Title:      order.Title,
				Quantity:   1,
				UnitPrice:  order.Amount,
				CurrencyID: "ARS", // Default to ARS, can be made configurable
			},
		},
		Payer: &preference.PayerRequest{
			Email: order.PayerEmail,
		},
		ExternalReference: externalRef,
		AutoReturn:        "approved",
		BackURLs: &preference.BackURLsRequest{
			Success: successURL,
			Failure: failureURL,
			Pending: pendingURL,
		},
		NotificationURL: notificationURL,
	}

	// Create the preference
	result, err := client.Create(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to create preference: %w", err)
	}

	return &domain.PaymentPreference{
		ID:        result.ID,
		InitPoint: result.InitPoint,
	}, nil
}

// GetPaymentInfo retrieves payment information from Mercado Pago.
// Used when processing webhooks to get the current payment status.
func (a *Adapter) GetPaymentInfo(ctx context.Context, accessToken string, paymentID string) (*domain.PaymentStatus, error) {
	// Create config with the gym's specific access token
	cfg, err := config.New(accessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create MP config: %w", err)
	}

	// Create payment client
	client := payment.NewClient(cfg)

	// Get payment info - SDK uses int for payment IDs
	var paymentIDInt int
	if _, err := fmt.Sscanf(paymentID, "%d", &paymentIDInt); err != nil {
		return nil, fmt.Errorf("invalid payment ID format: %w", err)
	}

	result, err := client.Get(ctx, paymentIDInt)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment info: %w", err)
	}

	// Use DateCreated directly (it's a time.Time, not a pointer)
	transactionDate := result.DateCreated
	if transactionDate.IsZero() {
		transactionDate = time.Now()
	}

	// Extract payer email (Payer is a struct, not a pointer)
	payerEmail := result.Payer.Email

	return &domain.PaymentStatus{
		PaymentID:       paymentID,
		Status:          result.Status,
		StatusDetail:    result.StatusDetail,
		ExternalRef:     result.ExternalReference,
		Amount:          result.TransactionAmount,
		PayerEmail:      payerEmail,
		TransactionDate: transactionDate,
	}, nil
}
