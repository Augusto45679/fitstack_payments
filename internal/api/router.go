// Package api contains the HTTP handlers and routing for the payment service.
package api

import (
	"github.com/gin-gonic/gin"
)

// SetupRouter configures the Gin router with all routes and middleware.
func SetupRouter(handler *Handler, ginMode string) *gin.Engine {
	// Set Gin mode
	gin.SetMode(ginMode)

	// Create router with default middleware (logger and recovery)
	router := gin.New()

	// Apply middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(CORSMiddleware())
	router.Use(RequestIDMiddleware())

	// Health check endpoint (no auth required)
	router.GET("/health", handler.Health)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Payment routes
		payments := v1.Group("/payments")
		{
			// Note: JWT validation middleware should be added here
			// payments.Use(JWTAuthMiddleware())
			payments.POST("/checkout", handler.CreateCheckout)
		}
	}

	// Webhook endpoint (uses gym slug in URL for identification)
	// This endpoint is called by Mercado Pago, so no JWT required
	// Security is handled by validating the webhook signature
	router.POST("/webhook/:gym_slug", handler.HandleWebhook)

	return router
}
