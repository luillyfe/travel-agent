package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"travel-agent/internal/handlers"
	"travel-agent/internal/models"
	"travel-agent/internal/service"
	"travel-agent/internal/service/ai"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Test data
var (
	validQuery    = "Book a flight for a 3-day trip to Paris"
	validDeadline = time.Now().Add(48 * time.Hour).Format(time.RFC3339)
)

// MockInferenceEngine mocks the AI inference engine
type MockInferenceEngine struct {
	mock.Mock
}

func (m *MockInferenceEngine) ExtractParameters(
	ctx context.Context,
	strategy *ai.TravelExtractionStrategy,
	request ai.ExtractionRequest,
	decoder ai.DecodingStrategy[ai.TravelParameters],
) (*ai.TravelParameters, error) {
	args := m.Called(ctx, strategy, request, decoder)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ai.TravelParameters), args.Error(1)
}

func TestBookingService_ProcessBooking(t *testing.T) {
	// Helper function for creating time pointers
	timePtr := func(t time.Time) *time.Time {
		return &t
	}

	// Setup common test data
	futureDate := time.Now().Add(24 * time.Hour)
	validDeadline := futureDate.Format(time.RFC3339)

	tests := []struct {
		name          string
		request       models.BookingRequest
		mockParams    *ai.TravelParameters
		mockError     error
		expectedError bool
		validateResp  func(*testing.T, *models.BookingResponse)
	}{
		{
			name: "successful booking",
			request: models.BookingRequest{
				Query:    "I want to fly from New York to London",
				Deadline: validDeadline,
			},
			mockParams: &ai.TravelParameters{
				DepartureCity: "New York",
				Destination:   "London",
				DepartureDate: timePtr(futureDate),
				ReturnDate:    timePtr(futureDate.Add(7 * 24 * time.Hour)),
				Preferences: ai.Preferences{
					BudgetRange: struct {
						Min *float64 `json:"min"`
						Max *float64 `json:"max"`
					}{
						Min: float64Ptr(1000),
						Max: float64Ptr(2000),
					},
					TravelClass: "economy",
				},
			},
			mockError:     nil,
			expectedError: false,
			validateResp: func(t *testing.T, resp *models.BookingResponse) {
				assert.NotNil(t, resp)
				assert.NotEmpty(t, resp.ID)
				assert.Equal(t, models.StatusProcessing, resp.Status)
				assert.Equal(t, "New York", resp.FlightDetails.DepartureCity)
				assert.Equal(t, "London", resp.FlightDetails.ArrivalCity)
			},
		},
		{
			name: "empty query",
			request: models.BookingRequest{
				Query:    "",
				Deadline: validDeadline,
			},
			mockParams:    nil,
			mockError:     assert.AnError,
			expectedError: true,
			validateResp:  nil,
		},
		{
			name: "invalid deadline format",
			request: models.BookingRequest{
				Query:    "Valid query",
				Deadline: "invalid-date",
			},
			mockParams:    nil,
			mockError:     nil,
			expectedError: true,
			validateResp:  nil,
		},
		{
			name: "AI extraction failure",
			request: models.BookingRequest{
				Query:    "Valid query",
				Deadline: validDeadline,
			},
			mockParams:    nil,
			mockError:     assert.AnError,
			expectedError: true,
			validateResp:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock
			mockEngine := new(MockInferenceEngine)
			service := service.NewBookingService(mockEngine)

			// Setup mock expectations if needed
			if tt.request.Query != "" && tt.request.Deadline == validDeadline {
				mockEngine.On("ExtractParameters",
					mock.Anything,
					mock.AnythingOfType("*ai.TravelExtractionStrategy"),
					mock.MatchedBy(func(req ai.ExtractionRequest) bool {
						return req.Query == tt.request.Query
					}),
					mock.AnythingOfType("*ai.TravelDecodingStrategy"),
				).Return(tt.mockParams, tt.mockError)
			}

			// Execute test
			resp, err := service.ProcessBooking(tt.request)

			// Verify results
			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				if tt.validateResp != nil {
					tt.validateResp(t, resp)
				}
			}

			// Verify mock expectations
			mockEngine.AssertExpectations(t)
		})
	}
}

// Helper function for float64 pointers
func float64Ptr(v float64) *float64 {
	return &v
}

func TestBookingHandler_CreateBooking(t *testing.T) {
	// Create mock inference engine
	mockEngine := new(MockInferenceEngine)

	// Setup mock expectations for valid requests
	mockEngine.On("ExtractParameters",
		mock.Anything,
		mock.AnythingOfType("*ai.TravelExtractionStrategy"),
		mock.MatchedBy(func(req ai.ExtractionRequest) bool {
			return req.Query == validQuery
		}),
		mock.AnythingOfType("*ai.TravelDecodingStrategy"),
	).Return(&ai.TravelParameters{
		DepartureCity: "New York", // example city
		Destination:   "Paris",
		DepartureDate: timePtr(time.Now().Add(24 * time.Hour)),
		ReturnDate:    timePtr(time.Now().Add(96 * time.Hour)),
		Preferences:   ai.Preferences{},
	}, nil)

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
				assert.NotEmpty(t, resp.ID, "Expected non-empty ID in response")
				assert.Equal(t, models.StatusProcessing, resp.Status, "Expected status 'processing'")
				assert.Equal(t, "Paris", resp.FlightDetails.ArrivalCity, "Expected arrival city to be Paris")
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
			svc := service.NewBookingService(mockEngine)
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
			assert.Equal(t, tt.wantStatus, rr.Code, "Unexpected status code")

			// Check response
			if tt.wantResponse {
				var resp models.BookingResponse
				err := json.NewDecoder(rr.Body).Decode(&resp)
				assert.NoError(t, err, "Failed to decode response")

				if tt.checkResponse != nil {
					tt.checkResponse(t, &resp)
				}
			}

			// Log response body if test fails
			if t.Failed() {
				t.Logf("Response Body: %s", rr.Body.String())
			}
		})
	}
}

