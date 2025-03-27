package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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
)

// Mock implementations
type MockTravelParameterExtractor struct {
	mock.Mock
}

func (m *MockTravelParameterExtractor) ProcessRequest(
	ctx context.Context,
	strategy ai.PromptStrategy[models.BookingRequest],
	request models.BookingRequest,
	decoder ai.DecodingStrategy[models.TravelParameters],
) (*models.TravelParameters, error) {
	args := m.Called(ctx, strategy, request, decoder)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TravelParameters), args.Error(1)
}

type MockFlightRecommender struct {
	mock.Mock
}

func (m *MockFlightRecommender) ProcessRequest(
	ctx context.Context,
	strategy ai.PromptStrategy[models.FlightRecommendationRequest],
	request models.FlightRecommendationRequest,
	decoder ai.DecodingStrategy[models.FlightRecommendation],
) (*models.FlightRecommendation, error) {
	args := m.Called(ctx, strategy, request, decoder)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.FlightRecommendation), args.Error(1)
}

func TestBookingService_ProcessBooking(t *testing.T) {
	tests := []struct {
		name          string
		request       models.BookingRequest
		setupMocks    func(*MockTravelParameterExtractor, *MockFlightRecommender)
		expectedError bool
		validate      func(*testing.T, *models.BookingResponse)
	}{
		{
			name: "Successful booking",
			request: models.BookingRequest{
				Query:    "I want to fly from NYC to London next week",
				Deadline: time.Now().Add(24 * time.Hour),
			},
			setupMocks: func(extractor *MockTravelParameterExtractor, recommender *MockFlightRecommender) {
				// Setup expected travel parameters
				travelParams := &models.TravelParameters{
					DepartureCity: "NYC",
					Destination:   "London",
					DepartureDate: &time.Time{}, // Set appropriate time
					ReturnDate:    &time.Time{}, // Set appropriate time
				}
				extractor.On("ProcessRequest", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(travelParams, nil)

				// Setup expected flight recommendation
				flightRec := &models.FlightRecommendation{
					Recommendations: []models.Flight{
						{
							Airline:       "British Airways",
							FlightNumber:  "BA123",
							Price:         800.0,
							DepartureCity: "NYC",
							ArrivalCity:   "London",
							DepartureTime: time.Now(),
							ArrivalTime:   time.Now().Add(7 * time.Hour),
						},
					},
				}
				recommender.On("ProcessRequest", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(flightRec, nil)
			},
			expectedError: false,
			validate: func(t *testing.T, response *models.BookingResponse) {
				assert.NotEmpty(t, response.ID)
				assert.Equal(t, models.StatusProcessing, response.Status)
				assert.Equal(t, "British Airways", response.FlightDetails.Airline)
				assert.Equal(t, "NYC", response.FlightDetails.DepartureCity)
				assert.Equal(t, "London", response.FlightDetails.ArrivalCity)
			},
		},
		{
			name: "Empty query",
			request: models.BookingRequest{
				Query:    "",
				Deadline: time.Now().Add(24 * time.Hour),
			},
			setupMocks: func(extractor *MockTravelParameterExtractor, recommender *MockFlightRecommender) {
				// No mock setup needed for this case
			},
			expectedError: true,
			validate:      nil,
		},
		// Add more test cases for different scenarios
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockExtractor := new(MockTravelParameterExtractor)
			mockRecommender := new(MockFlightRecommender)
			tt.setupMocks(mockExtractor, mockRecommender)

			svc := service.NewBookingService(mockExtractor, mockRecommender)

			// Execute
			response, err := svc.ProcessBooking(context.Background(), tt.request)

			// Verify
			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				if tt.validate != nil {
					tt.validate(t, response)
				}
			}

			// Verify all expectations on mocks were met
			mockExtractor.AssertExpectations(t)
			mockRecommender.AssertExpectations(t)
		})
	}
}

// MockBookingService mocks the booking service
type MockBookingService struct {
	mock.Mock
	processBookingFunc    func(ctx context.Context, req models.BookingRequest) (*models.BookingResponse, error)
	processGetBookingFunc func(w http.ResponseWriter, r *http.Request)
}

func (m *MockBookingService) ProcessBooking(ctx context.Context, req models.BookingRequest) (*models.BookingResponse, error) {
	return m.processBookingFunc(ctx, req)
}

