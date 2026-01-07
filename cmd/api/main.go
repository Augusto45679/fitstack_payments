// FitStack Payments Microservice
//
// Main entry point - wires up all dependencies and starts the server.
// Clean Architecture implementation with dependency injection.
package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	// Infrastructure layer
	"github.com/fitstack/fitstack-payments/internal/infrastructure/config"
	"github.com/fitstack/fitstack-payments/internal/infrastructure/external/django"
	"github.com/fitstack/fitstack-payments/internal/infrastructure/external/mercadopago"
	"github.com/fitstack/fitstack-payments/internal/infrastructure/persistence/multitenancy"

	// Core layer - Services
	"github.com/fitstack/fitstack-payments/internal/core/services"

	// Core layer - Use Cases
	"github.com/fitstack/fitstack-payments/internal/core/usecases/payment"

	// Interface layer
	"github.com/fitstack/fitstack-payments/internal/interfaces/http"
	"github.com/fitstack/fitstack-payments/internal/interfaces/http/handlers"
)

func main() {
	log.Println("Starting FitStack Payments Service (Clean Architecture v2.0)...")

	// ===========================
	// Load Configuration
	// ===========================
	cfg := config.Load()
	log.Printf("Config: Port=%s, Django=%s", cfg.Server.Port, cfg.Django.BaseURL)

	// ===========================
	// Infrastructure Layer Setup
	// ===========================

	// External adapters
	mpAdapter := mercadopago.NewAdapter()
	mpWebhookHandler := mercadopago.NewWebhookHandler()
	djangoClient := django.NewClient(cfg.Django.BaseURL, cfg.Django.APIKey)

	// Multi-tenancy infrastructure
	contextHandler := multitenancy.NewContextHandler()
	_ = contextHandler // Will be used in future middleware

	// ===========================
	// Core Layer - Services
	// ===========================

	// Domain services
	securityService := services.NewSecurityService()
	notificationService := services.NewNotificationService(djangoClient)

	// ===========================
	// Core Layer - Use Cases
	// ===========================

	// Payment use cases
	createPaymentUC := payment.NewCreatePaymentUseCase(
		mpAdapter,     // PaymentGateway
		djangoClient,  // TenantRepository
	)

	processWebhookUC := payment.NewProcessWebhookUseCase(
		mpAdapter,          // PaymentGateway
		djangoClient,       // TenantRepository
		mpWebhookHandler,   // WebhookValidator (or securityService)
		notificationService, // NotificationService
	)

	// ===========================
	// Interface Layer - HTTP Handlers
	// ===========================

	paymentHandler := handlers.NewPaymentHandler(createPaymentUC)
	webhookHandler := handlers.NewWebhookHandler(processWebhookUC)
	healthHandler := handlers.NewHealthHandler()

	// Setup router
	router := http.SetupRouter(
		paymentHandler,
		webhookHandler,
		healthHandler,
		cfg.Server.GinMode,
	)

	// ===========================
	// Start Server
	// ===========================

	serverAddr := fmt.Sprintf(":%s", cfg.Server.Port)
	go func() {
		log.Printf("Server listening on %s", serverAddr)
		if err := router.Run(serverAddr); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// ===========================
	// Graceful Shutdown
	// ===========================

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down gracefully...")
}