// Helper function for time pointers
func timePtr(t time.Time) *time.Time {
	return &t
}

func TestBookingHandler_GetBooking(t *testing.T) {
	mockEngine := &ai.InferenceEngine{}

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
			svc := service.NewBookingService(mockEngine)
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

func TestBookingFlow(t *testing.T) {
	// Create mock inference engine with expectations
	mockEngine := new(MockInferenceEngine)

	// Update mock to be more flexible with the request matching
	mockEngine.On("ExtractParameters",
		mock.Anything, // context
		mock.AnythingOfType("*ai.TravelExtractionStrategy"),
		mock.MatchedBy(func(req ai.ExtractionRequest) bool {
			// More flexible matching - just ensure it's not empty
			return req.Query != "" && !req.Deadline.IsZero()
		}),
		mock.AnythingOfType("*ai.TravelDecodingStrategy"),
	).Return(&ai.TravelParameters{
		DepartureCity: "Cúcuta",
		Destination:   "Paris",
		DepartureDate: timePtr(time.Date(2025, 3, 19, 5, 0, 0, 0, time.UTC)),
		ReturnDate:    timePtr(time.Date(2025, 3, 22, 5, 0, 0, 0, time.UTC)),
		Preferences: ai.Preferences{
			BudgetRange: struct {
				Min *float64 `json:"min"`
				Max *float64 `json:"max"`
			}{
				Min: float64Ptr(1000),
				Max: float64Ptr(2000),
			},
			TravelClass: "economy",
			Activities:  []string{"sightseeing"},
		},
	}, nil)

	// Create service and handler
	svc := service.NewBookingService(mockEngine)
	handler := handlers.NewBookingHandler(svc)

	// Test flow
	t.Run("complete booking flow", func(t *testing.T) {
		// 1. Create a booking
		createReq := models.BookingRequest{
			Query:    "Book a flight for a 3-day trip to Paris, from Cúcuta starting 2025-03-19T05:00:00Z",
			Deadline: validDeadline,
		}

		var body bytes.Buffer
		err := json.NewEncoder(&body).Encode(createReq)
		require.NoError(t, err, "Failed to encode request body")

		createReqHTTP := httptest.NewRequest(http.MethodPost, "/api/v1/bookings", &body)
		createReqHTTP.Header.Set("Content-Type", "application/json")
		createRR := httptest.NewRecorder()

		handler.CreateBooking(createRR, createReqHTTP)

		require.Equal(t, http.StatusOK, createRR.Code, "Unexpected status code for create booking")

		var createResp models.BookingResponse
		err = json.NewDecoder(createRR.Body).Decode(&createResp)
		require.NoError(t, err, "Failed to decode create response")

		// Verify create response
		assert.NotEmpty(t, createResp.ID, "Booking ID should not be empty")
		assert.Equal(t, models.StatusProcessing, createResp.Status, "Initial status should be processing")
		assert.Equal(t, createReq.Query, createResp.Query, "Query should match request")
		assert.NotNil(t, createResp.FlightDetails, "Flight details should not be nil")
		assert.Equal(t, "Paris", createResp.FlightDetails.ArrivalCity, "Destination should be Paris")

		// 2. Get booking status
		getRR := httptest.NewRecorder()
		getReqHTTP := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/bookings/status?id=%s", createResp.ID), nil)

		handler.GetBooking(getRR, getReqHTTP)

		require.Equal(t, http.StatusOK, getRR.Code, "Unexpected status code for get booking")

		var getResp models.BookingResponse
		err = json.NewDecoder(getRR.Body).Decode(&getResp)
		require.NoError(t, err, "Failed to decode get response")

		// Verify get response matches create response
		assert.Equal(t, createResp.ID, getResp.ID, "Booking ID should match")
		assert.Equal(t, createResp.Status, getResp.Status, "Status should match")
	})

	// TODO: Test non-existent booking cases

	t.Run("create booking with invalid request", func(t *testing.T) {
		invalidReq := models.BookingRequest{
			Query:    "", // Empty query
			Deadline: validDeadline,
		}

		var body bytes.Buffer
		err := json.NewEncoder(&body).Encode(invalidReq)
		require.NoError(t, err, "Failed to encode invalid request")

		reqHTTP := httptest.NewRequest(http.MethodPost, "/api/v1/bookings", &body)
		reqHTTP.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.CreateBooking(rr, reqHTTP)

		assert.Equal(t, http.StatusBadRequest, rr.Code, "Should return bad request for invalid booking")

		// Optionally verify error message
		var errorResp map[string]string
		err = json.NewDecoder(rr.Body).Decode(&errorResp)
		require.NoError(t, err, "Failed to decode error response")
		assert.Contains(t, errorResp["error"], "query cannot be empty", "Should return appropriate error message")
	})

	// Verify all mock expectations were met
	mockEngine.AssertExpectations(t)
}
