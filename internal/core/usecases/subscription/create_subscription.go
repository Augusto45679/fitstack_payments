// Package subscription contains subscription-related use cases.
package subscription

// CreateSubscriptionCommand represents the input for creating a subscription.
// This is a placeholder for future recurring payment functionality.
type CreateSubscriptionCommand struct {
	GymSlug       string
	CustomerEmail string
	PlanID        string
	Amount        float64
}

// CreateSubscriptionUseCase handles subscription creation.
// Future: This will handle recurring payment setup.
type CreateSubscriptionUseCase struct {
	// Future dependencies will be added here
}

// NewCreateSubscriptionUseCase creates a new instance.
func NewCreateSubscriptionUseCase() *CreateSubscriptionUseCase {
	return &CreateSubscriptionUseCase{}
}
