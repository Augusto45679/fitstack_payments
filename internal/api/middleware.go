// Package api contains the HTTP handlers and routing for the payment service.
package api

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CORSMiddleware handles Cross-Origin Resource Sharing.
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, X-Requested-With")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// RequestIDMiddleware adds a unique request ID to each request.
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// JWTAuthMiddleware validates JWT tokens from FitStack.
// This is a placeholder - actual implementation should validate tokens
// against FitStack Core or use a shared JWT secret.
func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(401, gin.H{
				"success": false,
				"error":   "Authorization header required",
				"code":    "UNAUTHORIZED",
			})
			return
		}

		// TODO: Implement actual JWT validation
		// Options:
		// 1. Validate signature using shared secret
		// 2. Call FitStack Core to validate token
		// 3. Use JWT library to decode and validate

		// For now, just check the header exists
		// In production, this should properly validate the token

		c.Next()
	}
}

// WebhookSecurityMiddleware validates Mercado Pago webhook signatures.
// This ensures webhooks are actually coming from Mercado Pago.
func WebhookSecurityMiddleware(webhookSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Mercado Pago sends these headers for webhook validation:
		// x-signature: ts=timestamp,v1=signature
		// x-request-id: unique request ID

		xSignature := c.GetHeader("x-signature")
		xRequestID := c.GetHeader("x-request-id")

		if webhookSecret == "" {
			// If no secret configured, skip validation (development mode)
			c.Next()
			return
		}

		if xSignature == "" || xRequestID == "" {
			// In production, you might want to reject requests without signatures
			// For now, we log and continue
			c.Next()
			return
		}

		// TODO: Implement proper signature validation
		// See: https://www.mercadopago.com.ar/developers/es/docs/your-integrations/notifications/webhooks

		c.Next()
	}
}
