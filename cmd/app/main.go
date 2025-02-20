package main

import (
	"log"
	"net/http"
	"travel-agent/internal/config"
	"travel-agent/internal/handlers"
	"travel-agent/internal/service"
	"travel-agent/internal/service/ai"
)

func main() {
	// Load configuration
	cfg, err := config.Load("config.json")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize AI inference module
	engine, err := ai.NewInferenceEngine(cfg.AIProvider.APIKey)
	if err != nil {
		log.Fatalf("Failed to initialize AI processor: %v", err)
	}

	// Initialize services
	bookingService := service.NewBookingService(engine)
	bookingHandler := handlers.NewBookingHandler(bookingService)

	// Setup routes
	http.HandleFunc("/api/v1/bookings", bookingHandler.CreateBooking)
	http.HandleFunc("/api/v1/bookings/status", bookingHandler.GetBooking)

	// Start server
	log.Printf("Server starting on port %s", cfg.ServerPort)
	if err := http.ListenAndServe(cfg.ServerPort, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
