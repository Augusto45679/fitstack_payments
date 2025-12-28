// Package django provides HTTP client for Django backend communication.
package django

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/fitstack/fitstack-payments/internal/core/domain"
)

// Client implements DjangoNotifier and GymCredentialProvider interfaces.
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewClient creates a new Django backend client.
func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// NotifyPaymentConfirmed sends payment confirmation to Django backend.
// POST /api/v1/payments/webhook-callback/
func (c *Client) NotifyPaymentConfirmed(ctx context.Context, payload domain.DjangoWebhookPayload) error {
	url := fmt.Sprintf("%s/api/v1/payments/webhook-callback/", c.baseURL)

	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return domain.NewServiceError(domain.ErrDjangoCallbackFailed,
			"failed to marshal payload", "MARSHAL_ERROR")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonBody))
	if err != nil {
		return domain.NewServiceError(domain.ErrDjangoCallbackFailed,
			"failed to create request", "REQUEST_ERROR")
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Secret", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return domain.NewServiceError(domain.ErrDjangoCallbackFailed,
			"request failed: "+err.Error(), "HTTP_ERROR")
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return domain.NewServiceError(domain.ErrDjangoCallbackFailed,
			fmt.Sprintf("Django returned status %d: %s", resp.StatusCode, string(body)),
			"DJANGO_ERROR")
	}

	return nil
}

// gymCredentialsResponse represents the response from Django for gym credentials.
type gymCredentialsResponse struct {
	GymSlug       string `json:"gym_slug"`
	WebhookSecret string `json:"webhook_secret"`
	AccessToken   string `json:"access_token"`
}

// GetWebhookSecret retrieves the webhook secret for a gym from Django.
// GET /api/v1/internal/gyms/:slug/credentials/
func (c *Client) GetWebhookSecret(ctx context.Context, gymSlug string) (string, error) {
	creds, err := c.getGymCredentials(ctx, gymSlug)
	if err != nil {
		return "", err
	}
	return creds.WebhookSecret, nil
}

// GetAccessToken retrieves the access token for a gym from Django.
func (c *Client) GetAccessToken(ctx context.Context, gymSlug string) (string, error) {
	creds, err := c.getGymCredentials(ctx, gymSlug)
	if err != nil {
		return "", err
	}
	return creds.AccessToken, nil
}

// getGymCredentials fetches gym credentials from Django.
func (c *Client) getGymCredentials(ctx context.Context, gymSlug string) (*gymCredentialsResponse, error) {
	url := fmt.Sprintf("%s/api/v1/internal/gyms/%s/credentials/", c.baseURL, gymSlug)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, domain.NewServiceError(domain.ErrGymNotFound,
			"failed to create request", "REQUEST_ERROR")
	}

	req.Header.Set("X-Internal-API-Key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, domain.NewServiceError(domain.ErrGymNotFound,
			"request failed: "+err.Error(), "HTTP_ERROR")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, domain.ErrGymNotFound
	}

	if resp.StatusCode != http.StatusOK {
		return nil, domain.NewServiceError(domain.ErrGymNotFound,
			fmt.Sprintf("Django returned status %d", resp.StatusCode), "DJANGO_ERROR")
	}

	var creds gymCredentialsResponse
	if err := json.NewDecoder(resp.Body).Decode(&creds); err != nil {
		return nil, domain.NewServiceError(domain.ErrGymNotFound,
			"failed to decode response", "DECODE_ERROR")
	}

	return &creds, nil
}
