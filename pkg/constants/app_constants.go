// Package constants provides application-wide constants.
package constants

// Payment status constants
const (
	PaymentStatusApproved   = "approved"
	PaymentStatusPending    = "pending"
	PaymentStatusInProcess  = "in_process"
	PaymentStatusRejected   = "rejected"
	PaymentStatusCancelled  = "cancelled"
	PaymentStatusRefunded   = "refunded"
	PaymentStatusChargedBack = "charged_back"
)

// Currency codes
const (
	CurrencyARS = "ARS" // Argentine Peso
	CurrencyUSD = "USD" // US Dollar
	CurrencyBRL = "BRL" // Brazilian Real
)

// Default URLs
const (
	DefaultWebhookBaseURL = "https://api.fitstackapp.com"
	DefaultFrontendBaseURL = "https://fitstackapp.com"
)

// Service metadata
const (
	ServiceName    = "fitstack-payments"
	ServiceVersion = "2.0.0"
	APIVersion     = "v1"
)
