// Package config provides Mercado Pago specific configuration.
package config

// MercadoPagoConfig holds Mercado Pago SDK settings.
type MercadoPagoConfig struct {
	WebhookBaseURL string
	Environment    string // "sandbox" or "production"
}

// LoadMercadoPagoConfig loads Mercado Pago configuration.
func LoadMercadoPagoConfig() *MercadoPagoConfig {
	return &MercadoPagoConfig{
		WebhookBaseURL: getEnv("WEBHOOK_BASE_URL", "https://api.fitstackapp.com"),
		Environment:    getEnv("MP_ENVIRONMENT", "production"),
	}
}
