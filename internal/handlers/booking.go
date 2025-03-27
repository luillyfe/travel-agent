package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"travel-agent/internal/models"
)

type BookingServiceInterface interface {
	ProcessBooking(ctx context.Context, req models.BookingRequest) (*models.BookingResponse, error)
}

type BookingHandler struct {
	bookingService BookingServiceInterface
}

func NewBookingHandler(bookingService BookingServiceInterface) *BookingHandler {
	return &BookingHandler{bookingService: bookingService}
}

func (h *BookingHandler) CreateBooking(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.BookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if err := validateBookingRequest(req); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Process the booking request
	response, err := h.bookingService.ProcessBooking(context.Background(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	// If writing fails for any reason(network issues, closed connection), respond with an error
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *BookingHandler) GetBooking(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract booking ID from URL
	bookingID := r.URL.Query().Get("id")
	if bookingID == "" {
		http.Error(w, "Booking ID is required", http.StatusBadRequest)
		return
	}

	// TODO: Implement booking status retrieval
	// For now, return dummy response
	response := &models.BookingResponse{
		ID:     bookingID,
		Status: models.StatusProcessing,
	}

	w.Header().Set("Content-Type", "application/json")
	// If writing fails for any reason(network issues, closed connection), respond with an error
	if err := json.NewEncoder(w).Encode(response); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to encode response")
		return
	}
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	}); err != nil {
		// If encoding fails, log the error but don't attempt to send another response
		// since headers have already been written
		fmt.Printf("Failed to encode error response: %v\n", err)
	}
}

func validateBookingRequest(req models.BookingRequest) error {
	if req.Query == "" {
		return fmt.Errorf("query cannot be empty")
	}
	if req.Deadline.IsZero() {
		return fmt.Errorf("deadline is required")
	}
	// Check for past deadlines
	if req.Deadline.Before(time.Now()) {
		return fmt.Errorf("deadline cannot be in the past")
	}

	return nil
}