func (m *MockBookingService) GetBooking(w http.ResponseWriter, r *http.Request) {
	m.processGetBookingFunc(w, r)
}

func TestBookingHandler_CreateBooking(t *testing.T) {
	tests := []struct {
		name             string
		requestBody      interface{}
		setupMock        func(*MockBookingService)
		expectedStatus   int
		expectedResponse *models.BookingResponse
	}{
		{
			name: "successful booking creation",
			requestBody: models.BookingRequest{
				Query:    "test query",
				Deadline: time.Now().Add(24 * time.Hour),
			},
			setupMock: func(m *MockBookingService) {
				m.processBookingFunc = func(ctx context.Context, req models.BookingRequest) (*models.BookingResponse, error) {
					return &models.BookingResponse{
						ID:     "123",
						Status: models.StatusProcessing,
					}, nil
				}
			},
			expectedStatus: http.StatusOK,
			expectedResponse: &models.BookingResponse{
				ID:     "123",
				Status: models.StatusProcessing,
			},
		},
		{
			name:        "Invalid request body",
			requestBody: "invalid json",
			setupMock: func(mockService *MockBookingService) {
				// No mock setup needed
			},
			expectedStatus:   http.StatusBadRequest,
			expectedResponse: nil,
		},
		{
			name: "Missing query",
			requestBody: models.BookingRequest{
				Query:    "",
				Deadline: time.Now().Add(48 * time.Hour),
			},
			setupMock: func(m *MockBookingService) {
				m.processBookingFunc = func(ctx context.Context, req models.BookingRequest) (*models.BookingResponse, error) {
					return nil, errors.New("query cannot be empty")
				}
			},
			expectedStatus:   http.StatusBadRequest,
			expectedResponse: nil,
		},
		{
			name: "Invalid deadline format",
			requestBody: struct {
				Query    string
				Deadline string
			}{
				Query:    "I want to fly from NYC to London",
				Deadline: "within 2 days", // Deadline must be a valid time
			},
			setupMock: func(m *MockBookingService) {
				m.processBookingFunc = func(ctx context.Context, req models.BookingRequest) (*models.BookingResponse, error) {
					return nil, errors.New("invalid deadline format")
				}
			},
			expectedStatus:   http.StatusBadRequest,
			expectedResponse: nil,
		},
		{
			name: "Service error",
			requestBody: models.BookingRequest{
				Query:    "I want to fly from NYC to London",
				Deadline: time.Now().Add(48 * time.Hour),
			},
			setupMock: func(m *MockBookingService) {
				m.processBookingFunc = func(ctx context.Context, req models.BookingRequest) (*models.BookingResponse, error) {
					return nil, errors.New("internal service error")
				}
			},
			expectedStatus:   http.StatusInternalServerError,
			expectedResponse: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock service
			mockService := &MockBookingService{}

			// Setup mock behavior
			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			handler := handlers.NewBookingHandler(mockService)

			// Create request
			var requestBody []byte
			var err error

			switch v := tt.requestBody.(type) {
			case string:
				requestBody = []byte(v)
			default:
				requestBody, err = json.Marshal(tt.requestBody)
				assert.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, "/bookings", bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Perform request
			handler.CreateBooking(w, req)

			// Check status code
			if status := w.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.expectedStatus)
			}

			// If we expect a response, verify it
			if tt.expectedResponse != nil {
				var response models.BookingResponse
				err := json.NewDecoder(w.Body).Decode(&response)
				if err != nil {
					t.Fatalf("Failed to decode response body: %v", err)
				}

				if response.ID != tt.expectedResponse.ID {
					t.Errorf("handler returned unexpected ID: got %v want %v",
						response.ID, tt.expectedResponse.ID)
				}

				if response.Status != tt.expectedResponse.Status {
					t.Errorf("handler returned unexpected status: got %v want %v",
						response.Status, tt.expectedResponse.Status)
				}
			}
		})
	}
}

