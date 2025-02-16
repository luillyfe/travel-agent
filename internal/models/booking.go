package models

import "time"

type BookingRequest struct {
	Query    string `json:"query"`    // Natural language query for the booking
	Deadline string `json:"deadline"` // When to stop looking for deals
}

type BookingResponse struct {
	ID            string    `json:"id"`               // Unique booking request ID
	Status        string    `json:"status"`           // Status of the booking (pending, completed, failed)
	Query         string    `json:"query"`            // Original query
	Deadline      time.Time `json:"deadline"`         // Original deadline
	FlightDetails *Flight   `json:"flight,omitempty"` // Flight details if found
	Message       string    `json:"message"`          // Additional information or error message
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type Flight struct {
	Airline       string    `json:"airline"`
	FlightNumber  string    `json:"flight_number"`
	DepartureCity string    `json:"departure_city"`
	ArrivalCity   string    `json:"arrival_city"`
	DepartureTime time.Time `json:"departure_time"`
	ArrivalTime   time.Time `json:"arrival_time"`
	Price         float64   `json:"price"`
	Currency      string    `json:"currency"`
}
