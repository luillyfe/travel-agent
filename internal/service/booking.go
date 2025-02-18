package service

import (
	"context"
	"fmt"
	"time"
	"travel-agent/internal/models"
	"travel-agent/internal/service/ai"

	"github.com/google/uuid"
)

type BookingService struct {
	aiInference *ai.InferenceEngine
}

func NewBookingService(aiInference *ai.InferenceEngine) *BookingService {
	return &BookingService{
		aiInference: aiInference,
	}
}

func (s *BookingService) ProcessBooking(req models.BookingRequest) (*models.BookingResponse, error) {
	// TODO: Should I create a middleware for type casting (string) to time.Time?
	// Cast req.Deadline (string) to time.Time
	deadline, err := time.Parse(time.RFC3339, req.Deadline)
	if err != nil {
		return nil, fmt.Errorf("failed to parse deadline: %w", err)
	}

	// Create AI request
	aiReq := ai.TravelRequest{
		Query:    req.Query,
		Deadline: deadline,
	}

	// Process with AI
	aiResp, err := s.aiInference.ProcessTravelRequest(context.Background(), aiReq)
	if err != nil {
		return nil, fmt.Errorf("AI processing failed: %w", err)
	}

	// Create booking response
	response := &models.BookingResponse{
		ID:     uuid.New().String(),
		Status: "processing",
		Query:  req.Query,
		FlightDetails: &models.Flight{
			DepartureCity: "User's City", // TODO: Extract from context
			ArrivalCity:   aiResp.Destination,
			DepartureTime: aiResp.DepartureDate,
			ArrivalTime:   aiResp.ReturnDate,
			// Other fields will be filled when actual flight is found
		},
		Deadline:  deadline,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Message:   fmt.Sprintf("Searching for flights to %s", aiResp.Destination),
	}

	return response, nil
}
