// Package dtos contains HTTP data transfer objects.
package dtos

// WebhookNotificationDTO represents a webhook notification from Mercado Pago.
type WebhookNotificationDTO struct {
	ID          int64  `json:"id"`
	LiveMode    bool   `json:"live_mode"`
	Type        string `json:"type"`
	DateCreated string `json:"date_created"`
	UserID      int64  `json:"user_id,omitempty"`
	APIVersion  string `json:"api_version"`
	Action      string `json:"action"`
	Data        struct {
		ID string `json:"id"`
	} `json:"data"`
}
