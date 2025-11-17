package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// Writes the given data as JSON to the HTTP response with the specified status code.
func writeJSON(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}
	return nil
}

// Validates a struct using the validator package and writes validation errors to the response.
func validateStruct(w http.ResponseWriter, s any) bool {
	return validateRequest(w, s)
}

// Validates that the Content-Type header is application/json.
func validateContentType(w http.ResponseWriter, r *http.Request) bool {
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
		return false
	}
	return true
}

// Limits the size of the request body to prevent large payloads.
func limitRequestBody(w http.ResponseWriter, r *http.Request, maxBytes int64) {
	if maxBytes == 0 {
		maxBytes = 1 << 20 // 1MB default
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
}

// isMaxBytesError checks if an error is due to request body size limit being exceeded
func isMaxBytesError(err error) bool {
	var maxBytesErr *http.MaxBytesError
	return errors.As(err, &maxBytesErr)
}

// isForeignKeyError checks if an error is due to a foreign key constraint violation
// MySQL error code 1452 indicates foreign key constraint failure
func isForeignKeyError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	// MySQL foreign key constraint violation error code
	return strings.Contains(errStr, "1452") ||
		strings.Contains(errStr, "foreign key constraint") ||
		strings.Contains(errStr, "Cannot add or update a child row")
}
