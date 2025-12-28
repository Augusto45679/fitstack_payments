// FitStack Payments Microservice
//
// This is the main entry point for the payment processing service.
// It wires up all dependencies and starts the HTTP server.
package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/fitstack/fitstack-payments/config"
	"github.com/fitstack/fitstack-payments/internal/api"
	"github.com/fitstack/fitstack-payments/internal/payment"
	"github.com/fitstack/fitstack-payments/internal/platform/fitstack_core"
	"github.com/fitstack/fitstack-payments/internal/platform/mercadopago"
)

func main() {
	log.Println("Starting FitStack Payments Service...")

	// Load configuration
	cfg := config.Load()
	log.Printf("Configuration loaded: Port=%s, CoreURL=%s", cfg.Server.Port, cfg.Core.BaseURL)

	// Validate required configuration
	if err := validateConfig(cfg); err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	// Wire up dependencies (manual dependency injection)
	//
	// Infrastructure Layer
	coreClient := fitstackcore.NewClient(cfg.Core.BaseURL, cfg.Core.APIKey)
	mpAdapter := mercadopago.NewAdapter()

	// Service Layer
	paymentService := payment.NewService(
		coreClient, // implements domain.GymRepository
		mpAdapter,  // implements domain.PaymentGateway
		coreClient, // implements domain.CoreNotifier
	)

	// API Layer
	handler := api.NewHandler(paymentService)
	router := api.SetupRouter(handler, cfg.Server.GinMode)

	// Start server in a goroutine
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

	log.Println("Shutting down server...")
}

// validateConfig checks that required configuration values are set.
func validateConfig(cfg *config.Config) error {
	if cfg.Core.BaseURL == "" {
		return fmt.Errorf("FITSTACK_CORE_URL is required")
	}
	if cfg.Core.APIKey == "" {
		log.Println("Warning: FITSTACK_CORE_API_KEY not set")
	}
	if cfg.Security.EncryptionKey == "" {
		log.Println("Warning: ENCRYPTION_KEY not set")
	}
	return nil
}
