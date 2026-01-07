// Package middlewares contains HTTP middlewares.
package middlewares

import (
	"github.com/fitstack/fitstack-payments/internal/core/domain/value_objects"
	"github.com/fitstack/fitstack-payments/internal/infrastructure/persistence/multitenancy"
	"github.com/gin-gonic/gin"
)

// TenantMiddleware extracts tenant information from the request.
// This is a placeholder for future multi-tenant request handling.
type TenantMiddleware struct {
	contextHandler *multitenancy.ContextHandler
}

// NewTenantMiddleware creates a new tenant middleware.
func NewTenantMiddleware(contextHandler *multitenancy.ContextHandler) *TenantMiddleware {
	return &TenantMiddleware{
		contextHandler: contextHandler,
	}
}

// ExtractTenant middleware extracts tenant from request path or header.
func (m *TenantMiddleware) ExtractTenant() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Future: Extract gym_slug from path or header
		// For now, we handle tenant in individual handlers via :gym_slug param
		
		// Example future usage:
		// gymSlug := c.Param("gym_slug")
		// if gymSlug != "" {
		//     tenantID, err := value_objects.NewTenantID(gymSlug)
		//     if err == nil {
		//         ctx := m.contextHandler.SetTenantID(c.Request.Context(), tenantID)
		//         c.Request = c.Request.WithContext(ctx)
		//     }
		// }

		c.Next()
	}
}
