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
│       │   └── travelParameterExtraction.go
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
  "query": "Book a flight for a 3-day trip to Paris",
  "deadline": "2024-02-18T15:04:05Z"
}
```

Response:

```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "status": "processing",
  "query": "Book a flight for a 3-day trip to Paris",
  "deadline": "2024-02-18T15:04:05Z",
  "flight": {
    "airline": "",
    "flight_number": "",
    "departure_city": "New York",
    "arrival_city": "Paris",
    "departure_time": "2024-03-01T10:00:00Z",
    "arrival_time": "2024-03-01T22:00:00Z",
    "price": 0,
    "currency": ""
  },
  "created_at": "2024-02-16T15:04:05Z",
  "updated_at": "2024-02-16T15:04:05Z",
  "message": "Processing your request"
}
```

### Get Booking Status

```
GET /api/v1/bookings/status?id={booking_id}
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
- ✅ Booking service implementation
- ✅ Test suite foundation
- ✅ Parameter extraction from natural language
- ✅ Health check endpoint

Pending:

- ⏳ Flight search integration
- ⏳ Database persistence
- ⏳ Complete booking status retrieval
- ⏳ Authentication/Authorization
- ⏳ Advanced error handling middleware
- ⏳ Metrics and monitoring
- ⏳ Rate limiting
- ⏳ Caching layer

## Contributing

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.
