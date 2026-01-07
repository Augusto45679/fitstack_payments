// Package multitenancy provides context handling for multi-tenant operations.
package multitenancy

import (
	"context"

	"github.com/fitstack/fitstack-payments/internal/core/domain/value_objects"
)

// Context key for tenant ID
type contextKey string

const tenantIDKey contextKey = "tenant_id"

// ContextHandler manages tenant context.
type ContextHandler struct {
}

// NewContextHandler creates a new context handler.
func NewContextHandler() *ContextHandler {
	return &ContextHandler{}
}

// SetTenantID adds tenant ID to context.
func (h *ContextHandler) SetTenantID(ctx context.Context, tenantID *value_objects.TenantID) context.Context {
	return context.WithValue(ctx, tenantIDKey, tenantID)
}

// GetTenantID retrieves tenant ID from context.
func (h *ContextHandler) GetTenantID(ctx context.Context) (*value_objects.TenantID, bool) {
	tenantID, ok := ctx.Value(tenantIDKey).(*value_objects.TenantID)
	return tenantID, ok
}
