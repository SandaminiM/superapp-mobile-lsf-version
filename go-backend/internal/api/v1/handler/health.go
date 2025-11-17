package handler

import (
	"net/http"
	"time"
)

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

// Health handles the health check endpoint
func Health(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
	}

	if err := writeJSON(w, http.StatusOK, response); err != nil {
		http.Error(w, "failed to write health response", http.StatusInternalServerError)
	}
}
