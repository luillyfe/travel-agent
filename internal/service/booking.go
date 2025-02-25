package service

import (
	"context"
	"fmt"
	"time"
	"travel-agent/internal/models"
	"travel-agent/internal/service/ai"

	"github.com/google/uuid"
)

type TravelParameterExtractor interface {
	ProcessRequest(
		ctx context.Context,
		strategy ai.PromptStrategy[models.BookingRequest],
		request models.BookingRequest,
		decoder ai.DecodingStrategy[models.TravelParameters],
	) (*models.TravelParameters, error)
}

type FlightRecommender interface {
	ProcessRequest(
		ctx context.Context,
		strategy ai.PromptStrategy[models.FlightRecommendationRequest],
		request models.FlightRecommendationRequest,
		decoder ai.DecodingStrategy[models.FlightRecommendation],
	) (*models.FlightRecommendation, error)
}

type BookingService struct {
	paramExtractor    TravelParameterExtractor
	flightRecommender FlightRecommender
}

func NewBookingService(
	paramExtractor TravelParameterExtractor,
	flightRecommender FlightRecommender,
) *BookingService {
	return &BookingService{
		paramExtractor:    paramExtractor,
		flightRecommender: flightRecommender,
	}
}

// ProcessBooking orchestrates the booking flow
func (s *BookingService) ProcessBooking(ctx context.Context, req models.BookingRequest) (*models.BookingResponse, error) {
	if req.Query == "" {
		return nil, fmt.Errorf("query cannot be empty")
	}

	// Extract travel parameters
	travelParams, err := s.extractTravelParameters(ctx, req.Query, req.Deadline)
	if err != nil {
		return nil, fmt.Errorf("parameter extraction failed: %w", err)
	}

	// Get flight recommendations
	recommendations, err := s.getFlightRecommendations(ctx, travelParams)
	if err != nil {
		return nil, fmt.Errorf("failed to get flight recommendations: %w", err)
	}

	// Create booking response
	response, err := s.createBookingResponse(req, recommendations, req.Deadline)
	if err != nil {
		return nil, fmt.Errorf("failed to create booking response: %w", err)
	}

	return response, nil
}

// getFlightRecommendations fetches flight recommendations from the AI engine
func (s *BookingService) getFlightRecommendations(ctx context.Context, params *models.TravelParameters) (*models.FlightRecommendation, error) {
	flightRecommendationStrategy := &ai.FlightRecommendationStrategy{}
	decodingStrategy := &ai.FlightRecommendationDecoder{}

	aiReq := models.FlightRecommendationRequest{
		DepartureCity: params.DepartureCity,
		Destination:   params.Destination,
		DepartureDate: *params.DepartureDate,
		ReturnDate:    *params.ReturnDate,
		// hardcoded values for now
		PreferredClass: "economy",
		MaxBudget:      2000.0,
		Passengers:     1,
	}

	recommendations, err := s.flightRecommender.ProcessRequest(
		ctx,
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
func (s *BookingService) extractTravelParameters(ctx context.Context, query string, deadline time.Time) (*models.TravelParameters, error) {
	extractionStrategy := &ai.ExtractionPromptStrategy{}
	decodingStrategy := &ai.ExtractionDecodingStrategy{}

	aiReq := models.BookingRequest{
		Query:    query,
		Deadline: deadline,
	}

	params, err := s.paramExtractor.ProcessRequest(
		ctx,
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
	params *models.FlightRecommendation,
	deadline time.Time,
) (*models.BookingResponse, error) {
	// check if params.Recommendations is empty first.
	if len(params.Recommendations) == 0 {
		return nil, fmt.Errorf("no flight recommendations found")
	}

	now := time.Now()
	response := &models.BookingResponse{
		ID:     uuid.New().String(),
		Status: models.StatusProcessing,
		Query:  req.Query,
		FlightDetails: &models.Flight{
			Airline:       params.Recommendations[0].Airline,
			FlightNumber:  params.Recommendations[0].FlightNumber,
			Price:         params.Recommendations[0].Price,
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
