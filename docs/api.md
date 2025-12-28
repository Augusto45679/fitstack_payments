# FitStack Payments API Documentation

## Base URL

- **Development**: `http://localhost:8080`
- **Production**: `https://api.fitstackapp.com/payments`

---

## Endpoints

### Create Checkout

Creates a Mercado Pago payment preference and returns the checkout URL.

```
POST /api/v1/payments/checkout
```

**Headers:**
| Header | Value | Required |
|--------|-------|----------|
| `Content-Type` | `application/json` | Yes |
| `Authorization` | `Bearer <jwt_token>` | Yes |

**Request Body:**
```json
{
  "gym_id": "string",       // Gym slug identifier (required)
  "amount": 5000.00,        // Amount in ARS (required, > 0)
  "title": "string",        // Payment title (required)
  "payer_email": "string"   // Payer email (required, valid email)
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "init_point": "https://www.mercadopago.com.ar/checkout/v1/redirect?pref_id=..."
}
```

**Error Responses:**

| Status | Code | Description |
|--------|------|-------------|
| `400` | `VALIDATION_ERROR` | Invalid request body |
| `403` | `PAYMENT_NOT_ENABLED` | Gym doesn't have payment integration |
| `404` | `GYM_NOT_FOUND` | Gym not found |
| `500` | `INVALID_TOKEN` | Gym has invalid MP token |
| `502` | `GATEWAY_ERROR` | Mercado Pago API error |

---

### Webhook

Receives payment notifications from Mercado Pago.

```
POST /webhook/:gym_slug
```

> ⚠️ This endpoint is called by Mercado Pago, not by clients.

**URL Parameters:**
| Parameter | Description |
|-----------|-------------|
| `gym_slug` | Gym identifier included in notification URL |

**Headers (from Mercado Pago):**
| Header | Description |
|--------|-------------|
| `x-signature` | HMAC signature for validation |
| `x-request-id` | Unique request identifier |

**Request Body (from Mercado Pago):**
```json
{
  "id": "12345",
  "type": "payment",
  "action": "payment.created",
  "data": {
    "id": "67890"
  },
  "live_mode": true,
  "date_created": "2024-01-15T10:30:00Z"
}
```

**Response:**
```json
{
  "status": "processed"
}
```

---

### Health Check

Check service health status.

```
GET /health
```

**Response (200 OK):**
```json
{
  "status": "ok",
  "service": "fitstack-payments"
}
```

---

## Error Response Format

All error responses follow this structure:

```json
{
  "success": false,
  "error": "Human-readable error message",
  "code": "ERROR_CODE"
}
```

### Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `VALIDATION_ERROR` | 400 | Invalid request data |
| `UNAUTHORIZED` | 401 | Missing or invalid JWT |
| `PAYMENT_NOT_ENABLED` | 403 | Gym payment feature disabled |
| `GYM_NOT_FOUND` | 404 | Gym doesn't exist |
| `INVALID_TOKEN` | 500 | Invalid MP access token |
| `GATEWAY_ERROR` | 502 | Mercado Pago API error |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

## Integration Examples

### JavaScript/TypeScript

```typescript
async function createCheckout(gymId: string, amount: number, title: string, email: string) {
  const response = await fetch('/api/v1/payments/checkout', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${getToken()}`
    },
    body: JSON.stringify({
      gym_id: gymId,
      amount: amount,
      title: title,
      payer_email: email
    })
  });

  const data = await response.json();
  
  if (data.success) {
    // Redirect user to Mercado Pago
    window.location.href = data.init_point;
  } else {
    console.error('Payment error:', data.error);
  }
}
```

### Python

```python
import requests

def create_checkout(gym_id: str, amount: float, title: str, email: str, token: str):
    response = requests.post(
        'https://api.fitstackapp.com/payments/api/v1/payments/checkout',
        headers={
            'Content-Type': 'application/json',
            'Authorization': f'Bearer {token}'
        },
        json={
            'gym_id': gym_id,
            'amount': amount,
            'title': title,
            'payer_email': email
        }
    )
    return response.json()
```
