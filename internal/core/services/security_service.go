// Package services contains domain services.
package services

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

// SecurityService handles security-related operations.
type SecurityService struct {
}

// NewSecurityService creates a new security service.
func NewSecurityService() *SecurityService {
	return &SecurityService{}
}

// ValidateSignature validates a Mercado Pago webhook signature.
// This implements the WebhookValidator interface.
func (s *SecurityService) ValidateSignature(xSignature, xRequestID, dataID, secret string) bool {
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
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(manifest))
	expectedHash := hex.EncodeToString(h.Sum(nil))

	return hmac.Equal([]byte(v1Hash), []byte(expectedHash))
}
