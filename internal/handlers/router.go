// Package handlers contains the HTTP handlers and routing.
package handlers

import (
	"github.com/gin-gonic/gin"
)

// SetupRouter configures the Gin router with all routes.
func SetupRouter(handler *PaymentHandler, ginMode string) *gin.Engine {
	gin.SetMode(ginMode)

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(CORSMiddleware())
	router.Use(RequestIDMiddleware())

	// Health check (public)
	router.GET("/health", handler.Health)

	// API v1 routes (requires Bearer auth)
	v1 := router.Group("/api/v1")
	{
		payments := v1.Group("/payments")
		payments.Use(ServiceAuthMiddleware())
		{
			payments.POST("/checkout", handler.CreateCheckout)
		}
	}

	// Webhook endpoint (public, validates x-signature)
	router.POST("/webhooks/:gym_slug", handler.HandleWebhook)

	return router
}
