# My Go Application

This is a basic Go application template following standard Go project layout.

## Project Structure

- `cmd/`: Contains the main applications
- `internal/`: Private application and library code
- `pkg/`: Library code that's ok to use by external applications
- `api/`: API specifications
- `tests/`: Additional external test apps and test data

## Getting Started

1. Clone the repository
2. Run `go mod tidy` to install dependencies
3. Start the application: `go run cmd/app/main.go`

## Testing

Run tests with: `go test ./...`
