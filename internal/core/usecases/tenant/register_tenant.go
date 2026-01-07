// Package tenant contains tenant-related use cases.
package tenant

// RegisterTenantCommand represents the input for registering a new tenant.
// This is a placeholder for future tenant management functionality.
type RegisterTenantCommand struct {
	GymSlug       string
	WebhookSecret string
	AccessToken   string
}

// RegisterTenantUseCase handles gym/tenant registration.
// Future: This will validate Mercado Pago credentials and store them.
type RegisterTenantUseCase struct {
	// Future dependencies will be added here
}

// NewRegisterTenantUseCase creates a new instance.
func NewRegisterTenantUseCase() *RegisterTenantUseCase {
	return &RegisterTenantUseCase{}
}
