# Plan de Implementación - Integración Mercado Pago
## FitStack Payments Microservice

**Fecha:** 2026-01-07  
**Versión:** 1.0  
**Arquitectura:** Clean Architecture (Go)

---

## Resumen Ejecutivo

Implementación completa de integración con Mercado Pago para plataforma multitenant FitStack. El microservicio en Go manejará OAuth, creación de preferencias de pago, webhooks y reconciliación, permitiendo que cada gimnasio conecte su cuenta de Mercado Pago para recibir pagos directamente.

**Duración estimada:** 6-8 semanas  
**Sprints:** 5 sprints de 1-2 semanas cada uno

---

## SPRINT 0: Setup & Fundamentos
**Duración:** 1 semana  
**Objetivos:** Configurar infraestructura base y validar accesos

### Tareas

#### 1. Configuración de Entorno
- [ ] Crear cuenta de desarrollador en Mercado Pago
- [ ] Obtener credenciales de sandbox (`MP_CLIENT_ID`, `MP_CLIENT_SECRET`)
- [ ] Configurar variables de entorno
  ```env
  MP_CLIENT_ID=TEST-xxxxx
  MP_CLIENT_SECRET=TEST-xxxxx
  MP_ENV=sandbox
  BASE_URL=http://localhost:8080
  DB_URL=postgresql://user:pass@localhost:5432/fitstack_payments
  KMS_KEY=base64_encoded_key_for_encryption
  DJANGO_NOTIFY_URL=http://localhost:8000/api/mp/notify
  LOG_LEVEL=debug
  ```

#### 2. Database Schema
- [ ] Crear migración inicial para `mp_accounts`
  ```sql
  CREATE TABLE mp_accounts (
    id SERIAL PRIMARY KEY,
    gimnasio_id UUID NOT NULL UNIQUE,
    mp_user_id BIGINT NOT NULL,
    access_token BYTEA NOT NULL,
    refresh_token BYTEA,
    token_expires_at TIMESTAMP WITH TIME ZONE,
    scope TEXT,
    key_version INT NOT NULL DEFAULT 1,      -- Para key rotation
    kms_key_id TEXT,                         -- Identificador de key en KMS
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT now()
  );
  CREATE INDEX idx_mp_accounts_gimnasio ON mp_accounts(gimnasio_id);
  CREATE INDEX idx_mp_accounts_key_version ON mp_accounts(key_version);
  ```

- [ ] Crear migración para `mp_preferences`
  ```sql
  CREATE TABLE mp_preferences (
    id SERIAL PRIMARY KEY,
    gimnasio_id UUID NOT NULL,
    orden_id TEXT NOT NULL UNIQUE,
    preference_id TEXT NOT NULL,
    init_point TEXT,
    sandbox_init_point TEXT,
    status TEXT DEFAULT 'created',
    external_reference TEXT,
    idempotency_key TEXT,                    -- Para rastrear requests duplicados
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT now()
  );
  CREATE INDEX idx_mp_preferences_orden ON mp_preferences(orden_id);
  CREATE INDEX idx_mp_preferences_preference ON mp_preferences(preference_id);
  CREATE INDEX idx_mp_preferences_idempotency ON mp_preferences(idempotency_key);
  ```

- [ ] Crear migración para `mp_webhook_requests` (auditoría e idempotencia)
  ```sql
  CREATE TABLE mp_webhook_requests (
    id SERIAL PRIMARY KEY,
    request_id TEXT NOT NULL UNIQUE,         -- x-request-id de MP
    event_type TEXT NOT NULL,                -- payment.created, payment.updated, etc.
    resource_id TEXT,                        -- payment_id, merchant_order_id
    raw_body JSONB NOT NULL,                 -- Payload completo
    raw_headers JSONB,                       -- Headers importantes
    signature TEXT,                          -- x-signature para validación
    status TEXT NOT NULL DEFAULT 'received', -- received, processing, processed, failed
    attempts INT DEFAULT 0,                  -- Número de intentos de procesamiento
    processed_at TIMESTAMP WITH TIME ZONE,
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT now()
  );
  CREATE INDEX idx_webhook_requests_request_id ON mp_webhook_requests(request_id);
  CREATE INDEX idx_webhook_requests_status ON mp_webhook_requests(status);
  CREATE INDEX idx_webhook_requests_created ON mp_webhook_requests(created_at);
  ```

#### 3. Estructura Clean Architecture
- [ ] Crear directorios según estructura:
  ```
  internal/
    core/
      domain/
        entities/          # MPAccount, MPPreference
        repositories/      # Interfaces
      usecases/           # OAuth, Payment, Webhook usecases
    infrastructure/
      persistence/        # PostgreSQL implementations
      external/           # MercadoPago API client
      crypto/             # KMS/encryption adapter
    interfaces/
      http/
        handlers/         # HTTP handlers
        middleware/       # Auth, logging
        dto/             # Request/response DTOs
  ```

