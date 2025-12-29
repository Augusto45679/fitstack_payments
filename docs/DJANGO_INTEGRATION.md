# FitStack Django - Payment Integration Guide

## Overview

This document describes the Django endpoints and models required to integrate with the `fitstack-payments` Go microservice.

---

## Architecture

```
┌─────────────┐      ┌─────────────┐      ┌─────────────┐      ┌─────────────┐
│   Frontend  │ ──── │   Django    │ ──── │ Go Payments │ ──── │ Mercado Pago│
│   (Next.js) │      │   Backend   │      │ Microservice│      │     API     │
└─────────────┘      └─────────────┘      └─────────────┘      └─────────────┘
                            │
                    ┌───────┴───────┐
                    │   PostgreSQL  │
                    │  (encrypted   │
                    │   MP creds)   │
                    └───────────────┘
```

---

## Required Model Changes

### Gym Model Updates

Add these fields to `apps/core/models.py`:

```python
class Gym(TimeStampedModel):
    # ... existing fields ...
    
    # Payment Configuration (Tier PRO+)
    is_payment_enabled = models.BooleanField(
        default=False,
        help_text="Enable Mercado Pago integration"
    )
    mp_access_token = models.TextField(
        blank=True, 
        null=True,
        help_text="Mercado Pago Access Token (encrypted)"
    )
    mp_webhook_secret = models.TextField(
        blank=True, 
        null=True,
        help_text="Mercado Pago Webhook Secret (encrypted)"
    )
    mp_configured_at = models.DateTimeField(
        blank=True, 
        null=True,
        help_text="When MP credentials were configured"
    )
```

### PackageRequest Model Updates

Add payment tracking fields:

```python
class PackageRequest(GymTenantModel):
    # Add new statuses
    PENDING_PAYMENT = 'PENDING_PAYMENT'
    PAYMENT_FAILED = 'PAYMENT_FAILED'
    
    STATUS_CHOICES = [
        # ... existing ...
        (PENDING_PAYMENT, 'Pending Payment'),
        (PAYMENT_FAILED, 'Payment Failed'),
    ]
    
    # Add payment tracking
    mp_preference_id = models.CharField(
        max_length=100, 
        blank=True, 
        null=True,
        help_text="Mercado Pago preference ID"
    )
    mp_payment_id = models.CharField(
        max_length=100, 
        blank=True, 
        null=True,
        help_text="Mercado Pago payment ID"
    )
```

### Voucher Model Updates

Add Mercado Pago as payment method:

```python
class Voucher(GymTenantModel):
    PAYMENT_METHODS = [
        ('CASH', 'Cash'),
        ('TRANSFER', 'Transfer'),
        ('DEBIT', 'Debit Card'),
        ('MERCADOPAGO', 'Mercado Pago'),  # New
        ('OTHER', 'Other'),
    ]
```

---

## Encryption Service

Create `apps/payments/encryption.py`:

```python
from cryptography.fernet import Fernet
from django.conf import settings

class CredentialEncryption:
    """Encrypt/decrypt sensitive payment credentials."""
    
    def __init__(self):
        key = settings.PAYMENT_ENCRYPTION_KEY
        if not key:
            raise ValueError("PAYMENT_ENCRYPTION_KEY not configured")
        self.fernet = Fernet(key.encode())
    
    def encrypt(self, value: str) -> str:
        """Encrypt a string value."""
        if not value:
            return ""
        return self.fernet.encrypt(value.encode()).decode()
    
    def decrypt(self, value: str) -> str:
        """Decrypt an encrypted value."""
        if not value:
            return ""
        return self.fernet.decrypt(value.encode()).decode()

# Singleton instance
encryption = CredentialEncryption()
```

**Generate encryption key:**
```python
from cryptography.fernet import Fernet
print(Fernet.generate_key().decode())
# Add to .env: PAYMENT_ENCRYPTION_KEY=<generated_key>
```

---

## API Endpoints

### 1. Configure Gym Payment Credentials

**Endpoint:** `POST /api/v1/gyms/:slug/payment-config/`

**Permission:** Gym Admin only

**View (`apps/payments/views.py`):**

