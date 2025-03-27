package main

import (
	"log"
	"travel-agent/internal/config"
	"travel-agent/internal/handlers"
	"travel-agent/internal/models"
	"travel-agent/internal/service"
	"travel-agent/internal/service/ai"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := config.Load("config.json")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize AI inference module
	extractionInference, err := ai.NewInferenceEngine[models.TravelParameters, models.BookingRequest](cfg.AIProvider.APIKey)
	if err != nil {
		log.Fatalf("Failed to initialize AI processor: %v", err)
	}
	recommendationInference, err := ai.NewInferenceEngine[models.FlightRecommendation, models.FlightRecommendationRequest](cfg.AIProvider.APIKey)
	if err != nil {
		log.Fatalf("Failed to initialize AI processor: %v", err)
	}

	// Initialize services
	bookingService := service.NewBookingService(extractionInference, recommendationInference)
	bookingHandler := handlers.NewBookingHandler(bookingService)

	// Create Gin router
	router := gin.Default()

	// Setup routes
	router.POST("/api/v1/bookings", func(c *gin.Context) {
		bookingHandler.CreateBooking(c.Writer, c.Request)
	})
	router.GET("/api/v1/bookings/status", func(c *gin.Context) {
		bookingHandler.GetBooking(c.Writer, c.Request)
	})

	// Start server
	log.Printf("Server starting on port %s", cfg.ServerPort)
	if err := router.Run(cfg.ServerPort); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
