// Package mercadopago provides webhook handling for Mercado Pago.
package mercadopago

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

// WebhookHandler validates Mercado Pago webhook signatures.
type WebhookHandler struct {
}

// NewWebhookHandler creates a new webhook handler.
func NewWebhookHandler() *WebhookHandler {
	return &WebhookHandler{}
}

// ValidateSignature validates the x-signature header from Mercado Pago.
// This implements the services.WebhookValidator interface.
func (h *WebhookHandler) ValidateSignature(xSignature, xRequestID, dataID, secret string) bool {
	if xSignature == "" || xRequestID == "" || dataID == "" || secret == "" {
		return false
	}

	// Parse x-signature header (format: "ts=123456,v1=hash")
	parts := strings.Split(xSignature, ",")
	if len(parts) < 2 {
		return false
	}

	var timestamp, v1Hash string
	for _, part := range parts {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}
		switch kv[0] {
		case "ts":
			timestamp = kv[1]
		case "v1":
			v1Hash = kv[1]
		}
	}

	if timestamp == "" || v1Hash == "" {
		return false
	}

	// Build manifest: id:{data.id};request-id:{x-request-id};ts:{timestamp};
	manifest := fmt.Sprintf("id:%s;request-id:%s;ts:%s;", dataID, xRequestID, timestamp)

	// Compute HMAC-SHA256
	h2 := hmac.New(sha256.New, []byte(secret))
	h2.Write([]byte(manifest))
	expectedHash := hex.EncodeToString(h2.Sum(nil))

	return hmac.Equal([]byte(v1Hash), []byte(expectedHash))
}
