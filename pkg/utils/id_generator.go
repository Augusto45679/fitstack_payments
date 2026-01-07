// Package utils provides utility functions.
package utils

import (
	"github.com/google/uuid"
)

// IDGenerator generates unique identifiers.
type IDGenerator struct {
}

// NewIDGenerator creates a new ID generator.
func NewIDGenerator() *IDGenerator {
	return &IDGenerator{}
}

// GenerateUUID generates a new UUID v4.
func (g *IDGenerator) GenerateUUID() string {
	return uuid.New().String()
}

// GeneratePaymentID generates a unique payment ID.
func (g *IDGenerator) GeneratePaymentID() string {
	return "pay_" + g.GenerateUUID()
}

// GenerateTransactionID generates a unique transaction ID.
func (g *IDGenerator) GenerateTransactionID() string {
	return "txn_" + g.GenerateUUID()
}