func TestBookingHandler_GetBooking(t *testing.T) {
	// Helper function to create a sample booking response
	createSampleBooking := func(id string) *models.BookingResponse {
		now := time.Now()
		return &models.BookingResponse{
			ID:     id,
			Status: models.StatusProcessing,
			Query:  "I want to fly from NYC to London next week",
			FlightDetails: &models.Flight{
				Airline:       "British Airways",
				FlightNumber:  "BA123",
				Price:         800.0,
				Currency:      "USD",
				DepartureCity: "NYC",
				ArrivalCity:   "London",
				DepartureTime: now.Add(24 * time.Hour),
				ArrivalTime:   now.Add(31 * time.Hour),
			},
			Deadline:  now.Add(48 * time.Hour),
			CreatedAt: now,
			UpdatedAt: now,
			Message:   "Searching for flights to London",
		}
	}

	tests := []struct {
		name             string
		bookingID        string
		setupMock        func(*MockBookingService)
		expectedStatus   int
		validateResponse func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:      "Successful booking retrieval",
			bookingID: "valid-booking-id",
			setupMock: func(m *MockBookingService) {
				m.processGetBookingFunc = func(w http.ResponseWriter, r *http.Request) {
					booking := createSampleBooking("valid-booking-id")
					if err := json.NewEncoder(w).Encode(booking); err != nil {
						t.Fatalf("Failed to encode response: %v", err)
					}
				}
			},
			expectedStatus: http.StatusOK,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response models.BookingResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "valid-booking-id", response.ID)
				assert.Equal(t, models.StatusProcessing, response.Status)
			},
		},
		{
			name:      "Empty booking ID",
			bookingID: "",
			setupMock: func(m *MockBookingService) {
				// No mock setup needed
			},
			expectedStatus: http.StatusBadRequest,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, "Booking ID is required\n", w.Body.String())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := new(MockBookingService)
			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			handler := handlers.NewBookingHandler(mockService)

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/bookings/status?id="+tt.bookingID, nil)
			w := httptest.NewRecorder()

			// Perform request
			handler.GetBooking(w, req)

			// Assert status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Validate response
			if tt.validateResponse != nil {
				tt.validateResponse(t, w)
			}

			// Verify mock expectations
			mockService.AssertExpectations(t)
		})
	}
}

