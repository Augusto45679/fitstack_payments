// Package entities contains core business entities.
package entities

// Tenant represents a gym (tenant) in the multi-tenant system.
// Each tenant has their own Mercado Pago credentials.
type Tenant struct {
	GymSlug       string
	WebhookSecret string
	AccessToken   string
	IsActive      bool
}

// GymCredentials holds the secrets needed for a gym.
// This represents the credentials required to interact with payment gateway.
type GymCredentials struct {
	GymSlug       string `json:"gym_slug"`
	WebhookSecret string `json:"webhook_secret"`
	AccessToken   string `json:"access_token,omitempty"`
}

// Validate checks if the credentials are complete.
func (c *GymCredentials) Validate() bool {
	return c.GymSlug != "" && c.WebhookSecret != ""
}
