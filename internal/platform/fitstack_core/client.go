// Package fitstackcore implements the GymRepository and CoreNotifier interfaces
// by communicating with the FitStack Core Django API.
package fitstackcore

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/fitstack/fitstack-payments/internal/domain"
)

// Client implements domain.GymRepository and domain.CoreNotifier
// by making HTTP requests to the FitStack Core API.
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewClient creates a new FitStack Core client.
func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// gymConfigResponse represents the JSON response from Core API.
type gymConfigResponse struct {
	ID               string `json:"id"`
	Slug             string `json:"slug"`
	Name             string `json:"name"`
	AccessToken      string `json:"mp_access_token"` // Decrypted by Core
	IsPaymentEnabled bool   `json:"is_payment_enabled"`
}

// GetGymConfig fetches the payment configuration for a gym from FitStack Core.
// The Core API is responsible for decrypting the access token.
func (c *Client) GetGymConfig(ctx context.Context, gymSlug string) (*domain.GymConfig, error) {
	url := fmt.Sprintf("%s/api/internal/gyms/%s/payment-config/", c.baseURL, gymSlug)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add internal API authentication
	req.Header.Set("X-Internal-API-Key", c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Handle response codes
	switch resp.StatusCode {
	case http.StatusOK:
		// Success - continue
	case http.StatusNotFound:
		return nil, domain.ErrGymNotFound
	case http.StatusUnauthorized, http.StatusForbidden:
		return nil, fmt.Errorf("authentication failed with Core API")
	default:
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var gymResp gymConfigResponse
	if err := json.NewDecoder(resp.Body).Decode(&gymResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &domain.GymConfig{
		ID:               gymResp.ID,
		Slug:             gymResp.Slug,
		Name:             gymResp.Name,
		AccessToken:      gymResp.AccessToken,
		IsPaymentEnabled: gymResp.IsPaymentEnabled,
	}, nil
}

// paymentStatusRequest represents the JSON request to notify Core about payment status.
type paymentStatusRequest struct {
	PaymentID       string  `json:"payment_id"`
	GymSlug         string  `json:"gym_slug"`
	Status          string  `json:"status"`
	StatusDetail    string  `json:"status_detail"`
	ExternalRef     string  `json:"external_ref"`
	Amount          float64 `json:"amount"`
	PayerEmail      string  `json:"payer_email"`
	TransactionDate string  `json:"transaction_date"`
}

// NotifyPaymentStatus sends payment status updates to FitStack Core.
// The Core will handle updating membership status, sending notifications, etc.
func (c *Client) NotifyPaymentStatus(ctx context.Context, status *domain.PaymentStatus) error {
	url := fmt.Sprintf("%s/api/internal/payments/webhook/", c.baseURL)

	payload := paymentStatusRequest{
		PaymentID:       status.PaymentID,
		GymSlug:         status.GymSlug,
		Status:          status.Status,
		StatusDetail:    status.StatusDetail,
		ExternalRef:     status.ExternalRef,
		Amount:          status.Amount,
		PayerEmail:      status.PayerEmail,
		TransactionDate: status.TransactionDate.Format(time.RFC3339),
	}

	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-Internal-API-Key", c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Core API returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
