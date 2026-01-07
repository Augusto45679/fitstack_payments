// Package payment contains payment use cases.
package payment

import (
	"context"
	"log"

	"github.com/fitstack/fitstack-payments/internal/core/domain"
	"github.com/fitstack/fitstack-payments/internal/core/domain/repository_interfaces"
	"github.com/fitstack/fitstack-payments/internal/core/domain/value_objects"
)

// CreatePaymentCommand represents the input for creating a payment.
type CreatePaymentCommand struct {
	GymSlug           string
	Amount            float64
	Title             string
	Description       string
	PayerEmail        string
	ExternalReference string
	MPAccessToken     string
	SuccessURL        string
	FailureURL        string
	PendingURL        string
}

// PaymentPreferenceResult represents the result of creating a payment preference.
type PaymentPreferenceResult struct {
	Success          bool
	PreferenceID     string
	InitPoint        string
	SandboxInitPoint string
	Error            string
	ErrorCode        string
}

// CreatePaymentUseCase handles the creation of payment preferences.
type CreatePaymentUseCase struct {
	gateway        PaymentGateway
	tenantRepo     repository_interfaces.TenantRepository
}

// NewCreatePaymentUseCase creates a new instance of CreatePaymentUseCase.
func NewCreatePaymentUseCase(
	gateway PaymentGateway,
	tenantRepo repository_interfaces.TenantRepository,
) *CreatePaymentUseCase {
	return &CreatePaymentUseCase{
		gateway:    gateway,
		tenantRepo: tenantRepo,
	}
}

// Execute creates a payment preference.
func (uc *CreatePaymentUseCase) Execute(ctx context.Context, cmd CreatePaymentCommand) (*PaymentPreferenceResult, error) {
	// Validate required fields
	if cmd.MPAccessToken == "" {
		return &PaymentPreferenceResult{
			Success:   false,
			Error:     "mp_access_token is required",
			ErrorCode: "VALIDATION_ERROR",
		}, nil
	}

	if cmd.GymSlug == "" || cmd.Amount <= 0 || cmd.Title == "" {
		return &PaymentPreferenceResult{
			Success:   false,
			Error:     "gym_slug, amount, and title are required",
			ErrorCode: "VALIDATION_ERROR",
		}, nil
	}

	// Validate tenant ID format
	tenantID, err := value_objects.NewTenantID(cmd.GymSlug)
	if err != nil {
		return &PaymentPreferenceResult{
			Success:   false,
			Error:     "Invalid gym_slug format: " + err.Error(),
			ErrorCode: "VALIDATION_ERROR",
		}, nil
	}

	// Validate amount
	money, err := value_objects.NewMoney(cmd.Amount, "ARS")
	if err != nil {
		return &PaymentPreferenceResult{
			Success:   false,
			Error:     "Invalid amount: " + err.Error(),
			ErrorCode: "VALIDATION_ERROR",
		}, nil
	}

	if !money.IsPositive() {
		return &PaymentPreferenceResult{
			Success:   false,
			Error:     "Amount must be positive",
			ErrorCode: "VALIDATION_ERROR",
		}, nil
	}

	log.Printf("Creating payment preference for tenant %s, amount: %s", tenantID.Value(), money.String())

	// Create preference using the payment gateway
	result, err := uc.gateway.CreatePreference(ctx, cmd.MPAccessToken, cmd)
	if err != nil {
		log.Printf("Failed to create preference for gym %s: %v", cmd.GymSlug, err)
		return &PaymentPreferenceResult{
			Success:   false,
			Error:     "Failed to create payment preference",
			ErrorCode: "GATEWAY_ERROR",
		}, nil
	}

	log.Printf("Created preference %s for gym %s, amount: %.2f", 
		result.PreferenceID, cmd.GymSlug, cmd.Amount)

	return result, nil
}