func TestBookingFlow(t *testing.T) {
	// Setup test time
	now := time.Now()
	departureTime := now.Add(24 * time.Hour)
	returnTime := departureTime.Add(7 * 24 * time.Hour)

	tests := []struct {
		name           string
		setupMocks     func(*MockTravelParameterExtractor, *MockFlightRecommender)
		bookingRequest models.BookingRequest
		validateFlow   func(*testing.T, *httptest.ResponseRecorder, *httptest.ResponseRecorder)
	}{
		{
			name: "Successful booking flow",
			bookingRequest: models.BookingRequest{
				Query:    "I want to fly from NYC to London next week",
				Deadline: now.Add(48 * time.Hour),
			},
			setupMocks: func(extractionEngine *MockTravelParameterExtractor, recommendationEngine *MockFlightRecommender) {
				// Setup extraction engine mock
				extractionEngine.On("ProcessRequest",
					mock.Anything,
					mock.AnythingOfType("*ai.ExtractionPromptStrategy"),
					mock.AnythingOfType("models.BookingRequest"),
					mock.AnythingOfType("*ai.ExtractionDecodingStrategy"),
				).Return(&models.TravelParameters{
					DepartureCity: "NYC",
					Destination:   "London",
					DepartureDate: &departureTime,
					ReturnDate:    &returnTime,
				}, nil)

				// Setup recommendation engine mock
				recommendationEngine.On("ProcessRequest",
					mock.Anything,
					mock.AnythingOfType("*ai.FlightRecommendationStrategy"),
					mock.AnythingOfType("models.FlightRecommendationRequest"),
					mock.AnythingOfType("*ai.FlightRecommendationDecoder"),
				).Return(&models.FlightRecommendation{
					Recommendations: []models.Flight{
						{
							Airline:       "British Airways",
							FlightNumber:  "BA123",
							Price:         800.0,
							DepartureCity: "NYC",
							ArrivalCity:   "London",
							DepartureTime: departureTime,
							ArrivalTime:   departureTime.Add(7 * time.Hour),
						},
					},
				}, nil)
			},
			validateFlow: func(t *testing.T, createResp, getResp *httptest.ResponseRecorder) {
				// Validate create booking response
				assert.Equal(t, http.StatusOK, createResp.Code)
				var createResponse models.BookingResponse
				err := json.Unmarshal(createResp.Body.Bytes(), &createResponse)
				assert.NoError(t, err)
				assert.NotEmpty(t, createResponse.ID)
				assert.Equal(t, models.StatusProcessing, createResponse.Status)
				assert.Equal(t, "British Airways", createResponse.FlightDetails.Airline)

				// Validate get booking response
				assert.Equal(t, http.StatusOK, getResp.Code)
				var getResponse models.BookingResponse
				err = json.Unmarshal(getResp.Body.Bytes(), &getResponse)
				assert.NoError(t, err)
				assert.Equal(t, createResponse.ID, getResponse.ID)
			},
		},
		{
			name: "Booking flow with AI extraction failure",
			bookingRequest: models.BookingRequest{
				Query:    "I want to fly from NYC to London next week",
				Deadline: now.Add(48 * time.Hour),
			},
			setupMocks: func(extractionEngine *MockTravelParameterExtractor, recommendationEngine *MockFlightRecommender) {
				extractionEngine.On("ProcessRequest",
					mock.Anything,
					mock.AnythingOfType("*ai.ExtractionPromptStrategy"),
					mock.AnythingOfType("models.BookingRequest"),
					mock.AnythingOfType("*ai.ExtractionDecodingStrategy"),
				).Return(nil, errors.New("AI extraction failed"))
			},
			validateFlow: func(t *testing.T, createResp, getResp *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusInternalServerError, createResp.Code)
				errorResponse := createResp.Body.String() // Get response as string
				assert.Contains(t, errorResponse, "AI extraction failed")
			},
		},
		{
			name: "Booking flow with invalid deadline",
			bookingRequest: models.BookingRequest{
				Query:    "I want to fly from NYC to London next week",
				Deadline: time.Now().Add(-24 * time.Hour),
			},
			setupMocks: func(extractionEngine *MockTravelParameterExtractor, recommendationEngine *MockFlightRecommender) {
				// No mock setup needed
			},
			validateFlow: func(t *testing.T, createResp, getResp *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, createResp.Code)
				var errorResponse map[string]string
				err := json.Unmarshal(createResp.Body.Bytes(), &errorResponse)
				assert.NoError(t, err)
				assert.Contains(t, errorResponse["error"], "deadline cannot be in the past")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockExtractionEngine := new(MockTravelParameterExtractor)
			mockRecommendationEngine := new(MockFlightRecommender)
			tt.setupMocks(mockExtractionEngine, mockRecommendationEngine)

			// Create services and handler
			bookingService := service.NewBookingService(
				mockExtractionEngine,
				mockRecommendationEngine,
			)
			handler := handlers.NewBookingHandler(bookingService)

			// Create booking request
			createReqBody, err := json.Marshal(tt.bookingRequest)
			assert.NoError(t, err)

			// Execute create booking request
			createReq := httptest.NewRequest(http.MethodPost, "/bookings", bytes.NewBuffer(createReqBody))
			createReq.Header.Set("Content-Type", "application/json")
			createResp := httptest.NewRecorder()
			handler.CreateBooking(createResp, createReq)

			// If creation was successful, test getting the booking
			var getResp *httptest.ResponseRecorder
			if createResp.Code == http.StatusOK {
				var createResponse models.BookingResponse
				err := json.Unmarshal(createResp.Body.Bytes(), &createResponse)
				assert.NoError(t, err)

				getReq := httptest.NewRequest(http.MethodGet, "/bookings/status?id="+createResponse.ID, nil)
				getResp = httptest.NewRecorder()
				handler.GetBooking(getResp, getReq)
			}

			// Validate the flow
			tt.validateFlow(t, createResp, getResp)

			// Verify all mock expectations were met
			mockExtractionEngine.AssertExpectations(t)
			mockRecommendationEngine.AssertExpectations(t)
		})
	}
}

// Helper function to create a valid booking response
// createValidBookingResponse = func() *models.BookingResponse {
// 	now := time.Now()
// 	return &models.BookingResponse{
// 		ID:     "test-booking-id",
// 		Status: models.StatusProcessing,
// 		Query:  "I want to fly from NYC to London next week",
// 		FlightDetails: &models.Flight{
// 			Airline:       "British Airways",
// 			FlightNumber:  "BA123",
// 			Price:         800.0,
// 			Currency:      "USD",
// 			DepartureCity: "NYC",
// 			ArrivalCity:   "London",
// 			DepartureTime: now.Add(24 * time.Hour),
// 			ArrivalTime:   now.Add(31 * time.Hour),
// 		},
// 		Deadline:  now.Add(48 * time.Hour),
// 		CreatedAt: now,
// 		UpdatedAt: now,
// 		Message:   "Searching for flights to London",
// 	}
// }
