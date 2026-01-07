// Package value_objects contains immutable value objects for the domain.
package value_objects

import (
	"fmt"
)

// Money represents a monetary value with currency.
// This is an immutable value object that ensures proper decimal handling.
type Money struct {
	Amount   float64
	Currency string
}

// NewMoney creates a new Money value object.
func NewMoney(amount float64, currency string) (*Money, error) {
	if amount < 0 {
		return nil, fmt.Errorf("amount cannot be negative: %f", amount)
	}
	if currency == "" {
		return nil, fmt.Errorf("currency is required")
	}
	return &Money{
		Amount:   amount,
		Currency: currency,
	}, nil
}

// Equals checks if two Money values are equal.
func (m *Money) Equals(other *Money) bool {
	if other == nil {
		return false
	}
	return m.Amount == other.Amount && m.Currency == other.Currency
}

// IsZero checks if the amount is zero.
func (m *Money) IsZero() bool {
	return m.Amount == 0
}

// IsPositive checks if the amount is positive.
func (m *Money) IsPositive() bool {
	return m.Amount > 0
}

// String returns a string representation of the money.
func (m *Money) String() string {
	return fmt.Sprintf("%.2f %s", m.Amount, m.Currency)
}

// NewMoneyError creates a validation error for money.
func NewMoneyError(message string) error {
	return fmt.Errorf("money validation error: %s", message)
}

