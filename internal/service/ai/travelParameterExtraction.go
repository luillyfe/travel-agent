package ai

import (
	"encoding/json"
	"fmt"
	"time"
)

// Input structure for the extraction
type ExtractionRequest struct {
	Query    string
	Deadline time.Time
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

// TravelExtractionStrategy handles the extraction of travel parameters from natural language
type TravelExtractionStrategy struct{}

// Make TravelExtractionStrategy implement PromptStrategy[ExtractionRequest]
var _ PromptStrategy[ExtractionRequest] = (*TravelExtractionStrategy)(nil) // Type assertion for interface compliance

// GetSystemPrompt returns the system prompt for parameter extraction
func (s *TravelExtractionStrategy) GetSystemPrompt() string {
	return `You are an AI travel assistant specialized in extracting structured travel information from natural language requests.

Output must be a valid JSON object with this exact structure:
{
    "departure_city": "city name",
    "destination": "city name",
    "departure_date": null,
    "return_date": null,
    "preferences": {
        "budget_range": {
            "min": null,
            "max": null
        },
        "travel_class": "",
        "activities": [],
        "dietary_restrictions": []
    }
}

Extraction Rules:
1. Use null for missing or uncertain values
2. Format dates as RFC3339 (e.g., "2024-01-15T12:00:00Z")
3. Use empty arrays [] for missing lists
4. Convert prices to numbers without currency symbols
5. Normalize city names to official names
6. Extract both explicit and implicit requirements
7. Omit unreferenced fields

Return only the JSON object, no additional text.`
}

// GetUserPrompt formats the user prompt with the request details
func (s *TravelExtractionStrategy) GetUserPrompt(req ExtractionRequest) string {
	return fmt.Sprintf(`Extract travel parameters from this request:

REQUEST TEXT:
%s

BOOKING DEADLINE:
%s

Required Parameters:
- Departure and destination cities
- Travel dates
- Budget information
- Travel preferences and requirements

Format as specified JSON structure.`,
		req.Query,
		req.Deadline.Format(time.RFC3339))
}

// validate checks if the required fields are present and valid
func (d *TravelDecodingStrategy) validate(params *TravelParameters) error {
	if params.DepartureCity == "" {
		return fmt.Errorf("departure city is required")
	}
	if params.Destination == "" {
		return fmt.Errorf("destination is required")
	}

	if params.DepartureDate == nil {
		return fmt.Errorf("departure date is required")
	}

	if params.ReturnDate == nil {
		return fmt.Errorf("return date is required")
	}

	// Validate dates if present
	if params.DepartureDate != nil {
		if params.DepartureDate.Before(time.Now()) {
			return fmt.Errorf("departure date cannot be in the past")
		}
	}
	if params.ReturnDate != nil {
		if params.DepartureDate != nil && params.ReturnDate.Before(*params.DepartureDate) {
			return fmt.Errorf("return date cannot be before departure date")
		}
	}

	return nil
}

// TravelDecodingStrategy implements DecodingStrategy for travel parameters
type TravelDecodingStrategy struct{}

func (d *TravelDecodingStrategy) DecodeResponse(content string) (*TravelParameters, error) {
	// Parse the JSON content into TravelParameters
	var params TravelParameters
	if err := json.Unmarshal([]byte(content), &params); err != nil {
		return nil, fmt.Errorf("failed to parse travel parameters: %w", err)
	}

	// Validate required fields
	if err := d.validate(&params); err != nil {
		return nil, fmt.Errorf("invalid travel parameters: %w", err)
	}

	return &params, nil
}
