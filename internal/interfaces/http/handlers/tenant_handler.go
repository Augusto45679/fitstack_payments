// Package handlers contains tenant management handlers (placeholder).
package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// TenantHandler handles tenant-related HTTP requests.
// This is a placeholder for future tenant management functionality.
type TenantHandler struct {
}

// NewTenantHandler creates a new tenant handler.
func NewTenantHandler() *TenantHandler {
	return &TenantHandler{}
}

// RegisterTenant handles POST /api/v1/tenants (future endpoint)
func (h *TenantHandler) RegisterTenant(c *gin.Context) {
	// Future: Handle gym registration with Mercado Pago credentials
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "not implemented yet",
	})
}
