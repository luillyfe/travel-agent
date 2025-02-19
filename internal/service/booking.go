package service

import (
	"context"
	"fmt"
	"time"
	"travel-agent/internal/models"
	"travel-agent/internal/service/ai"

	"github.com/google/uuid"
)

type AIInference interface {
	ExtractParameters(ctx context.Context, strategy *ai.TravelExtractionStrategy, request ai.ExtractionRequest, decoder ai.DecodingStrategy[ai.TravelParameters]) (*ai.TravelParameters, error)
}

type BookingService struct {
	aiInference AIInference
}

func NewBookingService(aiInference AIInference) *BookingService {
	return &BookingService{
		aiInference: aiInference,
	}
}

// ProcessBooking orchestrates the booking flow
func (s *BookingService) ProcessBooking(req models.BookingRequest) (*models.BookingResponse, error) {
	if req.Query == "" {
		return nil, fmt.Errorf("query cannot be empty")
	}

	// Parse and validate the request
	deadline, err := s.parseDeadline(req.Deadline)
	if err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Extract travel parameters
	travelParams, err := s.extractTravelParameters(req.Query, deadline)
	if err != nil {
		return nil, fmt.Errorf("parameter extraction failed: %w", err)
	}

	// Create booking response
	response, err := s.createBookingResponse(req, travelParams, deadline)
	if err != nil {
		return nil, fmt.Errorf("failed to create booking response: %w", err)
	}

	return response, nil
}

// TODO: parseDeadline needs to accept NL deadlines like "tomorrow at 5pm".
// parseDeadline handles deadline string to time.Time conversion
func (s *BookingService) parseDeadline(deadlineStr string) (time.Time, error) {
	deadline, err := time.Parse(time.RFC3339, deadlineStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid deadline format: %w", err)
	}
	return deadline, nil
}

// extractTravelParameters handles the AI parameter extraction
func (s *BookingService) extractTravelParameters(query string, deadline time.Time) (*ai.TravelParameters, error) {
	extractionStrategy := &ai.TravelExtractionStrategy{}
	decodingStrategy := &ai.TravelDecodingStrategy{}

	aiReq := ai.ExtractionRequest{
		Query:    query,
		Deadline: deadline,
	}

	params, err := s.aiInference.ExtractParameters(
		context.Background(),
		extractionStrategy,
		aiReq,
		decodingStrategy,
	)
	if err != nil {
		return nil, fmt.Errorf("AI extraction failed: %w", err)
	}

	return params, nil
}

// createBookingResponse creates the booking response from extracted parameters
func (s *BookingService) createBookingResponse(
	req models.BookingRequest,
	params *ai.TravelParameters,
	deadline time.Time,
) (*models.BookingResponse, error) {
	now := time.Now()
	response := &models.BookingResponse{
		ID:     uuid.New().String(),
		Status: models.StatusProcessing,
		Query:  req.Query,
		FlightDetails: &models.Flight{
			DepartureCity: params.DepartureCity,
			ArrivalCity:   params.Destination,
			DepartureTime: *params.DepartureDate,
			ArrivalTime:   *params.ReturnDate,
		},
		Deadline:  deadline,
		CreatedAt: now,
		UpdatedAt: now,
		Message:   fmt.Sprintf("Searching for flights to %s", params.Destination),
	}

	return response, nil
}
