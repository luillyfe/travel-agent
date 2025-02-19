package utils

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// TimeNow returns the current time in UTC
func TimeNow() time.Time {
	return time.Now().UTC()
}

func LogResponseWithoutConsuming(resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}

	// Log response details
	log.Printf("\nHTTP Response Details:")
	log.Printf("Status Code: %v", resp.Status)
	log.Printf("Body: %s", string(body))

	// Restore the body
	resp.Body = io.NopCloser(bytes.NewBuffer(body))
	return nil
}

func MustParseTime(timeStr string) time.Time {
	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		panic(fmt.Sprintf("invalid time format: %v", err))
	}
	return t
}
