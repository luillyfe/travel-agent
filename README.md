# AI Travel Agent API

A RESTful API service that processes travel booking requests using AI to find optimal flight options based on natural language queries.

## Project Structure

```
.
├── cmd/
│   └── app/
│       └── main.go           # Application entry point
├── internal/
│   ├── config/              # Configuration handling
│   │   └── config.go
│   ├── handlers/            # HTTP request handlers
│   │   └── booking.go
│   ├── models/              # Data models
│   │   └── booking.go
│   ├── server/             # Server implementation
│   │   └── server.go
│   └── service/            # Business logic
│       ├── ai/             # AI inference services
│       │   ├── inference.go
│       │   ├── travelParameterExtraction.go
│       │   └── flightRecommendation.go
│       └── booking.go
├── pkg/
│   └── utils/              # Shared utilities
│       └── utils.go
├── tests/                  # Test suites
│   ├── booking_test.go
│   ├── inference_test.go
│   └── server_test.go
└── api/
    └── openapi.yaml        # API specifications
```

## Core Components

### Models

- `BookingRequest`: Represents the initial booking request with query and deadline
- `BookingResponse`: Contains booking status and flight details
- `Flight`: Detailed flight information

### Services

- `BookingService`: Core business logic for processing booking requests
- `InferenceEngine`: Handles AI parameter extraction from natural language
- `TravelParameterExtraction`: Processes travel-specific parameters
- `FlightRecommendation`: AI-powered flight recommendations based on user preferences

### Configuration

- Environment-based configuration with JSON file support
- API keys and server settings management
- Default configurations with override capability

## API Endpoints

### Create Booking

```
POST /api/v1/bookings
```

Request body:

```json
{
  "query": "Book a flight from Cúcuta to Paris on March 10th for a 3-day trip",
  "deadline": "2025-03-18T15:04:05Z"
}
```

Response:

```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "status": "processing",
  "query": "Book a flight from Cúcuta to Paris on March 10th for a 3-day trip",
  "deadline": "2025-03-18T15:04:05Z",
  "flight": {
    "airline": "",
    "flight_number": "",
    "departure_city": "Cúcuta",
    "arrival_city": "Paris",
    "departure_time": "2025-03-01T10:00:00Z",
    "arrival_time": "2025-03-01T22:00:00Z",
    "price": 0,
    "currency": ""
  },
  "created_at": "2025-02-16T15:04:05Z",
  "updated_at": "2025-02-16T15:04:05Z",
  "message": "Processing your request"
}
```

### Get Booking Status

```
GET /api/v1/bookings/status?id={booking_id}
```

### Get Flight Recommendations

```
POST /api/v1/recommendations
```

Request body:

```json
{
  "query": "I want a flight with minimal layovers and preferably in the morning",
  "preferences": {
    "max_price": 1000,
    "preferred_airlines": ["AirFrance", "KLM"],
    "max_layovers": 1,
    "time_preference": "morning"
  }
}
```

Response:

```json
{
  "recommendations": [
    {
      "flight": {
        "airline": "AirFrance",
        "flight_number": "AF1234",
        "departure_city": "Cúcuta",
        "arrival_city": "Paris",
        "departure_time": "2025-03-01T08:00:00Z",
        "arrival_time": "2025-03-01T20:00:00Z",
        "price": 850,
        "currency": "USD"
      },
      "score": 0.95,
      "reasoning": "This flight matches your preferences with a morning departure and no layovers"
    }
  ]
}
```

## Getting Started

1. Clone the repository
2. Set up configuration:
   ```bash
   # Set environment variable for AI provider
   export AI_PROVIDER_API_KEY=your_api_key
   ```
3. Install dependencies:
   ```bash
   go mod tidy
   ```
4. Run the server:
   ```bash
   go run cmd/app/main.go
   ```

## Testing

Run the test suite:

```bash
go test ./...
```

## Implementation Status

Completed:

- ✅ Basic API structure and routing
- ✅ Request/Response models
- ✅ Configuration management
- ✅ AI integration framework
- ✅ Parameter extraction from natural language
- ✅ Health check endpoint
- ✅ Flight recommendation engine
- ✅ Preference-based scoring system

In Progress:

- 🔄 Booking service implementation
- 🔄 Test suite foundation
- 🔄 Flight search integration

Pending:

- ⏳ Database persistence
- ⏳ Complete booking status retrieval
- ⏳ Authentication/Authorization
- ⏳ Advanced error handling middleware
- ⏳ Metrics and monitoring
- ⏳ Rate limiting
- ⏳ Caching layer

## Contributing

1. Fork the repository
2. Create a feature branch following the format:
   - `feature/description` for new features
   - `fix/description` for bug fixes
   - `docs/description` for documentation changes
3. Commit your changes using conventional commits
4. Push to the branch
5. Create a Pull Request against the `develop` branch

Note: The `main` branch is protected. All changes must go through PR review.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
