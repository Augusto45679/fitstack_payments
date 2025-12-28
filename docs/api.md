# FitStack Payments API Documentation

## Overview

`fitstack-payments` is a Go microservice that handles payment processing for FitStack SaaS using Mercado Pago Checkout Pro.

**Technology Stack**: Go 1.22+, Gin Gonic, Mercado Pago SDK

---

## Security Model

### Server-to-Server Communication

The checkout endpoint is designed for **server-to-server** communication between Django and this microservice.

- **Authentication**: `Authorization: Bearer <SERVICE_API_KEY>`
- **Stateless Checkout**: The `mp_access_token` is passed in the request body (Django sends the decrypted token)
- **HTTPS Required**: All production traffic must use HTTPS

### Webhook Security

Mercado Pago webhooks are validated using:

- **URL-based identification**: `/webhooks/:gym_slug` identifies which gym's secret to use
- **Signature validation**: `x-signature` header is validated using HMAC-SHA256

---

## Base URLs

| Environment | URL |
|-------------|-----|
| Development | `http://localhost:8080` |
| Production | `https://payments.fitstackapp.com` |

---

## Endpoints

### `POST /api/v1/payments/checkout`

Creates a Mercado Pago payment preference.

**Authentication**: Required (`Authorization: Bearer <token>`)

**Headers:**
```
Content-Type: application/json
Authorization: Bearer <SERVICE_API_KEY>
```

**Request Body:**
```json
{
  "gym_slug": "sportlife",
  "amount": 15000.00,
  "title": "Plan Mensual Premium",
  "description": "Acceso ilimitado por 30 días",
  "payer_email": "cliente@email.com",
  "external_reference": "package_request_123",
  "mp_access_token": "APP_USR-xxxx-xxxx-xxxx",
  "success_url": "https://app.fitstackapp.com/payment/success",
  "failure_url": "https://app.fitstackapp.com/payment/failure",
  "pending_url": "https://app.fitstackapp.com/payment/pending"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `gym_slug` | string | Yes | Gym identifier |
| `amount` | float | Yes | Payment amount (> 0) |
| `title` | string | Yes | Payment title |
| `description` | string | No | Payment description |
| `payer_email` | string | Yes | Customer email |
| `external_reference` | string | Yes | Your reference ID |
| `mp_access_token` | string | Yes | Mercado Pago access token (decrypted) |
| `success_url` | string | No | Redirect on success |
| `failure_url` | string | No | Redirect on failure |
| `pending_url` | string | No | Redirect on pending |

**Response (200 OK):**
```json
{
  "success": true,
  "preference_id": "123456789-abc",
  "init_point": "https://www.mercadopago.com.ar/checkout/v1/redirect?pref_id=...",
  "sandbox_init_point": "https://sandbox.mercadopago.com.ar/checkout/v1/redirect?pref_id=..."
}
```

**Response (400 Bad Request):**
```json
{
  "success": false,
  "error": "mp_access_token is required",
  "error_code": "VALIDATION_ERROR"
}
```

**Response (401 Unauthorized):**
```json
{
  "success": false,
  "error": "Authorization header required",
  "code": "UNAUTHORIZED"
}
```

---

### `POST /webhooks/:gym_slug`

Receives Mercado Pago IPN (Instant Payment Notification).

**Authentication**: None (validates `x-signature` header)

**URL Parameter:**
- `:gym_slug` - Gym identifier (used to lookup webhook secret)

**Headers from Mercado Pago:**
```
x-signature: ts=1234567890,v1=abc123...
x-request-id: uuid-v4
```

**Request Body (from Mercado Pago):**
```json
{
  "id": 12345,
  "live_mode": true,
  "type": "payment",
  "date_created": "2025-12-28T10:30:00.000-03:00",
  "action": "payment.created",
  "data": {
    "id": "67890123456"
  }
}
```

**Response (200 OK):**
```json
{
  "status": "processed"
}
```

**Processing Flow:**
1. Extract `gym_slug` from URL
2. Fetch webhook secret from Django: `GET /api/v1/internal/gyms/:slug/credentials/`
3. Validate `x-signature` using HMAC-SHA256
4. Fetch payment details from Mercado Pago
5. Notify Django: `POST /api/v1/payments/webhook-callback/`

---

### `GET /health`

Health check endpoint.

**Authentication**: None

**Response (200 OK):**
```json
{
  "status": "ok",
  "service": "fitstack-payments",
  "version": "1.0.0"
}
```

---

## Django Endpoints (Required)

This microservice expects the following endpoints in Django:

### `GET /api/v1/internal/gyms/:slug/credentials/`

Returns gym credentials for webhook validation.

**Headers:**
```
X-Internal-API-Key: <DJANGO_API_KEY>
```

**Response:**
```json
{
  "gym_slug": "sportlife",
  "webhook_secret": "mp-webhook-secret",
  "access_token": "APP_USR-xxxx"
}
```

### `POST /api/v1/payments/webhook-callback/`

Receives payment confirmations from this microservice.

**Headers:**
```
Content-Type: application/json
X-Webhook-Secret: <DJANGO_API_KEY>
```

**Request Body:**
```json
{
  "event": "payment.approved",
  "gym_slug": "sportlife",
  "external_reference": "package_request_123",
  "payment_id": "67890123456",
  "payment_status": "approved",
  "payment_type": "credit_card",
  "amount": 15000.00,
  "payer_email": "cliente@email.com",
  "timestamp": "2025-12-28T10:30:00Z"
}
```

---

## Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `VALIDATION_ERROR` | 400 | Invalid request data |
| `UNAUTHORIZED` | 401 | Missing/invalid Authorization |
| `GYM_NOT_FOUND` | 404 | Gym not found |
| `GATEWAY_ERROR` | 500 | Mercado Pago API error |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

## Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `PORT` | No | 8080 | Server port |
| `GIN_MODE` | No | debug | Gin mode (debug/release) |
| `DJANGO_BACKEND_URL` | Yes | - | Django API base URL |
| `DJANGO_API_KEY` | Yes | - | API key for Django communication |
