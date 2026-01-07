// Package value_objects contains immutable value objects.
package value_objects

import (
	"fmt"
	"regexp"
)

// TenantID represents a unique identifier for a tenant (gym).
// This ensures type-safe tenant identification with validation.
type TenantID struct {
	value string
}

// Gym slug validation: lowercase alphanumeric with hyphens
var tenantIDPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

// NewTenantID creates a new TenantID with validation.
func NewTenantID(gymSlug string) (*TenantID, error) {
	if gymSlug == "" {
		return nil, fmt.Errorf("tenant ID cannot be empty")
	}
	if len(gymSlug) < 3 {
		return nil, fmt.Errorf("tenant ID must be at least 3 characters")
	}
	if len(gymSlug) > 100 {
		return nil, fmt.Errorf("tenant ID must be at most 100 characters")
	}
	if !tenantIDPattern.MatchString(gymSlug) {
		return nil, fmt.Errorf("tenant ID must be lowercase alphanumeric with hyphens: %s", gymSlug)
	}
	return &TenantID{value: gymSlug}, nil
}

// Value returns the string value of the tenant ID.
func (t *TenantID) Value() string {
	return t.value
}

// String returns the string representation.
func (t *TenantID) String() string {
	return t.value
}

// Equals checks if two tenant IDs are equal.
func (t *TenantID) Equals(other *TenantID) bool {
	if other == nil {
		return false
	}
	return t.value == other.value
}
