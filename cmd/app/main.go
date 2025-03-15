package main

import (
	"log"
	"net/http"
	"travel-agent/internal/config"
	"travel-agent/internal/handlers"
	"travel-agent/internal/models"
	"travel-agent/internal/service"
	"travel-agent/internal/service/ai"
)

func main() {
	// Load configuration
	cfg, err := config.Load("config.json")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize AI inference engines
	extractionInference, err := ai.NewInferenceEngine[models.TravelParameters, models.BookingRequest](cfg.AIProvider.APIKey)
	if err != nil {
		log.Fatalf("Failed to initialize extraction inference engine: %v", err)
	}

	recommendationInference, err := ai.NewInferenceEngine[models.FlightRecommendation, models.FlightRecommendationRequest](cfg.AIProvider.APIKey)
	if err != nil {
		log.Fatalf("Failed to initialize flight recommendation inference engine: %v", err)
	}

	// Register tools with inference engines
	// Example: Register a city validation tool with the extraction inference engine
	// extractionInference.RegisterTool(tools.NewCityValidationTool())

	// Example: Register a flight search tool with the recommendation inference engine
	// recommendationInference.RegisterTool(tools.NewFlightSearchTool())

	// Initialize services
	bookingService := service.NewBookingService(extractionInference, recommendationInference)
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
