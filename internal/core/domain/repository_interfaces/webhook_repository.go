// Package repository_interfaces defines repository contracts.
package repository_interfaces

import (
	"context"

	"github.com/fitstack/fitstack-payments/internal/core/domain/entities"
)

// WebhookRepository defines the interface for webhook event persistence.
// This enables webhook deduplication and audit trail.
type WebhookRepository interface {
	// SaveEvent persists a webhook event.
	SaveEvent(ctx context.Context, event *entities.WebhookEvent) error

	// FindByID retrieves a webhook event by ID.
	FindByID(ctx context.Context, id string) (*entities.WebhookEvent, error)

	// EventExists checks if a webhook event has already been processed.
	EventExists(ctx context.Context, eventID string) (bool, error)
}