#### 4. Investigación Técnica
- [ ] Estudiar documentación oficial:
  - [Checkout Pro](https://www.mercadopago.com.br/developers/en/docs/checkout-pro/overview)
  - [OAuth /oauth/token](https://www.mercadopago.com.ar/developers/en/reference/oauth/_oauth_token/post)
  - [Split Payments](https://www.mercadopago.com.br/developers/en/docs/split-payments/additional-content/security/oauth/introduction)
  - [Postman Collection](https://documenter.getpostman.com/view/15366798/2sAXjKasp4)

- [ ] Validar flujos en Postman/sandbox
- [ ] Documentar endpoints y payloads esperados

### Entregables
- ✅ Base de datos configurada con migraciones (incluyendo key_version y webhook audit table)
- ✅ Variables de entorno documentadas
- ✅ Estructura de proyecto creada
- ✅ Documento de investigación técnica

### Mejoras Prioritarias Implementadas
> [!IMPORTANT]
> **Key Rotation Support**: Campos `key_version` y `kms_key_id` agregados a `mp_accounts` para facilitar rotación de claves de cifrado sin downtime.

> [!IMPORTANT]
> **Webhook Idempotency**: Tabla `mp_webhook_requests` para prevenir procesamiento duplicado y mantener auditoría completa de notificaciones de Mercado Pago.

---

## SPRINT 1: OAuth & Gestión de Tokens
**Duración:** 2 semanas  
**Objetivos:** Implementar flujo OAuth completo para conectar cuentas de Mercado Pago

### Tareas

#### 1. Domain Layer - Entities
- [ ] Crear `MPAccount` entity
  ```go
  // internal/core/domain/entities/mp_account.go
  type MPAccount struct {
      ID             int64
      GimnasioID     uuid.UUID
      MPUserID       int64
      AccessToken    []byte  // encrypted
      RefreshToken   []byte  // encrypted
      TokenExpiresAt time.Time
      Scope          string
      CreatedAt      time.Time
      UpdatedAt      time.Time
  }
  ```

- [ ] Definir `MPAccountRepository` interface
  ```go
  // internal/core/domain/repositories/mp_account_repository.go
  type MPAccountRepository interface {
      Save(ctx context.Context, account *entities.MPAccount) error
      FindByGimnasioID(ctx context.Context, id uuid.UUID) (*entities.MPAccount, error)
      Update(ctx context.Context, account *entities.MPAccount) error
      Delete(ctx context.Context, gimnasioID uuid.UUID) error
  }
  ```

#### 2. Infrastructure - Crypto Adapter
- [ ] Implementar KMS/encryption
  ```go
  // internal/infrastructure/crypto/aes_encryptor.go
  type AESEncryptor struct {
      key []byte
  }
  
  func (e *AESEncryptor) Encrypt(plaintext string) ([]byte, error)
  func (e *AESEncryptor) Decrypt(ciphertext []byte) (string, error)
  ```

#### 3. Infrastructure - MercadoPago Client
- [ ] Crear cliente HTTP para MP API
  ```go
  // internal/infrastructure/external/mercadopago_client.go
  type MercadoPagoClient struct {
      clientID     string
      clientSecret string
      baseURL      string
      httpClient   *http.Client
  }
  
  func (c *MercadoPagoClient) ExchangeCodeForToken(ctx context.Context, code, redirectURI string) (*TokenResponse, error)
  func (c *MercadoPagoClient) RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error)
  func (c *MercadoPagoClient) RevokeToken(ctx context.Context, accessToken string) error
  ```

#### 4. Use Cases - OAuth
- [ ] Implementar `ConnectAccountUseCase`
  - Generar URL de autorización con state (HMAC de gimnasio_id)
  - Validar scopes: `offline_access` para refresh token

- [ ] Implementar `OAuthCallbackUseCase`
  - Validar state parameter
  - Intercambiar code por tokens
  - Cifrar tokens con KMS
  - Persistir en repositorio

- [ ] Implementar `RefreshTokenUseCase`
  - Obtener cuenta por gimnasio_id
  - Descifrar refresh_token
  - Llamar a MP para obtener nuevo access_token
  - Actualizar en BD

- [ ] Implementar `DisconnectAccountUseCase`
  - Revocar token en MP
  - Eliminar de BD

#### 5. HTTP Handlers
- [ ] `GET /mp/connect?gimnasio_id={id}`
  - **Rate Limiting**: 5 requests por gimnasio_id cada 10 minutos
  - Validar gimnasio_id format (UUID)
  - Generar state con HMAC-SHA256:
    ```go
    state = base64(gimnasio_id + "|" + timestamp + "|" + nonce + "|" + hmac)
    // hmac = HMAC-SHA256(secret, gimnasio_id + timestamp + nonce)
    // timestamp: Unix epoch (segundos)
    // nonce: 16 bytes random (base64)
    ```
  - State expira en 5 minutos
  - Response: redirect 302 a MP authorization URL
  
- [ ] `GET /mp/callback?code={code}&state={state}`
  - Validar state:
    1. Decodificar base64
    2. Extraer componentes (gimnasio_id, timestamp, nonce, hmac)
    3. Verificar HMAC
    4. Validar timestamp: `now() - timestamp < 300` segundos
  - Si validación falla: 400 Bad Request
  - Response: página HTML "Conectado correctamente"
  
- [ ] `POST /mp/tokens/refresh`
  - Request: `{"gimnasio_id": "uuid"}`
  - Response: `{"success": true, "expires_at": "timestamp"}`

- [ ] `POST /mp/disconnect`
  - Request: `{"gimnasio_id": "uuid"}`
  - Response: `{"success": true}`

#### 6. Repository Implementation
- [ ] Implementar `PostgresMPAccountRepository`
  - CRUD operations
  - Transacciones cuando sea necesario

#### 7. Testing
- [ ] Unit tests para usecases (mock MP client)
- [ ] Integration tests con DB de prueba
- [ ] Test de cifrado/descifrado
- [ ] Test de validación de state HMAC

### Entregables
- ✅ OAuth flow funcional end-to-end
- ✅ Tokens cifrados en BD
- ✅ Tests passing (>80% coverage)
- ✅ Documentación API para endpoints OAuth

### Detalles Técnicos a Investigar

1. **State Parameter Security** ✅ PRIORIZADO
   - Formato: `base64(gimnasio_id|timestamp|nonce|hmac)`
   - HMAC-SHA256 con secret de 32 bytes mínimo
   - Nonce: 16 bytes aleatorios (evita replay attacks)
   - Timestamp: Unix epoch en segundos
   - Expiración: 300 segundos (5 minutos)
   - Implementar en: `pkg/security/state_manager.go`
   ```go
   type StateManager struct {
       secret []byte // 32+ bytes
   }
   func (sm *StateManager) GenerateState(gimnasioID uuid.UUID) (string, error)
   func (sm *StateManager) ValidateState(state string, maxAge time.Duration) (uuid.UUID, error)
   ```

2. **Token Encryption**
   - AES-256-GCM para tokens
   - Key rotation strategy
   - Usar biblioteca `crypto/cipher`

3. **Error Handling**
   - Mapear errores de MP a códigos HTTP apropiados
   - Logging sin exponer tokens
   - Retry logic con exponential backoff

---

## SPRINT 2: Creación de Preferencias & Checkout
**Duración:** 2 semanas  
**Objetivos:** Implementar creación de preferencias de pago y integración con Checkout Pro

### Tareas

#### 1. Domain Layer - Entities
- [ ] Crear `MPPreference` entity
  ```go
  type MPPreference struct {
      ID                int64
      GimnasioID        uuid.UUID
      OrdenID           string
      PreferenceID      string
      InitPoint         string
      SandboxInitPoint  string
      Status            string // created, paid, expired, failed
      ExternalReference string
      CreatedAt         time.Time
      UpdatedAt         time.Time
  }
  ```

- [ ] Definir `MPPreferenceRepository` interface

#### 2. DTOs
- [ ] Request DTO para crear preferencia
  ```go
  // internal/interfaces/http/dto/create_preference_request.go
  type CreatePreferenceRequest struct {
      GimnasioID       string          `json:"gimnasio_id" validate:"required,uuid"`
      OrdenID          string          `json:"orden_id" validate:"required"`
      Items            []Item          `json:"items" validate:"required,min=1"`
      Payer            Payer           `json:"payer" validate:"required"`
      BackURLs         BackURLs        `json:"back_urls" validate:"required"`
      AutoReturn       string          `json:"auto_return"`
      ExternalReference string         `json:"external_reference"`
  }
  
  // Headers requeridos
  const (
      HeaderIdempotencyKey = "X-Idempotency-Key"
  )
  // Django DEBE enviar X-Idempotency-Key (UUID v4) en cada request
  
  type Item struct {
      ID          string  `json:"id" validate:"required"`
      Title       string  `json:"title" validate:"required,max=256"`
      Quantity    int     `json:"quantity" validate:"required,min=1"`
      UnitPrice   float64 `json:"unit_price" validate:"required,gt=0"`
  }
  ```

- [ ] Response DTO
  ```go
  type CreatePreferenceResponse struct {
      PreferenceID      string `json:"preference_id"`
      InitPoint         string `json:"init_point"`
      SandboxInitPoint  string `json:"sandbox_init_point,omitempty"`
  }
  ```

#### 3. MercadoPago Client - Preferences
- [ ] Implementar `CreatePreference` en MP client
  ```go
  func (c *MercadoPagoClient) CreatePreference(
      ctx context.Context, 
      accessToken string, 
      req *PreferenceRequest,
  ) (*PreferenceResponse, error)
  ```

- [ ] Mapear request de FitStack a formato MP
  - Incluir `statement_descriptor` (máx 22 chars)
  - Configurar `notification_url` para webhooks
  - Agregar metadata: `{"gimnasio_id": "...", "orden_id": "..."}`

#### 4. Use Cases - Payment Creation
- [ ] Implementar `CreatePaymentPreferenceUseCase`
  1. Validar que gimnasio tiene cuenta conectada
  2. Obtener y descifrar access_token
  3. Verificar si token está expirado, refrescar si es necesario
  4. Crear preferencia en MP usando access_token del gimnasio
  5. Persistir preference en BD
  6. Retornar init_point

- [ ] Manejar idempotencia ✅ PRIORIZADO
  - **Nivel 1 - Idempotency Key**: Verificar header `X-Idempotency-Key`
    - Si no existe: 400 Bad Request
    - Si existe en BD con misma key: retornar respuesta guardada (cache)
  - **Nivel 2 - Orden ID**: Verificar si `orden_id` ya existe
    - Si existe y key coincide: retornar preferencia existente
    - Si existe con diferente key: verificar estado
      - Estado "created": retornar preferencia (con warning log)
      - Estado "paid": 409 Conflict - orden ya pagada
  - **Race Condition**: Usar DB transaction con SELECT FOR UPDATE
    ```sql
    BEGIN;
    SELECT * FROM mp_preferences WHERE orden_id = $1 FOR UPDATE;
    -- Si no existe, crear
    INSERT INTO mp_preferences (...) VALUES (...);
    COMMIT;
    ```

#### 6. HTTP Handlers
- [ ] `POST /payments/create`
  - **Validar header**: `X-Idempotency-Key` (requerido, UUID v4)
  - Validar request body
  - Llamar a usecase con idempotency key
  - Manejar errores:
    - 400: validación fallida o falta idempotency key
    - 404: gimnasio sin cuenta conectada
    - 409: orden ya pagada (conflicto)
    - 401: token MP inválido o revocado
    - 500: error interno o timeout MP
  - Logging: `gimnasio_id`, `orden_id`, `preference_id`, `idempotency_key` (NO tokens)
  - Response headers: incluir `X-Idempotency-Key` en respuesta

#### 6. Testing
- [ ] Unit tests para usecase
  - Mock de repositorio
  - Mock de MP client
  - Test casos: token expirado, gimnasio no conectado, idempotencia

- [ ] Integration tests
  - Crear preferencia real en sandbox
  - Verificar persistence en BD
  - Validar init_point generado

- [ ] Test de error handling
  - MP devuelve 4xx/5xx
  - Timeout en llamada a MP
  - Token revocado

### Entregables
- ✅ Endpoint `/payments/create` funcional
- ✅ Preferencias creadas en MP sandbox
- ✅ Tests passing
- ✅ Documentación con ejemplos curl

### Detalles Técnicos a Investigar

1. **Checkout Pro Preference API**
   - Endpoint: `POST https://api.mercadopago.com/checkout/preferences`
   - Header: `Authorization: Bearer {access_token}`
   - Body completo con todos los campos opcionales útiles

2. **Idempotency Key - Contrato con Django** ✅ PRIORIZADO
   - Django DEBE generar UUID v4 único por intento de pago
   - Enviar en header: `X-Idempotency-Key: <uuid>`
   - Guardar en Django para retry logic
   - Si Django reintenta (error de red), usar MISMO key
   - Documentar en OpenAPI spec como header requerido

2. **Statement Descriptor**
   - Máximo 22 caracteres
   - Aparece en resumen de tarjeta del cliente
   - Formato: "FITSTACK-{GIMNOMBRE}"

3. **Back URLs**
   - Success, failure, pending
   - Django debe manejar estos callbacks
   - Pasar orden_id como query param

4. **Split Payments (Futuro)**
   - Investigar campo `application_fee`
   - Requiere app de Marketplace en MP
   - Documentar para Sprint 5

5. **Timeout & Retries**
   - Timeout de 10s para llamadas a MP
   - Retry 3 veces con backoff exponencial (1s, 2s, 4s)
   - Circuit breaker pattern

---

## SPRINT 3: Webhooks & Reconciliación
**Duración:** 2 semanas  
**Objetivos:** Recibir notificaciones de MP y actualizar estado de pagos

### Tareas

#### 1. Domain Layer
- [ ] Agregar estados a `MPPreference`
  ```go
  const (
      StatusCreated  = "created"
      StatusPending  = "pending"
      StatusApproved = "approved"
      StatusRejected = "rejected"
      StatusCancelled = "cancelled"
  )
  ```

- [ ] Crear entity `MPPaymentNotification`
  ```go
  type MPPaymentNotification struct {
      ID             int64
      Type           string // payment, merchant_order
      Data           json.RawMessage
      ProcessedAt    *time.Time
      Status         string // received, processing, processed, failed
      CreatedAt      time.Time
  }
  ```

#### 2. MercadoPago Client - Payment Info
- [ ] Implementar `GetPaymentInfo`
  ```go
  func (c *MercadoPagoClient) GetPaymentInfo(
      ctx context.Context,
      accessToken string,
      paymentID int64,
  ) (*PaymentInfo, error)
  ```

- [ ] Implementar `GetMerchantOrder`
  ```go
  func (c *MercadoPagoClient) GetMerchantOrder(
      ctx context.Context,
      accessToken string,
      orderID int64,
  ) (*MerchantOrder, error)
  ```

#### 3. Webhook Signature Validation ✅ PRIORIZADO
- [ ] **ANTES de implementar**: Probar en sandbox y capturar headers reales
  1. Crear preferencia de prueba
  2. Pagar en sandbox
  3. Capturar webhook request completo (headers + body)
  4. Documentar formato exacto de `x-signature` y `x-request-id`
  5. Validar algoritmo contra docs oficial más reciente

- [ ] Investigar método de validación de MP
  - Header `x-signature` - formato y algoritmo
  - Header `x-request-id` - para idempotencia
  - Posibles algoritmos: HMAC-SHA256, TS + V1
  - Verificar si MP envía webhook secret separado

- [ ] Implementar validador
  ```go
  // internal/infrastructure/external/webhook_validator.go
  type WebhookValidator struct {
      secret string // MP webhook secret
  }
  
  func (wv *WebhookValidator) ValidateSignature(
      requestID string,
      signature string,
      timestamp string,
      body []byte,
  ) error // error si inválido, nil si OK
  ```

#### 4. Use Cases - Webhook Processing
- [ ] Implementar `ProcessWebhookUseCase`
  1. Extraer `x-request-id` de headers
  2. **Idempotencia**: Verificar si ya procesamos este request_id
     - Si existe en `mp_webhook_requests`: retornar 200 OK (ya procesado)
  3. Validar signature
  4. Guardar en `mp_webhook_requests` (status=received)
  5. Parsear tipo de evento
  6. Encolar procesamiento asíncrono (o procesar inline)
  7. Retornar 200 OK inmediatamente
  
  > [!WARNING]
  > Mercado Pago puede reenviar webhooks si no recibe 200 OK rápido. Usar request_id para detectar duplicados.

- [ ] Implementar `ReconcilePaymentUseCase`
  1. Obtener payment info de MP usando payment_id
  2. Extraer external_reference (orden_id)
  3. Buscar preference en BD por orden_id
  4. Verificar que gimnasio_id coincida
  5. Actualizar status según payment.status de MP
  6. Si status = approved, notificar a Django

#### 5. Django Notification Client
- [ ] Implementar cliente HTTP para notificar Django
  ```go
  // internal/infrastructure/external/django_client.go
  func (c *DjangoClient) NotifyPaymentApproved(
      ctx context.Context,
      payload *PaymentApprovedNotification,
  ) error
  
  type PaymentApprovedNotification struct {
      OrdenID       string  `json:"orden_id"`
      GimnasioID    string  `json:"gimnasio_id"`
      PaymentID     int64   `json:"payment_id"`
      Status        string  `json:"status"`
      Amount        float64 `json:"amount"`
      ProcessedAt   string  `json:"processed_at"`
  }
  ```

#### 6. HTTP Handlers
- [ ] `POST /mp/webhook`
  - Validar signature
  - Llamar a usecase
  - Response: `200 OK` siempre (excepto validación fallida: 401)
  - Logging extensivo

- [ ] `GET /payments/{orden_id}` (para Django)
  - Consultar estado de preferencia
  - Response:
    ```json
    {
      "orden_id": "ORD-xxx",
      "status": "approved",
      "preference_id": "123456",
      "payment_id": 789,
      "updated_at": "2026-01-07T12:00:00Z"
    }
    ```

#### 7. Async Processing (Opcional)
- [ ] Evaluar usar worker pool en Go
- [ ] O implementar cola simple en memoria
- [ ] Evitar bloquear respuesta del webhook

#### 8. Testing ✅ PRIORIZADO - Usar Payloads Reales
- [ ] **Capturar ejemplos reales de sandbox**:
  1. Hacer pago de prueba en sandbox
  2. Guardar webhook request completo (headers + body)
  3. Crear fixtures en `testdata/webhooks/`:
     - `payment_created.json`
     - `payment_updated_approved.json`
     - `payment_updated_rejected.json`
     - Incluir headers reales: `x-signature`, `x-request-id`

- [ ] Unit tests para validación de signature
  - Usar fixtures con signatures reales de sandbox
  - Test casos: firma válida, inválida, timestamp expirado
  
- [ ] Test de idempotencia de webhooks
  - Enviar mismo request_id 2 veces
  - Verificar que solo se procesa una vez
  
- [ ] Test de reconciliación con diferentes estados
  - Mock de MP client con responses realistas
  
- [ ] Integration test: simular webhook completo
  - Endpoint real con DB de prueba
  - Validar insert en mp_webhook_requests

### Entregables
- ✅ Webhook endpoint funcional
- ✅ Validación de signature implementada
- ✅ Reconciliación de pagos
- ✅ Notificación a Django
- ✅ Tests passing
- ✅ Logs estructurados

### Detalles Técnicos a Investigar

1. **Webhook Signature Validation** ✅ ALTA PRIORIDAD
   - **PRIMER PASO**: Capturar headers reales de sandbox
   - Documentar formato exacto recibido
   - Consultar docs oficial de MP (versión más reciente)
   - Posibles formatos:
     - `HMAC-SHA256(secret, request_id + body)`
     - Formato con timestamp: `t={ts},v1={signature}`
   - Validar con ejemplos en Postman collection
   - Crear tabla de compatibilidad si formato cambia entre versiones

2. **Webhook Event Types**
   - `payment.created`
   - `payment.updated`
   - `merchant_order`
   - Filtrar eventos relevantes

3. **Idempotency in Webhook**
   - Guardar `request_id` de MP
   - Usar UNIQUE constraint para evitar reprocesar

4. **Payment Status Mapping**
   - MP statuses: approved, pending, rejected, cancelled, refunded, charged_back
   - Mapear a estados internos de FitStack

5. **Retry Mechanism**
   - Si notificación a Django falla, reintentar
   - Implementar dead-letter queue (o marcar como failed para revisión manual)

6. **Security**
   - Endpoint público, pero validado con signature
   - Rate limiting en webhook endpoint
   - Logging de IPs sospechosas

---

## SPRINT 4: Testing, Seguridad & Observabilidad
**Duración:** 1-2 semanas  
**Objetivos:** Asegurar robustez, seguridad y monitoreo

### Tareas

#### 1. Testing Completo
- [ ] **Unit Tests**
  - Coverage mínimo 80%
  - Todos los usecases con mocks
  - Crypto adapter
  - MP client (mock HTTP responses)

- [ ] **Integration Tests**
  - Tests contra BD real (containerizada)
  - Tests contra MP sandbox
  - OAuth flow completo
  - Payment creation + webhook

- [ ] **E2E Tests**
  - Script que simula flujo completo:
    1. Conectar cuenta gimnasio
    2. Crear preferencia
    3. Simular pago (manual o automatizado si MP lo permite)
    4. Recibir webhook
    5. Verificar estado final

#### 2. Seguridad

##### Token Management
- [ ] Implementar key rotation para encryption key
  - Documentar proceso manual (Sprint 5: automatizar)
  - Re-cifrar tokens con nueva key

- [ ] Implementar auto-refresh de tokens expirados
  - Worker que corre cada 1 hora
  - Refresca tokens que expiran en <24h

##### API Security
- [ ] Implementar autenticación para endpoints internos
  - Django debe enviar API key o JWT
  - Middleware de validación

- [ ] Rate limiting
  - Por IP para endpoints públicos (webhook, connect)
  - Por gimnasio_id para endpoints de pago

- [ ] Input validation estricta
  - Usar biblioteca `validator`
  - Sanitizar inputs antes de enviar a MP

##### Secrets Management
- [ ] Documentar uso de secrets manager (AWS Secrets Manager, HashiCorp Vault, etc.)
- [ ] Migrar de env vars a secrets manager (opcional para Sprint 5)

#### 3. Logging & Monitoring

##### Structured Logging
- [ ] Implementar logger estructurado (zerolog, zap)
  ```go
  log.Info().
      Str("gimnasio_id", id).
      Str("orden_id", ordenID).
      Msg("Preference created successfully")
  ```

- [ ] Niveles: DEBUG, INFO, WARN, ERROR
- [ ] NO loguear tokens o datos sensibles
- [ ] Loguear:
  - Request IDs (trace ID)
  - Latencias de llamadas a MP
  - Errores con stack traces

##### Metrics (Prometheus)
- [ ] Exponer endpoint `/metrics`
- [ ] Métricas clave:
  - `mp_api_calls_total` (counter) - labels: endpoint, status
  - `mp_api_latency_seconds` (histogram)
  - `preferences_created_total` (counter) - labels: gimnasio_id (hash)
  - `webhooks_received_total` (counter) - labels: event_type, status
  - `webhooks_duplicate_total` (counter) - detección de duplicados
  - `token_refresh_total` (counter) - labels: success/failure
  - `token_refresh_failures_total` (counter) - ⚠️ ALERTA
  - `webhook_processing_failures_total` (counter) - ⚠️ ALERTA
  - `errors_total` (counter por tipo)

##### Health Checks
- [ ] `GET /health`
  - Verificar conexión a BD
  - Verificar conectividad a MP API (opcional)
  - Response: `{"status": "healthy"}`

- [ ] `GET /ready`
  - Verificar que migraciones estén aplicadas
  - Response: `{"ready": true}`

#### 4. Error Handling & Resilience

- [ ] Implementar circuit breaker para MP API
  - Usar biblioteca `gobreaker`
  - Evitar saturar MP en caso de downtime

- [ ] Timeout configurables para todas las llamadas HTTP

- [ ] Retry logic con exponential backoff

- [ ] Graceful shutdown
  - Esperar a que requests activos terminen
  - Cerrar conexiones DB limpiamente

#### 5. Documentation

- [ ] OpenAPI 3.0 spec (Swagger)
  - Todos los endpoints documentados
  - **Header `X-Idempotency-Key` documentado como requerido en POST /payments/create**
  - Ejemplos de request/response
  - Códigos de error con descripciones detalladas

- [ ] README con:
  - Setup instructions
  - Environment variables
  - How to run locally
  - How to run tests

- [ ] Runbook operacional:
  - **Playbook: "Qué hacer si MP cambia la API"**
    1. Verificar changelog de MP
    2. Actualizar client según versión
    3. Tests de regresión
    4. Deploy gradual (canary)
  - **Playbook: "Qué hacer si token revocado"**
    1. Detectar error 401 de MP
    2. Marcar cuenta como disconnected
    3. Notificar a gimnasio vía Django
    4. Solicitar re-autenticación
  - Cómo refrescar tokens manualmente
  - Cómo revocar acceso de un gimnasio
  - Qué hacer si webhook falla (>N intentos)
  - Monitoreo y alertas (umbrales)

- [ ] **Integración con Sentry** ✅ PRIORIZADO
  - Configurar Sentry SDK para Go
  - Capturar errores críticos:
    - Token refresh failures
    - Webhook processing failures (después de retries)
    - MP API errors (5xx)
  - Tags: `gimnasio_id`, `error_type`, `mp_status_code`
  - Alertas en Sentry:
    - Token refresh failures > 5 en 10 min
    - Webhook failures > 10 en 5 min
    - MP API 5xx errors > 20 en 5 min

### Entregables
- ✅ Tests passing con >80% coverage
- ✅ Seguridad implementada y documentada
- ✅ Logging estructurado
- ✅ Métricas expuestas
- ✅ OpenAPI spec completo
- ✅ Runbook operacional

### Detalles Técnicos a Investigar

1. **Circuit Breaker Configuration**
   - Threshold: 5 fallos consecutivos
   - Timeout del breaker: 60s
   - Half-open state: permitir 1 request de prueba

2. **Encryption Key Rotation**
   - Mantener versiones de keys
   - Proceso: generar nueva key, cifrar con ambas (vieja y nueva), eventualmente deprecar vieja

3. **API Key Management**
   - Django debe tener API key específico
   - Rotar cada 90 días
   - Almacenar hash en BD del microservicio

4. **Alerting**
   - Integrar con Prometheus Alertmanager
   - Alertas críticas:
     - Webhook endpoint down
     - MP API errors > 10% en 5 min
     - Token refresh failures
     - DB connection errors

---

## SPRINT 5: Split Payments & Features Avanzados
**Duración:** 1-2 semanas  
**Objetivos:** Implementar comisión de FitStack (Split Payments) y features adicionales

### Tareas

#### 1. Split Payments / Marketplace

##### Investigación
- [ ] Configurar app como Marketplace en Mercado Pago
  - Requiere aprobación de MP
  - Documentar proceso de aplicación

- [ ] Estudiar documentación de `application_fee`

##### Implementación
- [ ] Modificar DTO de create preference
  ```go
  type CreatePreferenceRequest struct {
      // ... campos existentes
      UseSplitPayment  bool    `json:"use_split_payment"`
      ApplicationFee   float64 `json:"application_fee,omitempty"` // % o monto fijo
  }
  ```

- [ ] Modificar `CreatePreferenceUseCase`
  - Si `use_split_payment = true`:
    - Usar access_token de FitStack (app principal)
    - Agregar campo `marketplace_fee` o `application_fee` en request a MP
    - Especificar `collector_id` del gimnasio

- [ ] Testing con sandbox de Marketplace

#### 2. Features Adicionales

##### Consulta de Pagos
- [ ] `GET /payments/list?gimnasio_id={id}&limit=20&offset=0`
  - Listar preferencias de un gimnasio
  - Filtros: status, fecha

##### Estadísticas
- [ ] `GET /payments/stats?gimnasio_id={id}`
  - Total pagado este mes
  - Número de transacciones
  - Promedio de ticket

##### Desconexión Masiva (Admin)
- [ ] Endpoint para admin de FitStack
  - Revocar y desconectar múltiples gimnasios
  - Uso: gimnasio cancela suscripción

#### 3. Worker para Token Refresh
- [ ] Implementar worker/cron job
  - Correr cada 1 hora
  - Buscar cuentas con `token_expires_at < now() + 24h`
  - Refrescar tokens automáticamente
  - Loguear y alertar en caso de fallo

#### 4. Optimizaciones

##### Caching
- [ ] Cachear access_tokens en Redis (opcional)
  - TTL = `expires_at - now()`
  - Reducir queries a BD por cada pago

##### DB Indexes
- [ ] Revisar queries lentos
- [ ] Agregar índices adicionales si es necesario

##### Connection Pooling
- [ ] Configurar pool de conexiones a BD
  - Max connections: 20
  - Idle connections: 5

#### 5. Deployment Checklist

- [ ] Dockerfile optimizado
  - Multi-stage build
  - Imagen base: `golang:1.21-alpine`
  - Compilar estáticamente

- [ ] Docker Compose para local
  - Microservicio + PostgreSQL + Redis (opcional)

- [ ] CI/CD pipeline
  - GitHub Actions o GitLab CI
  - Stages: test, build, deploy

- [ ] Environment-specific configs
  - `config/dev.yaml`
  - `config/staging.yaml`
  - `config/production.yaml`

- [ ] Migración de DB en producción
  - Script de migración
  - Rollback plan

### Entregables
- ✅ Split Payments implementado (si aprobado por MP)
- ✅ Worker de token refresh
- ✅ Optimizaciones aplicadas
- ✅ Deployment listo para producción
- ✅ CI/CD pipeline

### Detalles Técnicos a Investigar

1. **Marketplace Application**
   - Requisitos de MP para aprobar app
   - Tiempo estimado de aprobación
   - Documentación necesaria

2. **Application Fee Calculation**
   - Porcentaje vs. monto fijo
   - Límites de comisión permitidos
   - Cómo se refleja en el estado de cuenta del gimnasio

3. **Token Refresh Strategy**
   - Usar `refresh_token` con `grant_type=refresh_token`
   - Manejar casos donde refresh también expiró (re-autenticar)

4. **Redis Caching**
   - Evaluar si es necesario según volumen
   - Estructura de key: `mp:token:{gimnasio_id}`

---

## Arquitectura Clean - Resumen de Capas

### Domain Layer (`internal/core/domain`)
**Responsabilidad:** Lógica de negocio pura, sin dependencias externas

- **Entities:** `MPAccount`, `MPPreference`, `MPPaymentNotification`
- **Repositories (interfaces):** `MPAccountRepository`, `MPPreferenceRepository`
- **Value Objects:** `EncryptedToken`, `GimnasioID`

### Use Cases Layer (`internal/core/usecases`)
**Responsabilidad:** Orquestación de lógica de negocio

- `ConnectAccountUseCase`
- `OAuthCallbackUseCase`
- `RefreshTokenUseCase`
- `DisconnectAccountUseCase`
- `CreatePaymentPreferenceUseCase`
- `ProcessWebhookUseCase`
- `ReconcilePaymentUseCase`

### Infrastructure Layer (`internal/infrastructure`)
**Responsabilidad:** Implementaciones concretas de interfaces

- **Persistence:** `PostgresMPAccountRepository`, `PostgresMPPreferenceRepository`
- **External:** `MercadoPagoClient`, `DjangoClient`, `WebhookValidator`
- **Crypto:** `AESEncryptor`, `KMSAdapter`

### Interfaces Layer (`internal/interfaces`)
**Responsabilidad:** Entrega/transporte (HTTP, gRPC, etc.)

- **HTTP Handlers:** `OAuthHandler`, `PaymentHandler`, `WebhookHandler`
- **Middleware:** `AuthMiddleware`, `LoggingMiddleware`, `RateLimitMiddleware`
- **DTOs:** Request y Response structs

### Pkg (`pkg/`)
**Responsabilidad:** Utilidades reutilizables

- `logger`
- `validator`
- `httputils`

---

## Detalles Técnicos a Investigar - Resumen

### 1. OAuth & Token Management
- [ ] Scopes exactos a pedir: `read:payment`, `write:payment`, `offline_access`
- [ ] Flujo de revocación: `POST /oauth/revoke`
- [ ] Manejo de `invalid_grant` error en refresh

### 2. Checkout Pro
- [ ] Campos opcionales útiles:
  - `payment_methods`: restricciones de métodos
  - `expires`: expiración de preferencia
  - `binary_mode`: true/false para simplificar status
  - `statement_descriptor`: custom label en resumen

- [ ] Diferencia entre `init_point` y `sandbox_init_point`

### 3. Webhook Security
- [ ] Algoritmo de verificación de firma
- [ ] Rotación de webhook secret
- [ ] IP whitelisting (si MP provee IPs fijas)

### 4. Split Payments
- [ ] Registro como Marketplace
- [ ] Campo `marketplace_fee` vs `application_fee`
- [ ] Collector ID vs User ID
- [ ] Cómo manejar disputes/chargebacks con split

### 5. Error Codes & Handling
- [ ] Mapear errores de MP a errores internos
  - `invalid_parameters` → 400
  - `unauthorized` → 401
  - `forbidden` → 403
  - `not_found` → 404
  - `internal_server_error` → 500

- [ ] Errores específicos:
  - `token_expired`
  - `token_revoked`
  - `insufficient_permissions`

### 6. Performance & Scalability
- [ ] Timeout recomendado para MP API: 10s
- [ ] Uso de HTTP/2 con keep-alive
- [ ] Batch processing de webhooks si volumen es alto

### 7. Compliance & Legal
- [ ] PCI compliance (MP maneja datos de tarjetas, no nosotros)
- [ ] GDPR/datos personales: qué guardamos (emails, nombres)
- [ ] Términos de servicio de MP: restricciones de uso

---

## Criterios de Aceptación por Sprint

### Sprint 0
- [x] Cuenta de developer creada en MP
- [x] Database schema aplicado
- [x] Estructura de proyecto creada
- [x] Documento de investigación completo

### Sprint 1
- [ ] Usuario puede iniciar OAuth desde frontend
- [ ] Callback procesa code y guarda tokens cifrados
- [ ] Refresh token funciona correctamente
- [ ] Tests de OAuth passing

### Sprint 2
- [ ] Django puede crear preferencia via API
- [ ] Init point generado redirige a Checkout Pro
- [ ] Idempotencia funciona (mismo orden_id no duplica)
- [ ] Tests de creación de preferencia passing

### Sprint 3
- [ ] Webhook recibe notificaciones de MP
- [ ] Signature validation implementada
- [ ] Estado de preferencia se actualiza a "approved"
- [ ] Django recibe notificación de pago exitoso

### Sprint 4
- [ ] Coverage >80%
- [ ] E2E test completo funciona
- [ ] Métricas expuestas en /metrics
- [ ] OpenAPI spec accesible en /swagger

### Sprint 5
- [ ] Split payments funciona (si aplicable)
- [ ] Worker de token refresh operativo
- [ ] Deployment en staging exitoso
- [ ] Runbook operacional completo

---

## Escenarios de Prueba Detallados

### Escenario 1: Gimnasio Conecta Cuenta
1. Staff hace clic en "Conectar Mercado Pago" en Django
2. Django llama `GET /mp/connect?gimnasio_id={uuid}`
3. Usuario es redirigido a MP, inicia sesión, autoriza
4. MP redirige a `/mp/callback?code=xxx&state=yyy`
5. Microservicio valida state, intercambia code
6. Tokens guardados cifrados en BD
7. Página confirma "Conectado correctamente"

**Validaciones:**
- State HMAC es válido
- Tokens están cifrados en BD
- `mp_user_id` guardado correctamente

### Escenario 2: Cliente Compra Paquete
1. Cliente selecciona paquete de 20 clases
2. Django crea orden interna
3. Django llama `POST /payments/create` con datos de orden
4. Microservicio:
   - Busca cuenta del gimnasio
   - Descifra access_token
   - Crea preferencia en MP
   - Guarda preference en BD
   - Retorna init_point
5. Django redirige cliente a init_point
6. Cliente paga en MP
7. MP envía webhook a `/mp/webhook`
8. Microservicio reconcilia, actualiza status a "approved"
9. Microservicio notifica Django
10. Django activa paquete para cliente

**Validaciones:**
- Preference creada con external_reference correcto
- Webhook signature es válida
- Status actualizado correctamente
- Django recibe notificación

### Escenario 3: Token Expirado
1. Django intenta crear preferencia
2. Microservicio detecta token expirado
3. Automáticamente refresca token usando refresh_token
4. Crea preferencia con nuevo token
5. Proceso continúa normalmente

**Validaciones:**
- Refresh token funciona
- Nuevo access_token guardado
- Preferencia se crea sin error

### Escenario 4: Gimnasio Desconecta Cuenta
1. Staff hace clic en "Desconectar Mercado Pago"
2. Django llama `POST /mp/disconnect`
3. Microservicio revoca token en MP
4. Microservicio elimina tokens de BD
5. Responde con éxito

**Validaciones:**
- Token revocado en MP
- Registro eliminado de BD
- Intentar crear preferencia falla con error apropiado

---

## Riesgos & Mitigaciones

### Riesgo 1: Tokens Comprometidos
**Impacto:** Alto  
**Probabilidad:** Media  
**Mitigación:**
- Cifrado AES-256
- Key rotation cada 6 meses
- Auditoría de accesos
- Revocación inmediata si se detecta

### Riesgo 2: Webhook Failures
**Impacto:** Alto (pagos no reconciliados)  
**Probabilidad:** Media  
**Mitigación:**
- Retry logic con backoff
- Dead-letter queue para fallos
- Endpoint de consulta manual (`GET /payments/{orden_id}`)
- Alertas en caso de fallos consecutivos

### Riesgo 3: MP API Downtime
**Impacto:** Alto  
**Probabilidad:** Baja  
**Mitigación:**
- Circuit breaker para evitar saturar
- Mensajes claros al usuario
- Fallback: permitir pagos offline (registro manual)

### Riesgo 4: Errores de Integración con Django
**Impacto:** Medio  
**Probabilidad:** Media  
**Mitigación:**
- Contrato claro de API (OpenAPI)
- Tests de integración compartidos
- Versionado de API

### Riesgo 5: Compliance / Regulatorio
**Impacto:** Alto  
**Probabilidad:** Baja  
**Mitigación:**
- Revisión legal de flujo
- Cumplir con términos de MP
- No almacenar datos de tarjetas (delegado a MP)

---

## Recursos & Referencias

### Documentación Oficial
1. [Checkout Pro - Mercado Pago](https://www.mercadopago.com.br/developers/en/docs/checkout-pro/overview)
2. [OAuth Token - Mercado Pago](https://www.mercadopago.com.ar/developers/en/reference/oauth/_oauth_token/post)
3. [Split Payments](https://www.mercadopago.com.br/developers/en/docs/split-payments/additional-content/security/oauth/introduction)
4. [Postman Collection](https://documenter.getpostman.com/view/15366798/2sAXjKasp4)

### Bibliotecas Go Recomendadas
- `github.com/gin-gonic/gin` - HTTP framework
- `github.com/lib/pq` - PostgreSQL driver
- `github.com/google/uuid` - UUID generation
- `github.com/go-playground/validator/v10` - Validation
- `github.com/rs/zerolog` - Structured logging
- `github.com/prometheus/client_golang` - Metrics
- `github.com/sony/gobreaker` - Circuit breaker
- `golang.org/x/crypto` - Crypto utilities

### Testing
- `github.com/stretchr/testify` - Assertions
- `github.com/DATA-DOG/go-sqlmock` - DB mocking
- `github.com/jarcoal/httpmock` - HTTP mocking

---

## Checklist Final de Entrega

### Código
- [ ] Todos los endpoints implementados y documentados
- [ ] Clean Architecture respetada
- [ ] Tests unitarios e integración (>80% coverage)
- [ ] Linting passing (`golangci-lint`)
- [ ] No secrets en código

### Base de Datos
- [ ] Migraciones SQL versionadas
- [ ] Índices optimizados
- [ ] Scripts de rollback

### Seguridad
- [ ] Tokens cifrados
- [ ] Validación de webhooks
- [ ] API key para Django
- [ ] Rate limiting implementado
- [ ] Secrets en secrets manager (no env vars en producción)

### Documentación
- [ ] OpenAPI 3.0 spec
- [ ] README completo
- [ ] Runbook operacional
- [ ] Diagramas de arquitectura
- [ ] Ejemplos de curl/Postman

### Deployment
- [ ] Dockerfile
- [ ] Docker Compose
- [ ] CI/CD pipeline
- [ ] Health checks
- [ ] Logging estructurado
- [ ] Métricas Prometheus

### Observabilidad
- [ ] Logs sin datos sensibles
- [ ] Trace IDs en requests
- [ ] Métricas clave expuestas
- [ ] Alertas configuradas

---

## Conclusión

Este plan de implementación cubre la integración completa con Mercado Pago siguiendo principios de Clean Architecture. Los 5 sprints están diseñados para entregar valor incremental, con cada sprint construyendo sobre el anterior.

**Próximos pasos:**
1. Validar plan con equipo de desarrollo
2. Obtener credenciales de MP sandbox
3. Iniciar Sprint 0
4. Revisiones al final de cada sprint

**Estimación total:** 6-8 semanas  
**Equipo recomendado:** 1-2 desarrolladores Go, 1 QA, 1 DevOps (part-time)