```python
from rest_framework.views import APIView
from rest_framework.response import Response
from rest_framework import status
from apps.core.permissions import IsGymAdmin
from apps.payments.encryption import encryption

class GymPaymentConfigView(APIView):
    permission_classes = [IsGymAdmin]
    
    def post(self, request, slug):
        """Configure MP credentials for a gym."""
        gym = get_object_or_404(Gym, slug=slug)
        
        # Verify user is admin of this gym
        if not request.user.gyms.filter(id=gym.id).exists():
            return Response(
                {"error": "Not authorized for this gym"},
                status=status.HTTP_403_FORBIDDEN
            )
        
        access_token = request.data.get('mp_access_token')
        webhook_secret = request.data.get('mp_webhook_secret')
        is_enabled = request.data.get('is_payment_enabled', True)
        
        if not access_token or not webhook_secret:
            return Response(
                {"error": "access_token and webhook_secret required"},
                status=status.HTTP_400_BAD_REQUEST
            )
        
        # Encrypt and save
        gym.mp_access_token = encryption.encrypt(access_token)
        gym.mp_webhook_secret = encryption.encrypt(webhook_secret)
        gym.is_payment_enabled = is_enabled
        gym.mp_configured_at = timezone.now()
        gym.save()
        
        return Response({
            "success": True,
            "message": "Payment configuration saved",
            "is_payment_enabled": gym.is_payment_enabled
        })
    
    def get(self, request, slug):
        """Get payment config status (not credentials)."""
        gym = get_object_or_404(Gym, slug=slug)
        
        return Response({
            "is_payment_enabled": gym.is_payment_enabled,
            "is_configured": bool(gym.mp_access_token),
            "configured_at": gym.mp_configured_at
        })
    
    def delete(self, request, slug):
        """Remove payment configuration."""
        gym = get_object_or_404(Gym, slug=slug)
        
        gym.mp_access_token = None
        gym.mp_webhook_secret = None
        gym.is_payment_enabled = False
        gym.mp_configured_at = None
        gym.save()
        
        return Response({"success": True})
```

---

### 2. Internal: Get Gym Credentials (Microservice Only)

**Endpoint:** `GET /api/v1/internal/gyms/:slug/credentials/`

**Auth:** `X-Internal-API-Key` header

**View:**

```python
from django.conf import settings

class InternalGymCredentialsView(APIView):
    """Internal endpoint for Go microservice."""
    
    def get(self, request, slug):
        # Validate internal API key
        api_key = request.headers.get('X-Internal-API-Key')
        if api_key != settings.INTERNAL_API_KEY:
            return Response(
                {"error": "Invalid API key"},
                status=status.HTTP_401_UNAUTHORIZED
            )
        
        gym = get_object_or_404(Gym, slug=slug)
        
        if not gym.is_payment_enabled:
            return Response(
                {"error": "Payments not enabled"},
                status=status.HTTP_403_FORBIDDEN
            )
        
        # Decrypt credentials
        return Response({
            "gym_slug": gym.slug,
            "access_token": encryption.decrypt(gym.mp_access_token),
            "webhook_secret": encryption.decrypt(gym.mp_webhook_secret)
        })
```

---

### 3. Internal: Receive Payment Webhook Callback

**Endpoint:** `POST /api/v1/payments/webhook-callback/`

**Auth:** `X-Webhook-Secret` header

**View:**

```python
class PaymentWebhookCallbackView(APIView):
    """Receive payment confirmations from Go microservice."""
    
    def post(self, request):
        # Validate webhook secret
        secret = request.headers.get('X-Webhook-Secret')
        if secret != settings.DJANGO_API_KEY:
            return Response(
                {"error": "Invalid webhook secret"},
                status=status.HTTP_401_UNAUTHORIZED
            )
        
        data = request.data
        event = data.get('event')
        external_ref = data.get('external_reference')
        gym_slug = data.get('gym_slug')
        
        # Parse external_reference (format: "package_request_123")
        try:
            request_id = int(external_ref.split('_')[-1])
            pkg_request = PackageRequest.objects.get(id=request_id)
        except (ValueError, PackageRequest.DoesNotExist):
            return Response(
                {"error": "Package request not found"},
                status=status.HTTP_404_NOT_FOUND
            )
        
        if event == 'payment.approved':
            # Approve the package request
            pkg_request.status = PackageRequest.APPROVED
            pkg_request.mp_payment_id = data.get('payment_id')
            pkg_request.save()
            
            # Create voucher
            voucher = Voucher.objects.create(
                gym=pkg_request.gym,
                user=pkg_request.user,
                package_type=pkg_request.package_type,
                payment_method='MERCADOPAGO',
                payment_reference=data.get('payment_id'),
                amount=data.get('amount'),
            )
            
            # Create package
            package = Package.objects.create(
                gym=pkg_request.gym,
                user=pkg_request.user,
                package_type=pkg_request.package_type,
                voucher=voucher,
            )
            
            return Response({
                "success": True,
                "package_id": package.id,
                "voucher_code": voucher.code
            })
        
        elif event == 'payment.rejected':
            pkg_request.status = PackageRequest.PAYMENT_FAILED
            pkg_request.save()
            return Response({"success": True, "status": "rejected"})
        
        return Response({"success": True, "status": "processed"})
```

---

### 4. Update Package Request Flow

**Endpoint:** `POST /api/v1/packages/request/` (modify existing)

