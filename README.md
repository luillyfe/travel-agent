# AI Travel Agent API

A RESTful API service that processes travel booking requests using AI to find optimal flight options based on natural language queries.

## Project Structure

```
.
├── cmd/
│   └── app/
│       └── main.go           # Application entry point
├── internal/
│   ├── models/
│   │   └── booking.go        # Data models and types
│   ├── service/
│   │   └── booking.go        # Core booking service logic
│   └── handlers/
│       └── booking.go        # HTTP request handlers
├── pkg/
│   └── utils/               # Shared utilities
├── api/
│   └── openapi.yaml         # API specifications
├── tests/
│   └── booking_test.go      # Integration and unit tests
└── go.mod                   # Module dependencies
```

## Core Components

### Models

- `BookingRequest`: Represents the initial booking request with query and deadline
- `BookingResponse`: Contains booking status and flight details
- `Flight`: Detailed flight information

### Services

- `BookingService`: Handles the core business logic for processing booking requests
  - Natural language processing
  - Flight search and optimization
  - Deadline-based monitoring

### Handlers

- `BookingHandler`: HTTP request handling for booking endpoints
  - Create new bookings
  - Check booking status

## API Endpoints

### Create Booking

```
POST /api/v1/bookings
```

Request body:

```json
{
  "query": "Book a flight for a 3-day trip to Paris",
  "deadline": "within 2 days"
}
```

Response:

```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "status": "pending",
  "query": "Book a flight for a 3-day trip to Paris",
  "deadline": "2024-02-18T15:04:05Z",
  "created_at": "2024-02-16T15:04:05Z",
  "updated_at": "2024-02-16T15:04:05Z",
  "message": "Processing your request"
}
```

### Get Booking Status

```
GET /api/v1/bookings/status?id={booking_id}
```

Response:

```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "status": "completed",
  "flight": {
    "airline": "Air France",
    "flight_number": "AF1234",
    "departure_city": "New York",
    "arrival_city": "Paris",
    "departure_time": "2024-03-01T10:00:00Z",
    "arrival_time": "2024-03-01T22:00:00Z",
    "price": 750.0,
    "currency": "USD"
  }
}
```

## Getting Started

1. Clone the repository
2. Install dependencies:
   ```bash
   go mod tidy
   ```
3. Run the server:
   ```bash
   go run cmd/app/main.go
   ```
4. The server will start on port 8080

## Testing

Run the test suite:

```bash
go test ./...
```

## Development Status

Current Implementation:

- ✅ Basic API structure
- ✅ Request/Response models
- ✅ Booking service framework
- ✅ HTTP handlers

Pending Implementation:

- ⏳ AI natural language processing
- ⏳ Flight search integration
- ⏳ Database persistence
- ⏳ Implement booking status retrieval
- ⏳ Authentication/Authorization
- ⏳ Request validation
- ⏳ Error handling middleware
- ⏳ Logging and monitoring

## Contributing

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.
