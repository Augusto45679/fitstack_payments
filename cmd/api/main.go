// FitStack Payments Microservice
//
// Main entry point - wires up all dependencies and starts the server.
package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/fitstack/fitstack-payments/config"
	"github.com/fitstack/fitstack-payments/internal/adapters/django"
	"github.com/fitstack/fitstack-payments/internal/adapters/mercadopago"
	"github.com/fitstack/fitstack-payments/internal/core/service"
	"github.com/fitstack/fitstack-payments/internal/handlers"
)

func main() {
	log.Println("Starting FitStack Payments Service...")

	// Load configuration
	cfg := config.Load()
	log.Printf("Config: Port=%s, Django=%s", cfg.Server.Port, cfg.Django.BaseURL)

	// Wire up dependencies (Clean Architecture)
	// ============================================

	// Adapters (Infrastructure Layer)
	mpAdapter := mercadopago.NewAdapter()
	mpValidator := mercadopago.NewWebhookValidator()
	djangoClient := django.NewClient(cfg.Django.BaseURL, cfg.Django.APIKey)

	// Service Layer
	paymentService := service.NewPaymentService(
		mpAdapter,      // PaymentGateway
		djangoClient,   // GymCredentialProvider
		djangoClient,   // DjangoNotifier
		mpValidator,    // WebhookValidator
	)

	// Handlers (Interface Layer)
	paymentHandler := handlers.NewPaymentHandler(paymentService)
	router := handlers.SetupRouter(paymentHandler, cfg.Server.GinMode)

	// Start server
	serverAddr := fmt.Sprintf(":%s", cfg.Server.Port)
	go func() {
		log.Printf("Server listening on %s", serverAddr)
		if err := router.Run(serverAddr); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down...")
}
