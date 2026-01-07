// Package repository_interfaces defines repository contracts.
package repository_interfaces

import (
	"context"

	"github.com/fitstack/fitstack-payments/internal/core/domain/entities"
	"github.com/fitstack/fitstack-payments/internal/core/domain/value_objects"
)

// TenantRepository defines the interface for tenant credential operations.
// This is implemented by the Django client in the infrastructure layer.
type TenantRepository interface {
	// GetCredentials retrieves the payment gateway credentials for a tenant.
	GetCredentials(ctx context.Context, tenantID *value_objects.TenantID) (*entities.GymCredentials, error)

	// GetWebhookSecret retrieves only the webhook secret for a tenant.
	GetWebhookSecret(ctx context.Context, tenantID *value_objects.TenantID) (string, error)

	// GetAccessToken retrieves only the access token for a tenant.
	GetAccessToken(ctx context.Context, tenantID *value_objects.TenantID) (string, error)
}
