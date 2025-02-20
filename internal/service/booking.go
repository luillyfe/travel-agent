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

	// Get flight recommendations
	recommendations, err := s.getFlightRecommendations(travelParams)
	if err != nil {
		return nil, fmt.Errorf("failed to get flight recommendations: %w", err)
	}

	// Create booking response
	response, err := s.createBookingResponse(req, recommendations, deadline)
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

// getFlightRecommendations fetches flight recommendations from the AI engine
func (s *BookingService) getFlightRecommendations(params *ai.TravelParameters) (*ai.FlightRecommendation, error) {
	flightRecommendationStrategy := &ai.FlightRecommendationStrategy{}
	decodingStrategy := &ai.FlightRecommendationDecoder{}

	aiReq := ai.FlightRecommendationRequest{
		DepartureCity: params.DepartureCity,
		Destination:   params.Destination,
		DepartureDate: *params.DepartureDate,
		ReturnDate:    *params.ReturnDate,
		// hardcoded values for now
		PreferredClass: "economy",
		MaxBudget:      2000.0,
		Passengers:     1,
	}

	recommendations, err := ai.ProcessRequest[ai.FlightRecommendation, ai.FlightRecommendationRequest](
		context.Background(),
		s.aiInference,
		flightRecommendationStrategy,
		aiReq,
		decodingStrategy,
	)
	if err != nil {
		return nil, fmt.Errorf("AI recommendation failed: %w", err)
	}

	return recommendations, nil
}

// extractTravelParameters handles the AI parameter extraction
func (s *BookingService) extractTravelParameters(query string, deadline time.Time) (*ai.TravelParameters, error) {
	extractionStrategy := &ai.ExtractionPromptStrategy{}
	decodingStrategy := &ai.ExtractionDecodingStrategy{}

	aiReq := ai.ExtractionRequest{
		Query:    query,
		Deadline: deadline,
	}

	params, err := ai.ProcessRequest[ai.TravelParameters, ai.ExtractionRequest](
		context.Background(),
		s.aiInference,
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
	params *ai.FlightRecommendation,
	deadline time.Time,
) (*models.BookingResponse, error) {
	now := time.Now()
	response := &models.BookingResponse{
		ID:     uuid.New().String(),
		Status: models.StatusProcessing,
		Query:  req.Query,
		FlightDetails: &models.Flight{
			Airline:       params.Recommendations[0].Airline,
			FlightNumber:  params.Recommendations[0].FlightNumber,
			Price:         params.Recommendations[0].EstimatedPrice,
			Currency:      "USD",
			DepartureCity: params.Recommendations[0].DepartureCity,
			ArrivalCity:   params.Recommendations[0].ArrivalCity,
			DepartureTime: params.Recommendations[0].DepartureTime,
			ArrivalTime:   params.Recommendations[0].ArrivalTime,
		},
		Deadline:  deadline,
		CreatedAt: now,
		UpdatedAt: now,
		Message:   fmt.Sprintf("Searching for flights to %s", params.Recommendations[0].ArrivalCity),
	}

	return response, nil
}
