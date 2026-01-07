// Package payment contains payment use cases.
package payment

import (
	"context"

	"github.com/fitstack/fitstack-payments/internal/core/domain/entities"
	"github.com/fitstack/fitstack-payments/internal/core/domain/repository_interfaces"
	"github.com/fitstack/fitstack-payments/internal/core/domain/value_objects"
)

// GetPaymentStatusQuery represents a query for payment status.
type GetPaymentStatusQuery struct {
	PaymentID string
	GymSlug   string
}

// GetPaymentStatusUseCase handles queries for payment status.
// This is a placeholder for future functionality.
type GetPaymentStatusUseCase struct {
	gateway    PaymentGateway
	tenantRepo repository_interfaces.TenantRepository
}

// NewGetPaymentStatusUseCase creates a new instance.
func NewGetPaymentStatusUseCase(
	gateway PaymentGateway,
	tenantRepo repository_interfaces.TenantRepository,
) *GetPaymentStatusUseCase {
	return &GetPaymentStatusUseCase{
		gateway:    gateway,
		tenantRepo: tenantRepo,
	}
}

// Execute retrieves the status of a payment.
func (uc *GetPaymentStatusUseCase) Execute(ctx context.Context, query GetPaymentStatusQuery) (*entities.PaymentInfo, error) {
	// Validate tenant
	tenantID, err := value_objects.NewTenantID(query.GymSlug)
	if err != nil {
		return nil, err
	}

	// Get access token
	accessToken, err := uc.tenantRepo.GetAccessToken(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	// Get payment info from gateway
	return uc.gateway.GetPaymentInfo(ctx, accessToken, query.PaymentID)
}
