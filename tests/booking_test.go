package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"travel-agent/internal/handlers"
	"travel-agent/internal/models"
	"travel-agent/internal/service"
)

// Test data
var (
	validQuery    = "Book a flight for a 3-day trip to Paris"
	validDeadline = "within 2 days"
)

func TestBookingService_ProcessBooking(t *testing.T) {
	tests := []struct {
		name    string
		req     models.BookingRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: models.BookingRequest{
				Query:    validQuery,
				Deadline: validDeadline,
			},
			wantErr: false,
		},
		{
			name: "empty query",
			req: models.BookingRequest{
				Query:    "",
				Deadline: validDeadline,
			},
			wantErr: true,
		},
		// TODO: Add test for deadline validation
		// {
		// 	name: "past deadline",
		// 	req: models.BookingRequest{
		// 		Query:    validQuery,
		// 		Deadline: "yesterday",
		// 	},
		// 	wantErr: true,
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := service.NewBookingService()
			resp, err := svc.ProcessBooking(tt.req)

			if (err != nil) != tt.wantErr {
				t.Errorf("ProcessBooking() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && resp == nil {
				t.Error("ProcessBooking() returned nil response for valid request")
			}

			if !tt.wantErr {
				if resp.ID == "" {
					t.Error("ProcessBooking() returned empty ID")
				}
				if resp.Status != "pending" {
					t.Errorf("ProcessBooking() returned status %v, want pending", resp.Status)
				}
			}
		})
	}
}

func TestBookingHandler_CreateBooking(t *testing.T) {
	tests := []struct {
		name          string
		requestBody   interface{}
		wantStatus    int
		wantResponse  bool
		checkResponse func(*testing.T, *models.BookingResponse)
	}{
		{
			name: "valid request",
			requestBody: models.BookingRequest{
				Query:    validQuery,
				Deadline: validDeadline,
			},
			wantStatus:   http.StatusOK,
			wantResponse: true,
			checkResponse: func(t *testing.T, resp *models.BookingResponse) {
				if resp.ID == "" {
					t.Error("Expected non-empty ID in response")
				}
				if resp.Status != "pending" {
					t.Errorf("Expected status 'pending', got %s", resp.Status)
				}
			},
		},
		{
			name: "invalid request - missing query",
			requestBody: models.BookingRequest{
				Deadline: validDeadline,
			},
			wantStatus:   http.StatusBadRequest,
			wantResponse: false,
		},
		{
			name:         "invalid request - malformed JSON",
			requestBody:  "{bad json}",
			wantStatus:   http.StatusBadRequest,
			wantResponse: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create handler
			svc := service.NewBookingService()
			handler := handlers.NewBookingHandler(svc)

			// Create request
			var body bytes.Buffer
			if err := json.NewEncoder(&body).Encode(tt.requestBody); err != nil {
				t.Fatalf("Failed to encode request body: %v", err)
			}

			// Create test request and response recorder
			req := httptest.NewRequest(http.MethodPost, "/api/v1/bookings", &body)
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			// Handle request
			handler.CreateBooking(rr, req)

			// Check status code
			if rr.Code != tt.wantStatus {
				t.Errorf("CreateBooking() status = %v, want %v", rr.Code, tt.wantStatus)
			}

			// Check response
			if tt.wantResponse {
				var resp models.BookingResponse
				if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}
				tt.checkResponse(t, &resp)
			}
		})
	}
}

func TestBookingHandler_GetBooking(t *testing.T) {
	tests := []struct {
		name       string
		bookingID  string
		wantStatus int
	}{
		{
			name:       "valid request",
			bookingID:  "123e4567-e89b-12d3-a456-426614174000",
			wantStatus: http.StatusOK,
		},
		{
			name:       "missing booking ID",
			bookingID:  "",
			wantStatus: http.StatusBadRequest,
		},
		// TODO: Add test for invalid booking ID
		// {
		// 	name:       "invalid booking ID",
		// 	bookingID:  "invalid-uuid",
		// 	wantStatus: http.StatusBadRequest,
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create handler
			svc := service.NewBookingService()
			handler := handlers.NewBookingHandler(svc)

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/api/v1/bookings/status?id="+tt.bookingID, nil)
			rr := httptest.NewRecorder()

			// Handle request
			handler.GetBooking(rr, req)

			// Check status code
			if rr.Code != tt.wantStatus {
				t.Errorf("GetBooking() status = %v, want %v", rr.Code, tt.wantStatus)
			}

			// For successful requests, verify response structure
			if tt.wantStatus == http.StatusOK {
				var resp models.BookingResponse
				if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				if resp.ID != tt.bookingID {
					t.Errorf("GetBooking() returned ID = %v, want %v", resp.ID, tt.bookingID)
				}
			}
		})
	}
}

// Integration test
func TestBookingFlow(t *testing.T) {
	// Create service and handler
	svc := service.NewBookingService()
	handler := handlers.NewBookingHandler(svc)

	// 1. Create a booking
	createReq := models.BookingRequest{
		Query:    validQuery,
		Deadline: validDeadline,
	}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(createReq); err != nil {
		t.Fatalf("Failed to encode request body: %v", err)
	}

	createReqHTTP := httptest.NewRequest(http.MethodPost, "/api/v1/bookings", &body)
	createReqHTTP.Header.Set("Content-Type", "application/json")
	createRR := httptest.NewRecorder()

	handler.CreateBooking(createRR, createReqHTTP)

	if createRR.Code != http.StatusOK {
		t.Fatalf("CreateBooking() status = %v, want %v", createRR.Code, http.StatusOK)
	}

	var createResp models.BookingResponse
	if err := json.NewDecoder(createRR.Body).Decode(&createResp); err != nil {
		t.Fatalf("Failed to decode create response: %v", err)
	}

	// 2. Get booking status
	getReqHTTP := httptest.NewRequest(http.MethodGet, "/api/v1/bookings/status?id="+createResp.ID, nil)
	getRR := httptest.NewRecorder()

	handler.GetBooking(getRR, getReqHTTP)

	if getRR.Code != http.StatusOK {
		t.Fatalf("GetBooking() status = %v, want %v", getRR.Code, http.StatusOK)
	}

	var getResp models.BookingResponse
	if err := json.NewDecoder(getRR.Body).Decode(&getResp); err != nil {
		t.Fatalf("Failed to decode get response: %v", err)
	}

	// Verify booking details match
	if getResp.ID != createResp.ID {
		t.Errorf("Booking ID mismatch: got %v, want %v", getResp.ID, createResp.ID)
	}
	if getResp.Query != createReq.Query {
		t.Errorf("Query mismatch: got %v, want %v", getResp.Query, createReq.Query)
	}
}
