// Package entities contains core business entities.
package entities

import "time"

// Subscription represents a recurring payment subscription.
// This is a placeholder for future subscription functionality.
type Subscription struct {
	ID                string
	GymSlug           string
	CustomerEmail     string
	PlanID            string
	Status            string
	Amount            float64
	Currency          string
	BillingCycle      string // "monthly", "yearly", etc.
	NextBillingDate   time.Time
	SubscriptionStart time.Time
	SubscriptionEnd   *time.Time
}

// IsActive checks if the subscription is currently active.
func (s *Subscription) IsActive() bool {
	return s.Status == "active"
}

// IsCancelled checks if the subscription has been cancelled.
func (s *Subscription) IsCancelled() bool {
	return s.Status == "cancelled"
}
