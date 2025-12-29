# FitStack Payments Microservice

Microservicio Go para procesamiento de pagos multi-tenant con Mercado Pago.

## 🎯 Propósito

Permite que **cada gimnasio** en FitStack reciba pagos directamente en su cuenta de Mercado Pago. FitStack actúa como intermediario técnico.

## 🏗️ Arquitectura

```
internal/
├── adapters/
│   ├── django/client.go           # HTTP client para Django
│   └── mercadopago/
│       ├── adapter.go             # SDK Mercado Pago
│       └── webhook_validator.go   # Validación x-signature
├── core/
│   ├── domain/                    # Entities + Errors
│   ├── ports/interfaces.go        # Interfaces
│   └── service/payment_service.go # Business logic
└── handlers/                      # HTTP handlers
```

## 🚀 Quick Start

### Requisitos
- Go 1.22+
- WSL (Ubuntu recomendado)

### Instalación

```bash
# En WSL
cd /mnt/h/fitstack_payments/fitstack_payments

# Instalar Go (si no está instalado)
chmod +x scripts/setup_go_wsl.sh
./scripts/setup_go_wsl.sh
source ~/.bashrc

# Descargar dependencias
go mod tidy

# Compilar
go build ./...

# Ejecutar
go run cmd/api/main.go
```

### Variables de Entorno

```env
PORT=8080
GIN_MODE=debug
DJANGO_BACKEND_URL=http://localhost:8000
DJANGO_API_KEY=your-api-key
```

## 📡 Endpoints

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| POST | `/api/v1/payments/checkout` | Bearer | Crear preferencia MP |
| POST | `/webhooks/:gym_slug` | x-signature | Webhook de MP |
| GET | `/health` | None | Health check |

## 📚 Documentación

| Documento | Contenido |
|-----------|-----------|
| [docs/api.md](docs/api.md) | API Reference |
| [docs/PAYMENTS_INTEGRATION.md](docs/PAYMENTS_INTEGRATION.md) | Flujo de integración |
| [docs/DJANGO_INTEGRATION.md](docs/DJANGO_INTEGRATION.md) | Guía para Django |

## 🔐 Seguridad

- Bearer token para checkout (server-to-server)
- HMAC-SHA256 para webhooks de MP
- Credenciales encriptadas en Django
