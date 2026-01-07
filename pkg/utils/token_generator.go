// Package utils provides utility functions.
package utils

import (
	"crypto/rand"
	"encoding/hex"
)

// TokenGenerator generates secure random tokens.
type TokenGenerator struct {
}

// NewTokenGenerator creates a new token generator.
func NewTokenGenerator() *TokenGenerator {
	return &TokenGenerator{}
}

// GenerateToken generates a secure random token of specified byte length.
func (g *TokenGenerator) GenerateToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// GenerateWebhookSecret generates a 32-byte webhook secret.
func (g *TokenGenerator) GenerateWebhookSecret() (string, error) {
	return g.GenerateToken(32)
}

// GenerateAPIKey generates a 32-byte API key.
func (g *TokenGenerator) GenerateAPIKey() (string, error) {
	return g.GenerateToken(32)
}
