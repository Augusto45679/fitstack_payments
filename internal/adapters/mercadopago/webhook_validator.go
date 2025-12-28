// Package mercadopago provides Mercado Pago webhook signature validation.
package mercadopago

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strings"
)

// WebhookValidator validates Mercado Pago webhook signatures.
type WebhookValidator struct{}

// NewWebhookValidator creates a new webhook validator.
func NewWebhookValidator() *WebhookValidator {
	return &WebhookValidator{}
}

// ValidateSignature validates the x-signature header from Mercado Pago.
// See: https://www.mercadopago.com.ar/developers/es/docs/your-integrations/notifications/webhooks
//
// The x-signature header contains: ts=<timestamp>,v1=<signature>
// The signature is HMAC-SHA256 of: id:<data.id>;request-id:<x-request-id>;ts:<timestamp>;
func (v *WebhookValidator) ValidateSignature(xSignature, xRequestID, dataID, secret string) bool {
	if xSignature == "" || secret == "" {
		return false
	}

	// Parse x-signature header
	ts, hash := parseSignatureHeader(xSignature)
	if ts == "" || hash == "" {
		return false
	}

	// Build the manifest string
	// Format: id:<data.id>;request-id:<x-request-id>;ts:<timestamp>;
	manifest := buildManifest(dataID, xRequestID, ts)

	// Calculate expected signature
	expectedHash := calculateHMAC(manifest, secret)

	// Compare signatures (constant-time comparison)
	return hmac.Equal([]byte(hash), []byte(expectedHash))
}

// parseSignatureHeader extracts ts and v1 values from x-signature header.
func parseSignatureHeader(header string) (ts, hash string) {
	// Match ts=value and v1=value patterns
	tsRegex := regexp.MustCompile(`ts=([^,]+)`)
	v1Regex := regexp.MustCompile(`v1=([^,]+)`)

	tsMatch := tsRegex.FindStringSubmatch(header)
	if len(tsMatch) > 1 {
		ts = tsMatch[1]
	}

	v1Match := v1Regex.FindStringSubmatch(header)
	if len(v1Match) > 1 {
		hash = v1Match[1]
	}

	return ts, hash
}

// buildManifest constructs the string to be signed.
func buildManifest(dataID, requestID, ts string) string {
	var parts []string

	if dataID != "" {
		parts = append(parts, "id:"+dataID)
	}
	if requestID != "" {
		parts = append(parts, "request-id:"+requestID)
	}
	if ts != "" {
		parts = append(parts, "ts:"+ts)
	}

	return strings.Join(parts, ";") + ";"
}

// calculateHMAC computes HMAC-SHA256 of the manifest.
func calculateHMAC(manifest, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(manifest))
	return hex.EncodeToString(h.Sum(nil))
}
