// Package multitenancy provides tenant resolution from requests.
package multitenancy

import (
	"context"

	"github.com/fitstack/fitstack-payments/internal/core/domain/value_objects"
)

// TenantResolver extracts tenant information from request context.
type TenantResolver struct {
}

// NewTenantResolver creates a new tenant resolver.
func NewTenantResolver() *TenantResolver {
	return &TenantResolver{}
}

// ResolveTenant extracts the tenant ID from context.
// This is a placeholder for future middleware integration.
func (r *TenantResolver) ResolveTenant(ctx context.Context) (*value_objects.TenantID, error) {
	// Future: Extract from context set by middleware
	return nil, nil
}
