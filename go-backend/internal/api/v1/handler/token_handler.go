package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"go-backend/internal/auth"
	"go-backend/internal/config"
	"go-backend/internal/tokenexchange"
)

type TokenHandler struct {
	cfg *config.Config
}

func NewTokenHandler(cfg *config.Config) *TokenHandler {
	return &TokenHandler{cfg: cfg}
}

// IssueToken handles POST /tokens endpoint for JWT token generation
func (h *TokenHandler) IssueToken(w http.ResponseWriter, r *http.Request) {
	// Get user info from context (set by auth middleware)
	userInfo, ok := auth.GetUserInfo(r.Context())
	if !ok {
		slog.Error("User info not found in context")
		http.Error(w, "user info not found in context", http.StatusUnauthorized)
		return
	}

	if !validateContentType(w, r) {
		return
	}

	limitRequestBody(w, r, 1<<20)
	var req tokenexchange.TokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("Invalid request body for token generation", "error", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if !validateStruct(w, &req) {
		return
	}

	// Issue JWT token
	token, err := tokenexchange.IssueJWT(h.cfg, userInfo.Email, req.MicroAppID)
	if err != nil {
		slog.Error("Error occurred while generating JWT token", "error", err, "email", userInfo.Email, "microAppId", req.MicroAppID)
		if writeErr := writeJSON(w, http.StatusInternalServerError, map[string]string{
			"message": "Error occurred while generating JWT token",
		}); writeErr != nil {
			slog.Error("Failed to write error response", "error", writeErr)
			http.Error(w, "failed to write error response", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(token)); err != nil {
		slog.Error("Failed to write token response", "error", err)
	}
}

// GetJWKS handles GET /.well-known/jwks endpoint for public key retrieval
func (h *TokenHandler) GetJWKS(w http.ResponseWriter, r *http.Request) {
	jwks, err := tokenexchange.GetJWKS(h.cfg)
	if err != nil {
		slog.Error("Failed to read JWKS", "error", err)
		if writeErr := writeJSON(w, http.StatusInternalServerError, map[string]string{
			"message": "Failed to read JWKS",
		}); writeErr != nil {
			slog.Error("Failed to write error response", "error", writeErr)
			http.Error(w, "failed to write error response", http.StatusInternalServerError)
		}
		return
	}

	if err := writeJSON(w, http.StatusOK, jwks); err != nil {
		slog.Error("Failed to write JWKS response", "error", err)
		http.Error(w, "failed to write response", http.StatusInternalServerError)
	}
}
