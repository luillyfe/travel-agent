package main

import (
	"log"
	"net/http"
	"travel-agent/internal/handlers"
	"travel-agent/internal/service"
)

func main() {
	// Initialize services
	bookingService := service.NewBookingService()
	bookingHandler := handlers.NewBookingHandler(bookingService)

	// Setup routes
	http.HandleFunc("/api/v1/bookings", bookingHandler.CreateBooking)
	http.HandleFunc("/api/v1/bookings/status", bookingHandler.GetBooking)

	// Start server
	log.Printf("Server starting on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
