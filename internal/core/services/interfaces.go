// Package services contains domain services.
package services

// WebhookValidator validates webhook signatures.
type WebhookValidator interface {
	// ValidateSignature validates the x-signature header from the payment gateway.
	ValidateSignature(xSignature, xRequestID, dataID, secret string) bool
}

// CredentialEncryptor handles encryption of sensitive credentials.
// This is a placeholder for future security enhancements.
type CredentialEncryptor interface {
	// Encrypt encrypts a credential value.
	Encrypt(plaintext string) (string, error)

	// Decrypt decrypts a credential value.
	Decrypt(ciphertext string) (string, error)
}
