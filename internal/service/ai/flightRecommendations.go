package ai

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

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
	EstimatedPrice      float64   `json:"estimated_price"`
	LayoverCount        int       `json:"layover_count"`
	TotalDuration       string    `json:"total_duration"`
	AvailableSeats      int       `json:"available_seats"`
	RecommendationScore float64   `json:"recommendation_score"`
}

// FlightRecommendationStrategy implements the PromptStrategy interface
type FlightRecommendationStrategy struct{}

func (s *FlightRecommendationStrategy) GetSystemPrompt() string {
	return `You are an AI Flight Recommendation Assistant specialized in analyzing travel requirements and suggesting optimal flight options. Your task is to recommend flights based on the provided criteria and explain your reasoning.

Output must be a valid JSON object with this exact structure:
{
    "recommendations": [
        {
            "airline": "string",
            "flight_number": "string",
            "departure_city": "string",
            "departure_time": "YYYY-MM-DDTHH:MM:SSZ",
            "arrival_city": "string",
            "arrival_time": "YYYY-MM-DDTHH:MM:SSZ",
            "class": "string",
            "estimated_price": number,
            "layover_count": number,
            "total_duration": "string",
            "available_seats": number,
            "recommendation_score": number
        }
    ],
    "reasoning": "string explaining why these flights were recommended",
}

Recommendation Rules:
1. Prioritize direct flights when available
2. Consider price-to-convenience ratio
3. Account for reasonable connection times (2-4 hours)
4. Factor in airline reliability and service quality
5. Consider time of day and arrival/departure convenience
6. Account for seasonal factors and typical delays
7. Consider airport-specific factors

Return only the JSON object, no additional text or explanation.`
}

func (s *FlightRecommendationStrategy) GetUserPrompt(req FlightRecommendationRequest) string {
	return fmt.Sprintf(`Analyze this flight request and provide recommendations:

TRAVEL DETAILS:
- Departure: %s
- Destination: %s
- Departure Date: %s
- Return Date: %s
- Preferred Class: %s
- Maximum Budget: %.2f
- Passengers: %d

Additional Context:
%s

Please recommend optimal flights considering:
1. Price within budget (%%..2f per passenger)
2. Convenient departure/arrival times
3. Airline reliability
4. Connection efficiency
5. Overall value

Format recommendations according to the specified JSON structure.`,
		req.DepartureCity,
		req.Destination,
		req.DepartureDate.Format(time.RFC3339),
		req.ReturnDate.Format(time.RFC3339),
		req.PreferredClass,
		req.MaxBudget,
		req.Passengers,
		"No additional context provided",
	)
}

// FlightRecommendationDecoder implements the DecodingStrategy interface
type FlightRecommendationDecoder struct{}

func (d *FlightRecommendationDecoder) DecodeResponse(content string) (*FlightRecommendation, error) {
	var recommendation FlightRecommendation
	if err := json.Unmarshal([]byte(content), &recommendation); err != nil {
		return nil, fmt.Errorf("failed to decode flight recommendations: %w", err)
	}

	// Validate recommendations
	if err := d.validate(&recommendation); err != nil {
		return nil, fmt.Errorf("invalid flight recommendations: %w", err)
	}

	return &recommendation, nil
}

func (d *FlightRecommendationDecoder) validate(rec *FlightRecommendation) error {
	if len(rec.Recommendations) == 0 {
		return errors.New("no flight recommendations provided")
	}

	for i, flight := range rec.Recommendations {
		if flight.Airline == "" {
			return fmt.Errorf("missing airline for recommendation %d", i+1)
		}
		if flight.FlightNumber == "" {
			return fmt.Errorf("missing flight number for recommendation %d", i+1)
		}
		if flight.EstimatedPrice <= 0 {
			return fmt.Errorf("invalid price for recommendation %d", i+1)
		}
		// Add more validation as needed
	}

	if rec.Reasoning == "" {
		return errors.New("missing recommendation reasoning")
	}

	return nil
}