```python
import requests
from django.conf import settings

class PackageRequestViewSet(viewsets.ModelViewSet):
    
    def create(self, request):
        # ... existing validation ...
        
        gym = request.user.current_gym
        
        # Check if gym has payment enabled
        if gym.is_payment_enabled:
            # Create request in PENDING_PAYMENT status
            pkg_request = PackageRequest.objects.create(
                gym=gym,
                user=request.user,
                package_type_id=request.data['package_type'],
                status=PackageRequest.PENDING_PAYMENT
            )
            
            # Call payment microservice
            payment_response = self._create_payment(pkg_request)
            
            if payment_response.get('success'):
                pkg_request.mp_preference_id = payment_response['preference_id']
                pkg_request.save()
                
                return Response({
                    "id": pkg_request.id,
                    "status": "PENDING_PAYMENT",
                    "payment_url": payment_response['init_point'],
                    "sandbox_url": payment_response.get('sandbox_init_point')
                })
            else:
                pkg_request.delete()
                return Response(
                    {"error": "Failed to create payment"},
                    status=status.HTTP_500_INTERNAL_SERVER_ERROR
                )
        else:
            # Traditional flow - staff approval
            pkg_request = PackageRequest.objects.create(
                gym=gym,
                user=request.user,
                package_type_id=request.data['package_type'],
                status=PackageRequest.PENDING
            )
            
            return Response({
                "id": pkg_request.id,
                "status": "PENDING",
                "message": "Request submitted for staff approval"
            })
    
    def _create_payment(self, pkg_request):
        """Call Go microservice to create payment."""
        gym = pkg_request.gym
        package_type = pkg_request.package_type
        
        payload = {
            "gym_slug": gym.slug,
            "amount": float(package_type.price),
            "title": package_type.name,
            "description": package_type.description or "",
            "payer_email": pkg_request.user.email,
            "external_reference": f"package_request_{pkg_request.id}",
            "mp_access_token": encryption.decrypt(gym.mp_access_token),
            "success_url": f"{settings.FRONTEND_URL}/gym/{gym.slug}/payment/success",
            "failure_url": f"{settings.FRONTEND_URL}/gym/{gym.slug}/payment/failure",
            "pending_url": f"{settings.FRONTEND_URL}/gym/{gym.slug}/payment/pending"
        }
        
        try:
            response = requests.post(
                f"{settings.PAYMENTS_SERVICE_URL}/api/v1/payments/checkout",
                json=payload,
                headers={
                    "Authorization": f"Bearer {settings.PAYMENTS_SERVICE_API_KEY}",
                    "Content-Type": "application/json"
                },
                timeout=10
            )
            return response.json()
        except Exception as e:
            logger.error(f"Payment service error: {e}")
            return {"success": False, "error": str(e)}
```

---

## URL Configuration

Add to `apps/payments/urls.py`:

```python
from django.urls import path
from .views import (
    GymPaymentConfigView,
    InternalGymCredentialsView,
    PaymentWebhookCallbackView
)

urlpatterns = [
    # Admin endpoint
    path('gyms/<slug:slug>/payment-config/', 
         GymPaymentConfigView.as_view(), 
         name='gym-payment-config'),
    
    # Internal endpoints (microservice only)
    path('internal/gyms/<slug:slug>/credentials/', 
         InternalGymCredentialsView.as_view(), 
         name='internal-gym-credentials'),
    path('payments/webhook-callback/', 
         PaymentWebhookCallbackView.as_view(), 
         name='payment-webhook-callback'),
]
```

---

## Environment Variables

Add to Django settings:

```python
# Payment Microservice
PAYMENTS_SERVICE_URL = env('PAYMENTS_SERVICE_URL', default='http://localhost:8080')
PAYMENTS_SERVICE_API_KEY = env('PAYMENTS_SERVICE_API_KEY')

# Internal API
INTERNAL_API_KEY = env('INTERNAL_API_KEY')
DJANGO_API_KEY = env('DJANGO_API_KEY')

# Encryption
PAYMENT_ENCRYPTION_KEY = env('PAYMENT_ENCRYPTION_KEY')
```

---

## Migrations

After adding model fields:

```bash
python manage.py makemigrations core
python manage.py migrate
```

---

## Frontend Requirements

### Gym Admin Panel
- Add "Payment Configuration" section
- Form fields: Access Token, Webhook Secret
- Show configuration status

### Client Checkout
- Detect if gym has payments enabled
- Redirect to Mercado Pago when purchasing
- Handle success/failure/pending redirects

---

## Testing

### Test Payment Configuration
```bash
curl -X POST http://localhost:8000/api/v1/gyms/level-gym/payment-config/ \
  -H "Authorization: Bearer <admin_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "mp_access_token": "APP_USR-xxx",
    "mp_webhook_secret": "xxx",
    "is_payment_enabled": true
  }'
```

### Test Package Purchase (with payments)
```bash
curl -X POST http://localhost:8000/api/v1/packages/request/ \
  -H "Authorization: Bearer <client_token>" \
  -H "X-Gym-Slug: level-gym" \
  -H "Content-Type: application/json" \
  -d '{"package_type": 1}'
```

Expected response when payments enabled:
```json
{
  "id": 123,
  "status": "PENDING_PAYMENT",
  "payment_url": "https://www.mercadopago.com.ar/checkout/v1/redirect?pref_id=..."
}
```
