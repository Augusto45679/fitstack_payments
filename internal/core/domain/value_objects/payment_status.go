// Package value_objects contains immutable value objects.
package value_objects

import "fmt"

// PaymentStatus represents the status of a payment.
// This is a type-safe enum for payment statuses.
type PaymentStatus string

const (
	PaymentStatusApproved   PaymentStatus = "approved"
	PaymentStatusPending    PaymentStatus = "pending"
	PaymentStatusInProcess  PaymentStatus = "in_process"
	PaymentStatusRejected   PaymentStatus = "rejected"
	PaymentStatusCancelled  PaymentStatus = "cancelled"
	PaymentStatusRefunded   PaymentStatus = "refunded"
	PaymentStatusChargedBack PaymentStatus = "charged_back"
)

// Valid payment statuses
var validStatuses = map[PaymentStatus]bool{
	PaymentStatusApproved:   true,
	PaymentStatusPending:    true,
	PaymentStatusInProcess:  true,
	PaymentStatusRejected:   true,
	PaymentStatusCancelled:  true,
	PaymentStatusRefunded:   true,
	PaymentStatusChargedBack: true,
}

// NewPaymentStatus creates a new PaymentStatus from a string.
func NewPaymentStatus(status string) (PaymentStatus, error) {
	ps := PaymentStatus(status)
	if !ps.IsValid() {
		return "", fmt.Errorf("invalid payment status: %s", status)
	}
	return ps, nil
}

// IsValid checks if the payment status is valid.
func (ps PaymentStatus) IsValid() bool {
	return validStatuses[ps]
}

// IsApproved checks if the payment is approved.
func (ps PaymentStatus) IsApproved() bool {
	return ps == PaymentStatusApproved
}

// IsPending checks if the payment is pending.
func (ps PaymentStatus) IsPending() bool {
	return ps == PaymentStatusPending || ps == PaymentStatusInProcess
}

// IsRejected checks if the payment is rejected.
func (ps PaymentStatus) IsRejected() bool {
	return ps == PaymentStatusRejected
}

// IsFinal checks if this is a final status (no more updates expected).
func (ps PaymentStatus) IsFinal() bool {
	return ps == PaymentStatusApproved ||
		ps == PaymentStatusRejected ||
		ps == PaymentStatusCancelled ||
		ps == PaymentStatusRefunded ||
		ps == PaymentStatusChargedBack
}

// String returns the string representation of the status.
func (ps PaymentStatus) String() string {
	return string(ps)
}

// ToEventName converts the payment status to an event name.
func (ps PaymentStatus) ToEventName() string {
	switch ps {
	case PaymentStatusApproved:
		return "payment.approved"
	case PaymentStatusPending, PaymentStatusInProcess:
		return "payment.pending"
	case PaymentStatusRejected:
		return "payment.rejected"
	case PaymentStatusCancelled:
		return "payment.cancelled"
	case PaymentStatusRefunded:
		return "payment.refunded"
	default:
		return "payment.updated"
	}
}
