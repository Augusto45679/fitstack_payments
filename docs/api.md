# FitStack Payments API Documentation

## Overview

`fitstack-payments` is a Go microservice that enables **multi-tenant payment processing** for FitStack SaaS using Mercado Pago Checkout Pro.

### Business Model

```
┌─────────────────────────────────────────────────────────────┐
│                    FitStack Platform                         │
│              (Technical Intermediary Only)                   │
│           Does NOT receive payments - facilitates            │
└─────────────────────────────────────────────────────────────┘
                              │
           ┌──────────────────┼──────────────────┐
           ▼                  ▼                  ▼
    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
    │  Level Gym  │    │ Sport Life  │    │  Iron Gym   │
    │  (Tenant)   │    │  (Tenant)   │    │  (Tenant)   │
    │             │    │             │    │             │
    │ MP Account: │    │ MP Account: │    │ MP Account: │
    │ @levelgym   │    │ @sportlife  │    │ @irongym    │
    └──────┬──────┘    └──────┬──────┘    └──────┬──────┘
           │                  │                  │
        Clients            Clients            Clients
        pay here           pay here           pay here
```

**Key Point**: When a client buys a package, the money goes **directly to the gym's Mercado Pago account**, not to FitStack.

---

## Technology Stack

- **Language**: Go 1.22+
- **Framework**: Gin Gonic
- **Payment SDK**: `github.com/mercadopago/sdk-go`

---

## Security Model

### User Roles

| Role | Description | Payment Actions |
|------|-------------|-----------------|
| **Client** | Gym member | Buys packages, redirected to MP checkout |
| **Staff** | Gym employee | Views payment status |
| **Admin** | Gym owner/manager | Configures MP credentials |
| **Superuser** | FitStack admin | Enables payment feature per gym |

### Authentication

| Endpoint | Auth Method |
|----------|-------------|
| `POST /api/v1/payments/checkout` | Bearer token (server-to-server) |
| `POST /webhooks/:gym_slug` | x-signature validation (HMAC-SHA256) |
| `GET /health` | None |

### Data Security

| Data | Storage | Notes |
|------|---------|-------|
| `mp_access_token` | Django (encrypted) | Decrypted only for API calls |
| `mp_webhook_secret` | Django (encrypted) | Used for signature validation |
| Card data | Never stored | Handled by Mercado Pago |

---

## Base URLs

| Environment | URL |
|-------------|-----|
| Development | `http://localhost:8080` |
| Production | `https://payments.fitstackapp.com` |

---

## Endpoints

### `POST /api/v1/payments/checkout`

Creates a Mercado Pago payment preference for a gym.

**Authentication**: `Authorization: Bearer <SERVICE_API_KEY>`

**Request:**
```json
{
  "gym_slug": "level-gym",
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
| `amount` | float | Yes | Payment amount in local currency |
| `title` | string | Yes | Payment title shown to client |
| `description` | string | No | Payment description |
| `payer_email` | string | Yes | Client email |
| `external_reference` | string | Yes | Your reference (e.g., package_request_id) |
| `mp_access_token` | string | Yes | Gym's MP access token (decrypted by Django) |
| `success_url` | string | No | Redirect URL on success |
| `failure_url` | string | No | Redirect URL on failure |
| `pending_url` | string | No | Redirect URL on pending |

**Response (200 OK):**
```json
{
  "success": true,
  "preference_id": "123456789-abc",
  "init_point": "https://www.mercadopago.com.ar/checkout/v1/redirect?pref_id=...",
  "sandbox_init_point": "https://sandbox.mercadopago.com.ar/checkout/v1/redirect?pref_id=..."
}
```

**Errors:**

| Code | Status | Description |
|------|--------|-------------|
| `VALIDATION_ERROR` | 400 | Missing required fields |
| `UNAUTHORIZED` | 401 | Missing/invalid Bearer token |
| `GATEWAY_ERROR` | 500 | Mercado Pago API error |

---

### `POST /webhooks/:gym_slug`

Receives Mercado Pago IPN (Instant Payment Notification).

**Authentication**: Validates `x-signature` header using HMAC-SHA256

**URL Parameter:**
- `:gym_slug` - Identifies which gym's webhook secret to use

**Headers (from Mercado Pago):**
```
x-signature: ts=1234567890,v1=abc123...
x-request-id: uuid-v4
```

**Request Body:**
```json
{
  "id": 12345,
  "live_mode": true,
  "type": "payment",
  "action": "payment.created",
  "data": {
    "id": "67890123456"
  }
}
```

**Response:**
```json
{
  "status": "processed"
}
```

**Processing Flow:**
1. Extract `gym_slug` from URL
2. Fetch `webhook_secret` from Django
3. Validate `x-signature` with HMAC-SHA256
4. Fetch payment details from Mercado Pago
5. Notify Django via webhook callback

---

### `GET /health`

Health check.

**Response:**
```json
{
  "status": "ok",
  "service": "fitstack-payments",
  "version": "1.0.0"
}
```

---

## Payment Flow

```
┌─────────┐     ┌─────────┐     ┌─────────┐     ┌─────────┐     ┌─────────┐
│ Client  │     │Frontend │     │ Django  │     │Go Micro │     │   MP    │
└────┬────┘     └────┬────┘     └────┬────┘     └────┬────┘     └────┬────┘
     │               │               │               │               │
     │ Select pkg    │               │               │               │
     │──────────────>│               │               │               │
     │               │ POST request  │               │               │
     │               │──────────────>│               │               │
     │               │               │ Get MP creds  │               │
     │               │               │──────────────>│               │
     │               │               │<──────────────│               │
     │               │               │ POST checkout │               │
     │               │               │──────────────>│               │
     │               │               │               │ Create pref   │
     │               │               │               │──────────────>│
     │               │               │               │<──────────────│
     │               │               │<──────────────│ init_point    │
     │               │<──────────────│               │               │
     │ Redirect      │               │               │               │
     │<──────────────│               │               │               │
     │               │               │               │               │
     │ Pay at MP     │               │               │               │
     │──────────────────────────────────────────────────────────────>│
     │               │               │               │               │
     │               │               │               │ Webhook       │
     │               │               │               │<──────────────│
     │               │               │ Callback      │               │
     │               │               │<──────────────│               │
     │               │               │ Approve pkg   │               │
     │               │               │               │               │
```

---

## Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `PORT` | No | 8080 | Server port |
| `GIN_MODE` | No | debug | Gin mode (debug/release) |
| `DJANGO_BACKEND_URL` | Yes | - | Django API base URL |
| `DJANGO_API_KEY` | Yes | - | API key for internal communication |

---

## Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `VALIDATION_ERROR` | 400 | Invalid request data |
| `UNAUTHORIZED` | 401 | Missing/invalid auth |
| `GYM_NOT_FOUND` | 404 | Gym not found |
| `GATEWAY_ERROR` | 500 | Mercado Pago error |
| `INTERNAL_ERROR` | 500 | Unexpected error |
