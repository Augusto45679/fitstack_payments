// Package repository_interfaces defines repository contracts for the domain layer.
// These interfaces are implemented by the infrastructure layer.
package repository_interfaces

import (
	"context"

	"github.com/fitstack/fitstack-payments/internal/core/domain/entities"
)

// PaymentRepository defines the interface for payment persistence operations.
// This is currently a placeholder for future database integration.
type PaymentRepository interface {
	// Save persists a payment to the data store.
	Save(ctx context.Context, payment *entities.Payment) error

	// FindByID retrieves a payment by its ID.
	FindByID(ctx context.Context, id string) (*entities.Payment, error)

	// FindByExternalReference retrieves a payment by external reference.
	FindByExternalReference(ctx context.Context, externalRef string) (*entities.Payment, error)

	// Update updates an existing payment.
	Update(ctx context.Context, payment *entities.Payment) error
}
