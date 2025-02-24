package models

import (
	"time"
)

type BookingStatus string

const (
	StatusProcessing BookingStatus = "processing"
	StatusConfirmed  BookingStatus = "confirmed"
	StatusFailed     BookingStatus = "failed"
)

// Input structure for the extraction
type BookingRequest struct {
	Query    string    `json:"query"`    // Natural language query for the booking
	Deadline time.Time `json:"deadline"` // When to stop looking for deals
	// Deadline string `json:"deadline"`
}

type BookingResponse struct {
	ID            string        `json:"id"`               // Unique booking request ID
	Status        BookingStatus `json:"status"`           // Status of the booking (pending, completed, failed)
	Query         string        `json:"query"`            // Original query
	Deadline      time.Time     `json:"deadline"`         // Original deadline
	FlightDetails *Flight       `json:"flight,omitempty"` // Flight details if found
	Message       string        `json:"message"`          // Additional information or error message
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
}

// Define the expected output structure
type TravelParameters struct {
	DepartureCity string      `json:"departure_city"`
	Destination   string      `json:"destination"`
	DepartureDate *time.Time  `json:"departure_date"`
	ReturnDate    *time.Time  `json:"return_date"`
	Preferences   Preferences `json:"preferences"`
}

type Preferences struct {
	BudgetRange struct {
		Min *float64 `json:"min"`
		Max *float64 `json:"max"`
	} `json:"budget_range"`
	TravelClass         string   `json:"travel_class"`
	Activities          []string `json:"activities"`
	DietaryRestrictions []string `json:"dietary_restrictions"`
}

// FlightRecommendationRequest represents the input for flight recommendations
type FlightRecommendationRequest struct {
	DepartureCity  string    `json:"departure_city"`
	Destination    string    `json:"destination"`
	DepartureDate  time.Time `json:"departure_date"`
	ReturnDate     time.Time `json:"return_date"`
	Passengers     int       `json:"passengers"`
	MaxBudget      float64   `json:"max_budget,omitempty"`
	PreferredClass string    `json:"preferred_class,omitempty"`
}

// FlightRecommendation represents the structured output
type FlightRecommendation struct {
	Recommendations []Flight `json:"recommendations"`
	Reasoning       string   `json:"reasoning"`
}

type Flight struct {
	Airline             string    `json:"airline"`
	FlightNumber        string    `json:"flight_number"`
	DepartureCity       string    `json:"departure_city"`
	DepartureTime       time.Time `json:"departure_time"`
	ArrivalCity         string    `json:"arrival_city"`
	ArrivalTime         time.Time `json:"arrival_time"`
	Class               string    `json:"class"`
	LayoverCount        int       `json:"layover_count"`
	TotalDuration       string    `json:"total_duration"`
	AvailableSeats      int       `json:"available_seats"`
	RecommendationScore float64   `json:"recommendation_score"`
	Price               float64   `json:"price"`
	Currency            string    `json:"currency"`
}

// Define mock response and request types
type MockTravelResponse struct{}
type MockTravelRequest struct{}

// Define a single type for all travel-related requests
type TravelInput interface {
	BookingRequest | FlightRecommendationRequest | MockTravelRequest
}

// Define a single type for all travel-related responses
type TravelOutput interface {
	TravelParameters | FlightRecommendation | MockTravelResponse
}
