// Package middlewares contains HTTP middlewares.
package middlewares

import (
	"strings"

	"github.com/gin-gonic/gin"
)

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
