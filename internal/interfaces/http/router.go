// Package http contains HTTP interface layer setup.
package http

import (
	"github.com/fitstack/fitstack-payments/internal/interfaces/http/handlers"
	"github.com/fitstack/fitstack-payments/internal/interfaces/http/middlewares"
	"github.com/gin-gonic/gin"
)

// SetupRouter configures the Gin router with all routes.
func SetupRouter(
	paymentHandler *handlers.PaymentHandler,
	webhookHandler *handlers.WebhookHandler,
	healthHandler *handlers.HealthHandler,
	ginMode string,
) *gin.Engine {
	gin.SetMode(ginMode)

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middlewares.CORSMiddleware())
	router.Use(middlewares.RequestIDMiddleware())

	// Health check (public)
	router.GET("/health", healthHandler.Health)

	// API v1 routes (requires Bearer auth)
	v1 := router.Group("/api/v1")
	{
		payments := v1.Group("/payments")
		payments.Use(middlewares.ServiceAuthMiddleware())
		{
			payments.POST("/checkout", paymentHandler.CreateCheckout)
		}
	}

	// Webhook endpoint (public, validates x-signature)
	router.POST("/webhooks/:gym_slug", webhookHandler.HandleWebhook)

	return router
}
