// Package handlers contains middleware for the payment service.
package handlers

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CORSMiddleware handles Cross-Origin Resource Sharing.
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, X-Request-ID")
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

// ServiceAuthMiddleware validates Bearer token for server-to-server communication.
// The checkout endpoint is called by Django, not by end users.
func ServiceAuthMiddleware() gin.HandlerFunc {
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

		// Expect: Bearer <token>
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(401, gin.H{
				"success": false,
				"error":   "Invalid authorization format",
				"code":    "UNAUTHORIZED",
			})
			return
		}

		// Store token for potential validation
		c.Set("service_token", parts[1])

		// TODO: Validate token against expected service key
		// In production, compare with PAYMENTS_SERVICE_API_KEY env var

		c.Next()
	}
}
